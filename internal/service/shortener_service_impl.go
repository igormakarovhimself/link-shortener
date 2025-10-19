package service

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
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

func (s *ShortenerServiceImpl) ShortenURL(originalURL string) (string, error) {
	_, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	shortenURL := s.generateShortURL(originalURL)

	if err := s.repo.Save(shortenURL, originalURL); err != nil {
		return "", fmt.Errorf("failed to save URL: %w", err)
	}

	return shortenURL, nil
}

func (s *ShortenerServiceImpl) GetOriginalURL(shortURL string) (string, error) {
	originalURL, err := s.repo.Get(shortURL)
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
