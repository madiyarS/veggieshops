package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"gorm.io/gorm"
)

type DistrictRepository interface {
	Create(ctx context.Context, district *models.District) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.District, error)
	GetByStoreID(ctx context.Context, storeID uuid.UUID, activeOnly bool) ([]*models.District, error)
	Update(ctx context.Context, district *models.District) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type districtRepository struct {
	db *gorm.DB
}

func NewDistrictRepository(db *gorm.DB) DistrictRepository {
	return &districtRepository{db: db}
}

func (r *districtRepository) Create(ctx context.Context, district *models.District) error {
	return r.db.WithContext(ctx).Create(district).Error
}

func (r *districtRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.District, error) {
	var district models.District
	err := r.db.WithContext(ctx).Preload("Streets").First(&district, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &district, nil
}

func (r *districtRepository) GetByStoreID(ctx context.Context, storeID uuid.UUID, activeOnly bool) ([]*models.District, error) {
	var districts []*models.District
	query := r.db.WithContext(ctx).Where("store_id = ?", storeID)
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	err := query.Preload("Streets").Find(&districts).Error
	return districts, err
}

func (r *districtRepository) Update(ctx context.Context, district *models.District) error {
	return r.db.WithContext(ctx).Save(district).Error
}

func (r *districtRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.District{}, "id = ?", id).Error
}
