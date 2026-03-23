package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/veggieshop/backend/internal/services"
	"github.com/veggieshop/backend/internal/utils"
)

type CategoryHandler struct {
	categorySvc *services.CategoryService
}

func NewCategoryHandler(categorySvc *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{categorySvc: categorySvc}
}

func (h *CategoryHandler) GetCategories(c *gin.Context) {
	categories, err := h.categorySvc.GetAll(c.Request.Context(), true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: categories})
}
