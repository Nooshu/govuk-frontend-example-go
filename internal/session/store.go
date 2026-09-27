package session

// Session values and the in-memory Store implementation.

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"

	"github.com/Nooshu/govuk-frontend-example-go/internal/service"
)

// Cookie banner and cookie settings choices.
const (
	ChoiceAccept = "accept"
	ChoiceReject = "reject"
)

// Errors carries validation failures from the POST that produced them to the GET that shows
// them, so a failed submit can redirect instead of re-rendering the page.
type Errors struct {
	Path  string
	Items []service.FieldError
}

// Notice is a one-off confirmation message shown on the next view of Path.
type Notice struct {
	Path string
	Text string
}

// Session is one applicant's state.
type Session struct {
	// ID identifies the session and is the value of the session cookie.
	ID string
	// CSRF is the token every form must post back.
	CSRF string
	// Application is the answers given so far.
	Application service.Application
	// CookieChoice is the saved analytics cookie decision, or empty when none has been made.
	CookieChoice string
	// CookieBanner is the decision to confirm in the banner, or empty when nothing to confirm.
	CookieBanner string
	// Errors are validation failures waiting to be shown once.
	Errors *Errors
	// Notice is a confirmation message waiting to be shown once.
	Notice *Notice
}

// Store creates, reads, and replaces sessions.
type Store interface {
	Create() *Session
	Get(id string) (*Session, bool)
	Save(s *Session)
}

// New returns a session with a fresh id, CSRF token, and empty application.
//
// It is not stored until a [Store] saves it.
func New() *Session {
	return &Session{
		ID:          randomHex(16),
		CSRF:        randomHex(16),
		Application: service.NewApplication(),
	}
}

// MemoryStore keeps sessions in memory for this process. It is safe for concurrent use.
type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewMemoryStore returns an empty in-memory store. Sessions disappear when the process stops.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{sessions: make(map[string]*Session)}
}

// Create stores and returns a new session.
func (s *MemoryStore) Create() *Session {
	created := New()
	s.Save(created)
	return created
}

// Get returns the stored session for id.
func (s *MemoryStore) Get(id string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	found, ok := s.sessions[id]
	return found, ok
}

// Save stores a session under its id, replacing any earlier copy.
func (s *MemoryStore) Save(session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
}

// ReferenceFor builds the confirmation reference shown to the applicant.
//
// It is derived from the session id so the same application always quotes the same reference.
func ReferenceFor(sessionID string) string {
	prefix := sessionID
	if len(prefix) > 6 {
		prefix = prefix[:6]
	}
	return "RL" + strings.ToUpper(prefix)
}

func randomHex(size int) string {
	bytes := make([]byte, size)
	// crypto/rand.Read never fails on the platforms Go supports; it panics internally instead.
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
