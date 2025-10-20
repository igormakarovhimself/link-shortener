package repository

import (
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

func (r *LocalRepository) Save(shortURL, originalURL string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.storage[shortURL] = originalURL
	return nil
}

func (r *LocalRepository) Get(shortURL string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	url, exists := r.storage[shortURL]
	if !exists {
		return "", fmt.Errorf("URL not found")
	}
	return url, nil
}
