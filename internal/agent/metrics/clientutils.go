package metrics

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"

	"go.uber.org/zap"
)

func calculateHash(data []byte, key string) string {
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

func getLocalIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP
}

// encryptWithPublicKey encrypts data using RSA public key
func encryptWithPublicKey(data []byte, publicKey *rsa.PublicKey) (string, error) {
	encryptedData, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, data, nil)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %w", err)
	}

	return base64.StdEncoding.EncodeToString(encryptedData), nil
}

// encryptWithAES encrypts data using AES-GCM
func encryptWithAES(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	nonce := make([]byte, 12)
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	encryptedData := aesGCM.Seal(nonce, nonce, data, nil)
	return encryptedData, nil
}

func encryptRequestBody(body []byte, publicKey *rsa.PublicKey) ([]byte, string, error) {
	// Generate AES key
	aesKey := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, aesKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate AES key: %w", err)
	}

	// Encrypt data with AES key
	encryptedData, err := encryptWithAES(body, aesKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to encrypt data with AES: %w", err)
	}

	// Encrypt AES key with RSA public key
	encryptedAESKey, err := encryptWithPublicKey(aesKey, publicKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to encrypt AES key: %w", err)
	}

	return encryptedData, encryptedAESKey, nil
}

func compressBody(compressedData *bytes.Buffer, jsonData []byte) error {
	gzipWriter := gzip.NewWriter(compressedData)
	_, err := gzipWriter.Write(jsonData)
	if err != nil {
		zap.L().Error("Error compressing data", zap.Error(err))
		return fmt.Errorf("error compressing data: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		zap.L().Error("Error closing gzip writer", zap.Error(err))
		return fmt.Errorf("error closing gzip writer: %w", err)
	}

	return nil
}
