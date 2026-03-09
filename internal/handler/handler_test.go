package handler

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"link-shortener/internal/audit"
	"link-shortener/internal/middleware"
	"link-shortener/internal/repository"
	"link-shortener/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHandlePost(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
	}
	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "valid URL",
			body: "https://practicum.yandex.ru/",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name: "invalid URL",
			body: "not-a-url",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "empty body",
			body: "",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := repository.NewLocalRepository()
			logger := zap.NewNop().Sugar()
			svc := service.NewShortenerService(repo, logger)
			handler := NewHandler(svc, "http://localhost:8080", audit.NewPublisher())

			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			ctx := context.WithValue(request.Context(), middleware.UserIDKey, "test-user-id")
			request = request.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.HandlePost(w, request)

			res := w.Result()
			defer func() { _ = res.Body.Close() }()

			assert.Equal(t, test.want.statusCode, res.StatusCode)

			if test.want.statusCode == http.StatusCreated {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Contains(t, string(resBody), "http://localhost:8080/")
			}
		})
	}
}

func TestGzipCompression(t *testing.T) {
	repo := repository.NewLocalRepository()
	logger := zap.NewNop().Sugar()
	svc := service.NewShortenerService(repo, logger)
	h := NewHandler(svc, "http://localhost:8080", audit.NewPublisher())

	r := chi.NewRouter()
	r.Use(middleware.WithGzip())
	r.Post("/api/shorten", h.HandleAPIShorten)

	srv := httptest.NewServer(r)
	defer srv.Close()

	requestBody := `{"url":"https://practicum.yandex.ru"}`

	successBody := `{"result":"http://localhost:8080/`

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL+"/api/shorten", buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Content-Type", "application/json")

		client := &http.Client{
			Transport: &http.Transport{
				DisableCompression: true,
			},
		}

		resp, err := client.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer func() { _ = resp.Body.Close() }()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(b), successBody)
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", srv.URL+"/api/shorten", buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")
		r.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer func() { _ = resp.Body.Close() }()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		assert.Contains(t, string(b), successBody)
	})
}

func TestHandleGet(t *testing.T) {
	type want struct {
		statusCode int
		location   string
	}
	tests := []struct {
		name     string
		shortURL string
		setupURL string
		want     want
	}{
		{
			name:     "existing URL",
			shortURL: "test1234",
			setupURL: "https://practicum.yandex.ru/",
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://practicum.yandex.ru/",
			},
		},
		{
			name:     "non-existing URL",
			shortURL: "notfound",
			setupURL: "",
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := repository.NewLocalRepository()
			logger := zap.NewNop().Sugar()
			svc := service.NewShortenerService(repo, logger)
			handler := NewHandler(svc, "http://localhost:8080", audit.NewPublisher())

			if test.setupURL != "" {
				err := repo.Save(context.Background(), test.shortURL, test.setupURL, "test-user-id")
				require.NoError(t, err)
			}

			r := chi.NewRouter()
			r.Get("/{id}", handler.HandleGet)

			request := httptest.NewRequest(http.MethodGet, "/"+test.shortURL, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			res := w.Result()
			defer func() { _ = res.Body.Close() }()

			assert.Equal(t, test.want.statusCode, res.StatusCode)

			if test.want.statusCode == http.StatusTemporaryRedirect {
				assert.Equal(t, test.want.location, res.Header.Get("Location"))
			}
		})
	}
}

func TestHandleAPIShorten(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
	}
	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "valid JSON",
			body: `{"url":"https://practicum.yandex.ru"}`,
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
			},
		},
		{
			name: "invalid JSON",
			body: `{"url":}`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid URL in JSON",
			body: `{"url":"not-a-url"}`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "empty URL in JSON",
			body: `{"url":""}`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := repository.NewLocalRepository()
			logger := zap.NewNop().Sugar()
			svc := service.NewShortenerService(repo, logger)
			handler := NewHandler(svc, "http://localhost:8080", audit.NewPublisher())

			request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(request.Context(), middleware.UserIDKey, "test-user-id")
			request = request.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.HandleAPIShorten(w, request)

			res := w.Result()
			defer func() { _ = res.Body.Close() }()

			assert.Equal(t, test.want.statusCode, res.StatusCode)

			if test.want.statusCode == http.StatusCreated {
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

				var response ShortenResponse
				err := json.NewDecoder(res.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response.Result, "http://localhost:8080/")
			}
		})
	}
}
