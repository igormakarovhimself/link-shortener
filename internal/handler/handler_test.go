package handler

import (
	"io"
	"link-shortener/internal/repository"
	"link-shortener/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			svc := service.NewShortenerService(repo)
			handler := NewHandler(svc, "http://localhost:8080")

			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			w := httptest.NewRecorder()

			handler.HandlePost(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.statusCode, res.StatusCode)

			if test.want.statusCode == http.StatusCreated {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Contains(t, string(resBody), "http://localhost:8080/")
			}
		})
	}
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
			svc := service.NewShortenerService(repo)
			handler := NewHandler(svc, "http://localhost:8080")

			if test.setupURL != "" {
				err := repo.Save(test.shortURL, test.setupURL)
				require.NoError(t, err)
			}

			r := chi.NewRouter()
			r.Get("/{id}", handler.HandleGet)

			request := httptest.NewRequest(http.MethodGet, "/"+test.shortURL, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.statusCode, res.StatusCode)

			if test.want.statusCode == http.StatusTemporaryRedirect {
				assert.Equal(t, test.want.location, res.Header.Get("Location"))
			}
		})
	}
}
