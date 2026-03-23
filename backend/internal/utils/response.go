package utils

import "github.com/gin-gonic/gin"

// SuccessResponse standard API success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse standard API error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

// GinError JSON-ошибка с опциональным кодом для поддержки и фронта.
func GinError(c *gin.Context, status int, message, code string) {
	c.JSON(status, ErrorResponse{Success: false, Error: message, Code: code})
}

// PaginatedResponse for paginated lists
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	Limit   int         `json:"limit"`
}
