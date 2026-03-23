package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/repositories"
	"github.com/veggieshop/backend/internal/utils"
)

type DeliveryAvailability struct {
	IsAvailable       bool              `json:"is_available"`
	DistrictID        *uuid.UUID        `json:"district_id,omitempty"`
	DistrictName      string             `json:"district_name,omitempty"`
	DistanceKm        float64            `json:"distance_km"`
	DeliveryFeeRegular int               `json:"delivery_fee_regular"`
	DeliveryFeeExpress int               `json:"delivery_fee_express"`
	MinOrderAmount    int                `json:"min_order_amount"`
	MaxOrderWeightKg  *float64           `json:"max_order_weight_kg,omitempty"`
	Districts         []*models.District `json:"districts,omitempty"`
	Error             string             `json:"error,omitempty"`
}

type TimeSlotAvailability struct {
	ID            uuid.UUID `json:"id"`
	StartTime     string    `json:"start_time"`
	EndTime       string    `json:"end_time"`
	AvailableSlots int       `json:"available_slots"`
	MaxOrders     int       `json:"max_orders"`
}

type DeliveryService struct {
	storeRepo   repositories.StoreRepository
	districtRepo repositories.DistrictRepository
	orderRepo   repositories.OrderRepository
	slotRepo   repositories.DeliveryTimeSlotRepository
}

func NewDeliveryService(
	sr repositories.StoreRepository,
	dr repositories.DistrictRepository,
	or repositories.OrderRepository,
	slr repositories.DeliveryTimeSlotRepository,
) *DeliveryService {
	return &DeliveryService{
		storeRepo:   sr,
		districtRepo: dr,
		orderRepo:   or,
		slotRepo:   slr,
	}
}

func (s *DeliveryService) CheckDeliveryAvailability(ctx context.Context, storeID uuid.UUID, customerLat, customerLon float64) (*DeliveryAvailability, error) {
	store, err := s.storeRepo.GetByID(ctx, storeID)
	if err != nil {
		slog.Error("store not found", "store_id", storeID, "error", err)
		return nil, utils.ErrNotFound
	}
	if !store.IsActive {
		return nil, utils.ErrNotFound
	}

	distance := utils.Haversine(store.Latitude, store.Longitude, customerLat, customerLon)

	if distance > store.DeliveryRadiusKm {
		slog.Info("delivery unavailable - distance exceeds radius",
			"distance", distance, "radius", store.DeliveryRadiusKm)
		return &DeliveryAvailability{
			IsAvailable: false,
			DistanceKm:  distance,
			Error:       "Доставка в ваш район недоступна. Расстояние превышает радиус доставки.",
		}, nil
	}

	districts, err := s.districtRepo.GetByStoreID(ctx, storeID, true)
	if err != nil {
		return nil, err
	}

	if len(districts) == 0 {
		return &DeliveryAvailability{
			IsAvailable:      true,
			DistanceKm:       distance,
			MinOrderAmount:   store.MinOrderAmount,
			MaxOrderWeightKg: store.MaxOrderWeightKg,
			Districts:        districts,
		}, nil
	}

	// Use first district for fee display; client selects from Districts list
	nearest := districts[0]
	for _, d := range districts {
		if d.DistanceKm <= distance && d.DistanceKm >= nearest.DistanceKm {
			nearest = d
		}
	}

	return &DeliveryAvailability{
		IsAvailable:        true,
		DistrictID:         &nearest.ID,
		DistrictName:       nearest.Name,
		DistanceKm:         distance,
		DeliveryFeeRegular: nearest.DeliveryFeeRegular,
		DeliveryFeeExpress: nearest.DeliveryFeeExpress,
		MinOrderAmount:     store.MinOrderAmount,
		MaxOrderWeightKg:   store.MaxOrderWeightKg,
		Districts:          districts,
	}, nil
}

func (s *DeliveryService) GetDistrictsByStore(ctx context.Context, storeID uuid.UUID) ([]*models.District, error) {
	return s.districtRepo.GetByStoreID(ctx, storeID, true)
}

func (s *DeliveryService) CalculateDeliveryFee(ctx context.Context, districtID uuid.UUID, deliveryType models.DeliveryType) (int, error) {
	district, err := s.districtRepo.GetByID(ctx, districtID)
	if err != nil {
		return 0, utils.ErrNotFound
	}
	if deliveryType == models.DeliveryExpress {
		return district.DeliveryFeeExpress, nil
	}
	return district.DeliveryFeeRegular, nil
}

func (s *DeliveryService) GetAvailableTimeSlots(ctx context.Context, storeID uuid.UUID, date string) ([]*TimeSlotAvailability, error) {
	// Календарная дата в часовом поясе Казахстана (как у клиентов); в БД day_of_week: 0=воскресенье … 6=суббота (как time.Weekday в Go).
	loc, err := time.LoadLocation("Asia/Almaty")
	if err != nil {
		loc = time.UTC
	}
	t, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return nil, err
	}
	dayOfWeek := int(t.Weekday())

	slots, err := s.slotRepo.GetByStoreID(ctx, storeID, &dayOfWeek)
	if err != nil {
		return nil, err
	}

	result := make([]*TimeSlotAvailability, 0, len(slots))
	for _, slot := range slots {
		count, _ := s.orderRepo.CountByTimeSlot(ctx, slot.ID, models.OrderCancelled)
		available := slot.MaxOrders - int(count)
		if available < 0 {
			available = 0
		}
		result = append(result, &TimeSlotAvailability{
			ID:            slot.ID,
			StartTime:     slot.StartTime,
			EndTime:       slot.EndTime,
			AvailableSlots: available,
			MaxOrders:     slot.MaxOrders,
		})
	}
	return result, nil
}
