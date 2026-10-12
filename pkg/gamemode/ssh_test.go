package gamemode

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSSHDefaultsOff(t *testing.T) {
	current = normalize(Config{})
	t.Cleanup(func() { current = normalize(Config{}) })
	cfg := Current()
	if cfg.SSH.Enabled || cfg.SSH.Keys.Enabled || cfg.SSH.Device.Enabled || cfg.SSH.Guest.Enabled {
		t.Fatalf("ssh must default off: %+v", cfg.SSH)
	}
	if cfg.Guests.PersistentEffects {
		t.Fatal("guest persistent effects must default off")
	}
	if !GuestEffectsAllowed(false) || GuestEffectsAllowed(true) {
		t.Fatal("guest effects gate")
	}
	if cfg.SSH.MaxConnections != 100 || cfg.SSH.AuthTimeout.Duration() != 30*time.Second {
		t.Fatalf("numeric defaults %+v", cfg.SSH)
	}
	if cfg.SSH.IdleTimeout.Duration() != 30*time.Minute || cfg.SSH.Guest.MaxSession.Duration() != 30*time.Minute {
		t.Fatalf("timeouts %+v guest %+v", cfg.SSH.IdleTimeout, cfg.SSH.Guest)
	}
	if !cfg.SSH.Keys.WebManageOn() || cfg.SSH.Mud.History != 20 || cfg.SSH.Door.CharsetDefault != "auto" {
		t.Fatalf("presentation defaults %+v", cfg.SSH)
	}
	if cfg.SSH.MaxSession.Duration() != 0 {
		t.Fatalf("account max session must stay unlimited, got %s", cfg.SSH.MaxSession.Duration())
	}
}

func TestApplyFileSSHAndGuestEffects(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ssh.yaml")
	body := []byte("presentation: classic\nauth: auth0\nssh:\n  enabled: true\n  listen: \":2232\"\n  auth_timeout: 12s\n  keys:\n    enabled: true\n    max_per_account: 4\n    web_manage: false\n  device:\n    ttl: 5m\n    activate_url: https://example.test/activate\n  guest:\n    max_session: 8m\nguests:\n  persistent_effects: true\n")
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ApplyFile(path); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { current = normalize(Config{}) })
	cfg := Current()
	if !cfg.SSH.Enabled || cfg.SSH.Listen != ":2232" || cfg.SSH.AuthTimeout.Duration() != 12*time.Second {
		t.Fatalf("%+v", cfg.SSH)
	}
	if !cfg.SSH.Keys.Enabled || cfg.SSH.Keys.MaxPerAccount != 4 || cfg.SSH.Keys.WebManageOn() {
		t.Fatalf("keys %+v", cfg.SSH.Keys)
	}
	if cfg.SSH.Device.TTL.Duration() != 5*time.Minute || cfg.SSH.Device.ActivateURL != "https://example.test/activate" {
		t.Fatalf("device %+v", cfg.SSH.Device)
	}
	if cfg.SSH.Guest.MaxSession.Duration() != 8*time.Minute || cfg.SSH.Guest.Enabled {
		t.Fatalf("guest %+v", cfg.SSH.Guest)
	}
	if !cfg.Guests.PersistentEffects || !GuestEffectsAllowed(true) {
		t.Fatal("persistent effects not loaded")
	}
}

func TestSSHEnvOverrides(t *testing.T) {
	current = normalize(Config{})
	t.Cleanup(func() {
		current = normalize(Config{})
	})
	t.Setenv("SSH_ENABLED", "true")
	t.Setenv("SSH_LISTEN", ":2232")
	t.Setenv("SSH_HOST_KEY_PATH", "/tmp/ssh-smoke/host")
	t.Setenv("SSH_PUBLIC_HOST", "mud.example")
	t.Setenv("SSH_PUBLIC_PORT", "2222")
	t.Setenv("SSH_ACTIVATE_URL", "https://mud.example/activate")
	t.Setenv("SSH_GUEST_ENABLED", "1")
	t.Setenv("SSH_DEVICE_ENABLED", "on")
	ApplyEnv()
	cfg := Current()
	if !cfg.SSH.Enabled || cfg.SSH.Listen != ":2232" || cfg.SSH.HostKeyPath != "/tmp/ssh-smoke/host" {
		t.Fatalf("%+v", cfg.SSH)
	}
	if cfg.SSH.PublicHost != "mud.example" || cfg.SSH.PublicPort != 2222 {
		t.Fatalf("public %+v", cfg.SSH)
	}
	if !cfg.SSH.Guest.Enabled || !cfg.SSH.Device.Enabled || cfg.SSH.Device.ActivateURL != "https://mud.example/activate" {
		t.Fatalf("flags %+v", cfg.SSH)
	}
	t.Setenv("SSH_ENABLED", "false")
	ApplyEnv()
	if Current().SSH.Enabled {
		t.Fatal("SSH_ENABLED=false did not win")
	}
}
