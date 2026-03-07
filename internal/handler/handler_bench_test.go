package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"

	"link-shortener/internal/audit"
	"link-shortener/internal/repository"
	"link-shortener/internal/service"
)

func BenchmarkHandlePost(b *testing.B) {
	repo := repository.NewLocalRepository()
	svc := service.NewShortenerService(repo, zap.NewNop().Sugar())
	h := NewHandler(svc, "http://localhost:8080", audit.NewPublisher())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru/go-developer/"))
		w := httptest.NewRecorder()
		b.StartTimer()
		h.HandlePost(w, req)
	}
}

func BenchmarkHandleAPIShorten(b *testing.B) {
	repo := repository.NewLocalRepository()
	svc := service.NewShortenerService(repo, zap.NewNop().Sugar())
	h := NewHandler(svc, "http://localhost:8080", audit.NewPublisher())
	body := `{"url":"https://practicum.yandex.ru/go-developer/"}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		b.StartTimer()
		h.HandleAPIShorten(w, req)
	}
}
