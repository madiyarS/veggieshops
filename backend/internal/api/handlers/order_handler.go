package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/api/middleware"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/services"
	"github.com/veggieshop/backend/internal/utils"
)

type OrderHandler struct {
	orderSvc *services.OrderService
}

func NewOrderHandler(orderSvc *services.OrderService) *OrderHandler {
	return &OrderHandler{orderSvc: orderSvc}
}

type CreateOrderRequest struct {
	StoreID            string              `json:"store_id" binding:"required"`
	DistrictID         string              `json:"district_id" binding:"required"`
	DeliveryType       string              `json:"delivery_type" binding:"required"`
	DeliveryTimeSlotID string              `json:"delivery_time_slot_id" binding:"required"`
	DeliveryAddress    string              `json:"delivery_address" binding:"required"`
	CustomerPhone      string              `json:"customer_phone" binding:"required"`
	CustomerName       string              `json:"customer_name" binding:"required"`
	PaymentMethod      string              `json:"payment_method" binding:"required"`
	Items              []CreateOrderItem   `json:"items" binding:"required"`
	Notes              string              `json:"notes"`
}

type CreateOrderItem struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinError(c, http.StatusBadRequest, "Неверные данные: "+err.Error(), utils.CodeValidation)
		return
	}
	storeID, _ := uuid.Parse(req.StoreID)
	districtID, _ := uuid.Parse(req.DistrictID)
	slotID, _ := uuid.Parse(req.DeliveryTimeSlotID)
	var deliveryType models.DeliveryType
	if req.DeliveryType == "express" {
		deliveryType = models.DeliveryExpress
	} else {
		deliveryType = models.DeliveryRegular
	}
	var paymentMethod models.PaymentMethod
	switch req.PaymentMethod {
	case "kaspi":
		paymentMethod = models.PaymentKaspi
	case "halyk":
		paymentMethod = models.PaymentHalyk
	default:
		paymentMethod = models.PaymentCash
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Success: false, Error: "Требуется авторизация"})
		return
	}
	items := make([]services.OrderItemInput, len(req.Items))
	for i, it := range req.Items {
		pid, _ := uuid.Parse(it.ProductID)
		items[i] = services.OrderItemInput{ProductID: pid, Quantity: it.Quantity}
	}
	uid := userID
	input := services.CreateOrderInput{
		UserID:             &uid,
		StoreID:            storeID,
		DistrictID:         districtID,
		DeliveryType:       deliveryType,
		DeliveryTimeSlotID: slotID,
		DeliveryAddress:    req.DeliveryAddress,
		CustomerPhone:      req.CustomerPhone,
		CustomerName:       req.CustomerName,
		PaymentMethod:      paymentMethod,
		Items:              items,
		Notes:              req.Notes,
	}
	order, err := h.orderSvc.CreateOrder(c.Request.Context(), input)
	if err != nil {
		if err == utils.ErrNotFound {
			utils.GinError(c, http.StatusNotFound, err.Error(), utils.CodeNotFound)
			return
		}
		if err == utils.ErrMinOrderAmount {
			utils.GinError(c, http.StatusBadRequest, "Минимальная сумма заказа не достигнута", utils.CodeMinOrderAmount)
			return
		}
		if err == utils.ErrInsufficientStock {
			utils.GinError(c, http.StatusBadRequest, "Недостаточно товара на складе. Уменьшите количество в корзине и обновите страницу.", utils.CodeInsufficientStock)
			return
		}
		if err == utils.ErrInvalidInput {
			utils.GinError(c, http.StatusBadRequest, "Товар временно недоступен или снят с продажи", utils.CodeProductUnavailable)
			return
		}
		utils.GinError(c, http.StatusInternalServerError, err.Error(), utils.CodeInternal)
		return
	}
	c.JSON(http.StatusCreated, utils.SuccessResponse{Success: true, Data: order})
}

func (h *OrderHandler) ListMyOrders(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.GinError(c, http.StatusUnauthorized, "Требуется авторизация", utils.CodeUnauthorized)
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	list, err := h.orderSvc.ListOrdersForCustomer(c.Request.Context(), userID, limit)
	if err != nil {
		utils.GinError(c, http.StatusInternalServerError, err.Error(), utils.CodeInternal)
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: list})
}

func (h *OrderHandler) GetMyOrder(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.GinError(c, http.StatusUnauthorized, "Требуется авторизация", utils.CodeUnauthorized)
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.GinError(c, http.StatusBadRequest, "Неверный ID заказа", utils.CodeValidation)
		return
	}
	order, err := h.orderSvc.GetOrderForCustomer(c.Request.Context(), userID, id)
	if err != nil {
		if err == utils.ErrNotFound {
			utils.GinError(c, http.StatusNotFound, "Заказ не найден", utils.CodeNotFound)
			return
		}
		utils.GinError(c, http.StatusInternalServerError, err.Error(), utils.CodeInternal)
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: order})
}
