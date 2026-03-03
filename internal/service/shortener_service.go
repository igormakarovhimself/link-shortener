package service

import (
	"context"

	"link-shortener/internal/model"
)

type ShortenerService interface {
	ShortenURL(ctx context.Context, originalURL, userID string) (string, error)
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
	SaveBatch(ctx context.Context, shortURLs, originalURLs []string, userID string) error
	GenerateShortURL(originalURL string) (string, error)
	GetURLsByUserID(ctx context.Context, userID string) ([]model.URLPair, error)
	DeleteURLsAsync(ctx context.Context, shortURLs []string, userID string)
	Ping(ctx context.Context) error
}
