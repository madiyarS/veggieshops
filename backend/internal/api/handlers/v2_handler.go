package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/repositories"
	"github.com/veggieshop/backend/internal/services"
	"github.com/veggieshop/backend/internal/utils"
)

type V2Handler struct {
	productSvc *services.ProductService
	orderSvc   *services.OrderService
	stockSvc   *services.StockService
}

func NewV2Handler(productSvc *services.ProductService, orderSvc *services.OrderService, stockSvc *services.StockService) *V2Handler {
	return &V2Handler{productSvc: productSvc, orderSvc: orderSvc, stockSvc: stockSvc}
}

// GET /api/v2/products?store_id=&limit=&after_id=&category_id=
func (h *V2Handler) ListProductsPaged(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	if storeIDStr == "" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "store_id обязателен"})
		return
	}
	storeID, err := uuid.Parse(storeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный store_id"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	var categoryID *uuid.UUID
	if cat := c.Query("category_id"); cat != "" {
		cid, e := uuid.Parse(cat)
		if e != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный category_id"})
			return
		}
		categoryID = &cid
	}
	var afterID *uuid.UUID
	if s := c.Query("after_id"); s != "" {
		id, e := uuid.Parse(s)
		if e != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный after_id"})
			return
		}
		afterID = &id
	}
	q := strings.TrimSpace(c.Query("q"))
	list, err := h.productSvc.GetProductsByStorePaged(c.Request.Context(), storeID, categoryID, q, limit, afterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	meta := gin.H{"has_more": len(list) == limit && limit > 0}
	if len(list) > 0 && len(list) == limit {
		meta["next_after_id"] = list[len(list)-1].ID.String()
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list, "meta": meta})
}

// GET /api/v2/admin/stores/:storeId/orders?limit=&after_created_at=&after_id=&status=
func (h *V2Handler) ListOrdersPaged(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный store_id"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	filters := repositories.OrderFilters{}
	if st := c.Query("status"); st != "" {
		s := models.OrderStatus(st)
		filters.Status = &s
	}
	if df := c.Query("date_from"); df != "" {
		filters.DateFrom = &df
	}
	if dt := c.Query("date_to"); dt != "" {
		filters.DateTo = &dt
	}
	var afterCreatedAt *time.Time
	var afterID *uuid.UUID
	if ac := c.Query("after_created_at"); ac != "" {
		t, e := time.Parse(time.RFC3339Nano, ac)
		if e != nil {
			t2, e2 := time.Parse(time.RFC3339, ac)
			if e2 != nil {
				c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "after_created_at в формате RFC3339"})
				return
			}
			t = t2
		}
		afterCreatedAt = &t
	}
	if aid := c.Query("after_id"); aid != "" {
		id, e := uuid.Parse(aid)
		if e != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный after_id"})
			return
		}
		afterID = &id
	}
	if (afterCreatedAt == nil) != (afterID == nil) {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "передайте оба поля курсора или ни одного"})
		return
	}
	list, err := h.orderSvc.ListOrdersForAdminPaged(c.Request.Context(), storeID, filters, limit, afterCreatedAt, afterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	meta := gin.H{"has_more": len(list) == limit && limit > 0}
	if len(list) > 0 && len(list) == limit {
		last := list[len(list)-1]
		meta["next_after_created_at"] = last.CreatedAt.Format(time.RFC3339Nano)
		meta["next_after_id"] = last.ID.String()
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list, "meta": meta})
}

// GET /api/v2/admin/stores/:storeId/stock/movements?limit=&after_created_at=&after_id=
func (h *V2Handler) ListStockMovementsPaged(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный store_id"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	var afterCreatedAt *time.Time
	var afterID *uuid.UUID
	if ac := c.Query("after_created_at"); ac != "" {
		t, e := time.Parse(time.RFC3339Nano, ac)
		if e != nil {
			t2, e2 := time.Parse(time.RFC3339, ac)
			if e2 != nil {
				c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "after_created_at в формате RFC3339"})
				return
			}
			t = t2
		}
		afterCreatedAt = &t
	}
	if aid := c.Query("after_id"); aid != "" {
		id, e := uuid.Parse(aid)
		if e != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный after_id"})
			return
		}
		afterID = &id
	}
	if (afterCreatedAt == nil) != (afterID == nil) {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "передайте оба поля курсора или ни одного"})
		return
	}
	list, err := h.stockSvc.ListStockMovementsPaged(c.Request.Context(), storeID, limit, afterCreatedAt, afterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	meta := gin.H{"has_more": len(list) == limit && limit > 0}
	if len(list) > 0 && len(list) == limit {
		last := list[len(list)-1]
		meta["next_after_created_at"] = last.CreatedAt.Format(time.RFC3339Nano)
		meta["next_after_id"] = last.ID.String()
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list, "meta": meta})
}
