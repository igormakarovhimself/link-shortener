package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"link-shortener/internal/model"
)

type urlFileData struct {
	originalURL string
	userID      string
	isDeleted   bool
}

// FileRepository хранит URL в в файле в JSON-формате.
type FileRepository struct {
	file    *os.File
	encoder *json.Encoder
	storage map[string]urlFileData
	counter int
	mutex   sync.RWMutex
}

// NewFileRepository открывает файл по пути filepath и загружает из него существующие URL.
func NewFileRepository(filepath string) (*FileRepository, error) {
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	repo := &FileRepository{
		file:    file,
		encoder: json.NewEncoder(file),
		storage: make(map[string]urlFileData),
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

		r.storage[record.ShortURL] = urlFileData{
			originalURL: record.OriginalURL,
			userID:      record.UserID,
		}

		if uuid, err := strconv.Atoi(record.UUID); err == nil {
			if uuid > r.counter {
				r.counter = uuid
			}
		}
	}

	return nil
}

func (r *FileRepository) Save(ctx context.Context, shortURL, originalURL, userID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.counter++
	record := &model.URLRecord{
		UUID:        strconv.Itoa(r.counter),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
	}

	if err := r.encoder.Encode(record); err != nil {
		return err
	}

	r.storage[shortURL] = urlFileData{
		originalURL: originalURL,
		userID:      userID,
	}

	return nil
}

func (r *FileRepository) Get(ctx context.Context, shortURL string) (string, error) {
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

func (r *FileRepository) GetByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for shortURL, data := range r.storage {
		if data.originalURL == originalURL {
			return shortURL, nil
		}
	}
	return "", fmt.Errorf("URL not found")
}

func (r *FileRepository) Close() error {
	return r.file.Close()
}

func (r *FileRepository) SaveBatch(ctx context.Context, shortURLs, originalURLs []string, userID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i := range shortURLs {
		r.counter++
		record := &model.URLRecord{
			UUID:        strconv.Itoa(r.counter),
			ShortURL:    shortURLs[i],
			OriginalURL: originalURLs[i],
			UserID:      userID,
		}

		if err := r.encoder.Encode(record); err != nil {
			return err
		}

		r.storage[shortURLs[i]] = urlFileData{
			originalURL: originalURLs[i],
			userID:      userID,
		}
	}

	return nil
}

func (r *FileRepository) GetURLsByUserID(ctx context.Context, userID string) ([]model.URLPair, error) {
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

func (r *FileRepository) DeleteURLs(ctx context.Context, shortURLs []string, userID string) error {
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

func (r *FileRepository) Ping(ctx context.Context) error {
	return nil
}
