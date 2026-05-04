package validate

import (
	"errors"
	"strings"
)

// EAN strips whitespace, validates digit-only + valid length (8/12/13/14).
// Empty string is allowed (field is optional).
func EAN(raw string) (string, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", nil
	}
	for _, c := range v {
		if c < '0' || c > '9' {
			return "", errors.New("EAN deve conter apenas dígitos")
		}
	}
	switch len(v) {
	case 8, 12, 13, 14:
		return v, nil
	default:
		return "", errors.New("EAN deve ter 8, 12, 13 ou 14 dígitos")
	}
}
