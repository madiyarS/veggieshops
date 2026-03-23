package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/services"
	"github.com/veggieshop/backend/internal/utils"
)

type AdminOrderHandler struct {
	orderSvc *services.OrderService
}

func NewAdminOrderHandler(orderSvc *services.OrderService) *AdminOrderHandler {
	return &AdminOrderHandler{orderSvc: orderSvc}
}

type patchOrderRequest struct {
	Action string `json:"action" binding:"required"` // cancel_pending | commit_stock | cancel_delivery
}

// PatchOrder отмена заказа в сборке, отмена доставки (вернуть в сборку), подтверждение списания без курьера.
func (h *AdminOrderHandler) PatchOrder(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный id заказа"})
		return
	}
	var req patchOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	switch req.Action {
	case "cancel_pending":
		if err := h.orderSvc.CancelPendingOrder(c.Request.Context(), id); err != nil {
			if err == utils.ErrNotFound {
				c.JSON(http.StatusNotFound, utils.ErrorResponse{Success: false, Error: "Заказ не найден"})
				return
			}
			if err == utils.ErrInvalidInput {
				c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Отмена возможна только для заказа в ожидании или в сборке"})
				return
			}
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: err.Error()})
			return
		}
	case "commit_stock":
		if err := h.orderSvc.ConfirmOrderStock(c.Request.Context(), id); err != nil {
			if err == utils.ErrNotFound {
				c.JSON(http.StatusNotFound, utils.ErrorResponse{Success: false, Error: "Заказ не найден"})
				return
			}
			if err == utils.ErrInvalidInput {
				c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Нельзя подтвердить списание для этого статуса"})
				return
			}
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: err.Error()})
			return
		}
	case "cancel_delivery":
		if err := h.orderSvc.AdminReturnFromDelivery(c.Request.Context(), id); err != nil {
			if err == utils.ErrNotFound {
				c.JSON(http.StatusNotFound, utils.ErrorResponse{Success: false, Error: "Заказ не найден"})
				return
			}
			if err == utils.ErrInvalidInput {
				c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Отмена доставки только для заказа «в доставке» до ввода кода клиентом"})
				return
			}
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: err.Error()})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неизвестное action"})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true})
}
