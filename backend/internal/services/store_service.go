package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/repositories"
	"github.com/veggieshop/backend/internal/utils"
)

type StoreService struct {
	storeRepo repositories.StoreRepository
}

func NewStoreService(sr repositories.StoreRepository) *StoreService {
	return &StoreService{storeRepo: sr}
}

func (s *StoreService) GetAllStores(ctx context.Context, activeOnly bool) ([]*models.Store, error) {
	return s.storeRepo.GetAll(ctx, activeOnly)
}

func (s *StoreService) GetStoreByID(ctx context.Context, storeID uuid.UUID) (*models.Store, error) {
	store, err := s.storeRepo.GetByID(ctx, storeID)
	if err != nil {
		return nil, utils.ErrNotFound
	}
	return store, nil
}

func (s *StoreService) CreateStore(ctx context.Context, store *models.Store) (*models.Store, error) {
	if err := s.storeRepo.Create(ctx, store); err != nil {
		return nil, err
	}
	return store, nil
}

// StorePatch частичное обновление магазина (только переданные поля).
type StorePatch struct {
	Name               *string  `json:"name"`
	Description        *string  `json:"description"`
	Address            *string  `json:"address"`
	Latitude           *float64 `json:"latitude"`
	Longitude          *float64 `json:"longitude"`
	Phone              *string  `json:"phone"`
	Email              *string  `json:"email"`
	DeliveryRadiusKm   *float64 `json:"delivery_radius_km"`
	MinOrderAmount     *int     `json:"min_order_amount"`
	MaxOrderWeightKg   *float64 `json:"max_order_weight_kg"`
	ClearMaxWeight     *bool    `json:"clear_max_weight"`
	IsActive           *bool    `json:"is_active"`
	WorkingHoursStart  *string  `json:"working_hours_start"`
	WorkingHoursEnd    *string  `json:"working_hours_end"`
	ClearWorkingStart  *bool    `json:"clear_working_hours_start"`
	ClearWorkingEnd    *bool    `json:"clear_working_hours_end"`
}

func (s *StoreService) PatchStore(ctx context.Context, storeID uuid.UUID, p *StorePatch) error {
	store, err := s.storeRepo.GetByID(ctx, storeID)
	if err != nil {
		return utils.ErrNotFound
	}
	if p.Name != nil {
		store.Name = *p.Name
	}
	if p.Description != nil {
		store.Description = *p.Description
	}
	if p.Address != nil {
		store.Address = *p.Address
	}
	if p.Latitude != nil {
		store.Latitude = *p.Latitude
	}
	if p.Longitude != nil {
		store.Longitude = *p.Longitude
	}
	if p.Phone != nil {
		store.Phone = *p.Phone
	}
	if p.Email != nil {
		store.Email = *p.Email
	}
	if p.DeliveryRadiusKm != nil {
		store.DeliveryRadiusKm = *p.DeliveryRadiusKm
	}
	if p.MinOrderAmount != nil {
		store.MinOrderAmount = *p.MinOrderAmount
	}
	if p.ClearMaxWeight != nil && *p.ClearMaxWeight {
		store.MaxOrderWeightKg = nil
	} else if p.MaxOrderWeightKg != nil {
		v := *p.MaxOrderWeightKg
		store.MaxOrderWeightKg = &v
	}
	if p.IsActive != nil {
		store.IsActive = *p.IsActive
	}
	if p.ClearWorkingStart != nil && *p.ClearWorkingStart {
		store.WorkingHoursStart = nil
	} else if p.WorkingHoursStart != nil {
		v := *p.WorkingHoursStart
		store.WorkingHoursStart = &v
	}
	if p.ClearWorkingEnd != nil && *p.ClearWorkingEnd {
		store.WorkingHoursEnd = nil
	} else if p.WorkingHoursEnd != nil {
		v := *p.WorkingHoursEnd
		store.WorkingHoursEnd = &v
	}
	return s.storeRepo.Update(ctx, store)
}

func (s *StoreService) UpdateDeliveryRadius(ctx context.Context, storeID uuid.UUID, radiusKm float64) error {
	store, err := s.storeRepo.GetByID(ctx, storeID)
	if err != nil {
		return utils.ErrNotFound
	}
	store.DeliveryRadiusKm = radiusKm
	return s.storeRepo.Update(ctx, store)
}
