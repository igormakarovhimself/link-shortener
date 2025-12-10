package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"link-shortener/internal/model"
	"os"
	"strconv"
	"sync"
)

type FileRepository struct {
	file    *os.File
	encoder *json.Encoder
	storage map[string]string
	counter int
	mutex   sync.RWMutex
}

func NewFileRepository(filepath string) (*FileRepository, error) {
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	repo := &FileRepository{
		file:    file,
		encoder: json.NewEncoder(file),
		storage: make(map[string]string),
		counter: 0,
	}

	if err := repo.loadData(); err != nil {
		file.Close()
		return nil, err
	}

	return repo, nil
}

func (r *FileRepository) loadData() error {
	decoder := json.NewDecoder(r.file)

	for {
		record := &model.URLRecord{}
		if err := decoder.Decode(record); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		r.storage[record.ShortURL] = record.OriginalURL

		if uuid, err := strconv.Atoi(record.UUID); err == nil {
			if uuid > r.counter {
				r.counter = uuid
			}
		}
	}

	return nil
}

func (r *FileRepository) Save(ctx context.Context, shortURL, originalURL string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.counter++
	record := &model.URLRecord{
		UUID:        strconv.Itoa(r.counter),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	if err := r.encoder.Encode(record); err != nil {
		return err
	}

	r.storage[shortURL] = originalURL

	return nil
}

func (r *FileRepository) Get(ctx context.Context, shortURL string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	url, exists := r.storage[shortURL]
	if !exists {
		return "", fmt.Errorf("URL not found")
	}

	return url, nil
}

func (r *FileRepository) GetByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for shortURL, origURL := range r.storage {
		if origURL == originalURL {
			return shortURL, nil
		}
	}
	return "", fmt.Errorf("URL not found")
}

func (r *FileRepository) Close() error {
	return r.file.Close()
}

func (r *FileRepository) SaveBatch(ctx context.Context, shortURLs, originalURLs []string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i := range shortURLs {
		r.counter++
		record := &model.URLRecord{
			UUID:        strconv.Itoa(r.counter),
			ShortURL:    shortURLs[i],
			OriginalURL: originalURLs[i],
		}

		if err := r.encoder.Encode(record); err != nil {
			return err
		}

		r.storage[shortURLs[i]] = originalURLs[i]
	}

	return nil
}

func (r *FileRepository) Ping(ctx context.Context) error {
	return nil
}
