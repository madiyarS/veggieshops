package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/utils"
)

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrorResponse{
				Success: false,
				Error:   "Требуется авторизация",
				Code:    "UNAUTHORIZED",
			})
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrorResponse{
				Success: false,
				Error:   "Неверный формат токена",
				Code:    "INVALID_TOKEN",
			})
			return
		}
		claims, err := utils.ParseToken(parts[1], secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrorResponse{
				Success: false,
				Error:   "Недействительный токен",
				Code:    "INVALID_TOKEN",
			})
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, utils.ErrorResponse{
				Success: false,
				Error:   "Доступ запрещен",
				Code:    "FORBIDDEN",
			})
			return
		}
		roleStr, _ := role.(string)
		if roleStr != "admin" && roleStr != "manager" {
			c.AbortWithStatusJSON(http.StatusForbidden, utils.ErrorResponse{
				Success: false,
				Error:   "Требуются права администратора",
				Code:    "FORBIDDEN",
			})
			return
		}
		c.Next()
	}
}

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	id, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	uid, ok := id.(uuid.UUID)
	return uid, ok
}

// CustomerOnly разрешает только роль customer (витрина: заказы и корзина).
func CustomerOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, utils.ErrorResponse{
				Success: false,
				Error:   "Доступ запрещен",
				Code:    "FORBIDDEN",
			})
			return
		}
		roleStr, _ := role.(string)
		if roleStr != string(models.RoleCustomer) {
			c.AbortWithStatusJSON(http.StatusForbidden, utils.ErrorResponse{
				Success: false,
				Error:   "Оформление заказа доступно только покупателям",
				Code:    "FORBIDDEN",
			})
			return
		}
		c.Next()
	}
}
