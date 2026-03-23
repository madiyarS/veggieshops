package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/services"
	"github.com/veggieshop/backend/internal/utils"
)

type AdminAnalyticsHandler struct {
	analyticsSvc *services.AnalyticsService
}

func NewAdminAnalyticsHandler(analyticsSvc *services.AnalyticsService) *AdminAnalyticsHandler {
	return &AdminAnalyticsHandler{analyticsSvc: analyticsSvc}
}

// GetRevenueReport GET /admin/analytics/revenue?store_id=&from=YYYY-MM-DD&to=YYYY-MM-DD
func (h *AdminAnalyticsHandler) GetRevenueReport(c *gin.Context) {
	var storeID *uuid.UUID
	if s := c.Query("store_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный store_id"})
			return
		}
		storeID = &id
	}

	var fromPtr, toPtr *string
	if f := c.Query("from"); f != "" {
		fromPtr = &f
	}
	if t := c.Query("to"); t != "" {
		toPtr = &t
	}

	report, err := h.analyticsSvc.AdminRevenueReport(c.Request.Context(), storeID, fromPtr, toPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: report})
}
