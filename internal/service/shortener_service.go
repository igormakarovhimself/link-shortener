package service

import "context"

type ShortenerService interface {
	ShortenURL(ctx context.Context, originalURL string) (string, error)
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
	SaveBatch(ctx context.Context, shortURLs, originalURLs []string) error
	GenerateShortURL(originalURL string) (string, error)
	Ping(ctx context.Context) error
}
