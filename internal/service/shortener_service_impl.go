package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"link-shortener/internal/model"
	"link-shortener/internal/repository"
	"net/url"
	"time"
)

type DeleteTask struct {
	UserID    string
	ShortURLs []string
}

type ShortenerServiceImpl struct {
	repo       repository.URLRepository
	deleteChan chan DeleteTask
}

func NewShortenerService(repo repository.URLRepository) *ShortenerServiceImpl {
	svc := &ShortenerServiceImpl{
		repo:       repo,
		deleteChan: make(chan DeleteTask, 1024),
	}
	go svc.startDeleteWorker()
	return svc
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

func (s *ShortenerServiceImpl) DeleteURLsAsync(shortURLs []string, userID string) {
	task := DeleteTask{
		UserID:    userID,
		ShortURLs: shortURLs,
	}
	s.deleteChan <- task
}

func (s *ShortenerServiceImpl) startDeleteWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var tasks []DeleteTask

	for {
		select {
		case task := <-s.deleteChan:
			tasks = append(tasks, task)
		case <-ticker.C:
			if len(tasks) == 0 {
				continue
			}

			s.processDeletions(tasks)
			tasks = nil
		}
	}
}

func (s *ShortenerServiceImpl) processDeletions(tasks []DeleteTask) {
	ctx := context.Background()

	for _, task := range tasks {
		err := s.repo.DeleteURLs(ctx, task.ShortURLs, task.UserID)
		if err != nil {
			continue
		}
	}
}

func (s *ShortenerServiceImpl) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
