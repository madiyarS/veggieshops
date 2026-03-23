package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/services"
	"github.com/veggieshop/backend/internal/utils"
)

type AdminStockHandler struct {
	stockSvc *services.StockService
}

func NewAdminStockHandler(stockSvc *services.StockService) *AdminStockHandler {
	return &AdminStockHandler{stockSvc: stockSvc}
}

func (h *AdminStockHandler) ListZones(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный store_id"})
		return
	}
	list, err := h.stockSvc.ListZones(c.Request.Context(), storeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: list})
}

func (h *AdminStockHandler) ListExpiring(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный store_id"})
		return
	}
	days := 2
	if d := c.Query("days"); d != "" {
		if n, e := strconv.Atoi(d); e == nil && n > 0 {
			days = n
		}
	}
	list, err := h.stockSvc.ListExpiring(c.Request.Context(), storeID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: list})
}

func (h *AdminStockHandler) ListReorderAlerts(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный store_id"})
		return
	}
	list, err := h.stockSvc.ListReorderAlerts(c.Request.Context(), storeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: list})
}

type writeOffRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
	Reason    string `json:"reason"`
	Type      string `json:"type" binding:"required"` // damage | shrink | resort
}

func (h *AdminStockHandler) WriteOff(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный store_id"})
		return
	}
	var req writeOffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	pid, err := uuid.Parse(req.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный product_id"})
		return
	}
	var mt models.StockMovementType
	switch req.Type {
	case "damage":
		mt = models.MovementWriteOffDamage
	case "shrink":
		mt = models.MovementWriteOffShrink
	case "resort":
		mt = models.MovementWriteOffResort
	default:
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "type: damage | shrink | resort"})
		return
	}
	if err := h.stockSvc.WriteOff(c.Request.Context(), storeID, pid, req.Quantity, mt, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	if err := h.stockSvc.MirrorProductStock(c.Request.Context(), storeID, pid); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true})
}

type receiptLineReq struct {
	ProductID string  `json:"product_id" binding:"required"`
	ZoneID    string  `json:"zone_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	ExpiresAt *string `json:"expires_at"`
}

type receiptRequest struct {
	SupplierID *string         `json:"supplier_id"`
	Note       string          `json:"note"`
	Lines      []receiptLineReq `json:"lines" binding:"required,min=1"`
}

func (h *AdminStockHandler) ApplyReceipt(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный store_id"})
		return
	}
	var req receiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	var sup *uuid.UUID
	if req.SupplierID != nil && *req.SupplierID != "" {
		sid, e := uuid.Parse(*req.SupplierID)
		if e != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный supplier_id"})
			return
		}
		sup = &sid
	}
	lines := make([]services.ReceiptLineInput, 0, len(req.Lines))
	for _, ln := range req.Lines {
		pid, e1 := uuid.Parse(ln.ProductID)
		zid, e2 := uuid.Parse(ln.ZoneID)
		if e1 != nil || e2 != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный product_id или zone_id"})
			return
		}
		var exp *time.Time
		if ln.ExpiresAt != nil && *ln.ExpiresAt != "" {
			t, e := time.Parse("2006-01-02", *ln.ExpiresAt)
			if e != nil {
				c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "expires_at формат YYYY-MM-DD"})
				return
			}
			exp = &t
		}
		lines = append(lines, services.ReceiptLineInput{
			ProductID: pid, ZoneID: zid, Quantity: ln.Quantity, ExpiresAt: exp,
		})
	}
	rec, err := h.stockSvc.ApplyReceipt(c.Request.Context(), storeID, sup, req.Note, lines)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	for _, ln := range lines {
		_ = h.stockSvc.MirrorProductStock(c.Request.Context(), storeID, ln.ProductID)
	}
	c.JSON(http.StatusCreated, utils.SuccessResponse{Success: true, Data: rec})
}

type auditCompleteRequest struct {
	Note  string `json:"note"`
	Lines []struct {
		ProductID  string  `json:"product_id" binding:"required"`
		ZoneID     *string `json:"zone_id"`
		CountedQty int     `json:"counted_qty" binding:"required"`
	} `json:"lines" binding:"required,min=1"`
}

func (h *AdminStockHandler) CompleteAudit(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный store_id"})
		return
	}
	var req auditCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	inputs := make([]services.AuditLineInput, 0, len(req.Lines))
	for _, ln := range req.Lines {
		pid, e := uuid.Parse(ln.ProductID)
		if e != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный product_id"})
			return
		}
		var zid *uuid.UUID
		if ln.ZoneID != nil && *ln.ZoneID != "" {
			z, e := uuid.Parse(*ln.ZoneID)
			if e != nil {
				c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный zone_id"})
				return
			}
			zid = &z
		}
		inputs = append(inputs, services.AuditLineInput{ProductID: pid, ZoneID: zid, CountedQty: ln.CountedQty})
	}
	sess, err := h.stockSvc.CompleteInventoryAudit(c.Request.Context(), storeID, req.Note, inputs)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, utils.SuccessResponse{Success: true, Data: sess})
}

type simpleReceiveRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
	Note      string `json:"note"`
}

// SimpleReceive приход в зону «Зал» одной кнопкой (UX как zakazik).
func (h *AdminStockHandler) SimpleReceive(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный store_id"})
		return
	}
	var req simpleReceiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	pid, err := uuid.Parse(req.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный product_id"})
		return
	}
	rec, err := h.stockSvc.SimpleReceiveToSalesFloor(c.Request.Context(), storeID, pid, req.Quantity, req.Note)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	_ = h.stockSvc.MirrorProductStock(c.Request.Context(), storeID, pid)
	c.JSON(http.StatusCreated, utils.SuccessResponse{Success: true, Data: rec})
}

// ListMovesJournal журнал движений с названиями товаров (последние N).
func (h *AdminStockHandler) ListMovesJournal(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный store_id"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	list, err := h.stockSvc.ListStockMovementsRecentWithNames(c.Request.Context(), storeID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: list})
}

type setInventoryActualRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Actual    int    `json:"actual" binding:"required,min=0"`
	Note      string `json:"note"`
}

// SetInventoryActual фактический остаток (как инвентаризация в zakazik).
func (h *AdminStockHandler) SetInventoryActual(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный store_id"})
		return
	}
	var req setInventoryActualRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	pid, err := uuid.Parse(req.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный product_id"})
		return
	}
	if err := h.stockSvc.SetInventoryActual(c.Request.Context(), storeID, pid, req.Actual, req.Note); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true})
}

func (h *AdminStockHandler) ListSuppliers(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный store_id"})
		return
	}
	list, err := h.stockSvc.ListSuppliers(c.Request.Context(), storeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: list})
}

type createSupplierRequest struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone"`
}

func (h *AdminStockHandler) CreateSupplier(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный store_id"})
		return
	}
	var req createSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	su, err := h.stockSvc.CreateSupplier(c.Request.Context(), storeID, req.Name, req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, utils.SuccessResponse{Success: true, Data: su})
}

func (h *AdminStockHandler) ListBatches(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный store_id"})
		return
	}
	pid, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный product_id"})
		return
	}
	list, err := h.stockSvc.ListBatchesForProduct(c.Request.Context(), storeID, pid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: list})
}
