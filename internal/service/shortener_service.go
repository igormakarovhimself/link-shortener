package service

import (
	"context"

	"link-shortener/internal/model"
)

// ShortenerService — интерфейс бизнес-логики сервиса.
type ShortenerService interface {
	// ShortenURL валидирует URL и сохраняет его, возвращает короткий идентификатор.
	ShortenURL(ctx context.Context, originalURL, userID string) (string, error)
	// GetOriginalURL возвращает оригинальный URL по короткому идентификатору.
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
	// SaveBatch сохраняет несколько пар URL за один раз.
	SaveBatch(ctx context.Context, shortURLs, originalURLs []string, userID string) error
	// GenerateShortURL генерирует короткий идентификатор для URL без сохранения.
	GenerateShortURL(originalURL string) (string, error)
	// GetURLsByUserID возвращает все URL пользователя.
	GetURLsByUserID(ctx context.Context, userID string) ([]model.URLPair, error)
	// DeleteURLsAsync отправляет URL на асинхронное удаление.
	DeleteURLsAsync(ctx context.Context, shortURLs []string, userID string)
	// Ping проверяет доступность хранилища.
	Ping(ctx context.Context) error
	// GetStats возвращает количество URL и уникальных пользователей.
	GetStats(ctx context.Context) (urls int, users int, err error)
}
