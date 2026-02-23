package audit

import (
	"sync"
	"time"
)

type AuditEvent struct {
	TS     int64  `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id"`
	URL    string `json:"url"`
}

type Observer interface {
	OnAudit(event AuditEvent)
}

type Publisher struct {
	observers map[string]Observer
	mu        sync.RWMutex
}

func NewPublisher() *Publisher {
	return &Publisher{
		observers: make(map[string]Observer),
	}
}

func (p *Publisher) Register(id string, o Observer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.observers[id] = o
}

func (p *Publisher) Deregister(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.observers, id)
}

func (p *Publisher) Notify(event AuditEvent) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, observer := range p.observers {
		observer.OnAudit(event)
	}
}

func NewAuditEvent(action, userID, url string) AuditEvent {
	return AuditEvent{
		TS:     time.Now().Unix(),
		Action: action,
		UserID: userID,
		URL:    url,
	}
}
