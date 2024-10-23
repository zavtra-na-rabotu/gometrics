package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// RequestLoggerMiddleware log requests and responses
func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Capture response details
		rw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(rw, r)

		zap.L().Info("Request and response",
			zap.String("URI", r.RequestURI),
			zap.String("Method", r.Method),
			zap.Int("Status", rw.Status()),
			zap.Int("Body size", rw.BytesWritten()),
			zap.Duration("Execution time", time.Since(start)),
		)
	})
}
