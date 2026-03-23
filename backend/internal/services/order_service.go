package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/repositories"
	"github.com/veggieshop/backend/internal/utils"
	"gorm.io/gorm"
)

type CreateOrderInput struct {
	UserID             *uuid.UUID
	StoreID            uuid.UUID
	DistrictID         uuid.UUID
	DeliveryType       models.DeliveryType
	DeliveryTimeSlotID uuid.UUID
	DeliveryAddress    string
	CustomerPhone      string
	CustomerName       string
	PaymentMethod      models.PaymentMethod
	Items              []OrderItemInput
	Notes              string
}

type OrderItemInput struct {
	ProductID uuid.UUID
	Quantity  int
}

type OrderService struct {
	orderRepo      repositories.OrderRepository
	orderItemRepo  repositories.OrderItemRepository
	productRepo    repositories.ProductRepository
	storeRepo      repositories.StoreRepository
	districtRepo   repositories.DistrictRepository
	slotRepo       repositories.DeliveryTimeSlotRepository
	deliverySvc    *DeliveryService
	invRepo        repositories.InventoryRepository
	stockSvc       *StockService
	workflow       *OrderWorkflow
	db             *gorm.DB
}

func NewOrderService(
	or repositories.OrderRepository,
	oir repositories.OrderItemRepository,
	qi repositories.ProductRepository,
	sr repositories.StoreRepository,
	dr repositories.DistrictRepository,
	slr repositories.DeliveryTimeSlotRepository,
	ds *DeliveryService,
	ir repositories.InventoryRepository,
	stockSvc *StockService,
	workflow *OrderWorkflow,
	db *gorm.DB,
) *OrderService {
	return &OrderService{
		orderRepo:     or,
		orderItemRepo: oir,
		productRepo:   qi,
		storeRepo:     sr,
		districtRepo:  dr,
		slotRepo:      slr,
		deliverySvc:   ds,
		invRepo:       ir,
		stockSvc:      stockSvc,
		workflow:      workflow,
		db:            db,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, input CreateOrderInput) (*models.Order, error) {
	store, err := s.storeRepo.GetByID(ctx, input.StoreID)
	if err != nil {
		return nil, utils.ErrNotFound
	}
	if _, err := s.districtRepo.GetByID(ctx, input.DistrictID); err != nil {
		return nil, utils.ErrNotFound
	}
	fee, err := s.deliverySvc.CalculateDeliveryFee(ctx, input.DistrictID, input.DeliveryType)
	if err != nil {
		return nil, err
	}

	var itemsTotal int
	var orderItems []*models.OrderItem
	for _, it := range input.Items {
		product, err := s.productRepo.GetByID(ctx, it.ProductID)
		if err != nil {
			return nil, utils.ErrNotFound
		}
		if product.StoreID != input.StoreID {
			return nil, utils.ErrNotFound
		}
		if !product.IsAvailable || !product.IsActive || product.TemporarilyUnavailable {
			return nil, utils.ErrInvalidInput
		}
		var subtotal int
		if product.InventoryUnit == models.InventoryUnitWeightGram {
			// price — тенге за 1 кг, quantity — граммы
			subtotal = product.Price * it.Quantity / 1000
			if subtotal <= 0 && it.Quantity > 0 {
				subtotal = 1
			}
		} else {
			subtotal = product.Price * it.Quantity
		}
		itemsTotal += subtotal
		orderItems = append(orderItems, &models.OrderItem{
			OrderID:      uuid.Nil,
			ProductID:    it.ProductID,
			Quantity:     it.Quantity,
			PriceAtOrder: product.Price,
			Subtotal:     subtotal,
		})
	}

	if itemsTotal < store.MinOrderAmount {
		return nil, utils.ErrMinOrderAmount
	}
	totalAmount := itemsTotal + fee

	orderNumber := fmt.Sprintf("ORD-%s-%03d", time.Now().Format("20060102"), time.Now().Unix()%1000)
	order := &models.Order{
		OrderNumber:        orderNumber,
		UserID:             input.UserID,
		StoreID:            input.StoreID,
		DistrictID:         input.DistrictID,
		Status:             models.OrderPreparing,
		DeliveryType:       input.DeliveryType,
		DeliveryTimeSlotID: input.DeliveryTimeSlotID,
		DeliveryAddress:    input.DeliveryAddress,
		CustomerPhone:      input.CustomerPhone,
		CustomerName:       input.CustomerName,
		TotalAmount:        totalAmount,
		DeliveryFee:        fee,
		PaymentMethod:      input.PaymentMethod,
		PaymentStatus:      models.PaymentPending,
		DeliveryCode:       utils.RandomDeliveryCode(),
		Notes:              input.Notes,
	}

	var out *models.Order
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lines := make([]OrderItemInput, len(input.Items))
		for i, it := range input.Items {
			lines[i] = OrderItemInput{ProductID: it.ProductID, Quantity: it.Quantity}
		}
		if err := s.stockSvc.ReserveLinesForNewOrder(tx, ctx, input.StoreID, lines); err != nil {
			if errors.Is(err, repositories.ErrInsufficientInventory) {
				return utils.ErrInsufficientStock
			}
			return err
		}
		if err := tx.WithContext(ctx).Create(order).Error; err != nil {
			return err
		}
		for _, oi := range orderItems {
			oi.OrderID = order.ID
		}
		if err := tx.WithContext(ctx).Create(&orderItems).Error; err != nil {
			return err
		}
		order.Items = orderItems
		out = order
		return nil
	})
	if err != nil {
		if errors.Is(err, utils.ErrInsufficientStock) {
			return nil, utils.ErrInsufficientStock
		}
		return nil, err
	}
	s.stockSvc.NotifyCatalogChanged(ctx, input.StoreID)
	return out, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*models.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, utils.ErrNotFound
	}
	return order, nil
}

func (s *OrderService) GetOrderByNumber(ctx context.Context, orderNumber, phone string) (*models.Order, error) {
	order, err := s.orderRepo.GetByOrderNumber(ctx, orderNumber)
	if err != nil {
		return nil, utils.ErrNotFound
	}
	if utils.NormalizePhone(order.CustomerPhone) != utils.NormalizePhone(phone) {
		return nil, utils.ErrNotFound
	}
	return order, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, newStatus models.OrderStatus) error {
	if newStatus == models.OrderDelivered {
		return utils.ErrDeliveredOnlyViaCode
	}
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return utils.ErrNotFound
	}
	order.Status = newStatus
	return s.orderRepo.Update(ctx, order)
}

func (s *OrderService) GetOrdersByStore(ctx context.Context, storeID uuid.UUID, filters repositories.OrderFilters) ([]*models.Order, error) {
	return s.orderRepo.GetByStoreID(ctx, storeID, filters)
}

// CancelPendingOrder отмена до отгрузки курьером: pending / сборка, снимает резерв.
func (s *OrderService) CancelPendingOrder(ctx context.Context, orderID uuid.UUID) error {
	return s.workflow.CancelPending(ctx, orderID)
}

// AdminReturnFromDelivery отмена доставки: заказ снова в сборке, курьер снят (пока склад не списан по коду).
func (s *OrderService) AdminReturnFromDelivery(ctx context.Context, orderID uuid.UUID) error {
	return s.workflow.AdminReturnFromDelivery(ctx, orderID)
}

// ConfirmOrderStock подтверждение списания без курьера (сборка в магазине).
func (s *OrderService) ConfirmOrderStock(ctx context.Context, orderID uuid.UUID) error {
	return s.workflow.CommitStock(ctx, orderID)
}

// ListOrdersForAdminPaged курсорная пагинация для API v2.
func (s *OrderService) ListOrdersForAdminPaged(ctx context.Context, storeID uuid.UUID, filters repositories.OrderFilters, limit int, afterCreatedAt *time.Time, afterID *uuid.UUID) ([]*models.Order, error) {
	return s.orderRepo.ListByStoreIDPaged(ctx, storeID, filters, limit, afterCreatedAt, afterID)
}

// ListOrdersForCustomer история заказов пользователя (по user_id).
func (s *OrderService) ListOrdersForCustomer(ctx context.Context, userID uuid.UUID, limit int) ([]*models.Order, error) {
	return s.orderRepo.ListByUserID(ctx, userID, limit)
}

// GetOrderForCustomer один заказ, если он принадлежит пользователю.
func (s *OrderService) GetOrderForCustomer(ctx context.Context, userID, orderID uuid.UUID) (*models.Order, error) {
	order, err := s.orderRepo.GetByIDForUser(ctx, orderID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return order, nil
}

// RunStalePendingSweep отменяет pending старше cutoff (фоновая задача).
func (s *OrderService) RunStalePendingSweep(ctx context.Context, olderThan time.Time) (int, error) {
	return s.workflow.ExpireStalePending(ctx, olderThan)
}
