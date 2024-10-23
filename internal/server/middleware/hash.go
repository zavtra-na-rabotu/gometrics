package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type hashWriter struct {
	http.ResponseWriter
	key string
}

func (w *hashWriter) Write(p []byte) (int, error) {
	w.Header().Set("HashSHA256", calculateHash(p, w.key))
	return w.ResponseWriter.Write(p)
}

func calculateHash(data []byte, key string) string {
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

// ResponseHashMiddleware calculate hash of response and set HashSHA256 header
func ResponseHashMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(&hashWriter{ResponseWriter: w, key: key}, r)
		})
	}
}

// RequestHashMiddleware calculate hash of request and compare with received hash from HashSHA256 header
func RequestHashMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHash := r.Header.Get("HashSHA256")
			if receivedHash == "" {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				zap.L().Error("Error reading body", zap.Error(err))
				http.Error(w, "Error reading body", http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			calculatedHash := calculateHash(body, key)

			if receivedHash != calculatedHash {
				zap.L().Error("Hash mismatch", zap.String("received hash", receivedHash), zap.String("calculated hash", calculatedHash))
				http.Error(w, "Hash mismatch", http.StatusBadRequest)
				return
			}

			zap.L().Info("Hashes are identical")

			next.ServeHTTP(w, r)
		})
	}
}
