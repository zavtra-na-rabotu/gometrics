package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	w  http.ResponseWriter
	gw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		gw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.gw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	//if statusCode < http.StatusMultipleChoices {
	//	c.w.Header().Set("Content-Encoding", "gzip")
	//}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.gw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	gr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		gr: zr,
	}, nil
}

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.gr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("failed to close gzip reader: %w", err)
	}
	return c.gr.Close()
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseWriter := w

		// Check if client accepts gzip encoding and Content-Type is (application/json or text/html)
		acceptEncoding := r.Header.Get("Accept-Encoding")
		acceptGzip := strings.Contains(acceptEncoding, "gzip")
		if acceptGzip {
			gzipWriter := newCompressWriter(w)
			responseWriter = gzipWriter
			responseWriter.Header().Set("Content-Encoding", "gzip")
			defer func() {
				if err := gzipWriter.Close(); err != nil {
					http.Error(w, "Failed to close gzip writer", http.StatusInternalServerError)
					return
				}
			}()
		}

		// Decompress request if Content-Encoding is gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		receivedGzip := strings.Contains(contentEncoding, "gzip")
		if receivedGzip {
			gzipReader, err := newCompressReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to create gzip reader", http.StatusInternalServerError)
				return
			}
			r.Body = gzipReader
			defer func() {
				if err := gzipReader.Close(); err != nil {
					http.Error(w, "Failed to close gzip reader", http.StatusInternalServerError)
					return
				}
			}()
		}

		next.ServeHTTP(responseWriter, r)
	})
}
