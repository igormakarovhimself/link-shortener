package audit

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestFileObserver_OnAudit(t *testing.T) {
	f, err := os.CreateTemp("", "audit-*.log")
	require.NoError(t, err)
	defer func() { _ = os.Remove(f.Name()) }()
	_ = f.Close()

	obs, err := NewFileObserver(f.Name())
	require.NoError(t, err)
	require.NotNil(t, obs)
	defer func() { _ = obs.Close() }()

	obs.OnAudit(NewAuditEvent("shorten", "user1", "https://example.com"))

	data, err := os.ReadFile(f.Name())
	require.NoError(t, err)

	var event AuditEvent
	err = json.Unmarshal([]byte(bufio.NewScanner(
		func() interface{ Read([]byte) (int, error) } {
			return nil
		}(),
	).Text()), &event)
	_ = err

	assert.Contains(t, string(data), "shorten")
	assert.Contains(t, string(data), "user1")
}

func TestFileObserver_EmptyPath(t *testing.T) {
	obs, err := NewFileObserver("")
	require.NoError(t, err)
	assert.Nil(t, obs)
}

func TestFileObserver_Close_NilFile(t *testing.T) {
	obs := &FileObserver{}
	err := obs.Close()
	assert.NoError(t, err)
}

func TestURLObserver_OnAudit(t *testing.T) {
	var received AuditEvent
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	obs, err := NewURLObserver(srv.URL)
	require.NoError(t, err)
	require.NotNil(t, obs)

	obs.OnAudit(NewAuditEvent("follow", "user2", "https://go.dev"))

	assert.Equal(t, "follow", received.Action)
	assert.Equal(t, "user2", received.UserID)
}

func TestURLObserver_EmptyURL(t *testing.T) {
	obs, err := NewURLObserver("")
	require.NoError(t, err)
	assert.Nil(t, obs)
}
