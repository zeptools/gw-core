package sec

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

// Read https://pkg.go.dev/golang.org/x/crypto/chacha20poly1305

type XChaCha20Poly1305Cipher struct {
	aead       cipher.AEAD
	encodeFunc func([]byte) string          // e.g. base64.RawURLEncoding.EncodeToString, hex.EncodeToString
	decodeFunc func(string) ([]byte, error) // e.g. base64.RawURLEncoding.DecodeString, hex.DecodeString
}

func NewXChaCha20Poly1305Cipher(
	key []byte,
	encodeFunc func([]byte) string,
	decodeFunc func(string) ([]byte, error),
) (*XChaCha20Poly1305Cipher, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("key must be %d bytes, got %d", chacha20poly1305.KeySize, len(key))
	}
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}
	return &XChaCha20Poly1305Cipher{
		aead:       aead,
		encodeFunc: encodeFunc,
		decodeFunc: decodeFunc,
	}, nil
}

func NewXChaCha20Poly1305CipherBase64(key []byte) (*XChaCha20Poly1305Cipher, error) {
	return NewXChaCha20Poly1305Cipher(
		key,
		base64.RawURLEncoding.EncodeToString,
		base64.RawURLEncoding.DecodeString,
	)
}

func (c *XChaCha20Poly1305Cipher) EncryptEncode(plaintext []byte) (string, error) {
	// Generate a random nonce every time, and leave capacity for the ciphertext
	nonce := make([]byte, c.aead.NonceSize(), c.aead.NonceSize()+len(plaintext)+c.aead.Overhead())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := c.aead.Seal(nonce, nonce, plaintext, nil)
	return c.encodeFunc(ciphertext), nil
}

func (c *XChaCha20Poly1305Cipher) DecodeDecrypt(encodedCiphertext string) ([]byte, error) {
	// Decode
	data, err := c.decodeFunc(encodedCiphertext)
	if err != nil {
		return nil, err
	}
	nonceSize := c.aead.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	// Split nonce and ciphertext
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	// Decrypt the message and check it wasn't tampered with
	return c.aead.Open(nil, nonce, ciphertext, nil)
}
