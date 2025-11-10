package session

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateWebSessionID generates 32 hex (0-9a-f) string from 16 random bytes for a Session ID
func GenerateWebSessionID() (string, error) {
	b := make([]byte, 16) // 128-bit random ID
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
