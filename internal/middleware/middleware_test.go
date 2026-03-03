package middleware

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"link-shortener/internal/auth"
)

func TestGetUserID(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, "user123")
	assert.Equal(t, "user123", GetUserID(ctx))

	assert.Equal(t, "", GetUserID(context.Background()))
}

func TestWithAuth_NewUser(t *testing.T) {
	handler := WithAuth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r.Context())
		assert.NotEmpty(t, userID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	result := w.Result()
	defer result.Body.Close()
	cookies := result.Cookies()
	assert.NotEmpty(t, cookies)
}

func TestWithAuth_ExistingCookie(t *testing.T) {
	userID := "known-user"
	cookieValue := auth.BuildCookieValue(userID)

	var capturedUserID string
	handler := WithAuth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "user_id", Value: cookieValue})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, userID, capturedUserID)
}

func TestWithGzip_AcceptsGzip(t *testing.T) {
	handler := WithGzip()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"ok"}`))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

	gr, err := gzip.NewReader(w.Body)
	require.NoError(t, err)
	body, err := io.ReadAll(gr)
	require.NoError(t, err)
	assert.Equal(t, `{"result":"ok"}`, string(body))
}

func TestWithGzip_NoAcceptEncoding(t *testing.T) {
	handler := WithGzip()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result":"ok"}`))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Empty(t, w.Header().Get("Content-Encoding"))
	assert.Equal(t, `{"result":"ok"}`, w.Body.String())
}

func TestWithGzip_SendsGzip(t *testing.T) {
	var received string
	handler := WithGzip()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received = string(body)
		w.WriteHeader(http.StatusOK)
	}))

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte("hello world"))
	gz.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hello world", received)
}

func TestSupportsCompression(t *testing.T) {
	assert.True(t, supportsCompression("application/json"))
	assert.True(t, supportsCompression("text/html; charset=utf-8"))
	assert.False(t, supportsCompression("image/png"))
	assert.False(t, supportsCompression("application/octet-stream"))
}

func TestWithLogging(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := WithLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", strings.NewReader(""))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
