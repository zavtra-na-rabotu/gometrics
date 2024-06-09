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

func (w *compressWriter) Write(b []byte) (int, error) {
	return w.gw.Write(b)
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Compress
		acceptEncoding := r.Header.Get("Accept-Encoding")
		acceptGzip := strings.Contains(acceptEncoding, "gzip")
		if acceptGzip {
			w.Header().Set("Content-Encoding", "gzip")

			gzWriter := gzip.NewWriter(w)
			defer func() {
				if err := gzWriter.Close(); err != nil {
					http.Error(w, "Failed to close gzip writer", http.StatusInternalServerError)
					return
				}
			}()

			cmpWriter := &compressWriter{ResponseWriter: w, gw: gzWriter}
			w = cmpWriter
		}

		// Decompress
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
