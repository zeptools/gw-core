package sec

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateOpaqueAccessRefreshTokenPair(byteLength int) (string, string, error) {
	var (
		err          error
		accessToken  string
		refreshToken string
	)
	if accessToken, err = GenerateOpaqueToken(byteLength); err != nil {
		return "", "", err
	}
	if refreshToken, err = GenerateOpaqueToken(byteLength); err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

type RefreshInfo struct {
	UserID     int
	ClientID   string
	ValidUntil int64  // Hardcap. [NOTE] Existence = at least in the Expiration Sliding Window
	AccessHash string // Hash of the access_token issued together
}

func (i RefreshInfo) String() string {
	return fmt.Sprintf("%d:%s:%d:%s", i.UserID, i.ClientID, i.ValidUntil, i.AccessHash)
}

func ParseRefreshInfo(s string) (*RefreshInfo, error) {
	segs := strings.Split(s, ":")
	if len(segs) != 4 {
		return nil, fmt.Errorf("invalid refresh info format")
	}
	uid, err := strconv.Atoi(segs[0])
	if err != nil {
		return nil, err
	}
	validUntil, err := strconv.ParseInt(segs[2], 10, 64)
	if err != nil {
		return nil, err
	}
	return &RefreshInfo{
		UserID:     uid,
		ClientID:   segs[1],
		ValidUntil: validUntil,
		AccessHash: segs[3],
	}, nil
}

// GenerateRSASignedJWTIDToken generates a jwt id_token signed by RS256
// sub: User ID
// email: Email Used for Authentication
func GenerateRSASignedJWTIDToken(iss string, sub string, email string, clientID string, privateKey *rsa.PrivateKey, kid string, expDuration time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   sub,
		"email": email,
		"iat":   now.Unix(),
		"exp":   now.Add(expDuration).Unix(),
		"iss":   iss,
		"aud":   clientID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	return token.SignedString(privateKey)
}

// ParseRSASignedToken verifies a signed token (string) into a parsed jwt.Token object
func ParseRSASignedToken(signedToken string, pubKey *rsa.PublicKey) (*jwt.Token, error) {
	return jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		// ensure alg is RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return pubKey, nil
	})
}

func GetClaimsFromParsedJWTToken(parsedToken *jwt.Token) (jwt.MapClaims, error) {
	if !parsedToken.Valid {
		return nil, errors.New("invalid token")
	}
	claimMap, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("failed to convert token claims to a map")
	}
	return claimMap, nil
}

// GenerateOpaqueToken generates a Base64-encoded, URL-safe, opaque random string
func GenerateOpaqueToken(byteLength int) (string, error) {
	if byteLength <= 0 {
		byteLength = 32 // default 32 bytes (256 bits)
	}
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("rand.Read: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func HashHexSHA256(data string) string {
	// SHA256 checksum (digest) of the data
	checksum := sha256.Sum256([]byte(data))
	// hexadecimal encoding
	return hex.EncodeToString(checksum[:])
}
