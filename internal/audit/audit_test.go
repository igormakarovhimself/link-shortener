package audit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockObserver struct {
	events []AuditEvent
}

func (m *mockObserver) OnAudit(event AuditEvent) {
	m.events = append(m.events, event)
}

func TestPublisher_RegisterAndNotify(t *testing.T) {
	pub := NewPublisher()
	obs := &mockObserver{}

	pub.Register("obs1", obs)
	event := NewAuditEvent("shorten", "user1", "https://example.com/")
	pub.Notify(event)

	assert.Len(t, obs.events, 1)
	assert.Equal(t, "shorten", obs.events[0].Action)
	assert.Equal(t, "user1", obs.events[0].UserID)
}

func TestPublisher_Deregister(t *testing.T) {
	pub := NewPublisher()
	obs := &mockObserver{}

	pub.Register("obs1", obs)
	pub.Deregister("obs1")

	pub.Notify(NewAuditEvent("follow", "user1", "https://example.com/"))
	assert.Empty(t, obs.events)
}

func TestPublisher_MultipleObservers(t *testing.T) {
	pub := NewPublisher()
	obs1 := &mockObserver{}
	obs2 := &mockObserver{}

	pub.Register("o1", obs1)
	pub.Register("o2", obs2)

	pub.Notify(NewAuditEvent("shorten", "u1", "https://a.com/"))

	assert.Len(t, obs1.events, 1)
	assert.Len(t, obs2.events, 1)
}

func TestNewAuditEvent(t *testing.T) {
	event := NewAuditEvent("shorten", "user42", "https://test.com/")
	assert.Equal(t, "shorten", event.Action)
	assert.Equal(t, "user42", event.UserID)
	assert.Equal(t, "https://test.com/", event.URL)
	assert.NotZero(t, event.TS)
}
