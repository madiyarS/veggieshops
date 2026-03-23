package repositories

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	GetByIDWithSubstitute(ctx context.Context, id uuid.UUID) (*models.Product, error)
	GetByStoreID(ctx context.Context, storeID uuid.UUID, filters ProductFilters) ([]*models.Product, error)
	GetByStoreIDActivePaged(ctx context.Context, storeID uuid.UUID, filters ProductFilters, limit int, afterID *uuid.UUID) ([]*models.Product, error)
	GetByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]*models.Product, error)
	Search(ctx context.Context, storeID uuid.UUID, query string) ([]*models.Product, error)
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByCategoryID(ctx context.Context, categoryID uuid.UUID) (int64, error)
}

type ProductFilters struct {
	CategoryID *uuid.UUID
	ActiveOnly bool
	// NameSearch подстрока по полю name (ILIKE), без % и _ от пользователя
	NameSearch string
}

func sanitizeProductNameSearch(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 100 {
		s = s[:100]
	}
	s = strings.ReplaceAll(s, "%", "")
	s = strings.ReplaceAll(s, "_", "")
	return strings.TrimSpace(s)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *models.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).Preload("Category").First(&product, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetByIDWithSubstitute(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("SubstituteProduct", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "store_id", "is_available", "is_active", "temporarily_unavailable", "price", "unit", "inventory_unit")
		}).
		First(&product, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetByStoreID(ctx context.Context, storeID uuid.UUID, filters ProductFilters) ([]*models.Product, error) {
	var products []*models.Product
	query := r.db.WithContext(ctx).Where("store_id = ?", storeID)
	if filters.CategoryID != nil {
		query = query.Where("category_id = ?", *filters.CategoryID)
	}
	if filters.ActiveOnly {
		query = query.Where("is_active = ? AND is_available = ?", true, true)
	}
	if q := sanitizeProductNameSearch(filters.NameSearch); q != "" {
		query = query.Where("name ILIKE ?", "%"+q+"%")
	}
	err := query.Preload("Category").
		Preload("SubstituteProduct", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "store_id", "is_available", "is_active", "temporarily_unavailable", "price", "unit", "inventory_unit")
		}).
		Find(&products).Error
	return products, err
}

func (r *productRepository) GetByStoreIDActivePaged(ctx context.Context, storeID uuid.UUID, filters ProductFilters, limit int, afterID *uuid.UUID) ([]*models.Product, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var products []*models.Product
	query := r.db.WithContext(ctx).Where("store_id = ?", storeID)
	if filters.CategoryID != nil {
		query = query.Where("category_id = ?", *filters.CategoryID)
	}
	if filters.ActiveOnly {
		query = query.Where("is_active = ? AND is_available = ?", true, true)
	}
	if q := sanitizeProductNameSearch(filters.NameSearch); q != "" {
		query = query.Where("name ILIKE ?", "%"+q+"%")
	}
	if afterID != nil {
		query = query.Where("id > ?", *afterID)
	}
	err := query.Preload("Category").
		Preload("SubstituteProduct", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "store_id", "is_available", "is_active", "temporarily_unavailable", "price", "unit", "inventory_unit")
		}).
		Order("id ASC").
		Limit(limit).
		Find(&products).Error
	return products, err
}

func (r *productRepository) GetByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]*models.Product, error) {
	var products []*models.Product
	err := r.db.WithContext(ctx).Where("category_id = ? AND is_active = ? AND is_available = ?",
		categoryID, true, true).Find(&products).Error
	return products, err
}

func (r *productRepository) Search(ctx context.Context, storeID uuid.UUID, query string) ([]*models.Product, error) {
	var products []*models.Product
	err := r.db.WithContext(ctx).Where("store_id = ? AND is_active = ? AND is_available = ?",
		storeID, true, true).Where("name LIKE ?", "%"+query+"%").Find(&products).Error
	return products, err
}

func (r *productRepository) Update(ctx context.Context, product *models.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

func (r *productRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Product{}, "id = ?", id).Error
}

func (r *productRepository) CountByCategoryID(ctx context.Context, categoryID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Product{}).Where("category_id = ?", categoryID).Count(&count).Error
	return count, err
}
