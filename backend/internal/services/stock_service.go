package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/catalogcache"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/repositories"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StockAfterChangeHook вызывается после изменения остатков (лог reorder и т.д.).
type StockAfterChangeHook func(ctx context.Context, storeID uuid.UUID)

// StockService партии FEFO, резервы, приходы, списания, инвентаризация.
type StockService struct {
	db           *gorm.DB
	invRepo      repositories.InventoryRepository
	catalog      *catalogcache.Store
	afterChangeH StockAfterChangeHook
}

func NewStockService(db *gorm.DB, invRepo repositories.InventoryRepository, catalog *catalogcache.Store, afterChange StockAfterChangeHook) *StockService {
	return &StockService{db: db, invRepo: invRepo, catalog: catalog, afterChangeH: afterChange}
}

func (s *StockService) afterChange(ctx context.Context, storeID uuid.UUID) {
	if s.catalog != nil {
		s.catalog.Bump(ctx, storeID)
	}
	if s.afterChangeH != nil {
		s.afterChangeH(ctx, storeID)
	}
}

// NotifyCatalogChanged публичный вызов после внешних транзакций (заказ, курьер).
func (s *StockService) NotifyCatalogChanged(ctx context.Context, storeID uuid.UUID) {
	s.afterChange(ctx, storeID)
}

func (s *StockService) dbx(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return s.db
}

// ZoneDefaultSalesFloor код зоны по умолчанию для ручной правки остатка.
const ZoneCodeSalesFloor = "sales_floor"

func (s *StockService) getZoneIDByCode(tx *gorm.DB, ctx context.Context, storeID uuid.UUID, code string) (uuid.UUID, error) {
	var z models.StoreStockZone
	err := s.dbx(tx).WithContext(ctx).Where("store_id = ? AND code = ?", storeID, code).First(&z).Error
	if err != nil {
		return uuid.Nil, err
	}
	return z.ID, nil
}

func (s *StockService) logMovement(tx *gorm.DB, ctx context.Context, m *models.StockMovement) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	m.CreatedAt = time.Now()
	return s.dbx(tx).WithContext(ctx).Create(m).Error
}

// consumeFEFO списывает quantity с партий (срок, затем дата прихода).
func (s *StockService) consumeFEFO(tx *gorm.DB, ctx context.Context, storeID, productID uuid.UUID, qty int,
	movType models.StockMovementType, refOrderID *uuid.UUID, reason string,
) error {
	if qty <= 0 {
		return nil
	}
	remaining := qty
	for remaining > 0 {
		var b models.StockBatch
		err := s.dbx(tx).WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("store_id = ? AND product_id = ? AND quantity > 0", storeID, productID).
			Order("CASE WHEN expires_at IS NULL THEN 1 ELSE 0 END ASC, expires_at ASC NULLS LAST, received_at ASC, id ASC").
			First(&b).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repositories.ErrInsufficientInventory
			}
			return err
		}
		take := b.Quantity
		if take > remaining {
			take = remaining
		}
		newQ := b.Quantity - take
		zid := b.ZoneID
		bid := b.ID
		if err := s.logMovement(tx, ctx, &models.StockMovement{
			StoreID: storeID, ProductID: productID, BatchID: &bid, ZoneID: &zid,
			Delta: -take, MovementType: movType, RefOrderID: refOrderID, Reason: reason,
		}); err != nil {
			return err
		}
		if newQ == 0 {
			if err := s.dbx(tx).WithContext(ctx).Delete(&models.StockBatch{}, "id = ?", b.ID).Error; err != nil {
				return err
			}
		} else {
			if err := s.dbx(tx).WithContext(ctx).Model(&models.StockBatch{}).Where("id = ?", b.ID).Update("quantity", newQ).Error; err != nil {
				return err
			}
		}
		remaining -= take
	}
	return nil
}

// CommitOrderStockTx FEFO + резерв внутри переданной транзакции (курьер / админ).
func (s *StockService) CommitOrderStockTx(tx *gorm.DB, ctx context.Context, orderID uuid.UUID) error {
	var fresh models.Order
	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Items").First(&fresh, "id = ?", orderID).Error; err != nil {
		return err
	}
	if fresh.StockCommitted {
		return nil
	}
	if len(fresh.Items) == 0 {
		return tx.Model(&models.Order{}).Where("id = ?", fresh.ID).Updates(map[string]interface{}{
			"stock_committed": true,
			"updated_at":      time.Now(),
		}).Error
	}
	for _, oi := range fresh.Items {
		if err := s.consumeFEFO(tx, ctx, fresh.StoreID, oi.ProductID, oi.Quantity,
			models.MovementSale, &fresh.ID, ""); err != nil {
			return err
		}
		if err := s.invRepo.CommitReservedAfterSale(tx, ctx, fresh.StoreID, oi.ProductID, oi.Quantity); err != nil {
			return err
		}
		if err := s.mirrorProductStockTx(tx, ctx, fresh.StoreID, oi.ProductID); err != nil {
			return err
		}
	}
	return tx.Model(&models.Order{}).Where("id = ?", fresh.ID).Updates(map[string]interface{}{
		"stock_committed": true,
		"updated_at":      time.Now(),
	}).Error
}

// CommitOrderStock отдельная транзакция (админ «подтвердить списание»).
func (s *StockService) CommitOrderStock(ctx context.Context, orderID uuid.UUID) error {
	var storeID uuid.UUID
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var o models.Order
		if err := tx.WithContext(ctx).First(&o, "id = ?", orderID).Error; err != nil {
			return err
		}
		storeID = o.StoreID
		return s.CommitOrderStockTx(tx, ctx, orderID)
	})
	if err != nil {
		return err
	}
	s.afterChange(ctx, storeID)
	return nil
}

func (s *StockService) mirrorProductStockTx(tx *gorm.DB, ctx context.Context, storeID, productID uuid.UUID) error {
	var q int
	if err := tx.WithContext(ctx).Raw(
		`SELECT COALESCE(quantity, 0) FROM store_inventory WHERE store_id = ? AND product_id = ?`,
		storeID, productID,
	).Scan(&q).Error; err != nil {
		return err
	}
	return tx.WithContext(ctx).Model(&models.Product{}).Where("id = ?", productID).Update("stock_quantity", q).Error
}

// MirrorProductStockTx то же внутри транзакции (клон каталога).
func (s *StockService) MirrorProductStockTx(tx *gorm.DB, ctx context.Context, storeID, productID uuid.UUID) error {
	return s.mirrorProductStockTx(tx, ctx, storeID, productID)
}

// ReleaseOrderReservationsTx снятие резерва внутри транзакции.
func (s *StockService) ReleaseOrderReservationsTx(tx *gorm.DB, ctx context.Context, orderID uuid.UUID) error {
	var fresh models.Order
	if err := tx.WithContext(ctx).Preload("Items").First(&fresh, "id = ?", orderID).Error; err != nil {
		return err
	}
	if fresh.StockCommitted {
		return fmt.Errorf("нельзя снять резерв: склад уже списан по заказу")
	}
	for _, oi := range fresh.Items {
		if err := s.invRepo.ReleaseReserved(tx, ctx, fresh.StoreID, oi.ProductID, oi.Quantity); err != nil {
			return err
		}
	}
	return nil
}

// ReleaseOrderReservations отмена заказа в pending: только резерв.
func (s *StockService) ReleaseOrderReservations(ctx context.Context, order *models.Order) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.ReleaseOrderReservationsTx(tx, ctx, order.ID)
	})
}

// ReserveLinesForNewOrder проверка и резерв по строкам (внутри общей транзакции заказа).
func (s *StockService) ReserveLinesForNewOrder(tx *gorm.DB, ctx context.Context, storeID uuid.UUID, lines []OrderItemInput) error {
	for _, it := range lines {
		if err := s.invRepo.ReserveIfEnough(tx, ctx, storeID, it.ProductID, it.Quantity); err != nil {
			return err
		}
	}
	return nil
}

// RebuildInventoryToAbsoluteTx одна партия «зал» + строка склада (внутри внешней транзакции).
func (s *StockService) RebuildInventoryToAbsoluteTx(tx *gorm.DB, ctx context.Context, storeID, productID uuid.UUID, absoluteQty int) error {
	if absoluteQty < 0 {
		absoluteQty = 0
	}
	zid, err := s.getZoneIDByCode(tx, ctx, storeID, ZoneCodeSalesFloor)
	if err != nil {
		return err
	}
	if err := tx.WithContext(ctx).Where("store_id = ? AND product_id = ?", storeID, productID).
		Delete(&models.StockBatch{}).Error; err != nil {
		return err
	}
	if absoluteQty > 0 {
		b := models.StockBatch{
			StoreID:    storeID,
			ProductID:  productID,
			ZoneID:     zid,
			Quantity:   absoluteQty,
			ReceivedAt: time.Now(),
		}
		if err := tx.WithContext(ctx).Create(&b).Error; err != nil {
			return err
		}
	}
	if err := tx.WithContext(ctx).Exec(`
UPDATE store_inventory
SET quantity = ?,
    reserved_quantity = LEAST(reserved_quantity, ?),
    updated_at = NOW()
WHERE store_id = ? AND product_id = ?
`, absoluteQty, absoluteQty, storeID, productID).Error; err != nil {
		return err
	}
	if err := tx.WithContext(ctx).Exec(`
INSERT INTO store_inventory (id, store_id, product_id, quantity, reserved_quantity, updated_at)
SELECT uuid_generate_v4(), ?, ?, ?, 0, NOW()
WHERE NOT EXISTS (SELECT 1 FROM store_inventory si WHERE si.store_id = ? AND si.product_id = ?)
`, storeID, productID, absoluteQty, storeID, productID).Error; err != nil {
		return err
	}
	return nil
}

// RebuildInventoryToAbsolute выравнивает партии в одну партию «зал» и store_inventory (для ручной правки остатка).
func (s *StockService) RebuildInventoryToAbsolute(ctx context.Context, storeID, productID uuid.UUID, absoluteQty int) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.RebuildInventoryToAbsoluteTx(tx, ctx, storeID, productID, absoluteQty)
	})
	if err != nil {
		return err
	}
	s.afterChange(ctx, storeID)
	return nil
}

// WriteOff списание брака / усушки / пересорта (FEFO).
func (s *StockService) WriteOff(ctx context.Context, storeID, productID uuid.UUID, qty int, movType models.StockMovementType, reason string) error {
	if qty <= 0 {
		return errors.New("количество должно быть > 0")
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.consumeFEFO(tx, ctx, storeID, productID, qty, movType, nil, reason); err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Exec(`
UPDATE store_inventory AS si
SET quantity = COALESCE((SELECT SUM(b.quantity) FROM stock_batches b WHERE b.store_id = si.store_id AND b.product_id = si.product_id), 0),
    updated_at = NOW()
WHERE si.store_id = ? AND si.product_id = ?
`, storeID, productID).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.afterChange(ctx, storeID)
	return nil
}

type ReceiptLineInput struct {
	ProductID uuid.UUID
	ZoneID    uuid.UUID
	Quantity  int
	ExpiresAt *time.Time
}

// ApplyReceipt приход на склад (+партии, +движения).
func (s *StockService) ApplyReceipt(ctx context.Context, storeID uuid.UUID, supplierID *uuid.UUID, note string, lines []ReceiptLineInput) (*models.StockReceipt, error) {
	if len(lines) == 0 {
		return nil, errors.New("нет строк прихода")
	}
	var out *models.StockReceipt
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		r := models.StockReceipt{StoreID: storeID, SupplierID: supplierID, Note: note}
		if err := tx.WithContext(ctx).Create(&r).Error; err != nil {
			return err
		}
		for _, ln := range lines {
			if ln.Quantity <= 0 {
				continue
			}
			line := models.StockReceiptLine{
				ReceiptID: r.ID, ProductID: ln.ProductID, ZoneID: ln.ZoneID,
				Quantity: ln.Quantity, ExpiresAt: ln.ExpiresAt,
			}
			if err := tx.WithContext(ctx).Create(&line).Error; err != nil {
				return err
			}
			b := models.StockBatch{
				StoreID: storeID, ProductID: ln.ProductID, ZoneID: ln.ZoneID,
				Quantity: ln.Quantity, ReceivedAt: time.Now(), ExpiresAt: ln.ExpiresAt,
				SupplierID: supplierID,
				Note:       note,
			}
			if err := tx.WithContext(ctx).Create(&b).Error; err != nil {
				return err
			}
			zid := ln.ZoneID
			if err := s.logMovement(tx, ctx, &models.StockMovement{
				StoreID: storeID, ProductID: ln.ProductID, BatchID: &b.ID, ZoneID: &zid,
				Delta: ln.Quantity, MovementType: models.MovementReceipt, Reason: note,
			}); err != nil {
				return err
			}
			if err := tx.WithContext(ctx).Exec(`
INSERT INTO store_inventory (id, store_id, product_id, quantity, reserved_quantity, updated_at)
VALUES (uuid_generate_v4(), ?, ?, 0, 0, NOW())
ON CONFLICT (store_id, product_id) DO NOTHING
`, storeID, ln.ProductID).Error; err != nil {
				return err
			}
			if err := tx.WithContext(ctx).Exec(`
UPDATE store_inventory AS si
SET quantity = COALESCE((SELECT SUM(b2.quantity) FROM stock_batches b2 WHERE b2.store_id = si.store_id AND b2.product_id = si.product_id), 0),
    updated_at = NOW()
WHERE si.store_id = ? AND si.product_id = ?
`, storeID, ln.ProductID).Error; err != nil {
				return err
			}
		}
		out = &r
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.afterChange(ctx, storeID)
	return out, nil
}

type ExpiringBatchRow struct {
	BatchID     uuid.UUID  `json:"batch_id"`
	ProductID   uuid.UUID  `json:"product_id"`
	ProductName string     `json:"product_name"`
	ZoneName    string     `json:"zone_name"`
	Quantity    int        `json:"quantity"`
	ExpiresAt   *time.Time `json:"expires_at"`
	ReceivedAt  time.Time  `json:"received_at"`
}

// SimpleReceiveToSalesFloor быстрый приход в зону «Зал» (аналог «Приход» в zakazik).
func (s *StockService) SimpleReceiveToSalesFloor(ctx context.Context, storeID, productID uuid.UUID, qty int, note string) (*models.StockReceipt, error) {
	if qty <= 0 {
		return nil, errors.New("количество должно быть > 0")
	}
	zid, err := s.getZoneIDByCode(nil, ctx, storeID, ZoneCodeSalesFloor)
	if err != nil {
		return nil, fmt.Errorf("зона «Зал» (sales_floor) не найдена: %w", err)
	}
	return s.ApplyReceipt(ctx, storeID, nil, note, []ReceiptLineInput{
		{ProductID: productID, ZoneID: zid, Quantity: qty},
	})
}

// StockMovementView строка журнала с названием товара (удобно для UI).
type StockMovementView struct {
	ID            uuid.UUID                 `json:"id"`
	StoreID       uuid.UUID                 `json:"store_id"`
	ProductID     uuid.UUID                 `json:"product_id"`
	BatchID       *uuid.UUID                `json:"batch_id,omitempty"`
	ZoneID        *uuid.UUID                `json:"zone_id,omitempty"`
	Delta         int                       `json:"delta"`
	MovementType  models.StockMovementType  `json:"movement_type"`
	RefOrderID    *uuid.UUID                `json:"ref_order_id,omitempty"`
	Reason        string                    `json:"reason,omitempty"`
	CreatedAt     time.Time                 `json:"created_at"`
	ProductName   string                    `json:"product_name"`
}

// ListStockMovementsRecentWithNames последние движения с JOIN на products (как вкладка «Движение» в zakazik).
func (s *StockService) ListStockMovementsRecentWithNames(ctx context.Context, storeID uuid.UUID, limit int) ([]StockMovementView, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var rows []StockMovementView
	err := s.db.WithContext(ctx).Raw(`
SELECT m.id, m.store_id, m.product_id, m.batch_id, m.zone_id, m.delta, m.movement_type,
       m.ref_order_id, m.reason, m.created_at, p.name AS product_name
FROM stock_movements m
JOIN products p ON p.id = m.product_id AND p.store_id = m.store_id
WHERE m.store_id = ?
ORDER BY m.created_at DESC, m.id DESC
LIMIT ?
`, storeID, limit).Scan(&rows).Error
	return rows, err
}

// SetInventoryActual выставляет остаток в шт/г (как «факт» в zakazik); движение audit_adjustment с дельтой.
func (s *StockService) SetInventoryActual(ctx context.Context, storeID, productID uuid.UUID, actual int, note string) error {
	if actual < 0 {
		actual = 0
	}
	if note == "" {
		note = "Инвентаризация"
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var prev int
		if err := tx.WithContext(ctx).Raw(`
SELECT COALESCE(GREATEST(si.quantity - si.reserved_quantity, 0), 0)
FROM store_inventory si
WHERE si.store_id = ? AND si.product_id = ?
`, storeID, productID).Scan(&prev).Error; err != nil {
			return err
		}
		delta := actual - prev
		if err := s.RebuildInventoryToAbsoluteTx(tx, ctx, storeID, productID, actual); err != nil {
			return err
		}
		if delta != 0 {
			return s.logMovement(tx, ctx, &models.StockMovement{
				StoreID:      storeID,
				ProductID:    productID,
				Delta:        delta,
				MovementType: models.MovementAudit,
				Reason:       note,
			})
		}
		return nil
	})
	if err != nil {
		return err
	}
	_ = s.MirrorProductStock(ctx, storeID, productID)
	s.afterChange(ctx, storeID)
	return nil
}

// ListStockMovementsPaged журнал движений, курсор (created_at DESC, id DESC).
func (s *StockService) ListStockMovementsPaged(ctx context.Context, storeID uuid.UUID, limit int, afterCreatedAt *time.Time, afterID *uuid.UUID) ([]models.StockMovement, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := s.db.WithContext(ctx).Where("store_id = ?", storeID)
	if afterCreatedAt != nil && afterID != nil {
		q = q.Where("created_at < ? OR (created_at = ? AND id < ?)", *afterCreatedAt, *afterCreatedAt, *afterID)
	}
	var rows []models.StockMovement
	err := q.Order("created_at DESC, id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (s *StockService) ListExpiring(ctx context.Context, storeID uuid.UUID, withinDays int) ([]ExpiringBatchRow, error) {
	if withinDays < 1 {
		withinDays = 2
	}
	until := time.Now().AddDate(0, 0, withinDays)
	var rows []ExpiringBatchRow
	err := s.db.WithContext(ctx).Raw(`
SELECT b.id AS batch_id, b.product_id, p.name AS product_name, z.name AS zone_name,
       b.quantity, b.expires_at, b.received_at
FROM stock_batches b
JOIN products p ON p.id = b.product_id
JOIN store_stock_zones z ON z.id = b.zone_id
WHERE b.store_id = ? AND b.quantity > 0 AND b.expires_at IS NOT NULL AND b.expires_at <= ?
ORDER BY b.expires_at ASC
`, storeID, until).Scan(&rows).Error
	return rows, err
}

// MinExpiresAtByProduct для витрины: минимальный expires_at по каждому товару (только партии со сроком и quantity > 0).
func (s *StockService) MinExpiresAtByProduct(ctx context.Context, storeID uuid.UUID) (map[uuid.UUID]time.Time, error) {
	if s.db == nil {
		return map[uuid.UUID]time.Time{}, nil
	}
	var rows []struct {
		ProductID uuid.UUID `gorm:"column:product_id"`
		MinExp    time.Time `gorm:"column:min_exp"`
	}
	err := s.db.WithContext(ctx).Raw(`
SELECT product_id, MIN(expires_at) AS min_exp
FROM stock_batches
WHERE store_id = ? AND quantity > 0 AND expires_at IS NOT NULL
GROUP BY product_id
`, storeID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[uuid.UUID]time.Time, len(rows))
	for _, r := range rows {
		m[r.ProductID] = r.MinExp
	}
	return m, nil
}

type ReorderAlertRow struct {
	ProductID     uuid.UUID `json:"product_id"`
	Name          string    `json:"name"`
	Available     int       `json:"available"`
	ReorderMin    int       `json:"reorder_min_qty"`
	InventoryUnit string    `json:"inventory_unit"`
}

func (s *StockService) ListReorderAlerts(ctx context.Context, storeID uuid.UUID) ([]ReorderAlertRow, error) {
	var rows []ReorderAlertRow
	err := s.db.WithContext(ctx).Raw(`
SELECT p.id AS product_id, p.name,
       GREATEST(si.quantity - si.reserved_quantity, 0) AS available,
       p.reorder_min_qty,
       p.inventory_unit AS inventory_unit
FROM products p
JOIN store_inventory si ON si.product_id = p.id AND si.store_id = p.store_id
WHERE p.store_id = ? AND p.is_active AND p.reorder_min_qty > 0
  AND GREATEST(si.quantity - si.reserved_quantity, 0) < p.reorder_min_qty
ORDER BY p.name
`, storeID).Scan(&rows).Error
	return rows, err
}

func (s *StockService) ListZones(ctx context.Context, storeID uuid.UUID) ([]models.StoreStockZone, error) {
	var z []models.StoreStockZone
	err := s.db.WithContext(ctx).Where("store_id = ?", storeID).Order("sort_order, name").Find(&z).Error
	return z, err
}

type AuditLineInput struct {
	ProductID  uuid.UUID
	ZoneID     *uuid.UUID
	CountedQty int
}

// CompleteInventoryAudit применяет расхождения: diff > 0 — приход в зону зал; diff < 0 — FEFO списание.
func (s *StockService) CompleteInventoryAudit(ctx context.Context, storeID uuid.UUID, note string, lines []AuditLineInput) (*models.InventoryAuditSession, error) {
	if len(lines) == 0 {
		return nil, errors.New("нет строк пересчёта")
	}
	var session *models.InventoryAuditSession
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		sess := models.InventoryAuditSession{StoreID: storeID, Note: note, StartedAt: time.Now()}
		if err := tx.WithContext(ctx).Create(&sess).Error; err != nil {
			return err
		}
		zidDefault, zerr := s.getZoneIDByCode(tx, ctx, storeID, ZoneCodeSalesFloor)
		if zerr != nil {
			return zerr
		}
		now := time.Now()
		for _, ln := range lines {
			var sysQty int
			if err := tx.WithContext(ctx).Raw(`
SELECT COALESCE(SUM(quantity),0) FROM stock_batches WHERE store_id = ? AND product_id = ?
`, storeID, ln.ProductID).Scan(&sysQty).Error; err != nil {
				return err
			}
			diff := ln.CountedQty - sysQty
			al := models.InventoryAuditLine{
				SessionID: sess.ID, ProductID: ln.ProductID, ZoneID: ln.ZoneID,
				CountedQty: ln.CountedQty, SystemQtySnapshot: sysQty, DiffQty: diff,
				CreatedAt: now,
			}
			if err := tx.WithContext(ctx).Create(&al).Error; err != nil {
				return err
			}
			if diff == 0 {
				continue
			}
			if diff > 0 {
				zone := zidDefault
				if ln.ZoneID != nil {
					zone = *ln.ZoneID
				}
				b := models.StockBatch{
					StoreID: storeID, ProductID: ln.ProductID, ZoneID: zone,
					Quantity: diff, ReceivedAt: now, Note: "инвентаризация +",
				}
				if err := tx.WithContext(ctx).Create(&b).Error; err != nil {
					return err
				}
				z := zone
				if err := s.logMovement(tx, ctx, &models.StockMovement{
					StoreID: storeID, ProductID: ln.ProductID, BatchID: &b.ID, ZoneID: &z,
					Delta: diff, MovementType: models.MovementAudit, Reason: note,
				}); err != nil {
					return err
				}
			} else {
				if err := s.consumeFEFO(tx, ctx, storeID, ln.ProductID, -diff,
					models.MovementAudit, nil, note); err != nil {
					return err
				}
			}
			if err := tx.WithContext(ctx).Exec(`
INSERT INTO store_inventory (id, store_id, product_id, quantity, reserved_quantity, updated_at)
VALUES (uuid_generate_v4(), ?, ?, 0, 0, NOW())
ON CONFLICT (store_id, product_id) DO NOTHING
`, storeID, ln.ProductID).Error; err != nil {
				return err
			}
			if err := tx.WithContext(ctx).Exec(`
UPDATE store_inventory AS si
SET quantity = COALESCE((SELECT SUM(b2.quantity) FROM stock_batches b2 WHERE b2.store_id = si.store_id AND b2.product_id = si.product_id), 0),
    reserved_quantity = LEAST(si.reserved_quantity, COALESCE((SELECT SUM(b2.quantity) FROM stock_batches b2 WHERE b2.store_id = si.store_id AND b2.product_id = si.product_id), 0)),
    updated_at = NOW()
WHERE si.store_id = ? AND si.product_id = ?
`, storeID, ln.ProductID).Error; err != nil {
				return err
			}
		}
		if err := tx.WithContext(ctx).Model(&models.InventoryAuditSession{}).Where("id = ?", sess.ID).
			Update("completed_at", now).Error; err != nil {
			return err
		}
		session = &sess
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.afterChange(ctx, storeID)
	return session, nil
}

func (s *StockService) ListSuppliers(ctx context.Context, storeID uuid.UUID) ([]models.Supplier, error) {
	var list []models.Supplier
	err := s.db.WithContext(ctx).Where("store_id = ? AND is_active = ?", storeID, true).Order("name").Find(&list).Error
	return list, err
}

func (s *StockService) CreateSupplier(ctx context.Context, storeID uuid.UUID, name, phone string) (*models.Supplier, error) {
	su := models.Supplier{StoreID: storeID, Name: name, Phone: phone, IsActive: true}
	if err := s.db.WithContext(ctx).Create(&su).Error; err != nil {
		return nil, err
	}
	return &su, nil
}

// ListBatchesForProduct админка: партии по товару.
func (s *StockService) ListBatchesForProduct(ctx context.Context, storeID, productID uuid.UUID) ([]models.StockBatch, error) {
	var list []models.StockBatch
	err := s.db.WithContext(ctx).Preload("Zone").Where("store_id = ? AND product_id = ? AND quantity > 0", storeID, productID).
		Order("expires_at ASC NULLS LAST, received_at ASC").Find(&list).Error
	return list, err
}

// MirrorProductStock обновляет products.stock_quantity из физического склада (после операций).
func (s *StockService) MirrorProductStock(ctx context.Context, storeID, productID uuid.UUID) error {
	var q int
	if err := s.db.WithContext(ctx).Raw(
		`SELECT COALESCE(quantity, 0) FROM store_inventory WHERE store_id = ? AND product_id = ?`,
		storeID, productID,
	).Scan(&q).Error; err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&models.Product{}).Where("id = ?", productID).Update("stock_quantity", q).Error
}

// CopyBatchesBetweenProductsTx копирует партии при клонировании каталога (зоны по code внутри магазина).
func (s *StockService) CopyBatchesBetweenProductsTx(tx *gorm.DB, ctx context.Context, fromStore, toStore, fromPID, toPID uuid.UUID) error {
	var batches []models.StockBatch
	if err := tx.WithContext(ctx).Where("store_id = ? AND product_id = ? AND quantity > 0", fromStore, fromPID).Find(&batches).Error; err != nil {
		return err
	}
	for _, b := range batches {
		var z models.StoreStockZone
		if err := tx.WithContext(ctx).First(&z, "id = ?", b.ZoneID).Error; err != nil {
			return err
		}
		var tz models.StoreStockZone
		if err := tx.WithContext(ctx).Where("store_id = ? AND code = ?", toStore, z.Code).First(&tz).Error; err != nil {
			return err
		}
		nb := models.StockBatch{
			StoreID: toStore, ProductID: toPID, ZoneID: tz.ID,
			Quantity: b.Quantity, ReceivedAt: b.ReceivedAt, ExpiresAt: b.ExpiresAt,
			Note: b.Note,
		}
		if err := tx.WithContext(ctx).Create(&nb).Error; err != nil {
			return err
		}
	}
	return tx.WithContext(ctx).Exec(`
UPDATE store_inventory AS si
SET quantity = COALESCE((SELECT SUM(b2.quantity) FROM stock_batches b2 WHERE b2.store_id = si.store_id AND b2.product_id = si.product_id), 0),
    reserved_quantity = LEAST(si.reserved_quantity, COALESCE((SELECT SUM(b2.quantity) FROM stock_batches b2 WHERE b2.store_id = si.store_id AND b2.product_id = si.product_id), 0)),
    updated_at = NOW()
WHERE si.store_id = ? AND si.product_id = ?
`, toStore, toPID).Error
}
