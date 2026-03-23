package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/repositories"
	"github.com/veggieshop/backend/internal/services"
	"github.com/veggieshop/backend/internal/utils"
	"gorm.io/gorm"
)

type AdminDeliveryHandler struct {
	storeSvc     *services.StoreService
	districtRepo repositories.DistrictRepository
	slotRepo     repositories.DeliveryTimeSlotRepository
	orderRepo    repositories.OrderRepository
	streetRepo   repositories.DistrictStreetRepository
}

func NewAdminDeliveryHandler(
	storeSvc *services.StoreService,
	districtRepo repositories.DistrictRepository,
	slotRepo repositories.DeliveryTimeSlotRepository,
	orderRepo repositories.OrderRepository,
	streetRepo repositories.DistrictStreetRepository,
) *AdminDeliveryHandler {
	return &AdminDeliveryHandler{
		storeSvc:     storeSvc,
		districtRepo: districtRepo,
		slotRepo:     slotRepo,
		orderRepo:    orderRepo,
		streetRepo:   streetRepo,
	}
}

func normalizeTimeHHMMSS(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if len(s) == 5 && s[2] == ':' {
		return s + ":00"
	}
	return s
}

func (h *AdminDeliveryHandler) parseStoreID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный ID магазина"})
		return uuid.Nil, false
	}
	return id, true
}

func (h *AdminDeliveryHandler) ensureStore(c *gin.Context, storeID uuid.UUID) bool {
	_, err := h.storeSvc.GetStoreByID(c.Request.Context(), storeID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Success: false, Error: "Магазин не найден"})
			return false
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return false
	}
	return true
}

// ListDistricts GET /admin/stores/:storeId/districts
func (h *AdminDeliveryHandler) ListDistricts(c *gin.Context) {
	storeID, ok := h.parseStoreID(c)
	if !ok {
		return
	}
	if !h.ensureStore(c, storeID) {
		return
	}
	list, err := h.districtRepo.GetByStoreID(c.Request.Context(), storeID, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: list})
}

type createDistrictRequest struct {
	Name               string   `json:"name" binding:"required"`
	DistanceKm         float64  `json:"distance_km" binding:"required"`
	DeliveryFeeRegular int      `json:"delivery_fee_regular" binding:"required"`
	DeliveryFeeExpress int      `json:"delivery_fee_express" binding:"required"`
	IsActive           *bool    `json:"is_active"`
	Streets            []string `json:"streets"`
}

// CreateDistrict POST /admin/stores/:storeId/districts
func (h *AdminDeliveryHandler) CreateDistrict(c *gin.Context) {
	storeID, ok := h.parseStoreID(c)
	if !ok {
		return
	}
	if !h.ensureStore(c, storeID) {
		return
	}
	var req createDistrictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	d := &models.District{
		StoreID:            storeID,
		Name:               strings.TrimSpace(req.Name),
		DistanceKm:         req.DistanceKm,
		DeliveryFeeRegular: req.DeliveryFeeRegular,
		DeliveryFeeExpress: req.DeliveryFeeExpress,
		IsActive:           true,
	}
	if req.IsActive != nil {
		d.IsActive = *req.IsActive
	}
	if err := h.districtRepo.Create(c.Request.Context(), d); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	for _, raw := range req.Streets {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		st := &models.DistrictStreet{DistrictID: d.ID, StreetName: name}
		if err := h.streetRepo.Create(c.Request.Context(), st); err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
			return
		}
	}
	created, err := h.districtRepo.GetByID(c.Request.Context(), d.ID)
	if err != nil {
		c.JSON(http.StatusCreated, utils.SuccessResponse{Success: true, Data: d})
		return
	}
	c.JSON(http.StatusCreated, utils.SuccessResponse{Success: true, Data: created})
}

type patchDistrictRequest struct {
	Name               *string   `json:"name"`
	DistanceKm         *float64  `json:"distance_km"`
	DeliveryFeeRegular *int      `json:"delivery_fee_regular"`
	DeliveryFeeExpress *int      `json:"delivery_fee_express"`
	IsActive           *bool     `json:"is_active"`
	Streets            *[]string `json:"streets"`
}

// PatchDistrict PATCH /admin/districts/:id
func (h *AdminDeliveryHandler) PatchDistrict(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный ID"})
		return
	}
	var req patchDistrictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	d, err := h.districtRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Success: false, Error: "Район не найден"})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	if req.Name != nil {
		t := strings.TrimSpace(*req.Name)
		if t == "" {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Название не может быть пустым"})
			return
		}
		d.Name = t
	}
	if req.DistanceKm != nil {
		d.DistanceKm = *req.DistanceKm
	}
	if req.DeliveryFeeRegular != nil {
		d.DeliveryFeeRegular = *req.DeliveryFeeRegular
	}
	if req.DeliveryFeeExpress != nil {
		d.DeliveryFeeExpress = *req.DeliveryFeeExpress
	}
	if req.IsActive != nil {
		d.IsActive = *req.IsActive
	}
	if err := h.districtRepo.Update(c.Request.Context(), d); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	if req.Streets != nil {
		if err := h.streetRepo.DeleteByDistrictID(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
			return
		}
		for _, raw := range *req.Streets {
			name := strings.TrimSpace(raw)
			if name == "" {
				continue
			}
			st := &models.DistrictStreet{DistrictID: id, StreetName: name}
			if err := h.streetRepo.Create(c.Request.Context(), st); err != nil {
				c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
				return
			}
		}
	}
	updated, err := h.districtRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusOK, utils.SuccessResponse{Success: true})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: updated})
}

// DeleteDistrict DELETE /admin/districts/:id
func (h *AdminDeliveryHandler) DeleteDistrict(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный ID"})
		return
	}
	d, err := h.districtRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Success: false, Error: "Район не найден"})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	n, err := h.orderRepo.CountByDistrictID(c.Request.Context(), id, models.OrderCancelled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	if n > 0 {
		d.IsActive = false
		if err := h.districtRepo.Update(c.Request.Context(), d); err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Message: "Район отключён: есть заказы, удаление невозможно"})
		return
	}
	if err := h.districtRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Message: "Район удалён"})
}

// ListTimeSlots GET /admin/stores/:storeId/time-slots
func (h *AdminDeliveryHandler) ListTimeSlots(c *gin.Context) {
	storeID, ok := h.parseStoreID(c)
	if !ok {
		return
	}
	if !h.ensureStore(c, storeID) {
		return
	}
	list, err := h.slotRepo.ListAllByStoreID(c.Request.Context(), storeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: list})
}

type createTimeSlotRequest struct {
	DayOfWeek int    `json:"day_of_week" binding:"required,min=0,max=6"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
	MaxOrders int    `json:"max_orders"`
	IsActive  *bool  `json:"is_active"`
}

// CreateTimeSlot POST /admin/stores/:storeId/time-slots
func (h *AdminDeliveryHandler) CreateTimeSlot(c *gin.Context) {
	storeID, ok := h.parseStoreID(c)
	if !ok {
		return
	}
	if !h.ensureStore(c, storeID) {
		return
	}
	var req createTimeSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	start := normalizeTimeHHMMSS(req.StartTime)
	end := normalizeTimeHHMMSS(req.EndTime)
	if start == "" || end == "" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Укажите время начала и конца"})
		return
	}
	slot := &models.DeliveryTimeSlot{
		StoreID:   storeID,
		DayOfWeek: req.DayOfWeek,
		StartTime: start,
		EndTime:   end,
		MaxOrders: 10,
		IsActive:  true,
	}
	if req.MaxOrders > 0 {
		slot.MaxOrders = req.MaxOrders
	}
	if req.IsActive != nil {
		slot.IsActive = *req.IsActive
	}
	if err := h.slotRepo.Create(c.Request.Context(), slot); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, utils.SuccessResponse{Success: true, Data: slot})
}

type patchTimeSlotRequest struct {
	DayOfWeek *int    `json:"day_of_week"`
	StartTime *string `json:"start_time"`
	EndTime   *string `json:"end_time"`
	MaxOrders *int    `json:"max_orders"`
	IsActive  *bool   `json:"is_active"`
}

// PatchTimeSlot PATCH /admin/time-slots/:id
func (h *AdminDeliveryHandler) PatchTimeSlot(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный ID"})
		return
	}
	var req patchTimeSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	slot, err := h.slotRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Success: false, Error: "Слот не найден"})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	if req.DayOfWeek != nil {
		if *req.DayOfWeek < 0 || *req.DayOfWeek > 6 {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "day_of_week: 0–6"})
			return
		}
		slot.DayOfWeek = *req.DayOfWeek
	}
	if req.StartTime != nil {
		t := normalizeTimeHHMMSS(*req.StartTime)
		if t == "" {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Некорректное время начала"})
			return
		}
		slot.StartTime = t
	}
	if req.EndTime != nil {
		t := normalizeTimeHHMMSS(*req.EndTime)
		if t == "" {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Некорректное время окончания"})
			return
		}
		slot.EndTime = t
	}
	if req.MaxOrders != nil {
		slot.MaxOrders = *req.MaxOrders
	}
	if req.IsActive != nil {
		slot.IsActive = *req.IsActive
	}
	if err := h.slotRepo.Update(c.Request.Context(), slot); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: slot})
}

// DeleteTimeSlot DELETE /admin/time-slots/:id
func (h *AdminDeliveryHandler) DeleteTimeSlot(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный ID"})
		return
	}
	slot, err := h.slotRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Success: false, Error: "Слот не найден"})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	n, err := h.orderRepo.CountByTimeSlot(c.Request.Context(), id, models.OrderCancelled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	if n > 0 {
		slot.IsActive = false
		if err := h.slotRepo.Update(c.Request.Context(), slot); err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Message: "Слот отключён: есть заказы, удаление невозможно"})
		return
	}
	if err := h.slotRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Message: "Слот удалён"})
}
