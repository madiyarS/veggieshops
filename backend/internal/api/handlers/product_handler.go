package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/services"
	"github.com/veggieshop/backend/internal/utils"
)

type ProductHandler struct {
	productSvc *services.ProductService
}

func NewProductHandler(productSvc *services.ProductService) *ProductHandler {
	return &ProductHandler{productSvc: productSvc}
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	if storeIDStr == "" {
		utils.GinError(c, http.StatusBadRequest, "store_id обязателен", utils.CodeValidation)
		return
	}
	storeID, err := uuid.Parse(storeIDStr)
	if err != nil {
		utils.GinError(c, http.StatusBadRequest, "Неверный store_id", utils.CodeValidation)
		return
	}
	var categoryID *uuid.UUID
	if catStr := c.Query("category_id"); catStr != "" {
		catID, err := uuid.Parse(catStr)
		if err == nil {
			categoryID = &catID
		}
	}
	q := strings.TrimSpace(c.Query("q"))
	inStock := c.Query("in_stock_only") == "1" || strings.EqualFold(c.Query("in_stock_only"), "true")
	sortKey := strings.TrimSpace(strings.ToLower(c.Query("sort")))
	switch sortKey {
	case "name", "price_asc", "price_desc", "expiry_asc":
	default:
		sortKey = ""
	}
	opts := services.ProductListOpts{InStockOnly: inStock, Sort: sortKey}
	products, err := h.productSvc.GetProductsByStore(c.Request.Context(), storeID, categoryID, q, opts)
	if err != nil {
		utils.GinError(c, http.StatusInternalServerError, "Не удалось загрузить каталог", utils.CodeInternal)
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: products})
}

// GET /api/v1/products/availability?store_id=&product_ids=uuid,uuid (до 50 id)
func (h *ProductHandler) GetProductAvailability(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	if storeIDStr == "" {
		utils.GinError(c, http.StatusBadRequest, "Укажите store_id", utils.CodeValidation)
		return
	}
	storeID, err := uuid.Parse(storeIDStr)
	if err != nil {
		utils.GinError(c, http.StatusBadRequest, "Неверный store_id", utils.CodeValidation)
		return
	}
	raw := strings.TrimSpace(c.Query("product_ids"))
	if raw == "" {
		utils.GinError(c, http.StatusBadRequest, "Укажите product_ids (через запятую)", utils.CodeValidation)
		return
	}
	parts := strings.Split(raw, ",")
	var ids []uuid.UUID
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, e := uuid.Parse(p)
		if e != nil {
			utils.GinError(c, http.StatusBadRequest, "Неверный product_id в списке", utils.CodeValidation)
			return
		}
		ids = append(ids, id)
		if len(ids) >= 50 {
			break
		}
	}
	if len(ids) == 0 {
		utils.GinError(c, http.StatusBadRequest, "Нет ни одного product_id", utils.CodeValidation)
		return
	}
	m, err := h.productSvc.GetAvailableByProductIDs(c.Request.Context(), storeID, ids)
	if err != nil {
		utils.GinError(c, http.StatusInternalServerError, "Не удалось получить остатки", utils.CodeInternal)
		return
	}
	out := make(map[string]int, len(m))
	for id, q := range m {
		out[id.String()] = q
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": out})
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.GinError(c, http.StatusBadRequest, "Неверный ID", utils.CodeValidation)
		return
	}
	product, err := h.productSvc.GetProductByID(c.Request.Context(), id)
	if err != nil {
		utils.GinError(c, http.StatusNotFound, "Товар не найден", utils.CodeNotFound)
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: product})
}
