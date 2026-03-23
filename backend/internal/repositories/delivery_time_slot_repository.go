package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"gorm.io/gorm"
)

type DeliveryTimeSlotRepository interface {
	Create(ctx context.Context, slot *models.DeliveryTimeSlot) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.DeliveryTimeSlot, error)
	GetByStoreID(ctx context.Context, storeID uuid.UUID, dayOfWeek *int) ([]*models.DeliveryTimeSlot, error)
	ListAllByStoreID(ctx context.Context, storeID uuid.UUID) ([]*models.DeliveryTimeSlot, error)
	Update(ctx context.Context, slot *models.DeliveryTimeSlot) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type deliveryTimeSlotRepository struct {
	db *gorm.DB
}

func NewDeliveryTimeSlotRepository(db *gorm.DB) DeliveryTimeSlotRepository {
	return &deliveryTimeSlotRepository{db: db}
}

func (r *deliveryTimeSlotRepository) Create(ctx context.Context, slot *models.DeliveryTimeSlot) error {
	return r.db.WithContext(ctx).Create(slot).Error
}

func (r *deliveryTimeSlotRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.DeliveryTimeSlot, error) {
	var slot models.DeliveryTimeSlot
	err := r.db.WithContext(ctx).First(&slot, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &slot, nil
}

func (r *deliveryTimeSlotRepository) GetByStoreID(ctx context.Context, storeID uuid.UUID, dayOfWeek *int) ([]*models.DeliveryTimeSlot, error) {
	var slots []*models.DeliveryTimeSlot
	query := r.db.WithContext(ctx).Where("store_id = ? AND is_active = ?", storeID, true)
	if dayOfWeek != nil {
		query = query.Where("day_of_week = ?", *dayOfWeek)
	}
	err := query.Order("day_of_week, start_time").Find(&slots).Error
	return slots, err
}

func (r *deliveryTimeSlotRepository) ListAllByStoreID(ctx context.Context, storeID uuid.UUID) ([]*models.DeliveryTimeSlot, error) {
	var slots []*models.DeliveryTimeSlot
	err := r.db.WithContext(ctx).Where("store_id = ?", storeID).Order("day_of_week, start_time").Find(&slots).Error
	return slots, err
}

func (r *deliveryTimeSlotRepository) Update(ctx context.Context, slot *models.DeliveryTimeSlot) error {
	return r.db.WithContext(ctx).Save(slot).Error
}

func (r *deliveryTimeSlotRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.DeliveryTimeSlot{}, "id = ?", id).Error
}
