package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/services"
	"github.com/veggieshop/backend/internal/utils"
)

type AdminCategoryHandler struct {
	catSvc  *services.CategoryService
	prodSvc *services.ProductService
}

func NewAdminCategoryHandler(catSvc *services.CategoryService, prodSvc *services.ProductService) *AdminCategoryHandler {
	return &AdminCategoryHandler{catSvc: catSvc, prodSvc: prodSvc}
}

func (h *AdminCategoryHandler) List(c *gin.Context) {
	list, err := h.catSvc.GetAll(c.Request.Context(), false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: list})
}

type createCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
	Order       int    `json:"order"`
	IsActive    *bool  `json:"is_active"`
}

func (h *AdminCategoryHandler) Create(c *gin.Context) {
	var req createCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	cat := &models.Category{
		Name:        req.Name,
		Description: req.Description,
		IconURL:     req.IconURL,
		Order:       req.Order,
		IsActive:    true,
	}
	if req.IsActive != nil {
		cat.IsActive = *req.IsActive
	}
	created, err := h.catSvc.Create(c.Request.Context(), cat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, utils.SuccessResponse{Success: true, Data: created})
}

type patchCategoryRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IconURL     *string `json:"icon_url"`
	Order       *int    `json:"order"`
	IsActive    *bool   `json:"is_active"`
}

func (h *AdminCategoryHandler) Patch(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный ID"})
		return
	}
	var req patchCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	patch := &services.CategoryPatch{
		Name:        req.Name,
		Description: req.Description,
		IconURL:     req.IconURL,
		Order:       req.Order,
		IsActive:    req.IsActive,
	}
	if err := h.catSvc.Update(c.Request.Context(), id, patch); err != nil {
		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Success: false, Error: "Категория не найдена"})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true})
}

func (h *AdminCategoryHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверный ID"})
		return
	}
	n, err := h.prodSvc.CountByCategoryID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	if n > 0 {
		c.JSON(http.StatusConflict, utils.ErrorResponse{
			Success: false,
			Error:   "Нельзя удалить категорию с товарами. Перенесите или скройте товары.",
		})
		return
	}
	if err := h.catSvc.Delete(c.Request.Context(), id); err != nil {
		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{Success: false, Error: "Категория не найдена"})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true})
}
