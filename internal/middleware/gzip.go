package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w *compressWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *compressWriter) WriteHeader(statusCode int) {
	contentType := w.ResponseWriter.Header().Get("Content-Type")
	if supportsCompression(contentType) {
		w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *compressWriter) Close() error {
	if closer, ok := w.Writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

type compressReader struct {
	io.ReadCloser
	Reader io.ReadCloser
}

func (r *compressReader) Read(p []byte) (n int, err error) {
	return r.Reader.Read(p)
}

func (r *compressReader) Close() error {
	if err := r.Reader.Close(); err != nil {
		return err
	}
	return r.ReadCloser.Close()
}

func supportsCompression(contentType string) bool {
	compressibleTypes := []string{
		"application/json",
		"text/html",
	}
	for _, ct := range compressibleTypes {
		if strings.Contains(contentType, ct) {
			return true
		}
	}
	return false
}

func WithGzip() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w

			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				cw := &compressWriter{
					ResponseWriter: w,
					Writer:         gz,
				}
				ow = cw
				defer cw.Close()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				gz, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				cr := &compressReader{
					ReadCloser: r.Body,
					Reader:     gz,
				}
				r.Body = cr
				defer cr.Close()
			}

			next.ServeHTTP(ow, r)
		})
	}
}
