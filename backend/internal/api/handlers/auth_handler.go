package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/veggieshop/backend/internal/services"
	"github.com/veggieshop/backend/internal/utils"
)

type AuthHandler struct {
	authSvc *services.AuthService
}

func NewAuthHandler(authSvc *services.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type RegisterRequest struct {
	Phone     string `json:"phone" binding:"required"`
	Password  string `json:"password" binding:"required"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	_, err := h.authSvc.RegisterUser(c.Request.Context(), req.Phone, req.Password, req.FirstName, req.LastName)
	if err != nil {
		if err == utils.ErrInvalidInput {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: "Ошибка регистрации"})
		return
	}
	tokens, err := h.authSvc.LoginUser(c.Request.Context(), req.Phone, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Success: false, Error: "Регистрация прошла, но вход не удался"})
		return
	}
	c.JSON(http.StatusCreated, utils.SuccessResponse{Success: true, Data: tokens})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	tokens, err := h.authSvc.LoginUser(c.Request.Context(), req.Phone, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Success: false, Error: "Неверный телефон или пароль"})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: tokens})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Success: false, Error: "Неверные данные"})
		return
	}
	tokens, err := h.authSvc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Success: false, Error: "Недействительный токен"})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Success: true, Data: tokens})
}
