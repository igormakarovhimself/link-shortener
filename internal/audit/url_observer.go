package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type URLObserver struct {
	url string
}

func NewURLObserver(url string) (*URLObserver, error) {
	if url == "" {
		return nil, nil
	}

	return &URLObserver{
		url: url,
	}, nil
}

func (u *URLObserver) OnAudit(event AuditEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	resp, err := http.Post(u.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
