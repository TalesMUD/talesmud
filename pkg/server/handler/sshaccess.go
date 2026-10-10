package handler

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/talesmud/talesmud/pkg/devicecode"
	e "github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/presentation/textline"
	"github.com/talesmud/talesmud/pkg/sshgate"
	"github.com/talesmud/talesmud/pkg/sshkeys"
)

// SSHSettings is the process SSH API. Zero values keep every route closed.
type SSHSettings struct {
	Keys          *sshkeys.Store
	Devices       *devicecode.Store
	ActivateURL   string
	KeysEnabled   bool
	DeviceEnabled bool
	WebManage     bool
	MaxKeys       int
}

// SSHAPI serves the account key and device-code routes.
// The listener fills it after the routes are registered.
type SSHAPI struct {
	mu          sync.RWMutex
	on          bool
	keys        *sshkeys.Store
	devices     *devicecode.Store
	activateURL string
	keysOn      bool
	deviceOn    bool
	web         bool
	maxKeys     int
	mutate      *ipLimiter
}

// NewSSHAPI is closed until Set is called.
func NewSSHAPI() *SSHAPI {
	return &SSHAPI{mutate: newIPLimiter(20, 10*time.Minute)}
}

// Set installs the stores for a running listener.
func (a *SSHAPI) Set(cfg SSHSettings) {
	if a == nil {
		return
	}
	a.mu.Lock()
	a.on = true
	a.keys = cfg.Keys
	a.devices = cfg.Devices
	a.activateURL = strings.TrimSpace(cfg.ActivateURL)
	a.keysOn = cfg.KeysEnabled && cfg.Keys != nil
	a.deviceOn = cfg.DeviceEnabled && cfg.Devices != nil
	a.web = cfg.WebManage
	a.maxKeys = cfg.MaxKeys
	a.mu.Unlock()
}

type sshKeyBody struct {
	PublicKey string `json:"public_key"`
	Label     string `json:"label"`
}

type sshCodeBody struct {
	UserCode string `json:"user_code"`
	CSRF     string `json:"csrf"`
}

type sshKeyView struct {
	ID          string    `json:"id"`
	Fingerprint string    `json:"fingerprint"`
	KeyType     string    `json:"key_type"`
	Label       string    `json:"label"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedVia  string    `json:"created_via"`
	LastUsedAt  time.Time `json:"last_used_at,omitempty"`
}

// ListKeys returns the signed-in account's linked keys.
func (a *SSHAPI) ListKeys(c *gin.Context) {
	store, _, _, ok := a.keyStore()
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	user, ok := accountUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	if refuseAccount(c, user) {
		return
	}
	rows, err := store.List(user.RefID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unavailable"})
		return
	}
	out := make([]sshKeyView, 0, len(rows))
	for _, row := range rows {
		out = append(out, sshKeyView{
			ID:          row.ID,
			Fingerprint: row.Fingerprint,
			KeyType:     row.KeyType,
			Label:       row.Label,
			CreatedAt:   row.CreatedAt,
			CreatedVia:  row.CreatedVia,
			LastUsedAt:  row.LastUsedAt,
		})
	}
	c.JSON(http.StatusOK, out)
}

// AddKey links a public key the signed-in account pasted.
// Possession of the private key is not required: the account owner is already signed in.
func (a *SSHAPI) AddKey(c *gin.Context) {
	store, max, web, ok := a.keyStore()
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if !web {
		c.JSON(http.StatusForbidden, gin.H{"error": "key management is disabled"})
		return
	}
	user, ok := accountUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	if refuseAccount(c, user) {
		return
	}
	if a.mutate != nil && !a.mutate.allow(user.RefID) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many attempts"})
		return
	}
	var body sshKeyBody
	if !bindSSH(c, &body) {
		return
	}
	row, err := store.Add(user.RefID, body.PublicKey, body.Label, "web", max)
	if err != nil {
		writeKeyErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, sshKeyView{
		ID:          row.ID,
		Fingerprint: row.Fingerprint,
		KeyType:     row.KeyType,
		Label:       row.Label,
		CreatedAt:   row.CreatedAt,
		CreatedVia:  row.CreatedVia,
	})
}

// DeleteKey revokes one key owned by the signed-in account.
func (a *SSHAPI) DeleteKey(c *gin.Context) {
	store, _, web, ok := a.keyStore()
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if !web {
		c.JSON(http.StatusForbidden, gin.H{"error": "key management is disabled"})
		return
	}
	user, ok := accountUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	if refuseAccount(c, user) {
		return
	}
	if err := store.Delete(user.RefID, c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

// Lookup describes a pending SSH sign-in. The response never contains the raw code.
func (a *SSHAPI) Lookup(c *gin.Context) {
	a.device(c, func(store *devicecode.Store, user *e.User, body sshCodeBody) {
		view, err := store.Lookup(body.UserCode, user.RefID, c.ClientIP(), allowUser(user))
		if err != nil {
			writeDeviceErr(c, err)
			return
		}
		var key any
		if view.KeyFP != "" {
			key = sshgate.MaskFingerprint(view.KeyFP)
		}
		c.JSON(http.StatusOK, gin.H{
			"ip":             view.IP,
			"created":        view.Created,
			"mode":           view.Mode,
			"client_version": textline.Sanitize(view.ClientVersion),
			"key":            key,
			"csrf":           view.CSRF,
		})
	})
}

// Confirm binds the signed-in account to the waiting SSH lobby.
func (a *SSHAPI) Confirm(c *gin.Context) {
	a.device(c, func(store *devicecode.Store, user *e.User, body sshCodeBody) {
		if err := store.Confirm(body.UserCode, user.RefID, c.ClientIP(), body.CSRF, allowUser(user)); err != nil {
			writeDeviceErr(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}

// Deny rejects a pending SSH sign-in.
func (a *SSHAPI) Deny(c *gin.Context) {
	a.device(c, func(store *devicecode.Store, user *e.User, body sshCodeBody) {
		if err := store.Deny(body.UserCode, user.RefID, c.ClientIP(), body.CSRF, allowUser(user)); err != nil {
			writeDeviceErr(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}

func (a *SSHAPI) device(c *gin.Context, fn func(*devicecode.Store, *e.User, sshCodeBody)) {
	store, ok := a.deviceStore()
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if !a.originOK(c.Request) {
		c.JSON(http.StatusForbidden, gin.H{"error": "origin refused"})
		return
	}
	user, ok := accountUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	if refuseAccount(c, user) {
		return
	}
	var body sshCodeBody
	if !bindSSH(c, &body) {
		return
	}
	fn(store, user, body)
}

func (a *SSHAPI) keyStore() (*sshkeys.Store, int, bool, bool) {
	if a == nil {
		return nil, 0, false, false
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	if !a.on || !a.keysOn || a.keys == nil {
		return nil, 0, false, false
	}
	return a.keys, a.maxKeys, a.web, true
}

func (a *SSHAPI) deviceStore() (*devicecode.Store, bool) {
	if a == nil {
		return nil, false
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	if !a.on || !a.deviceOn || a.devices == nil {
		return nil, false
	}
	return a.devices, true
}

func (a *SSHAPI) originOK(r *http.Request) bool {
	if a == nil {
		return false
	}
	a.mu.RLock()
	raw := a.activateURL
	a.mu.RUnlock()
	return originAllowed(r, raw)
}

func originAllowed(r *http.Request, activateURL string) bool {
	if r == nil {
		return false
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	referer := strings.TrimSpace(r.Header.Get("Referer"))
	if origin == "" && referer == "" {
		return true
	}
	want, err := url.Parse(activateURL)
	if err != nil || want.Scheme == "" || want.Host == "" {
		return false
	}
	allowed := strings.ToLower(want.Scheme + "://" + want.Host)
	if origin != "" {
		return strings.EqualFold(strings.TrimRight(origin, "/"), allowed)
	}
	ref, err := url.Parse(referer)
	if err != nil || ref.Scheme == "" || ref.Host == "" {
		return false
	}
	return strings.EqualFold(ref.Scheme+"://"+ref.Host, allowed)
}

func accountUser(c *gin.Context) (*e.User, bool) {
	v, ok := c.Get("user")
	if !ok || v == nil {
		return nil, false
	}
	user, ok := v.(*e.User)
	if !ok || user == nil || strings.TrimSpace(user.RefID) == "" {
		return nil, false
	}
	return user, true
}

func refuseAccount(c *gin.Context, user *e.User) bool {
	if user == nil || user.IsGuest || user.IsBanned {
		c.JSON(http.StatusForbidden, gin.H{"error": "account cannot confirm"})
		return true
	}
	return false
}

func allowUser(user *e.User) func(string) error {
	return func(ref string) error {
		if user == nil || user.IsGuest || user.IsBanned || ref != user.RefID {
			return devicecode.ErrForbidden
		}
		return nil
	}
}

func bindSSH(c *gin.Context, dst any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	if err := c.ShouldBindJSON(dst); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return false
	}
	return true
}

func writeKeyErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, sshkeys.ErrExists):
		c.JSON(http.StatusConflict, gin.H{"error": "key already linked"})
	case errors.Is(err, sshkeys.ErrFull):
		c.JSON(http.StatusConflict, gin.H{"error": "too many keys"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "key rejected"})
	}
}

func writeDeviceErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, devicecode.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "account cannot confirm"})
	case errors.Is(err, devicecode.ErrLimited):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many attempts"})
	default:
		c.JSON(http.StatusNotFound, gin.H{"error": "code not found or expired"})
	}
}
