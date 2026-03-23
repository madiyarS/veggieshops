package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"gorm.io/gorm"
)

// ErrInsufficientInventory не хватает остатка на складе (или строки склада нет).
var ErrInsufficientInventory = errors.New("insufficient inventory")

type InventoryRepository interface {
	Upsert(ctx context.Context, storeID, productID uuid.UUID, quantity int) error
	UpsertTx(tx *gorm.DB, ctx context.Context, storeID, productID uuid.UUID, quantity int) error
	DecrementIfEnough(tx *gorm.DB, ctx context.Context, storeID, productID uuid.UUID, qty int) error
	GetQuantitiesByStore(ctx context.Context, storeID uuid.UUID) (map[uuid.UUID]int, error)
	GetQuantity(ctx context.Context, storeID, productID uuid.UUID) (int, error)
	// Доступно к продаже = quantity - reserved_quantity
	GetAvailableByStore(ctx context.Context, storeID uuid.UUID) (map[uuid.UUID]int, error)
	GetRowsByStore(ctx context.Context, storeID uuid.UUID) ([]models.StoreInventory, error)
	ReserveIfEnough(tx *gorm.DB, ctx context.Context, storeID, productID uuid.UUID, qty int) error
	ReleaseReserved(tx *gorm.DB, ctx context.Context, storeID, productID uuid.UUID, qty int) error
	CommitReservedAfterSale(tx *gorm.DB, ctx context.Context, storeID, productID uuid.UUID, qty int) error
}

type inventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) dbx(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *inventoryRepository) Upsert(ctx context.Context, storeID, productID uuid.UUID, quantity int) error {
	return r.UpsertTx(nil, ctx, storeID, productID, quantity)
}

func (r *inventoryRepository) UpsertTx(tx *gorm.DB, ctx context.Context, storeID, productID uuid.UUID, quantity int) error {
	if quantity < 0 {
		quantity = 0
	}
	return r.dbx(tx).WithContext(ctx).Exec(`
INSERT INTO store_inventory (id, store_id, product_id, quantity, updated_at)
VALUES (uuid_generate_v4(), ?, ?, ?, NOW())
ON CONFLICT (store_id, product_id)
DO UPDATE SET quantity = EXCLUDED.quantity, updated_at = NOW()
`, storeID, productID, quantity).Error
}

// DecrementIfEnough атомарно уменьшает остаток; при нехватке — ошибка ErrInsufficientStock через gorm.ErrRecordNotFound не используем — смотрим RowsAffected.
func (r *inventoryRepository) DecrementIfEnough(tx *gorm.DB, ctx context.Context, storeID, productID uuid.UUID, qty int) error {
	if qty <= 0 {
		return nil
	}
	res := r.dbx(tx).WithContext(ctx).Exec(`
UPDATE store_inventory
SET quantity = quantity - ?, updated_at = NOW()
WHERE store_id = ? AND product_id = ? AND quantity >= ?
`, qty, storeID, productID, qty)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrInsufficientInventory
	}
	return nil
}

func (r *inventoryRepository) GetQuantitiesByStore(ctx context.Context, storeID uuid.UUID) (map[uuid.UUID]int, error) {
	var rows []models.StoreInventory
	if err := r.db.WithContext(ctx).Where("store_id = ?", storeID).Find(&rows).Error; err != nil {
		return nil, err
	}
	m := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		m[row.ProductID] = row.Quantity
	}
	return m, nil
}

func (r *inventoryRepository) GetQuantity(ctx context.Context, storeID, productID uuid.UUID) (int, error) {
	var row models.StoreInventory
	err := r.db.WithContext(ctx).Where("store_id = ? AND product_id = ?", storeID, productID).First(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return row.Quantity, nil
}

func (r *inventoryRepository) GetAvailableByStore(ctx context.Context, storeID uuid.UUID) (map[uuid.UUID]int, error) {
	rows, err := r.GetRowsByStore(ctx, storeID)
	if err != nil {
		return nil, err
	}
	m := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		avail := row.Quantity - row.ReservedQuantity
		if avail < 0 {
			avail = 0
		}
		m[row.ProductID] = avail
	}
	return m, nil
}

func (r *inventoryRepository) GetRowsByStore(ctx context.Context, storeID uuid.UUID) ([]models.StoreInventory, error) {
	var rows []models.StoreInventory
	err := r.db.WithContext(ctx).Where("store_id = ?", storeID).Find(&rows).Error
	return rows, err
}

func (r *inventoryRepository) ReserveIfEnough(tx *gorm.DB, ctx context.Context, storeID, productID uuid.UUID, qty int) error {
	if qty <= 0 {
		return nil
	}
	res := r.dbx(tx).WithContext(ctx).Exec(`
UPDATE store_inventory
SET reserved_quantity = reserved_quantity + ?, updated_at = NOW()
WHERE store_id = ? AND product_id = ? AND (quantity - reserved_quantity) >= ?
`, qty, storeID, productID, qty)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrInsufficientInventory
	}
	return nil
}

func (r *inventoryRepository) ReleaseReserved(tx *gorm.DB, ctx context.Context, storeID, productID uuid.UUID, qty int) error {
	if qty <= 0 {
		return nil
	}
	res := r.dbx(tx).WithContext(ctx).Exec(`
UPDATE store_inventory
SET reserved_quantity = reserved_quantity - ?, updated_at = NOW()
WHERE store_id = ? AND product_id = ? AND reserved_quantity >= ?
`, qty, storeID, productID, qty)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrInsufficientInventory
	}
	return nil
}

// CommitReservedAfterSale после FEFO-списания партий: пересчитать физический остаток из партий и уменьшить резерв.
func (r *inventoryRepository) CommitReservedAfterSale(tx *gorm.DB, ctx context.Context, storeID, productID uuid.UUID, qty int) error {
	if qty <= 0 {
		return nil
	}
	res := r.dbx(tx).WithContext(ctx).Exec(`
UPDATE store_inventory AS si
SET quantity = COALESCE((SELECT SUM(b.quantity) FROM stock_batches b WHERE b.store_id = si.store_id AND b.product_id = si.product_id), 0),
    reserved_quantity = si.reserved_quantity - ?,
    updated_at = NOW()
WHERE si.store_id = ? AND si.product_id = ? AND si.reserved_quantity >= ?
`, qty, storeID, productID, qty)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrInsufficientInventory
	}
	return nil
}
