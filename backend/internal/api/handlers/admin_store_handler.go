package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/services"
	"github.com/veggieshop/backend/internal/utils"
)

type AdminStoreHandler struct {
	storeSvc *services.StoreService
	prodSvc  *services.ProductService
}

func NewAdminStoreHandler(storeSvc *services.StoreService, prodSvc *services.ProductService) *AdminStoreHandler {
	return &AdminStoreHandler{storeSvc: storeSvc, prodSvc: prodSvc}
}

type CreateStoreRequest struct {
	Name             string   `json:"name" binding:"required"`
	Description      string   `json:"description"`
	Address          string   `json:"address" binding:"required"`
	Latitude         float64  `json:"latitude" binding:"required"`
	Longitude        float64  `json:"longitude" binding:"required"`
	Phone            string   `json:"phone"`
	Email            string   `json:"email"`
	DeliveryRadiusKm float64  `json:"delivery_radius_km"`
	MinOrderAmount   int      `json:"min_order_amount"`
	MaxOrderWeightKg *float64 `json:"max_order_weight_kg"`
	// CopyCatalogFromStoreID — скопировать номенклатуру и остатки склада из другого магазина (новые id товаров).
	CopyCatalogFromStoreID *string `json:"copy_catalog_from_store_id"`
}

func (h *AdminStoreHandler) CreateStore(c *gin.Context) {
	var req CreateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	store := &models.Store{
		Name:             req.Name,
		Description:      req.Description,
		Address:          req.Address,
		Latitude:         req.Latitude,
		Longitude:        req.Longitude,
		Phone:            req.Phone,
		Email:            req.Email,
		DeliveryRadiusKm: 3,
		MinOrderAmount:   2500,
		IsActive:         true,
	}
	if req.DeliveryRadiusKm > 0 {
		store.DeliveryRadiusKm = req.DeliveryRadiusKm
	}
	if req.MinOrderAmount > 0 {
		store.MinOrderAmount = req.MinOrderAmount
	}
	store.MaxOrderWeightKg = req.MaxOrderWeightKg

	var copySourceID uuid.UUID
	wantCopy := false
	if req.CopyCatalogFromStoreID != nil && strings.TrimSpace(*req.CopyCatalogFromStoreID) != "" {
		sid, perr := uuid.Parse(strings.TrimSpace(*req.CopyCatalogFromStoreID))
		if perr != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный copy_catalog_from_store_id"})
			return
		}
		copySourceID = sid
		wantCopy = true
	}

	created, err := h.storeSvc.CreateStore(c.Request.Context(), store)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	msg := ""
	if wantCopy {
		n, cerr := h.prodSvc.CloneCatalogFromStore(c.Request.Context(), created.ID, copySourceID)
		if cerr != nil {
			if errors.Is(cerr, utils.ErrNotFound) {
				msg = "Магазин создан. Исходный магазин для копирования не найден."
			} else {
				msg = "Магазин создан. Ошибка копирования каталога: " + cerr.Error()
			}
		} else if n > 0 {
			msg = fmt.Sprintf("Скопировано товаров: %d", n)
		}
	}
	c.JSON(http.StatusCreated, utils.SuccessResponse{Success: true, Data: created, Message: msg})
}

func (h *AdminStoreHandler) GetStores(c *gin.Context) {
	stores, err := h.storeSvc.GetAllStores(c.Request.Context(), false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: stores})
}

func (h *AdminStoreHandler) UpdateStore(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный ID"})
		return
	}
	var patch services.StorePatch
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	if err := h.storeSvc.PatchStore(c.Request.Context(), id, &patch); err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Success: false, Error: "Магазин не найден"})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	updated, err := h.storeSvc.GetStoreByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusOK, utils.SuccessResponse{Success: true})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: updated})
}
