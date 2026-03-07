package repository

import (
	"context"

	"link-shortener/internal/model"
)

// URLRepository — интерфейс для работы с хранилищем URL.
type URLRepository interface {
	// Save сохраняет новую пару короткий/оригинальный URL.
	Save(ctx context.Context, shortURL, originalURL, userID string) error
	// Get возвращает оригинальный URL по короткому.
	Get(ctx context.Context, shortURL string) (string, error)
	// GetByOriginalURL ищет короткий URL по оригинальному.
	GetByOriginalURL(ctx context.Context, originalURL string) (string, error)
	// SaveBatch сохраняет несколько пар URL за один вызов
	SaveBatch(ctx context.Context, shortURLs, originalURLs []string, userID string) error
	// GetURLsByUserID возвращает все URL пользователя.
	GetURLsByUserID(ctx context.Context, userID string) ([]model.URLPair, error)
	// DeleteURLs помечает указанные URL как удаленные
	DeleteURLs(ctx context.Context, shortURLs []string, userID string) error
	// Ping проверяет доступность хранилища.
	Ping(ctx context.Context) error
}
