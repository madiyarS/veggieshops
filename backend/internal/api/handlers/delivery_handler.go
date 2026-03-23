package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/services"
	"github.com/veggieshop/backend/internal/utils"
)

type DeliveryHandler struct {
	deliverySvc *services.DeliveryService
}

func NewDeliveryHandler(deliverySvc *services.DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{deliverySvc: deliverySvc}
}

type CheckDeliveryRequest struct {
	StoreID   string  `json:"store_id" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Address   string  `json:"address"`
}

func (h *DeliveryHandler) CheckDelivery(c *gin.Context) {
	var req CheckDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	storeID, err := uuid.Parse(req.StoreID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный ID магазина"})
		return
	}
	result, err := h.deliverySvc.CheckDeliveryAvailability(c.Request.Context(), storeID, req.Latitude, req.Longitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

type TimeSlotsRequest struct {
	Date string `form:"date" binding:"required"`
}

func (h *DeliveryHandler) GetTimeSlots(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный ID магазина"})
		return
	}
	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Параметр date обязателен"})
		return
	}
	slots, err := h.deliverySvc.GetAvailableTimeSlots(c.Request.Context(), storeID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: slots})
}
