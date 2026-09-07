package server

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	e "github.com/talesmud/talesmud/pkg/entities"
)

const wsTicketTTL = 60 * time.Second

type ticketEntry struct {
	userID string
	exp    time.Time
}

// TicketStore issues single-use, short-lived WebSocket tickets so browsers
// never put a session JWT in the WebSocket URL.
type TicketStore struct {
	mu      sync.Mutex
	tickets map[string]ticketEntry
}

func NewTicketStore() *TicketStore {
	return &TicketStore{tickets: make(map[string]ticketEntry)}
}

func (s *TicketStore) Issue(user *e.User) (string, time.Duration, error) {
	if s == nil || user == nil {
		return "", 0, errNoTicketUser
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", 0, err
	}
	id := hex.EncodeToString(b)
	now := time.Now()
	s.mu.Lock()
	s.pruneLocked(now)
	s.tickets[id] = ticketEntry{userID: user.ID, exp: now.Add(wsTicketTTL)}
	s.mu.Unlock()
	return id, wsTicketTTL, nil
}

func (s *TicketStore) Consume(id string) (userID string, ok bool) {
	if s == nil || id == "" {
		return "", false
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.tickets[id]
	if !exists {
		return "", false
	}
	delete(s.tickets, id)
	if now.After(entry.exp) {
		return "", false
	}
	return entry.userID, true
}

func (s *TicketStore) pruneLocked(now time.Time) {
	for id, entry := range s.tickets {
		if now.After(entry.exp) {
			delete(s.tickets, id)
		}
	}
}

type ticketError string

func (e ticketError) Error() string { return string(e) }

const errNoTicketUser ticketError = "user required to issue websocket ticket"

// IssueWSTicket handles POST /api/ws-ticket.
func (s *TicketStore) IssueWSTicket(c *gin.Context) {
	usr, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	user, ok := usr.(*e.User)
	if !ok || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	ticket, ttl, err := s.Issue(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue ticket"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ticket":    ticket,
		"expiresIn": int(ttl.Seconds()),
	})
}
