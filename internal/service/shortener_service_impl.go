package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"link-shortener/internal/model"
	"link-shortener/internal/repository"
	"net/url"
	"sync"

	"go.uber.org/zap"
)

const (
	numWorkers     = 3
	deleteChanSize = 1024
)

type DeleteTask struct {
	UserID    string
	ShortURLs []string
}

type DeleteResult struct {
	Success bool
	Error   error
	UserID  string
	Count   int
}

type ShortenerServiceImpl struct {
	repo     repository.URLRepository
	logger   *zap.SugaredLogger
	deleteCh chan DeleteTask
	doneCh   chan struct{}
	wg       sync.WaitGroup
}

func NewShortenerService(repo repository.URLRepository, logger *zap.SugaredLogger) *ShortenerServiceImpl {
	svc := &ShortenerServiceImpl{
		repo:     repo,
		logger:   logger,
		deleteCh: make(chan DeleteTask, deleteChanSize),
		doneCh:   make(chan struct{}),
	}

	resultChannels := make([]chan DeleteResult, numWorkers)
	for i := 0; i < numWorkers; i++ {
		resultChannels[i] = svc.deleteWorker(i)
	}

	finalResultCh := svc.fanIn(resultChannels...)
	svc.resultLogger(finalResultCh)

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
	hash := sha256.Sum256([]byte(originalURL))
	result := base64.RawURLEncoding.EncodeToString(hash[:])
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

func (s *ShortenerServiceImpl) deleteWorker(workerID int) chan DeleteResult {
	resultCh := make(chan DeleteResult)
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		defer close(resultCh)

		for task := range s.deleteCh {
			err := s.repo.DeleteURLs(context.Background(), task.ShortURLs, task.UserID)

			result := DeleteResult{
				Success: err == nil,
				Error:   err,
				UserID:  task.UserID,
				Count:   len(task.ShortURLs),
			}

			select {
			case <-s.doneCh:
				return
			case resultCh <- result:
			}
		}
	}()

	return resultCh
}

func (s *ShortenerServiceImpl) fanIn(resultChs ...chan DeleteResult) chan DeleteResult {
	finalCh := make(chan DeleteResult)
	var wg sync.WaitGroup

	for _, ch := range resultChs {
		chClosure := ch
		wg.Add(1)

		go func() {
			defer wg.Done()

			for result := range chClosure {
				select {
				case <-s.doneCh:
					return
				case finalCh <- result:
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	return finalCh
}

func (s *ShortenerServiceImpl) resultLogger(resultCh chan DeleteResult) {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		for result := range resultCh {
			if result.Error != nil {
				s.logger.Errorf("Failed to delete %d URLs for user %s: %v", result.Count, result.UserID, result.Error)
			}
		}
	}()
}

func (s *ShortenerServiceImpl) DeleteURLsAsync(ctx context.Context, shortURLs []string, userID string) {
	task := DeleteTask{
		UserID:    userID,
		ShortURLs: shortURLs,
	}

	select {
	case s.deleteCh <- task:
	case <-ctx.Done():
	}
}

func (s *ShortenerServiceImpl) Shutdown(ctx context.Context) error {
	close(s.deleteCh)
	close(s.doneCh)

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *ShortenerServiceImpl) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
