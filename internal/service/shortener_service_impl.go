package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"link-shortener/internal/model"
	"link-shortener/internal/repository"
	"net/url"
)

type ShortenerServiceImpl struct {
	repo repository.URLRepository
}

func NewShortenerService(repo repository.URLRepository) *ShortenerServiceImpl {
	return &ShortenerServiceImpl{
		repo: repo,
	}
}

func (s *ShortenerServiceImpl) ShortenURL(ctx context.Context, originalURL, userID string) (string, error) {
	_, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	shortenURL := s.generateShortURL(originalURL)

	if err := s.repo.Save(ctx, shortenURL, originalURL, userID); err != nil {
		return "", err
	}

	return shortenURL, nil
}

func (s *ShortenerServiceImpl) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	originalURL, err := s.repo.Get(ctx, shortURL)
	if err != nil {
		return "", fmt.Errorf("URL not found: %w", err)
	}
	return originalURL, nil
}

func (s *ShortenerServiceImpl) generateShortURL(originalURL string) string {
	hash := sha256.New()
	hash.Write([]byte(originalURL))
	resultHash := hash.Sum(nil)
	result := base64.RawURLEncoding.EncodeToString(resultHash)

	return result[:8]
}

func (s *ShortenerServiceImpl) SaveBatch(ctx context.Context, shortURLs, originalURLs []string, userID string) error {
	return s.repo.SaveBatch(ctx, shortURLs, originalURLs, userID)
}

func (s *ShortenerServiceImpl) GenerateShortURL(originalURL string) (string, error) {
	_, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	return s.generateShortURL(originalURL), nil
}

func (s *ShortenerServiceImpl) GetURLsByUserID(ctx context.Context, userID string) ([]model.URLPair, error) {
	return s.repo.GetURLsByUserID(ctx, userID)
}

func (s *ShortenerServiceImpl) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
