package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/repositories"
	"github.com/veggieshop/backend/internal/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OrderWorkflow единая точка переходов заказа и связанных операций склада.
type OrderWorkflow struct {
	db        *gorm.DB
	orderRepo repositories.OrderRepository
	stockSvc  *StockService
}

func NewOrderWorkflow(db *gorm.DB, orderRepo repositories.OrderRepository, stockSvc *StockService) *OrderWorkflow {
	return &OrderWorkflow{db: db, orderRepo: orderRepo, stockSvc: stockSvc}
}

// CancelPending pending → cancelled, снятие резерва.
func (w *OrderWorkflow) CancelPending(ctx context.Context, orderID uuid.UUID) error {
	var storeID uuid.UUID
	err := w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.WithContext(ctx).First(&order, "id = ?", orderID).Error; err != nil {
			return utils.ErrNotFound
		}
		storeID = order.StoreID
		if order.Status != models.OrderPending && order.Status != models.OrderPreparing {
			return utils.ErrInvalidInput
		}
		if err := w.stockSvc.ReleaseOrderReservationsTx(tx, ctx, orderID); err != nil {
			return err
		}
		order.Status = models.OrderCancelled
		order.UpdatedAt = time.Now()
		return tx.WithContext(ctx).Save(&order).Error
	})
	if err != nil {
		return err
	}
	w.stockSvc.NotifyCatalogChanged(ctx, storeID)
	return nil
}

// CommitStock списание FEFO + снятие резерва (идемпотентно по stock_committed).
// Запрещено только для отменённых и доставленных.
func (w *OrderWorkflow) CommitStock(ctx context.Context, orderID uuid.UUID) error {
	order, err := w.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return utils.ErrNotFound
	}
	if order.Status == models.OrderCancelled || order.Status == models.OrderDelivered {
		return utils.ErrInvalidInput
	}
	if order.Status == models.OrderInDelivery {
		return utils.ErrInvalidInput
	}
	return w.stockSvc.CommitOrderStock(ctx, orderID)
}

// AcceptByCourier курьер забирает заказ: статус «в доставке», резерв ещё не списан (списание при вводе кода).
func (w *OrderWorkflow) AcceptByCourier(ctx context.Context, courierStoreID, courierUserID, orderID uuid.UUID) error {
	err := w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&order, "id = ?", orderID).Error; err != nil {
			return utils.ErrNotFound
		}
		if order.StoreID != courierStoreID {
			return utils.ErrForbidden
		}
		if order.Status == models.OrderDelivered || order.Status == models.OrderCancelled {
			return utils.ErrInvalidInput
		}
		if order.Status == models.OrderInDelivery {
			return utils.ErrInvalidInput
		}
		switch order.Status {
		case models.OrderPending, models.OrderPreparing, models.OrderConfirmed:
		default:
			return utils.ErrInvalidInput
		}
		if order.CourierID != nil && *order.CourierID != courierUserID {
			return utils.ErrForbidden
		}
		uid := courierUserID
		order.CourierID = &uid
		order.Status = models.OrderInDelivery
		order.UpdatedAt = time.Now()
		return tx.WithContext(ctx).Save(&order).Error
	})
	if err != nil {
		return err
	}
	w.stockSvc.NotifyCatalogChanged(ctx, courierStoreID)
	return nil
}

// CompleteDeliveryForCourier списание FEFO + доставлен (после верного кода).
func (w *OrderWorkflow) CompleteDeliveryForCourier(ctx context.Context, courierUserID, orderID uuid.UUID, code string) error {
	var storeID uuid.UUID
	err := w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&order, "id = ?", orderID).Error; err != nil {
			return utils.ErrNotFound
		}
		storeID = order.StoreID
		if order.Status != models.OrderInDelivery {
			return utils.ErrOrderNotInDelivery
		}
		if order.CourierID == nil || *order.CourierID != courierUserID {
			return utils.ErrForbidden
		}
		if utils.NormalizeDeliveryCode(code) != utils.NormalizeDeliveryCode(order.DeliveryCode) {
			return utils.ErrWrongDeliveryCode
		}
		if err := w.stockSvc.CommitOrderStockTx(tx, ctx, orderID); err != nil {
			return err
		}
		updates := map[string]interface{}{
			"status":     models.OrderDelivered,
			"updated_at": time.Now(),
		}
		if order.PaymentMethod == models.PaymentCash {
			updates["payment_status"] = models.PaymentCompleted
		}
		return tx.WithContext(ctx).Model(&models.Order{}).Where("id = ?", orderID).Updates(updates).Error
	})
	if err != nil {
		return err
	}
	w.stockSvc.NotifyCatalogChanged(ctx, storeID)
	return nil
}

// AdminReturnFromDelivery отмена доставки: заказ снова «собирается», курьер снят, резерв возвращён (пока склад не списан).
func (w *OrderWorkflow) AdminReturnFromDelivery(ctx context.Context, orderID uuid.UUID) error {
	var storeID uuid.UUID
	err := w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&order, "id = ?", orderID).Error; err != nil {
			return utils.ErrNotFound
		}
		storeID = order.StoreID
		if order.Status != models.OrderInDelivery {
			return utils.ErrInvalidInput
		}
		if order.StockCommitted {
			return errors.New("склад уже списан по заказу — отмена доставки недоступна")
		}
		if err := w.stockSvc.ReleaseOrderReservationsTx(tx, ctx, orderID); err != nil {
			return err
		}
		order.CourierID = nil
		order.Status = models.OrderPreparing
		order.UpdatedAt = time.Now()
		return tx.WithContext(ctx).Save(&order).Error
	})
	if err != nil {
		return err
	}
	w.stockSvc.NotifyCatalogChanged(ctx, storeID)
	return nil
}

// ExpireStalePending отменяет устаревшие pending (тот же эффект, что ручная отмена).
func (w *OrderWorkflow) ExpireStalePending(ctx context.Context, olderThan time.Time) (int, error) {
	ids, err := w.orderRepo.ListPendingOrderIDsCreatedBefore(ctx, olderThan, 100)
	if err != nil {
		return 0, err
	}
	var n int
	for _, id := range ids {
		if err := w.CancelPending(ctx, id); err != nil {
			continue
		}
		n++
	}
	return n, nil
}
