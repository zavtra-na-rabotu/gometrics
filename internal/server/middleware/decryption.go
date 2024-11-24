package middleware

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

// DecryptMiddleware middleware to decrypt request with private key
func DecryptMiddleware(privateKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get encrypted AES from header
			encryptedAESKey := r.Header.Get("Encrypted-AES-Key")
			if encryptedAESKey == "" {
				http.Error(w, "Missing Encrypted-AES-Key header", http.StatusBadRequest)
				return
			}

			// Decrypt AES key
			aesKey, err := decryptWithPrivateKey(encryptedAESKey, privateKey)
			if err != nil {
				zap.L().Error("Failed to decrypt AES key", zap.Error(err))
				http.Error(w, "Failed to decrypt AES key", http.StatusInternalServerError)
				return
			}

			// Decrypt request body using decrypted AES key
			decryptedData, err := decryptWithAES(r.Body, aesKey)
			if err != nil {
				zap.L().Error("Failed to decrypt request body", zap.Error(err))
				http.Error(w, "Failed to decrypt request body", http.StatusInternalServerError)
				return
			}

			// Change encrypted body on decrypted body
			r.Body = io.NopCloser(bytes.NewReader(decryptedData))
			r.ContentLength = int64(len(decryptedData))

			next.ServeHTTP(w, r)
		})
	}
}

// decryptWithPrivateKey decrypt AES key using RSA private key
func decryptWithPrivateKey(encryptedKey string, privateKey *rsa.PrivateKey) ([]byte, error) {
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted AES key: %w", err)
	}

	aesKey, err := rsa.DecryptOAEP(sha256.New(), nil, privateKey, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES key: %w", err)
	}

	return aesKey, nil
}

// decryptWithAES decrypt request body using AES key
func decryptWithAES(encryptedBody io.ReadCloser, aesKey []byte) ([]byte, error) {
	defer encryptedBody.Close()

	encryptedData, err := io.ReadAll(encryptedBody)
	if err != nil {
		return nil, fmt.Errorf("failed to read encrypted body: %w", err)
	}

	nonceSize := 12
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("encrypted data is too short")
	}

	nonce := encryptedData[:nonceSize]
	ciphertext := encryptedData[nonceSize:]

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}
