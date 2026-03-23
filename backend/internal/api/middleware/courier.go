package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/veggieshop/backend/internal/utils"
)

func CourierOnly() gin.HandlerFunc {
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
		if roleStr != "courier" {
			c.AbortWithStatusJSON(http.StatusForbidden, utils.ErrorResponse{
				Success: false,
				Error:   "Требуется роль курьера",
				Code:    "FORBIDDEN",
			})
			return
		}
		c.Next()
	}
}
