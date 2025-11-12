package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	http.ResponseWriter
	Writer         io.Writer
	wroteHeader    bool
	shouldCompress bool
}

func (w *compressWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if w.shouldCompress {
		return w.Writer.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (w *compressWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	contentType := w.ResponseWriter.Header().Get("Content-Type")
	if contentType != "" && supportsCompression(contentType) {
		w.shouldCompress = true
		w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

type compressReader struct {
	io.ReadCloser
	Reader io.Reader
}

func (r compressReader) Read(p []byte) (n int, err error) {
	return r.Reader.Read(p)
}

func (r *compressReader) Close() error {
	if err := r.ReadCloser.Close(); err != nil {
		return err
	}
	return r.Reader.(*gzip.Reader).Close()
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		ResponseWriter: w,
		Writer:         gzip.NewWriter(w),
	}
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		ReadCloser: r,
		Reader:     gr,
	}, nil
}

func supportsCompression(contentType string) bool {
	return strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/html")
}

func WithGzip() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w

			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := acceptEncoding != "" && strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				cw := newCompressWriter(w)
				ow = cw
				defer cw.Writer.(*gzip.Writer).Close()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer cr.Close()
			}

			h.ServeHTTP(ow, r)
		})
	}
}
