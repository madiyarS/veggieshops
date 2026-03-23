package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"gorm.io/gorm"
)

type CourierRepository interface {
	Create(ctx context.Context, courier *models.Courier) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Courier, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Courier, error)
	GetByStoreID(ctx context.Context, storeID uuid.UUID, activeOnly bool) ([]*models.Courier, error)
	Update(ctx context.Context, courier *models.Courier) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type courierRepository struct {
	db *gorm.DB
}

func NewCourierRepository(db *gorm.DB) CourierRepository {
	return &courierRepository{db: db}
}

func (r *courierRepository) Create(ctx context.Context, courier *models.Courier) error {
	return r.db.WithContext(ctx).Create(courier).Error
}

func (r *courierRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Courier, error) {
	var courier models.Courier
	err := r.db.WithContext(ctx).First(&courier, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &courier, nil
}

func (r *courierRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Courier, error) {
	var courier models.Courier
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID, true).First(&courier).Error
	if err != nil {
		return nil, err
	}
	return &courier, nil
}

func (r *courierRepository) GetByStoreID(ctx context.Context, storeID uuid.UUID, activeOnly bool) ([]*models.Courier, error) {
	var couriers []*models.Courier
	query := r.db.WithContext(ctx).Where("store_id = ?", storeID)
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	err := query.Find(&couriers).Error
	return couriers, err
}

func (r *courierRepository) Update(ctx context.Context, courier *models.Courier) error {
	return r.db.WithContext(ctx).Save(courier).Error
}

func (r *courierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Courier{}, "id = ?", id).Error
}
