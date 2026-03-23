package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"gorm.io/gorm"
)

type StoreRepository interface {
	Create(ctx context.Context, store *models.Store) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Store, error)
	GetAll(ctx context.Context, activeOnly bool) ([]*models.Store, error)
	Update(ctx context.Context, store *models.Store) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type storeRepository struct {
	db *gorm.DB
}

func NewStoreRepository(db *gorm.DB) StoreRepository {
	return &storeRepository{db: db}
}

func (r *storeRepository) Create(ctx context.Context, store *models.Store) error {
	return r.db.WithContext(ctx).Create(store).Error
}

func (r *storeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Store, error) {
	var store models.Store
	err := r.db.WithContext(ctx).First(&store, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &store, nil
}

func (r *storeRepository) GetAll(ctx context.Context, activeOnly bool) ([]*models.Store, error) {
	var stores []*models.Store
	query := r.db.WithContext(ctx)
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	err := query.Find(&stores).Error
	return stores, err
}

func (r *storeRepository) Update(ctx context.Context, store *models.Store) error {
	return r.db.WithContext(ctx).Save(store).Error
}

func (r *storeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Store{}, "id = ?", id).Error
}
