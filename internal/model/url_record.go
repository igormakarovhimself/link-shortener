// Package model содержит типы данных, которые используются в разных слоях сервиса.
package model

// URLRecord — запись о сокращенном URL.
type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	IsDeleted   bool   `json:"is_deleted"`
}

// URLPair — пара из короткого и оригинального URL, используется в ответах API.
type URLPair struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
