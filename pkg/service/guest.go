package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	mathrand "math/rand"
	"os"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	e "github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
)

const (
	GuestSessionDuration = 30 * time.Minute
	GuestMaxLevel        = 5
	GuestCleanupInterval = 5 * time.Minute
	GuestRateLimitPerIP  = 10 // max guest creations per IP per hour
)

// GuestService handles temporary guest account creation, token management, and cleanup.
type GuestService interface {
	CreateGuestSession(remoteIP string) (token string, err error)
	ValidateGuestToken(tokenStr string) (userID string, err error)
	CleanupExpiredGuests()
	StartCleanupLoop()
}

type guestService struct {
	facade     Facade
	secret     []byte
	rateLimits map[string][]time.Time
	mu         sync.Mutex
}

// NewGuestService creates a new guest service. Reads GUEST_SECRET from env.
func NewGuestService(facade Facade) GuestService {
	secret := os.Getenv("GUEST_SECRET")
	if secret == "" {
		b := make([]byte, 32)
		rand.Read(b)
		secret = hex.EncodeToString(b)
		log.Warn("GUEST_SECRET not set, generated random secret (guest tokens won't survive server restart)")
	}
	return &guestService{
		facade:     facade,
		secret:     []byte(secret),
		rateLimits: make(map[string][]time.Time),
	}
}

var guestNamePrefixes = []string{
	"Wanderer", "Traveler", "Stranger", "Drifter", "Nomad",
	"Pilgrim", "Vagrant", "Rover", "Wayfarer", "Outsider",
}

func generateGuestName() string {
	prefix := guestNamePrefixes[mathrand.Intn(len(guestNamePrefixes))]
	suffix := make([]byte, 2)
	rand.Read(suffix)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(suffix))
}

func generateGuestRefID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return "guest:" + hex.EncodeToString(b)
}

// CreateGuestSession creates a temporary guest user + character and returns a signed JWT.
func (gs *guestService) CreateGuestSession(remoteIP string) (string, error) {
	// Check server settings
	settings, err := gs.facade.ServerSettingsService().Get()
	if err != nil {
		return "", errors.New("could not load server settings")
	}
	if !settings.GuestsAllowed {
		return "", errors.New("guest mode is disabled")
	}

	// Check max concurrent guests
	if settings.MaxGuestAccounts > 0 {
		count := gs.countActiveGuests()
		if count >= settings.MaxGuestAccounts {
			return "", errors.New("server is full, try again later")
		}
	}

	// Check rate limit
	if !gs.checkRateLimit(remoteIP) {
		return "", errors.New("rate limit exceeded")
	}

	// Generate guest identity
	guestName := generateGuestName()
	refID := generateGuestRefID()

	// Ensure name uniqueness (retry up to 5 times)
	for attempt := 0; attempt < 5; attempt++ {
		if !gs.facade.CharactersService().IsCharacterNameTaken(guestName) {
			break
		}
		guestName = generateGuestName()
		if attempt == 4 {
			return "", errors.New("could not generate unique guest name")
		}
	}

	// Create guest user
	now := time.Now()
	user := &e.User{
		Entity:         e.NewEntity(),
		RefID:          refID,
		Nickname:       guestName,
		Created:        now,
		LastSeen:       now,
		IsNewUser:      false,
		Role:           e.RolePlayer,
		IsGuest:        true,
		GuestExpiresAt: now.Add(GuestSessionDuration),
	}

	user, err = gs.facade.UsersService().Create(user)
	if err != nil {
		return "", fmt.Errorf("could not create guest user: %v", err)
	}

	// Pick a random class template
	templates := characters.SystemCharacterTemplatePresets()
	template := templates[mathrand.Intn(len(templates))]

	// Create character from template
	character := &characters.Character{
		Entity:           e.NewEntity(),
		Name:             guestName,
		Description:      "A mysterious guest adventurer.",
		Race:             template.Race,
		Class:            template.Class,
		CurrentHitPoints: template.CurrentHitPoints,
		MaxHitPoints:     template.MaxHitPoints,
		CurrentMana:      template.CurrentMana,
		MaxMana:          template.MaxMana,
		XP:               0,
		Level:            template.Level,
		Attributes:       template.Attributes,
		MaxLevelCap:      GuestMaxLevel,
		Created:          now,
	}
	character.BelongsUserID = user.ID

	if len(template.DefaultSkills) > 0 {
		character.EquippedSkills = make([]string, len(template.DefaultSkills))
		copy(character.EquippedSkills, template.DefaultSkills)
	}

	// Equip starter items from template
	if len(template.StartingItems) > 0 {
		character.EquippedItems = make(map[items.ItemSlot]*items.Item)
		for _, si := range template.StartingItems {
			itemTemplate := items.StarterItemTemplateByName(si.ItemTemplateName)
			if itemTemplate != nil {
				// Create a copy of the item for this character
				itemCopy := *itemTemplate
				itemCopy.Entity = e.NewEntity()
				itemCopy.IsTemplate = false
				if si.Slot != "" {
					character.EquippedItems[si.Slot] = &itemCopy
				}
			}
		}
	}

	// Store character (bypasses name-taken check since we already verified above)
	storedChar, err := gs.facade.CharactersService().Store(character)
	if err != nil {
		// Clean up user on failure
		gs.facade.UsersService().Delete(user.ID)
		return "", fmt.Errorf("could not create guest character: %v", err)
	}

	// Link character to user
	user.LastCharacter = storedChar.ID
	gs.facade.UsersService().Update(user.RefID, user)

	// Sign JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   refID,
		"uid":   user.ID,
		"exp":   now.Add(GuestSessionDuration).Unix(),
		"iat":   now.Unix(),
		"guest": true,
	})

	tokenStr, err := token.SignedString(gs.secret)
	if err != nil {
		return "", fmt.Errorf("could not sign guest token: %v", err)
	}

	log.WithFields(log.Fields{
		"guestName": guestName,
		"userID":    user.ID,
		"class":     template.Class.Name,
		"expiresAt": user.GuestExpiresAt.Format(time.RFC3339),
	}).Info("Guest session created")

	return tokenStr, nil
}

// ValidateGuestToken validates a guest JWT and returns the user entity ID.
func (gs *guestService) ValidateGuestToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return gs.secret, nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token claims")
	}

	// Verify guest claim
	isGuest, ok := claims["guest"].(bool)
	if !ok || !isGuest {
		return "", errors.New("not a guest token")
	}

	// Extract user ID
	uid, ok := claims["uid"].(string)
	if !ok || uid == "" {
		return "", errors.New("missing user ID in token")
	}

	return uid, nil
}

// CleanupExpiredGuests removes expired guest users and their characters.
func (gs *guestService) CleanupExpiredGuests() {
	users, err := gs.facade.UsersService().FindAll()
	if err != nil {
		log.WithError(err).Error("Guest cleanup: could not load users")
		return
	}

	now := time.Now()
	cleaned := 0

	for _, user := range users {
		if !user.IsGuest {
			continue
		}
		if user.GuestExpiresAt.IsZero() || now.Before(user.GuestExpiresAt) {
			continue
		}

		// Delete all characters for this guest
		if chars, err := gs.facade.CharactersService().FindAllForUser(user.ID); err == nil {
			for _, ch := range chars {
				gs.facade.CharactersService().Delete(ch.ID)
			}
		}

		// Delete the guest user
		gs.facade.UsersService().Delete(user.ID)
		cleaned++
	}

	if cleaned > 0 {
		log.WithField("count", cleaned).Info("Guest cleanup: removed expired guest accounts")
	}

	// Prune old rate limit entries
	gs.pruneRateLimits()
}

// StartCleanupLoop starts a background goroutine that periodically cleans up expired guests.
func (gs *guestService) StartCleanupLoop() {
	go func() {
		ticker := time.NewTicker(GuestCleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			gs.CleanupExpiredGuests()
		}
	}()
	log.Info("Guest cleanup loop started")
}

// countActiveGuests counts the number of currently active (non-expired) guest users.
func (gs *guestService) countActiveGuests() int {
	users, err := gs.facade.UsersService().FindAll()
	if err != nil {
		return 0
	}

	now := time.Now()
	count := 0
	for _, user := range users {
		if user.IsGuest && !user.GuestExpiresAt.IsZero() && now.Before(user.GuestExpiresAt) {
			count++
		}
	}
	return count
}

// checkRateLimit returns true if the IP is within the rate limit.
func (gs *guestService) checkRateLimit(ip string) bool {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Hour)

	// Filter to recent entries
	recent := []time.Time{}
	for _, t := range gs.rateLimits[ip] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	if len(recent) >= GuestRateLimitPerIP {
		return false
	}

	gs.rateLimits[ip] = append(recent, now)
	return true
}

// pruneRateLimits removes stale rate limit entries (older than 1 hour).
func (gs *guestService) pruneRateLimits() {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	cutoff := time.Now().Add(-1 * time.Hour)
	for ip, times := range gs.rateLimits {
		recent := []time.Time{}
		for _, t := range times {
			if t.After(cutoff) {
				recent = append(recent, t)
			}
		}
		if len(recent) == 0 {
			delete(gs.rateLimits, ip)
		} else {
			gs.rateLimits[ip] = recent
		}
	}
}
