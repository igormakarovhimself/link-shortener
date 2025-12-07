package repository

import (
	"context"
	"fmt"
	"sync"
)

type LocalRepository struct {
	storage map[string]string
	mutex   sync.RWMutex
}

func NewLocalRepository() *LocalRepository {
	return &LocalRepository{
		storage: make(map[string]string),
	}
}

func (r *LocalRepository) Save(ctx context.Context, shortURL, originalURL string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.storage[shortURL] = originalURL
	return nil
}

func (r *LocalRepository) Get(ctx context.Context, shortURL string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	url, exists := r.storage[shortURL]
	if !exists {
		return "", fmt.Errorf("URL not found")
	}
	return url, nil
}

func (r *LocalRepository) GetByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for shortURL, origURL := range r.storage {
		if origURL == originalURL {
			return shortURL, nil
		}
	}
	return "", fmt.Errorf("URL not found")
}

func (r *LocalRepository) SaveBatch(ctx context.Context, shortURLs, originalURLs []string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i := range shortURLs {
		r.storage[shortURLs[i]] = originalURLs[i]
	}

	return nil
}

func (r *LocalRepository) Ping(ctx context.Context) error {
	return nil
}
