package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*models.Order, error)
	GetByStoreID(ctx context.Context, storeID uuid.UUID, filters OrderFilters) ([]*models.Order, error)
	Update(ctx context.Context, order *models.Order) error
	CountByTimeSlot(ctx context.Context, slotID uuid.UUID, excludeStatus models.OrderStatus) (int64, error)
	CountByDistrictID(ctx context.Context, districtID uuid.UUID, excludeStatus models.OrderStatus) (int64, error)
	AdminRevenueSummary(ctx context.Context, storeID *uuid.UUID, dateFrom, dateTo *string) (*AdminRevenueSummaryResult, error)
	AdminRevenueByDay(ctx context.Context, storeID *uuid.UUID, dateFrom, dateTo *string) ([]AdminRevenueDayRow, error)
	ListActiveForStore(ctx context.Context, storeID uuid.UUID) ([]*models.Order, error)
	ListPendingOrderIDsCreatedBefore(ctx context.Context, before time.Time, limit int) ([]uuid.UUID, error)
	ListByStoreIDPaged(ctx context.Context, storeID uuid.UUID, filters OrderFilters, limit int, afterCreatedAt *time.Time, afterID *uuid.UUID) ([]*models.Order, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*models.Order, error)
	GetByIDForUser(ctx context.Context, id, userID uuid.UUID) (*models.Order, error)
}

// AdminRevenueSummaryResult агрегаты по заказам (без отменённых).
type AdminRevenueSummaryResult struct {
	TotalRevenue     int64 `gorm:"column:total_revenue"`
	OrdersCount      int64 `gorm:"column:orders_count"`
	TotalDeliveryFee int64 `gorm:"column:total_delivery_fee"`
}

// AdminRevenueDayRow выручка за календарный день.
type AdminRevenueDayRow struct {
	Date    string `json:"date" gorm:"column:date"`
	Revenue int64  `json:"revenue" gorm:"column:revenue"`
	Orders  int64  `json:"orders" gorm:"column:orders"`
}

type OrderFilters struct {
	Status   *models.OrderStatus
	DateFrom *string
	DateTo   *string
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *models.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).Preload("Items").First(&order, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByOrderNumber(ctx context.Context, orderNumber string) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).Preload("Items").First(&order, "order_number = ?", orderNumber).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByStoreID(ctx context.Context, storeID uuid.UUID, filters OrderFilters) ([]*models.Order, error) {
	var orders []*models.Order
	query := r.db.WithContext(ctx).Where("store_id = ?", storeID)
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.DateFrom != nil {
		query = query.Where("DATE(created_at) >= ?", *filters.DateFrom)
	}
	if filters.DateTo != nil {
		query = query.Where("DATE(created_at) <= ?", *filters.DateTo)
	}
	err := query.Preload("Items").Order("created_at DESC").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) Update(ctx context.Context, order *models.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *orderRepository) CountByTimeSlot(ctx context.Context, slotID uuid.UUID, excludeStatus models.OrderStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Order{}).
		Where("delivery_time_slot_id = ? AND status != ?", slotID, excludeStatus).
		Count(&count).Error
	return count, err
}

func (r *orderRepository) CountByDistrictID(ctx context.Context, districtID uuid.UUID, excludeStatus models.OrderStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Order{}).
		Where("district_id = ? AND status != ?", districtID, excludeStatus).
		Count(&count).Error
	return count, err
}

func (r *orderRepository) adminRevenueBaseQuery(ctx context.Context, storeID *uuid.UUID, dateFrom, dateTo *string) *gorm.DB {
	q := r.db.WithContext(ctx).Model(&models.Order{}).Where("status != ?", models.OrderCancelled)
	if storeID != nil {
		q = q.Where("store_id = ?", *storeID)
	}
	if dateFrom != nil {
		q = q.Where("DATE(created_at) >= ?", *dateFrom)
	}
	if dateTo != nil {
		q = q.Where("DATE(created_at) <= ?", *dateTo)
	}
	return q
}

func (r *orderRepository) AdminRevenueSummary(ctx context.Context, storeID *uuid.UUID, dateFrom, dateTo *string) (*AdminRevenueSummaryResult, error) {
	var res AdminRevenueSummaryResult
	err := r.adminRevenueBaseQuery(ctx, storeID, dateFrom, dateTo).
		Select("COALESCE(SUM(total_amount),0) as total_revenue, COUNT(*) as orders_count, COALESCE(SUM(delivery_fee),0) as total_delivery_fee").
		Scan(&res).Error
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *orderRepository) AdminRevenueByDay(ctx context.Context, storeID *uuid.UUID, dateFrom, dateTo *string) ([]AdminRevenueDayRow, error) {
	var rows []AdminRevenueDayRow
	err := r.adminRevenueBaseQuery(ctx, storeID, dateFrom, dateTo).
		Select("DATE(created_at)::text as date, COALESCE(SUM(total_amount),0) as revenue, COUNT(*) as orders").
		Group("DATE(created_at)").
		Order("DATE(created_at) ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *orderRepository) ListActiveForStore(ctx context.Context, storeID uuid.UUID) ([]*models.Order, error) {
	var orders []*models.Order
	err := r.db.WithContext(ctx).
		Where("store_id = ? AND status NOT IN ?", storeID, []models.OrderStatus{models.OrderDelivered, models.OrderCancelled}).
		Preload("Items").
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

func (r *orderRepository) ListPendingOrderIDsCreatedBefore(ctx context.Context, before time.Time, limit int) ([]uuid.UUID, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []struct {
		ID uuid.UUID `gorm:"column:id"`
	}
	err := r.db.WithContext(ctx).Model(&models.Order{}).
		Select("id").
		Where("status IN ? AND created_at < ?", []models.OrderStatus{models.OrderPending, models.OrderPreparing}, before).
		Order("created_at ASC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ID)
	}
	return out, nil
}

// ListByStoreIDPaged курсор по (created_at DESC, id DESC): передайте afterCreatedAt+afterID с последней строки предыдущей страницы.
func (r *orderRepository) ListByStoreIDPaged(ctx context.Context, storeID uuid.UUID, filters OrderFilters, limit int, afterCreatedAt *time.Time, afterID *uuid.UUID) ([]*models.Order, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := r.db.WithContext(ctx).Where("store_id = ?", storeID)
	if filters.Status != nil {
		q = q.Where("status = ?", *filters.Status)
	}
	if filters.DateFrom != nil {
		q = q.Where("DATE(created_at) >= ?", *filters.DateFrom)
	}
	if filters.DateTo != nil {
		q = q.Where("DATE(created_at) <= ?", *filters.DateTo)
	}
	if afterCreatedAt != nil && afterID != nil {
		q = q.Where("created_at < ? OR (created_at = ? AND id < ?)", *afterCreatedAt, *afterCreatedAt, *afterID)
	}
	var orders []*models.Order
	err := q.Preload("Items").Order("created_at DESC, id DESC").Limit(limit).Find(&orders).Error
	return orders, err
}

func (r *orderRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*models.Order, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var orders []*models.Order
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Items").
		Order("created_at DESC").
		Limit(limit).
		Find(&orders).Error
	return orders, err
}

func (r *orderRepository) GetByIDForUser(ctx context.Context, id, userID uuid.UUID) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		First(&order, "id = ? AND user_id = ?", id, userID).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}
