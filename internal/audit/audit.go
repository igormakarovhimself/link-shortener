package audit

import (
	"sync"
	"time"
)

// AuditEvent — событие о действии пользователя.
type AuditEvent struct {
	TS     int64  `json:"ts"`      // unix-timestamp события
	Action string `json:"action"`  // тип действия: shorten, follow
	UserID string `json:"user_id"` // идентификатор пользователя
	URL    string `json:"url"`     // URL, с которым связано действие
}

// Observer — интерфейс получателя аудит-событий.
type Observer interface {
	OnAudit(event AuditEvent)
}

// Publisher рассылает аудит-события зарегистрированным наблюдателям.
type Publisher struct {
	observers map[string]Observer
	mu        sync.RWMutex
}

// NewPublisher создает новый Publisher.
func NewPublisher() *Publisher {
	return &Publisher{
		observers: make(map[string]Observer),
	}
}

// Register добавляет наблюдателя с указанным идентификатором.
func (p *Publisher) Register(id string, o Observer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.observers[id] = o
}

// Deregister удаляет наблюдателя по идентификатору.
func (p *Publisher) Deregister(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.observers, id)
}

// Notify отправляет событие всем зарегистрированным наблюдателям.
func (p *Publisher) Notify(event AuditEvent) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, observer := range p.observers {
		observer.OnAudit(event)
	}
}

// NewAuditEvent создает событие аудита.
func NewAuditEvent(action, userID, url string) AuditEvent {
	return AuditEvent{
		TS:     time.Now().Unix(),
		Action: action,
		UserID: userID,
		URL:    url,
	}
}
