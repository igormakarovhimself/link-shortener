package service

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"link-shortener/internal/repository"
)

func BenchmarkGenerateShortURL(b *testing.B) {
	repo := repository.NewLocalRepository()
	svc := NewShortenerService(repo, zap.NewNop().Sugar())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.generateShortURL("https://practicum.yandex.ru/go-developer/")
	}
}

func BenchmarkShortenURL(b *testing.B) {
	repo := repository.NewLocalRepository()
	svc := NewShortenerService(repo, zap.NewNop().Sugar())
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.ShortenURL(ctx, "https://practicum.yandex.ru/go-developer/", "user1")
	}
}
