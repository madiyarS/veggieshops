package utils

// Коды ошибок API (поле code в ErrorResponse).
const (
	CodeValidation       = "VALIDATION_ERROR"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeNotFound         = "NOT_FOUND"
	CodeMinOrderAmount   = "MIN_ORDER_AMOUNT"
	CodeInsufficientStock = "INSUFFICIENT_STOCK"
	CodeProductUnavailable = "PRODUCT_UNAVAILABLE"
	CodeInternal         = "INTERNAL_ERROR"
)
