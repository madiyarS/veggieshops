package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/services"
	"github.com/veggieshop/backend/internal/utils"
)

type AdminProductHandler struct {
	prodSvc *services.ProductService
}

func NewAdminProductHandler(prodSvc *services.ProductService) *AdminProductHandler {
	return &AdminProductHandler{prodSvc: prodSvc}
}

func (h *AdminProductHandler) ListByStore(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный store_id"})
		return
	}
	var categoryID *uuid.UUID
	if catStr := c.Query("category_id"); catStr != "" {
		cid, err := uuid.Parse(catStr)
		if err == nil {
			categoryID = &cid
		}
	}
	list, err := h.prodSvc.ListProductsForAdmin(c.Request.Context(), storeID, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: list})
}

type createProductRequest struct {
	CategoryID             string `json:"category_id" binding:"required"`
	Name                   string `json:"name" binding:"required"`
	Description            string `json:"description"`
	Price                  int    `json:"price" binding:"required"`
	WeightGram             int    `json:"weight_gram"`
	Unit                   string `json:"unit"`
	StockQuantity          int    `json:"stock_quantity"`
	ImageURL               string `json:"image_url"`
	Origin                 string `json:"origin"`
	ShelfLifeDays          *int   `json:"shelf_life_days"`
	IsAvailable            *bool  `json:"is_available"`
	IsActive               *bool  `json:"is_active"`
	InventoryUnit          string `json:"inventory_unit"`
	PackageGrams           *int   `json:"package_grams"`
	IsSeasonal             *bool  `json:"is_seasonal"`
	TemporarilyUnavailable *bool  `json:"temporarily_unavailable"`
	SubstituteProductID    string `json:"substitute_product_id"`
	ReorderMinQty          *int   `json:"reorder_min_qty"`
	CartStepGrams          *int   `json:"cart_step_grams"`
}

func (h *AdminProductHandler) Create(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный store_id"})
		return
	}
	var req createProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	if req.Price <= 0 {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Цена должна быть больше 0"})
		return
	}
	catID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный category_id"})
		return
	}
	p := &models.Product{
		CategoryID:      catID,
		Name:            req.Name,
		Description:     req.Description,
		Price:           req.Price,
		WeightGram:      req.WeightGram,
		Unit:            req.Unit,
		StockQuantity:   req.StockQuantity,
		ImageURL:        req.ImageURL,
		Origin:          req.Origin,
		ShelfLifeDays:   req.ShelfLifeDays,
		IsAvailable:     true,
		IsActive:        true,
	}
	if p.Unit == "" {
		p.Unit = "шт"
	}
	switch req.InventoryUnit {
	case string(models.InventoryUnitWeightGram):
		p.InventoryUnit = models.InventoryUnitWeightGram
	default:
		p.InventoryUnit = models.InventoryUnitPiece
	}
	if req.PackageGrams != nil {
		p.PackageGrams = req.PackageGrams
	}
	if req.IsSeasonal != nil {
		p.IsSeasonal = *req.IsSeasonal
	}
	if req.TemporarilyUnavailable != nil {
		p.TemporarilyUnavailable = *req.TemporarilyUnavailable
	}
	if req.SubstituteProductID != "" {
		sid, err := uuid.Parse(req.SubstituteProductID)
		if err == nil {
			p.SubstituteProductID = &sid
		}
	}
	if req.ReorderMinQty != nil {
		p.ReorderMinQty = *req.ReorderMinQty
	}
	if req.CartStepGrams != nil {
		p.CartStepGrams = *req.CartStepGrams
	}
	if p.InventoryUnit == models.InventoryUnitWeightGram && p.CartStepGrams <= 0 {
		p.CartStepGrams = 250
	}
	if req.IsAvailable != nil {
		p.IsAvailable = *req.IsAvailable
	}
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}
	created, err := h.prodSvc.CreateProduct(c.Request.Context(), storeID, p)
	if err != nil {
		if err == utils.ErrNotFound {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Магазин или категория не найдены"})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, utils.SuccessResponse{Success: true, Data: created})
}

type patchProductRequest struct {
	CategoryID               *string `json:"category_id"`
	Name                     *string `json:"name"`
	Description              *string `json:"description"`
	Price                    *int    `json:"price"`
	WeightGram               *int    `json:"weight_gram"`
	Unit                     *string `json:"unit"`
	StockQuantity            *int    `json:"stock_quantity"`
	ImageURL                 *string `json:"image_url"`
	Origin                   *string `json:"origin"`
	ShelfLifeDays            *int    `json:"shelf_life_days"`
	ClearShelfLife           bool    `json:"clear_shelf_life"`
	IsAvailable              *bool   `json:"is_available"`
	IsActive                 *bool   `json:"is_active"`
	InventoryUnit            *string `json:"inventory_unit"`
	PackageGrams             *int    `json:"package_grams"`
	ClearPackageGrams        bool    `json:"clear_package_grams"`
	IsSeasonal               *bool   `json:"is_seasonal"`
	TemporarilyUnavailable   *bool   `json:"temporarily_unavailable"`
	SubstituteProductID      *string `json:"substitute_product_id"`
	ClearSubstitute          bool    `json:"clear_substitute"`
	ReorderMinQty            *int    `json:"reorder_min_qty"`
	CartStepGrams            *int    `json:"cart_step_grams"`
}

func (h *AdminProductHandler) Patch(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный ID"})
		return
	}
	var req patchProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	patch := &services.ProductPatch{
		Name:                   req.Name,
		Description:            req.Description,
		Price:                  req.Price,
		WeightGram:             req.WeightGram,
		Unit:                   req.Unit,
		StockQuantity:          req.StockQuantity,
		ImageURL:               req.ImageURL,
		Origin:                 req.Origin,
		ShelfLifeDays:          req.ShelfLifeDays,
		ClearShelfLife:         req.ClearShelfLife,
		IsAvailable:            req.IsAvailable,
		IsActive:               req.IsActive,
		ClearPackageGrams:      req.ClearPackageGrams,
		ClearSubstitute:        req.ClearSubstitute,
		IsSeasonal:             req.IsSeasonal,
		TemporarilyUnavailable: req.TemporarilyUnavailable,
		ReorderMinQty:          req.ReorderMinQty,
		CartStepGrams:          req.CartStepGrams,
	}
	if req.InventoryUnit != nil {
		switch *req.InventoryUnit {
		case string(models.InventoryUnitWeightGram):
			u := models.InventoryUnitWeightGram
			patch.InventoryUnit = &u
		case string(models.InventoryUnitPiece):
			u := models.InventoryUnitPiece
			patch.InventoryUnit = &u
		}
	}
	if req.PackageGrams != nil {
		patch.PackageGrams = req.PackageGrams
	}
	if req.SubstituteProductID != nil && *req.SubstituteProductID != "" {
		sid, err := uuid.Parse(*req.SubstituteProductID)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный substitute_product_id"})
			return
		}
		patch.SubstituteProductID = &sid
	}
	if req.CategoryID != nil {
		cid, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный category_id"})
			return
		}
		patch.CategoryID = &cid
	}
	if err := h.prodSvc.PatchProduct(c.Request.Context(), id, patch); err != nil {
		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Success: false, Error: "Товар не найден"})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true})
}

func (h *AdminProductHandler) Deactivate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный ID"})
		return
	}
	if err := h.prodSvc.DeactivateProduct(c.Request.Context(), id); err != nil {
		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Success: false, Error: "Товар не найден"})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true})
}
