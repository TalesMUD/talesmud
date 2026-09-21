package handler

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/authlocal"
	e "github.com/talesmud/talesmud/pkg/entities"
)

// LocalAuthHandler exposes username/password registration for door mode.
type LocalAuthHandler struct {
	Auth *authlocal.Service
}

type registerBody struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type forgotBody struct {
	Email string `json:"email"`
}

type resetBody struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

var authLimiter = newIPLimiter(12, time.Minute)

// Register creates an account and session. The response never includes a password hash.
func (h *LocalAuthHandler) Register(c *gin.Context) {
	if !authLimiter.allow(c.ClientIP()) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many attempts"})
		return
	}
	var body registerBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	token, user, err := h.Auth.Register(body.Username, body.Email, body.Password)
	if err != nil {
		writeAuthErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"token": token, "user": user})
}

// Login verifies a password and returns a session token.
func (h *LocalAuthHandler) Login(c *gin.Context) {
	if !authLimiter.allow(c.ClientIP()) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many attempts"})
		return
	}
	var body loginBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	token, user, err := h.Auth.Login(body.Username, body.Password)
	if err != nil {
		writeAuthErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

// Forgot accepts an email and, when the account exists, emails a one-time token.
// The HTTP body never contains the token.
func (h *LocalAuthHandler) Forgot(c *gin.Context) {
	if !authLimiter.allow(c.ClientIP()) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many attempts"})
		return
	}
	var body forgotBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.Auth.Forgot(body.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not send reset token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Reset consumes a one-time token and sets a new password.
func (h *LocalAuthHandler) Reset(c *gin.Context) {
	if !authLimiter.allow(c.ClientIP()) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many attempts"})
		return
	}
	var body resetBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.Auth.Reset(body.Token, body.Password); err != nil {
		writeAuthErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Me returns the authenticated user without secrets.
func (h *LocalAuthHandler) Me(c *gin.Context) {
	usr, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	user, ok := usr.(*e.User)
	if !ok || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": authlocal.Public(user)})
}

func writeAuthErr(c *gin.Context, err error) {
	switch err {
	case authlocal.ErrValidation, authlocal.ErrToken:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case authlocal.ErrExists:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case authlocal.ErrCredentials:
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case authlocal.ErrBanned:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

type ipLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	hits   map[string][]time.Time
}

func newIPLimiter(limit int, window time.Duration) *ipLimiter {
	return &ipLimiter{limit: limit, window: window, hits: map[string][]time.Time{}}
}

func (l *ipLimiter) allow(ip string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	cut := now.Add(-l.window)
	kept := l.hits[ip][:0]
	for _, ts := range l.hits[ip] {
		if ts.After(cut) {
			kept = append(kept, ts)
		}
	}
	if len(kept) >= l.limit {
		l.hits[ip] = kept
		return false
	}
	l.hits[ip] = append(kept, now)
	return true
}
