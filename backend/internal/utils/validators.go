package utils

import (
	"regexp"
	"strings"
)

var phoneRegex = regexp.MustCompile(`^\+?[0-9]{10,15}$`)

// ValidatePhone validates Kazakstan phone format
func ValidatePhone(phone string) bool {
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	return phoneRegex.MatchString(cleaned)
}

// ValidatePassword checks password strength
func ValidatePassword(password string) (bool, string) {
	if len(password) < 6 {
		return false, "пароль должен быть не менее 6 символов"
	}
	return true, ""
}

// NormalizePhone normalizes phone for storage and comparison (77476951662 или +7 747... → 77476951662)
func NormalizePhone(phone string) string {
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.TrimPrefix(cleaned, "+")
	var digits strings.Builder
	for _, r := range cleaned {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	s := digits.String()
	if len(s) == 11 && s[0] == '8' {
		return "7" + s[1:]
	}
	if len(s) == 10 {
		return "7" + s
	}
	return s
}
