package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/repositories"
	"github.com/veggieshop/backend/internal/utils"
)

type CourierService struct {
	courierRepo repositories.CourierRepository
	orderRepo   repositories.OrderRepository
	workflow    *OrderWorkflow
}

func NewCourierService(cr repositories.CourierRepository, or repositories.OrderRepository, workflow *OrderWorkflow) *CourierService {
	return &CourierService{
		courierRepo: cr,
		orderRepo:   or,
		workflow:    workflow,
	}
}

func (s *CourierService) GetCouriersByStore(ctx context.Context, storeID uuid.UUID, activeOnly bool) ([]*models.Courier, error) {
	return s.courierRepo.GetByStoreID(ctx, storeID, activeOnly)
}

func (s *CourierService) AssignCourierToOrder(ctx context.Context, orderID, courierID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return utils.ErrNotFound
	}
	courier, err := s.courierRepo.GetByID(ctx, courierID)
	if err != nil {
		return utils.ErrNotFound
	}
	if courier.StoreID != order.StoreID {
		return utils.ErrForbidden
	}
	order.CourierID = &courier.UserID
	return s.orderRepo.Update(ctx, order)
}

func (s *CourierService) GetCourierOrders(ctx context.Context, courierID uuid.UUID) ([]*models.Order, error) {
	courier, err := s.courierRepo.GetByID(ctx, courierID)
	if err != nil {
		return nil, utils.ErrNotFound
	}
	return s.orderRepo.GetByStoreID(ctx, courier.StoreID, repositories.OrderFilters{})
}

func (s *CourierService) courierByUser(ctx context.Context, userID uuid.UUID) (*models.Courier, error) {
	c, err := s.courierRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, utils.ErrCourierProfile
	}
	return c, nil
}

// ListMyOrders активные заказы магазина курьера.
func (s *CourierService) ListMyOrders(ctx context.Context, userID uuid.UUID) ([]*models.Order, error) {
	c, err := s.courierByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.orderRepo.ListActiveForStore(ctx, c.StoreID)
}

// AcceptOrder курьер забирает заказ: списание склада (FEFO) + статус «в доставке».
func (s *CourierService) AcceptOrder(ctx context.Context, userID, orderID uuid.UUID) error {
	c, err := s.courierByUser(ctx, userID)
	if err != nil {
		return err
	}
	return s.workflow.AcceptByCourier(ctx, c.StoreID, userID, orderID)
}

// CompleteDelivery завершение: списание со склада + доставлен (код от клиента).
func (s *CourierService) CompleteDelivery(ctx context.Context, userID, orderID uuid.UUID, code string) error {
	if _, err := s.courierByUser(ctx, userID); err != nil {
		return err
	}
	return s.workflow.CompleteDeliveryForCourier(ctx, userID, orderID, code)
}
