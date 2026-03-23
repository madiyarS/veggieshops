package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"gorm.io/gorm"
)

type OrderItemRepository interface {
	Create(ctx context.Context, item *models.OrderItem) error
	CreateBatch(ctx context.Context, items []*models.OrderItem) error
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*models.OrderItem, error)
}

type orderItemRepository struct {
	db *gorm.DB
}

func NewOrderItemRepository(db *gorm.DB) OrderItemRepository {
	return &orderItemRepository{db: db}
}

func (r *orderItemRepository) Create(ctx context.Context, item *models.OrderItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *orderItemRepository) CreateBatch(ctx context.Context, items []*models.OrderItem) error {
	return r.db.WithContext(ctx).Create(&items).Error
}

func (r *orderItemRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*models.OrderItem, error) {
	var items []*models.OrderItem
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&items).Error
	return items, err
}
