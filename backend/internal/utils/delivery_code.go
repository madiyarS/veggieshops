package utils

import (
	"crypto/rand"
	"io"
	"strings"
)

// NormalizeDeliveryCode убирает пробелы для сравнения кода.
func NormalizeDeliveryCode(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\u00a0", ""))
}

// RandomDeliveryCode — 6 цифр для передачи курьеру устно.
func RandomDeliveryCode() string {
	const digits = "0123456789"
	b := make([]byte, 6)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "000000"
	}
	for i := range b {
		b[i] = digits[int(b[i])%len(digits)]
	}
	return string(b)
}
