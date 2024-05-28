package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type compressWriter struct {
	http.ResponseWriter
	gw *gzip.Writer
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip encoding and Content-Type is (application/json or text/html)
		acceptEncoding := r.Header.Get("Accept-Encoding")
		contentType := r.Header.Get("Content-Type")

		acceptGzip := strings.Contains(acceptEncoding, "gzip")
		compressAllowed := strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")
		if acceptGzip && compressAllowed {
			w.Header().Set("Content-Encoding", "gzip")

			gzWriter := gzip.NewWriter(w)
			defer func() {
				if err := gzWriter.Close(); err != nil {
					http.Error(w, "Failed to close gzip writer", http.StatusInternalServerError)
					return
				}
			}()

			cmpWriter := &compressWriter{gw: gzWriter, ResponseWriter: w}
			w = cmpWriter
		}

		// Decompress request if Content-Encoding is gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		receivedGzip := strings.Contains(contentEncoding, "gzip")
		if receivedGzip {
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to create gzip reader", http.StatusInternalServerError)
				return
			}
			defer func() {
				if err := gzipReader.Close(); err != nil {
					http.Error(w, "Failed to close gzip reader", http.StatusInternalServerError)
					return
				}
			}()

			r.Body = gzipReader
		}

		next.ServeHTTP(w, r)
	})
}
