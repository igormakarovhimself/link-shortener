package repository

import (
	"context"
	"link-shortener/internal/model"
)

type URLRepository interface {
	Save(ctx context.Context, shortURL, originalURL, userID string) error
	Get(ctx context.Context, shortURL string) (string, error)
	GetByOriginalURL(ctx context.Context, originalURL string) (string, error)
	SaveBatch(ctx context.Context, shortURLs, originalURLs []string, userID string) error
	GetURLsByUserID(ctx context.Context, userID string) ([]model.URLPair, error)
	Ping(ctx context.Context) error
}
