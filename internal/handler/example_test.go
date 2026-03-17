package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"link-shortener/internal/audit"
	"link-shortener/internal/middleware"
	"link-shortener/internal/repository"
	"link-shortener/internal/service"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	. "link-shortener/internal/handler"
)

func ExampleURLHandler_HandlePost() {
	repo := repository.NewLocalRepository()
	svc := service.NewShortenerService(repo, zap.NewNop().Sugar())
	h := NewHandler(svc, "http://localhost:8080", audit.NewPublisher(), "")

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	r = r.WithContext(context.WithValue(r.Context(), middleware.UserIDKey, "user1"))
	w := httptest.NewRecorder()

	h.HandlePost(w, r)

	fmt.Println(w.Code)
	// Output:
	// 201
}

func ExampleURLHandler_HandleGet() {
	repo := repository.NewLocalRepository()
	svc := service.NewShortenerService(repo, zap.NewNop().Sugar())
	h := NewHandler(svc, "http://localhost:8080", audit.NewPublisher(), "")

	_ = repo.Save(context.Background(), "abc12345", "https://example.com", "user1")

	router := chi.NewRouter()
	router.Get("/{id}", h.HandleGet)

	r := httptest.NewRequest(http.MethodGet, "/abc12345", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	fmt.Println(w.Code)
	fmt.Println(w.Header().Get("Location"))
	// Output:
	// 307
	// https://example.com
}

func ExampleURLHandler_HandleAPIShorten() {
	repo := repository.NewLocalRepository()
	svc := service.NewShortenerService(repo, zap.NewNop().Sugar())
	h := NewHandler(svc, "http://localhost:8080", audit.NewPublisher(), "")

	body := `{"url":"https://example.com"}`
	r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = r.WithContext(context.WithValue(r.Context(), middleware.UserIDKey, "user1"))
	w := httptest.NewRecorder()

	h.HandleAPIShorten(w, r)

	fmt.Println(w.Code)
	fmt.Println(w.Header().Get("Content-Type"))
	// Output:
	// 201
	// application/json
}

func ExampleURLHandler_HandleAPIBatch() {
	repo := repository.NewLocalRepository()
	svc := service.NewShortenerService(repo, zap.NewNop().Sugar())
	h := NewHandler(svc, "http://localhost:8080", audit.NewPublisher(), "")

	body := `[
		{"correlation_id":"1","original_url":"https://example.com"},
		{"correlation_id":"2","original_url":"https://google.com"}
	]`
	r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = r.WithContext(context.WithValue(r.Context(), middleware.UserIDKey, "user1"))
	w := httptest.NewRecorder()

	h.HandleAPIBatch(w, r)

	fmt.Println(w.Code)
	// Output:
	// 201
}
