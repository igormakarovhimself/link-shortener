package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	http.ResponseWriter
	zw             *gzip.Writer
	headerWritten  bool
	shouldCompress bool
	checkedType    bool
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	zw, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
	return &compressWriter{
		ResponseWriter: w,
		zw:             zw,
	}
}

func (w *compressWriter) Write(b []byte) (int, error) {
	if !w.checkedType {
		w.checkedType = true
		contentType := w.ResponseWriter.Header().Get("Content-Type")
		if contentType == "" {
			contentType = http.DetectContentType(b)
		}
		w.shouldCompress = supportsCompression(contentType)
	}

	if !w.headerWritten {
		if w.shouldCompress {
			w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		}
		w.WriteHeader(http.StatusOK)
	}

	if w.shouldCompress {
		return w.zw.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (w *compressWriter) WriteHeader(statusCode int) {
	if w.headerWritten {
		return
	}
	w.headerWritten = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *compressWriter) Close() error {
	if w.shouldCompress {
		return w.zw.Close()
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
			supportsGzip := acceptEncoding != "" && strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				cw := newCompressWriter(w)
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
