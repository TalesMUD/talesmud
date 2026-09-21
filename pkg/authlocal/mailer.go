package authlocal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Mailer delivers one-time password reset tokens. Implementations must not
// log the token at info level. The HTTP API never returns the raw token.
type Mailer interface {
	SendPasswordReset(to, rawToken string) error
}

// OutboxMailer appends reset tokens to a local file when SMTP is not configured.
// This is the P0 stand-in for email so operators can complete a reset in dev
// without the API echoing the secret.
type OutboxMailer struct {
	Path string
}

func (m OutboxMailer) SendPasswordReset(to, rawToken string) error {
	path := m.Path
	if path == "" {
		path = "data/door-outbox.log"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s\treset\t%s\t%s\n", time.Now().UTC().Format(time.RFC3339), to, rawToken)
	return err
}

// CaptureMailer records the last reset token for tests.
type CaptureMailer struct {
	LastTo    string
	LastToken string
	Err       error
}

func (m *CaptureMailer) SendPasswordReset(to, rawToken string) error {
	if m.Err != nil {
		return m.Err
	}
	m.LastTo = to
	m.LastToken = rawToken
	return nil
}
