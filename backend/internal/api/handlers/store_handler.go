package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/services"
	"github.com/veggieshop/backend/internal/utils"
)

type StoreHandler struct {
	storeSvc    *services.StoreService
	districtSvc *services.DeliveryService
}

func NewStoreHandler(storeSvc *services.StoreService, districtSvc *services.DeliveryService) *StoreHandler {
	return &StoreHandler{storeSvc: storeSvc, districtSvc: districtSvc}
}

func (h *StoreHandler) GetStores(c *gin.Context) {
	stores, err := h.storeSvc.GetAllStores(c.Request.Context(), true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: stores})
}

func (h *StoreHandler) GetStoreByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный ID"})
		return
	}
	store, err := h.storeSvc.GetStoreByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{Success: false, Error: "Магазин не найден"})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: store})
}

func (h *StoreHandler) GetDistricts(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный ID магазина"})
		return
	}
	districts, err := h.districtSvc.GetDistrictsByStore(c.Request.Context(), storeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: districts})
}
