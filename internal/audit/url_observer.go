package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

// URLObserver отправляет аудит-события POST-запросом на внешний URL.
type URLObserver struct {
	url    string
	client *http.Client
}

// NewURLObserver создает URLObserver, который шлет события на url.
// Если url пустой — возвращает nil без ошибки.
func NewURLObserver(url string) (*URLObserver, error) {
	if url == "" {
		return nil, nil
	}

	return &URLObserver{
		url: url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

// OnAudit сериализует событие в JSON и отправляет POST-запросом.
func (u *URLObserver) OnAudit(event AuditEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	resp, err := u.client.Post(u.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
