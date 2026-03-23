package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/catalogcache"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/repositories"
	"github.com/veggieshop/backend/internal/utils"
	"gorm.io/gorm"
)

type ProductService struct {
	productRepo  repositories.ProductRepository
	storeRepo    repositories.StoreRepository
	categoryRepo repositories.CategoryRepository
	invRepo      repositories.InventoryRepository
	stockSvc     *StockService
	catalog      *catalogcache.Store
	db           *gorm.DB
}

func NewProductService(
	pr repositories.ProductRepository,
	sr repositories.StoreRepository,
	cr repositories.CategoryRepository,
	ir repositories.InventoryRepository,
	stockSvc *StockService,
	catalog *catalogcache.Store,
	db *gorm.DB,
) *ProductService {
	return &ProductService{
		productRepo:  pr,
		storeRepo:    sr,
		categoryRepo: cr,
		invRepo:      ir,
		stockSvc:     stockSvc,
		catalog:      catalog,
		db:           db,
	}
}

func catalogCachePart(categoryID *uuid.UUID) string {
	if categoryID == nil {
		return "all"
	}
	return categoryID.String()
}

func (s *ProductService) bumpCatalog(ctx context.Context, storeID uuid.UUID) {
	if s.catalog != nil {
		s.catalog.Bump(ctx, storeID)
	}
}

// Для витрины: доступно = физический остаток − резерв.
func (s *ProductService) attachAvailableInventory(ctx context.Context, storeID uuid.UUID, products []*models.Product) error {
	m, err := s.invRepo.GetAvailableByStore(ctx, storeID)
	if err != nil {
		return nil
	}
	for _, p := range products {
		if q, ok := m[p.ID]; ok {
			p.StockQuantity = q
		} else {
			p.StockQuantity = 0
		}
	}
	return nil
}

func (s *ProductService) attachAvailableOne(ctx context.Context, p *models.Product) error {
	m, err := s.invRepo.GetAvailableByStore(ctx, p.StoreID)
	if err != nil {
		return nil
	}
	if q, ok := m[p.ID]; ok {
		p.StockQuantity = q
	} else {
		p.StockQuantity = 0
	}
	return nil
}

func (s *ProductService) attachAdminInventory(ctx context.Context, storeID uuid.UUID, products []*models.Product) error {
	rows, err := s.invRepo.GetRowsByStore(ctx, storeID)
	if err != nil {
		return nil
	}
	byPID := make(map[uuid.UUID]models.StoreInventory, len(rows))
	for _, r := range rows {
		byPID[r.ProductID] = r
	}
	for _, p := range products {
		if r, ok := byPID[p.ID]; ok {
			p.StockPhysical = r.Quantity
			p.StockReserved = r.ReservedQuantity
			p.StockQuantity = r.Quantity
		} else {
			p.StockPhysical = 0
			p.StockReserved = 0
			p.StockQuantity = 0
		}
	}
	return nil
}

// ProductListOpts фильтры и сортировка витрины (после загрузки и обогащения остатками).
type ProductListOpts struct {
	InStockOnly bool
	Sort        string // name, price_asc, price_desc, expiry_asc — иначе порядок из БД
}

func sortCatalogProducts(products []*models.Product, sortKey string) {
	switch sortKey {
	case "price_asc":
		sort.SliceStable(products, func(i, j int) bool { return products[i].Price < products[j].Price })
	case "price_desc":
		sort.SliceStable(products, func(i, j int) bool { return products[i].Price > products[j].Price })
	case "name":
		sort.SliceStable(products, func(i, j int) bool {
			return strings.ToLower(products[i].Name) < strings.ToLower(products[j].Name)
		})
	case "expiry_asc":
		sort.SliceStable(products, func(i, j int) bool {
			a, b := products[i].NearestBatchExpiresAt, products[j].NearestBatchExpiresAt
			if a == nil && b == nil {
				return false
			}
			if a == nil {
				return false // без срока — в конец
			}
			if b == nil {
				return true
			}
			return a.Before(*b)
		})
	}
}

func filterCatalogInStock(products []*models.Product) []*models.Product {
	var out []*models.Product
	for _, p := range products {
		if p.StockQuantity > 0 && !p.TemporarilyUnavailable {
			out = append(out, p)
		}
	}
	return out
}

func applyCatalogListOpts(products []*models.Product, opts ProductListOpts) []*models.Product {
	out := products
	if opts.InStockOnly {
		out = filterCatalogInStock(out)
	}
	if opts.Sort != "" {
		sortCatalogProducts(out, opts.Sort)
	}
	return out
}

func (s *ProductService) attachCatalogDisplayHints(ctx context.Context, storeID uuid.UUID, products []*models.Product) error {
	if len(products) == 0 {
		return nil
	}
	var expMin map[uuid.UUID]time.Time
	if s.stockSvc != nil {
		m, err := s.stockSvc.MinExpiresAtByProduct(ctx, storeID)
		if err == nil {
			expMin = m
		}
	}
	for _, p := range products {
		p.NearestBatchExpiresAt = nil
		p.CatalogLowStock = false
		if expMin != nil {
			if t, ok := expMin[p.ID]; ok {
				t2 := t
				p.NearestBatchExpiresAt = &t2
			}
		}
		if p.ReorderMinQty > 0 && p.StockQuantity > 0 && p.StockQuantity < p.ReorderMinQty {
			p.CatalogLowStock = true
		}
	}
	return nil
}

// GetAvailableByProductIDs доступный остаток (физический − резерв) по списку товаров магазина.
func (s *ProductService) GetAvailableByProductIDs(ctx context.Context, storeID uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]int, error) {
	m, err := s.invRepo.GetAvailableByStore(ctx, storeID)
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]int, len(ids))
	for _, id := range ids {
		out[id] = m[id]
	}
	return out, nil
}

func (s *ProductService) GetProductsByStore(ctx context.Context, storeID uuid.UUID, categoryID *uuid.UUID, nameSearch string, opts ProductListOpts) ([]*models.Product, error) {
	nameSearch = strings.TrimSpace(nameSearch)
	part := catalogCachePart(categoryID)
	if nameSearch == "" && s.catalog != nil {
		if raw, ok := s.catalog.GetProductsJSON(ctx, storeID, part); ok {
			var cached []*models.Product
			if err := json.Unmarshal(raw, &cached); err == nil {
				if err := s.attachAvailableInventory(ctx, storeID, cached); err != nil {
					return nil, err
				}
				if err := s.attachCatalogDisplayHints(ctx, storeID, cached); err != nil {
					return nil, err
				}
				base := append([]*models.Product(nil), cached...)
				return applyCatalogListOpts(base, opts), nil
			}
		}
	}
	filters := repositories.ProductFilters{CategoryID: categoryID, ActiveOnly: true, NameSearch: nameSearch}
	list, err := s.productRepo.GetByStoreID(ctx, storeID, filters)
	if err != nil {
		return nil, err
	}
	if err := s.attachAvailableInventory(ctx, storeID, list); err != nil {
		return nil, err
	}
	if err := s.attachCatalogDisplayHints(ctx, storeID, list); err != nil {
		return nil, err
	}
	if nameSearch == "" && s.catalog != nil {
		_ = s.catalog.SetProductsJSON(ctx, storeID, part, list)
	}
	base := append([]*models.Product(nil), list...)
	return applyCatalogListOpts(base, opts), nil
}

// GetProductsByStorePaged витрина с keyset по id ASC (для API v2).
func (s *ProductService) GetProductsByStorePaged(ctx context.Context, storeID uuid.UUID, categoryID *uuid.UUID, nameSearch string, limit int, afterID *uuid.UUID) ([]*models.Product, error) {
	filters := repositories.ProductFilters{CategoryID: categoryID, ActiveOnly: true, NameSearch: strings.TrimSpace(nameSearch)}
	list, err := s.productRepo.GetByStoreIDActivePaged(ctx, storeID, filters, limit, afterID)
	if err != nil {
		return nil, err
	}
	if err := s.attachAvailableInventory(ctx, storeID, list); err != nil {
		return nil, err
	}
	return list, nil
}

func (s *ProductService) GetProductsByCategory(ctx context.Context, categoryID uuid.UUID) ([]*models.Product, error) {
	return s.productRepo.GetByCategoryID(ctx, categoryID)
}

func (s *ProductService) GetProductByID(ctx context.Context, productID uuid.UUID) (*models.Product, error) {
	product, err := s.productRepo.GetByIDWithSubstitute(ctx, productID)
	if err != nil {
		return nil, utils.ErrNotFound
	}
	if err := s.attachAvailableOne(ctx, product); err != nil {
		return nil, err
	}
	_ = s.attachCatalogDisplayHints(ctx, product.StoreID, []*models.Product{product})
	return product, nil
}

func (s *ProductService) CreateProduct(ctx context.Context, storeID uuid.UUID, product *models.Product) (*models.Product, error) {
	if _, err := s.storeRepo.GetByID(ctx, storeID); err != nil {
		return nil, utils.ErrNotFound
	}
	if _, err := s.categoryRepo.GetByID(ctx, product.CategoryID); err != nil {
		return nil, utils.ErrNotFound
	}
	product.StoreID = storeID
	qty := product.StockQuantity
	if qty < 0 {
		qty = 0
	}
	product.StockQuantity = qty
	if product.Unit == "" {
		product.Unit = "шт"
	}
	if product.InventoryUnit == "" {
		product.InventoryUnit = models.InventoryUnitPiece
	}
	if product.CartStepGrams <= 0 {
		product.CartStepGrams = 250
	}
	if product.SubstituteProductID != nil {
		sub, err := s.productRepo.GetByID(ctx, *product.SubstituteProductID)
		if err != nil || sub.StoreID != storeID {
			return nil, utils.ErrNotFound
		}
	}
	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}
	if err := s.invRepo.Upsert(ctx, storeID, product.ID, qty); err != nil {
		return nil, err
	}
	if err := s.stockSvc.RebuildInventoryToAbsolute(ctx, storeID, product.ID, qty); err != nil {
		return nil, err
	}
	if err := s.stockSvc.MirrorProductStock(ctx, storeID, product.ID); err != nil {
		return nil, err
	}
	return product, nil
}

// ListProductsForAdmin все товары магазина (включая скрытые); физический остаток и резерв.
func (s *ProductService) ListProductsForAdmin(ctx context.Context, storeID uuid.UUID, categoryID *uuid.UUID) ([]*models.Product, error) {
	filters := repositories.ProductFilters{CategoryID: categoryID, ActiveOnly: false}
	list, err := s.productRepo.GetByStoreID(ctx, storeID, filters)
	if err != nil {
		return nil, err
	}
	if err := s.attachAdminInventory(ctx, storeID, list); err != nil {
		return nil, err
	}
	return list, nil
}

func (s *ProductService) CountByCategoryID(ctx context.Context, categoryID uuid.UUID) (int64, error) {
	return s.productRepo.CountByCategoryID(ctx, categoryID)
}

// ProductPatch частичное обновление товара в админке.
type ProductPatch struct {
	CategoryID               *uuid.UUID
	Name                     *string
	Description              *string
	Price                    *int
	WeightGram               *int
	Unit                     *string
	StockQuantity            *int
	ImageURL                 *string
	Origin                   *string
	ShelfLifeDays            *int
	ClearShelfLife           bool
	IsAvailable              *bool
	IsActive                 *bool
	InventoryUnit            *models.InventoryUnit
	PackageGrams             *int
	ClearPackageGrams        bool
	IsSeasonal               *bool
	TemporarilyUnavailable   *bool
	SubstituteProductID      *uuid.UUID
	ClearSubstitute          bool
	ReorderMinQty            *int
	CartStepGrams            *int
}

func (s *ProductService) PatchProduct(ctx context.Context, productID uuid.UUID, patch *ProductPatch) error {
	if patch == nil {
		return nil
	}
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return utils.ErrNotFound
	}
	if patch.CategoryID != nil {
		if _, err := s.categoryRepo.GetByID(ctx, *patch.CategoryID); err != nil {
			return utils.ErrNotFound
		}
		product.CategoryID = *patch.CategoryID
	}
	if patch.Name != nil {
		product.Name = *patch.Name
	}
	if patch.Description != nil {
		product.Description = *patch.Description
	}
	if patch.Price != nil {
		product.Price = *patch.Price
	}
	if patch.WeightGram != nil {
		product.WeightGram = *patch.WeightGram
	}
	if patch.Unit != nil {
		product.Unit = *patch.Unit
	}
	if patch.StockQuantity != nil {
		q := *patch.StockQuantity
		if q < 0 {
			q = 0
		}
		if err := s.invRepo.Upsert(ctx, product.StoreID, productID, q); err != nil {
			return err
		}
		if err := s.stockSvc.RebuildInventoryToAbsolute(ctx, product.StoreID, productID, q); err != nil {
			return err
		}
		if err := s.stockSvc.MirrorProductStock(ctx, product.StoreID, productID); err != nil {
			return err
		}
		product.StockQuantity = q
	}
	if patch.ImageURL != nil {
		product.ImageURL = *patch.ImageURL
	}
	if patch.Origin != nil {
		product.Origin = *patch.Origin
	}
	if patch.ClearShelfLife {
		product.ShelfLifeDays = nil
	} else if patch.ShelfLifeDays != nil {
		product.ShelfLifeDays = patch.ShelfLifeDays
	}
	if patch.IsAvailable != nil {
		product.IsAvailable = *patch.IsAvailable
	}
	if patch.IsActive != nil {
		product.IsActive = *patch.IsActive
	}
	if patch.InventoryUnit != nil {
		product.InventoryUnit = *patch.InventoryUnit
	}
	if patch.ClearPackageGrams {
		product.PackageGrams = nil
	} else if patch.PackageGrams != nil {
		product.PackageGrams = patch.PackageGrams
	}
	if patch.IsSeasonal != nil {
		product.IsSeasonal = *patch.IsSeasonal
	}
	if patch.TemporarilyUnavailable != nil {
		product.TemporarilyUnavailable = *patch.TemporarilyUnavailable
	}
	if patch.ClearSubstitute {
		product.SubstituteProductID = nil
	} else if patch.SubstituteProductID != nil {
		sub, err := s.productRepo.GetByID(ctx, *patch.SubstituteProductID)
		if err != nil || sub.StoreID != product.StoreID {
			return utils.ErrNotFound
		}
		product.SubstituteProductID = patch.SubstituteProductID
	}
	if patch.ReorderMinQty != nil {
		product.ReorderMinQty = *patch.ReorderMinQty
		if product.ReorderMinQty < 0 {
			product.ReorderMinQty = 0
		}
	}
	if patch.CartStepGrams != nil {
		product.CartStepGrams = *patch.CartStepGrams
		if product.CartStepGrams < 50 {
			product.CartStepGrams = 50
		}
	}
	if err := s.productRepo.Update(ctx, product); err != nil {
		return err
	}
	s.bumpCatalog(ctx, product.StoreID)
	return nil
}

// DeactivateProduct скрывает товар (заказы ссылаются на product_id — физическое удаление невозможно).
func (s *ProductService) DeactivateProduct(ctx context.Context, productID uuid.UUID) error {
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return utils.ErrNotFound
	}
	product.IsActive = false
	product.IsAvailable = false
	if err := s.productRepo.Update(ctx, product); err != nil {
		return err
	}
	s.bumpCatalog(ctx, product.StoreID)
	return nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, productID uuid.UUID, updates *models.Product) error {
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return utils.ErrNotFound
	}
	if updates.Name != "" {
		product.Name = updates.Name
	}
	if updates.Description != "" {
		product.Description = updates.Description
	}
	if updates.Price > 0 {
		product.Price = updates.Price
	}
	if updates.WeightGram > 0 {
		product.WeightGram = updates.WeightGram
	}
	if updates.Unit != "" {
		product.Unit = updates.Unit
	}
	if updates.StockQuantity >= 0 {
		if err := s.invRepo.Upsert(ctx, product.StoreID, productID, updates.StockQuantity); err != nil {
			return err
		}
		if err := s.stockSvc.RebuildInventoryToAbsolute(ctx, product.StoreID, productID, updates.StockQuantity); err != nil {
			return err
		}
		if err := s.stockSvc.MirrorProductStock(ctx, product.StoreID, productID); err != nil {
			return err
		}
		product.StockQuantity = updates.StockQuantity
	}
	if updates.ImageURL != "" {
		product.ImageURL = updates.ImageURL
	}
	if updates.Origin != "" {
		product.Origin = updates.Origin
	}
	if updates.ShelfLifeDays != nil {
		product.ShelfLifeDays = updates.ShelfLifeDays
	}
	product.IsAvailable = updates.IsAvailable
	if err := s.productRepo.Update(ctx, product); err != nil {
		return err
	}
	s.bumpCatalog(ctx, product.StoreID)
	return nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	p, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}
	if err := s.productRepo.Delete(ctx, productID); err != nil {
		return err
	}
	s.bumpCatalog(ctx, p.StoreID)
	return nil
}

func (s *ProductService) SearchProducts(ctx context.Context, storeID uuid.UUID, query string) ([]*models.Product, error) {
	list, err := s.productRepo.Search(ctx, storeID, query)
	if err != nil {
		return nil, err
	}
	if err := s.attachAvailableInventory(ctx, storeID, list); err != nil {
		return nil, err
	}
	return list, nil
}

// CloneCatalogFromStore копирует номенклатуру, поля витрины/склада и партии (или одну партию по остатку).
func (s *ProductService) CloneCatalogFromStore(ctx context.Context, targetStoreID, sourceStoreID uuid.UUID) (int, error) {
	if targetStoreID == sourceStoreID {
		return 0, fmt.Errorf("нельзя копировать каталог в тот же магазин")
	}
	if _, err := s.storeRepo.GetByID(ctx, targetStoreID); err != nil {
		return 0, utils.ErrNotFound
	}
	if _, err := s.storeRepo.GetByID(ctx, sourceStoreID); err != nil {
		return 0, utils.ErrNotFound
	}
	src, err := s.productRepo.GetByStoreID(ctx, sourceStoreID, repositories.ProductFilters{ActiveOnly: false})
	if err != nil {
		return 0, err
	}
	qtyMap, err := s.invRepo.GetQuantitiesByStore(ctx, sourceStoreID)
	if err != nil {
		return 0, err
	}
	var n int
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, p := range src {
			q := qtyMap[p.ID]
			if q == 0 && p.StockQuantity > 0 {
				q = p.StockQuantity
			}
			np := models.Product{
				StoreID:                targetStoreID,
				CategoryID:             p.CategoryID,
				Name:                   p.Name,
				Description:            p.Description,
				Price:                  p.Price,
				WeightGram:             p.WeightGram,
				Unit:                   p.Unit,
				StockQuantity:          q,
				ImageURL:               p.ImageURL,
				Origin:                 p.Origin,
				ShelfLifeDays:          p.ShelfLifeDays,
				IsAvailable:            p.IsAvailable,
				IsActive:               p.IsActive,
				InventoryUnit:          p.InventoryUnit,
				PackageGrams:           p.PackageGrams,
				IsSeasonal:             p.IsSeasonal,
				TemporarilyUnavailable: p.TemporarilyUnavailable,
				SubstituteProductID:    nil,
				ReorderMinQty:          p.ReorderMinQty,
				CartStepGrams:          p.CartStepGrams,
			}
			if np.Unit == "" {
				np.Unit = "шт"
			}
			if np.InventoryUnit == "" {
				np.InventoryUnit = models.InventoryUnitPiece
			}
			if np.CartStepGrams <= 0 {
				np.CartStepGrams = 250
			}
			if err := tx.WithContext(ctx).Create(&np).Error; err != nil {
				return err
			}
			if err := s.invRepo.UpsertTx(tx, ctx, targetStoreID, np.ID, q); err != nil {
				return err
			}
			var batchCount int64
			if err := tx.WithContext(ctx).Model(&models.StockBatch{}).
				Where("store_id = ? AND product_id = ? AND quantity > 0", sourceStoreID, p.ID).
				Count(&batchCount).Error; err != nil {
				return err
			}
			if batchCount > 0 {
				if err := s.stockSvc.CopyBatchesBetweenProductsTx(tx, ctx, sourceStoreID, targetStoreID, p.ID, np.ID); err != nil {
					return err
				}
			} else if q > 0 {
				if err := s.stockSvc.RebuildInventoryToAbsoluteTx(tx, ctx, targetStoreID, np.ID, q); err != nil {
					return err
				}
			}
			if err := s.stockSvc.MirrorProductStockTx(tx, ctx, targetStoreID, np.ID); err != nil {
				return err
			}
			n++
		}
		return nil
	})
	if err != nil {
		return n, err
	}
	s.stockSvc.NotifyCatalogChanged(ctx, targetStoreID)
	return n, nil
}
