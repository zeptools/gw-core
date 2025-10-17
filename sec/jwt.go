package sec

import (
	"encoding/base64"
	"errors"
	"strings"
)

func SplitSignedJwtTokenRawParts(signedToken string) (string, string, string, error) {
	parts := strings.Split(signedToken, ".")
	if len(parts) != 3 {
		return "", "", "", errors.New("invalid token format")
	}
	return parts[0], parts[1], parts[2], nil // header, payload(claims), signature
}

func DecodeJwtHeader(headerEncoded string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(headerEncoded)
}
