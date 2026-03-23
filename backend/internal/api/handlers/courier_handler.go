package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/api/middleware"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/services"
	"github.com/veggieshop/backend/internal/utils"
)

// courierOrderView — без кода выдачи (курьер получает его только от клиента).
type courierOrderView struct {
	ID               uuid.UUID           `json:"id"`
	OrderNumber      string              `json:"order_number"`
	Status           models.OrderStatus  `json:"status"`
	TotalAmount      int                 `json:"total_amount"`
	DeliveryAddress  string              `json:"delivery_address"`
	CustomerName     string              `json:"customer_name"`
	CustomerPhone    string              `json:"customer_phone"`
	CourierID        *uuid.UUID          `json:"courier_id"`
	CreatedAt        time.Time           `json:"created_at"`
	Items            []*models.OrderItem `json:"items,omitempty"`
}

func toCourierOrderViews(list []*models.Order) []courierOrderView {
	out := make([]courierOrderView, 0, len(list))
	for _, o := range list {
		out = append(out, courierOrderView{
			ID:              o.ID,
			OrderNumber:     o.OrderNumber,
			Status:          o.Status,
			TotalAmount:     o.TotalAmount,
			DeliveryAddress: o.DeliveryAddress,
			CustomerName:    o.CustomerName,
			CustomerPhone:   o.CustomerPhone,
			CourierID:       o.CourierID,
			CreatedAt:       o.CreatedAt,
			Items:           o.Items,
		})
	}
	return out
}

type CourierHandler struct {
	courierSvc *services.CourierService
}

func NewCourierHandler(courierSvc *services.CourierService) *CourierHandler {
	return &CourierHandler{courierSvc: courierSvc}
}

func (h *CourierHandler) ListOrders(c *gin.Context) {
	uid, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Success: false, Error: "Не авторизован"})
		return
	}
	list, err := h.courierSvc.ListMyOrders(c.Request.Context(), uid)
	if err != nil {
		if err == utils.ErrCourierProfile {
			c.JSON(http.StatusForbidden, utils.ErrorResponse{Success: false, Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: toCourierOrderViews(list)})
}

func (h *CourierHandler) AcceptOrder(c *gin.Context) {
	uid, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Success: false, Error: "Не авторизован"})
		return
	}
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный ID заказа"})
		return
	}
	if err := h.courierSvc.AcceptOrder(c.Request.Context(), uid, orderID); err != nil {
		switch err {
		case utils.ErrNotFound:
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Success: false, Error: "Заказ не найден"})
		case utils.ErrForbidden:
			c.JSON(http.StatusForbidden, utils.ErrorResponse{Success: false, Error: err.Error()})
		case utils.ErrCourierProfile:
			c.JSON(http.StatusForbidden, utils.ErrorResponse{Success: false, Error: err.Error()})
		case utils.ErrInvalidInput:
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Заказ уже завершён или недоступен"})
		default:
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true})
}

type completeDeliveryBody struct {
	Code string `json:"code" binding:"required"`
}

func (h *CourierHandler) CompleteDelivery(c *gin.Context) {
	uid, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Success: false, Error: "Не авторизован"})
		return
	}
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный ID заказа"})
		return
	}
	var body completeDeliveryBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Укажите код"})
		return
	}
	code := strings.TrimSpace(body.Code)
	if err := h.courierSvc.CompleteDelivery(c.Request.Context(), uid, orderID, code); err != nil {
		switch err {
		case utils.ErrNotFound:
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Success: false, Error: "Заказ не найден"})
		case utils.ErrForbidden:
			c.JSON(http.StatusForbidden, utils.ErrorResponse{Success: false, Error: err.Error()})
		case utils.ErrWrongDeliveryCode:
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: err.Error()})
		case utils.ErrOrderNotInDelivery:
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: err.Error()})
		case utils.ErrCourierProfile:
			c.JSON(http.StatusForbidden, utils.ErrorResponse{Success: false, Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true})
}
