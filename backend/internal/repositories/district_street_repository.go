package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"gorm.io/gorm"
)

type DistrictStreetRepository interface {
	Create(ctx context.Context, street *models.DistrictStreet) error
	GetByDistrictID(ctx context.Context, districtID uuid.UUID) ([]*models.DistrictStreet, error)
	DeleteByDistrictID(ctx context.Context, districtID uuid.UUID) error
}

type districtStreetRepository struct {
	db *gorm.DB
}

func NewDistrictStreetRepository(db *gorm.DB) DistrictStreetRepository {
	return &districtStreetRepository{db: db}
}

func (r *districtStreetRepository) Create(ctx context.Context, street *models.DistrictStreet) error {
	return r.db.WithContext(ctx).Create(street).Error
}

func (r *districtStreetRepository) GetByDistrictID(ctx context.Context, districtID uuid.UUID) ([]*models.DistrictStreet, error) {
	var streets []*models.DistrictStreet
	err := r.db.WithContext(ctx).Where("district_id = ?", districtID).Find(&streets).Error
	return streets, err
}

func (r *districtStreetRepository) DeleteByDistrictID(ctx context.Context, districtID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("district_id = ?", districtID).Delete(&models.DistrictStreet{}).Error
}
