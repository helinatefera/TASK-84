package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func GenerateShareToken() (string, error) {
	return GenerateToken(32) // 64 hex characters
}

func GenerateCSRFToken() (string, error) {
	return GenerateToken(32)
}
