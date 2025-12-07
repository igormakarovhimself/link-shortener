package repository

import "context"

type URLRepository interface {
	Save(ctx context.Context, shortURL, originalURL string) error
	Get(ctx context.Context, shortURL string) (string, error)
	GetByOriginalURL(ctx context.Context, originalURL string) (string, error)
	SaveBatch(ctx context.Context, shortURLs, originalURLs []string) error
	Ping(ctx context.Context) error
}
