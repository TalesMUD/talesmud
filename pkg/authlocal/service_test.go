package authlocal

import (
	"path/filepath"
	"strings"
	"testing"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

func TestRegisterLoginResetPersistsHashOnly(t *testing.T) {
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })

	users := service.NewUsersService(repository.NewSQLiteFactory(client).Users())
	mailer := &CaptureMailer{}
	svc := New(client.DB(), users, "test-session-secret", mailer)

	const password = "s3cret-horse-battery"
	token, pub, err := svc.Register("Mara_Quinn", "Mara@Example.com", password)
	if err != nil {
		t.Fatal(err)
	}
	if pub.Username != "mara_quinn" || pub.Email != "mara@example.com" {
		t.Fatalf("public user = %+v", pub)
	}
	if _, _, err := svc.Register("mara_quinn", "other@example.com", password); !errorsIs(err, ErrExists) {
		t.Fatalf("duplicate username err = %v", err)
	}

	stored, err := users.FindByUsername("mara_quinn")
	if err != nil {
		t.Fatal(err)
	}
	if stored.PasswordHash == "" || strings.Contains(stored.PasswordHash, password) || !strings.HasPrefix(stored.PasswordHash, "$argon2id$") {
		t.Fatalf("password not stored as argon2id hash: %q", stored.PasswordHash)
	}

	var raw string
	if err := client.DB().QueryRow(`SELECT data FROM users WHERE id = ?`, stored.ID).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(raw, password) {
		t.Fatal("sqlite payload contains the plaintext password")
	}

	if _, err := svc.UserFromToken(token); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Login("mara_quinn", "wrong-password-1"); !errorsIs(err, ErrCredentials) {
		t.Fatalf("bad password err = %v", err)
	}
	if _, _, err := svc.Login("nobody_user", password); !errorsIs(err, ErrCredentials) {
		t.Fatalf("missing user err = %v", err)
	}

	if err := svc.Forgot("missing@example.com"); err != nil {
		t.Fatal(err)
	}
	if mailer.LastToken != "" {
		t.Fatal("unknown email produced a reset token")
	}
	if err := svc.Forgot("mara@example.com"); err != nil {
		t.Fatal(err)
	}
	if mailer.LastToken == "" || mailer.LastTo != "mara@example.com" {
		t.Fatalf("mailer = %+v", mailer)
	}
	var tokenRow string
	if err := client.DB().QueryRow(`SELECT data FROM auth_tokens`).Scan(&tokenRow); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(tokenRow, mailer.LastToken) {
		t.Fatal("reset token stored in plaintext")
	}

	if err := svc.Reset(mailer.LastToken, "new-password-22"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UserFromToken(token); err == nil {
		t.Fatal("old session survived password reset")
	}
	if _, _, err := svc.Login("mara_quinn", "new-password-22"); err != nil {
		t.Fatal(err)
	}
	if err := svc.Reset(mailer.LastToken, "another-password"); !errorsIs(err, ErrToken) {
		t.Fatalf("reused token err = %v", err)
	}
}

func errorsIs(err, target error) bool {
	return err != nil && (err == target || strings.Contains(err.Error(), target.Error()))
}
