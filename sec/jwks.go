package sec

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json/v2"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
)

// JWK JSON Web Group
type JWK struct {
	Kty string `json:"kty"` // Group Type
	Use string `json:"use"` // Usage
	Kid string `json:"kid"` // Group ID
	Alg string `json:"alg"` // Algorithm
	N   string `json:"n"`   // Modulus
	E   string `json:"e"`   // Exponent
}

// ToPublicKey Convert JWK to an rsa.PublicKey
func (j *JWK) ToPublicKey() (*rsa.PublicKey, error) {
	nb, err := base64.RawURLEncoding.DecodeString(j.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode N: %w", err)
	}
	eb, err := base64.RawURLEncoding.DecodeString(j.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode E: %w", err)
	}
	e := 0
	for _, b := range eb {
		e = e<<8 + int(b)
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nb),
		E: e,
	}, nil
}

func NewJWKFromPublicKey(kid string, pub *rsa.PublicKey) JWK {
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes())
	return JWK{
		Kty: "RSA",
		Use: "sig",
		Kid: kid,
		Alg: "RS256",
		N:   n,
		E:   e,
	}
}

// JWKS JSON Web Group Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

func (s *JWKS) CreateJSONFile(path string) error {
	var (
		err  error
		file *os.File
	)
	if file, err = os.Create(path); err != nil {
		return err
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Printf("[ERROR] %v", err)
		}
	}()
	if err = json.MarshalWrite(file, s); err != nil {
		return err
	}
	return nil
}

func (s *JWKS) GetJWKByKID(kid string) (*JWK, error) {
	for _, key := range s.Keys {
		if key.Kid == kid {
			return &key, nil // copy
		}
	}
	return nil, errors.New("key not found")
}

func LoadPublicPEMKeysAsJWKS(dirPath string) (*JWKS, error) {
	var (
		err        error
		keys       []JWK
		dirEntries []os.DirEntry
		pemBytes   []byte
		rest       []byte
		pemBlock   *pem.Block
		pub        any
		ok         bool
		publicKey  *rsa.PublicKey
		kid        string
	)
	if dirEntries, err = os.ReadDir(dirPath); err != nil {
		return nil, fmt.Errorf("failed to read key directory: %v", err)
	}
	for _, entry := range dirEntries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".pem" || !strings.HasSuffix(entry.Name(), "_public.pem") {
			continue
		}
		if pemBytes, err = os.ReadFile(filepath.Join(dirPath, entry.Name())); err != nil {
			return nil, fmt.Errorf("failed to read pem file %s: %w", entry.Name(), err)
		}
		if pemBlock, rest = pem.Decode(pemBytes); pemBlock == nil || pemBlock.Type != "PUBLIC KEY" {
			continue
		}
		if len(rest) > 0 {
			// We need a single public pem key for each key_id
			return nil, fmt.Errorf("extra data found after PEM block in %s", entry.Name())
		}
		if pub, err = x509.ParsePKIXPublicKey(pemBlock.Bytes); err != nil {
			return nil, fmt.Errorf("failed to parse public key %s: %w", entry.Name(), err)
		}
		if publicKey, ok = pub.(*rsa.PublicKey); !ok {
			// Skips non-RSA keys gracefully
			continue
		}
		kid = strings.TrimSuffix(entry.Name(), "_public.pem")
		keys = append(keys, NewJWKFromPublicKey(kid, publicKey))
	}
	return &JWKS{Keys: keys}, nil
}
