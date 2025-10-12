package repository

import (
	"fmt"
)

type LocalRepository struct {
	storage map[string]string
}

func NewLocalRepository() *LocalRepository {
	return &LocalRepository{
		storage: make(map[string]string),
	}
}

func (r *LocalRepository) Save(shortURL, originalURL string) error {
	r.storage[shortURL] = originalURL
	return nil
}

func (r *LocalRepository) Get(shortURL string) (string, error) {
	url, exists := r.storage[shortURL]
	if !exists {
		return "", fmt.Errorf("URL not found")
	}
	return url, nil
}
