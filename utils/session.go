package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// Generate a random session token
func GenerateSessionToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
