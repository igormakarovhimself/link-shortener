package repository

import (
	"context"
	"fmt"
	"sync"

	"link-shortener/internal/model"
)

type urlData struct {
	originalURL string
	userID      string
	isDeleted   bool
}

type LocalRepository struct {
	storage map[string]urlData
	mutex   sync.RWMutex
}

func NewLocalRepository() *LocalRepository {
	return &LocalRepository{
		storage: make(map[string]urlData),
	}
}

func (r *LocalRepository) Save(ctx context.Context, shortURL, originalURL, userID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.storage[shortURL] = urlData{
		originalURL: originalURL,
		userID:      userID,
	}
	return nil
}

func (r *LocalRepository) Get(ctx context.Context, shortURL string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	data, exists := r.storage[shortURL]
	if !exists {
		return "", fmt.Errorf("URL not found")
	}
	if data.isDeleted {
		return "", ErrURLDeleted
	}
	return data.originalURL, nil
}

func (r *LocalRepository) GetByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for shortURL, data := range r.storage {
		if data.originalURL == originalURL {
			return shortURL, nil
		}
	}
	return "", fmt.Errorf("URL not found")
}

func (r *LocalRepository) SaveBatch(ctx context.Context, shortURLs, originalURLs []string, userID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i := range shortURLs {
		r.storage[shortURLs[i]] = urlData{
			originalURL: originalURLs[i],
			userID:      userID,
		}
	}

	return nil
}

func (r *LocalRepository) GetURLsByUserID(ctx context.Context, userID string) ([]model.URLPair, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var urls []model.URLPair
	for shortURL, data := range r.storage {
		if data.userID == userID {
			urls = append(urls, model.URLPair{
				ShortURL:    shortURL,
				OriginalURL: data.originalURL,
			})
		}
	}

	return urls, nil
}

func (r *LocalRepository) DeleteURLs(ctx context.Context, shortURLs []string, userID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, shortURL := range shortURLs {
		if data, exists := r.storage[shortURL]; exists && data.userID == userID {
			data.isDeleted = true
			r.storage[shortURL] = data
		}
	}

	return nil
}

func (r *LocalRepository) Ping(ctx context.Context) error {
	return nil
}
