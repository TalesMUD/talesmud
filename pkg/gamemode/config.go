// Package gamemode holds the process flags that select a presentation and auth mode.
// The zero value is classic play with external auth.
package gamemode

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	PresentationClassic = "classic"
	PresentationANSI    = "door_tui"

	AuthExternal = "auth0"
	AuthLocal    = "local"
)

// Config is the process-wide mode. Zero value matches an unset classic server.
type Config struct {
	Presentation  string `yaml:"presentation"`
	Auth          string `yaml:"auth"`
	Port          string `yaml:"port"`
	SQLitePath    string `yaml:"sqlite_path"`
	WorldPack     string `yaml:"world_pack"`
	Timezone      string `yaml:"timezone"`
	Title         string `yaml:"title"`
	SessionSecret string `yaml:"session_secret"`
	SecretPath    string `yaml:"secret_path"`
	OutboxPath    string `yaml:"outbox_path"`
}

var current = normalize(Config{})

// Current returns the loaded mode.
func Current() Config {
	return current
}

// ANSI reports whether this process paints the text client.
func ANSI() bool {
	return current.Presentation == PresentationANSI
}

// LocalAuth reports whether username/password sessions are enabled.
func LocalAuth() bool {
	return current.Auth == AuthLocal
}

// Location returns the configured timezone. Invalid names fall back to UTC.
func (c Config) Location() *time.Location {
	name := strings.TrimSpace(c.Timezone)
	if name == "" {
		name = "UTC"
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}

// ApplyEnv reads mode flags from the environment. It does not change PORT or SQLITE_PATH.
func ApplyEnv() {
	cfg := current
	if v := strings.TrimSpace(os.Getenv("PRESENTATION")); v != "" {
		cfg.Presentation = v
	}
	if v := strings.TrimSpace(os.Getenv("AUTH_MODE")); v != "" {
		cfg.Auth = v
	}
	if v := strings.TrimSpace(os.Getenv("WORLD_PACK")); v != "" {
		cfg.WorldPack = v
	}
	if v := strings.TrimSpace(os.Getenv("GAME_TIMEZONE")); v != "" {
		cfg.Timezone = v
	}
	if v := strings.TrimSpace(os.Getenv("SESSION_SECRET")); v != "" {
		cfg.SessionSecret = v
	}
	if v := strings.TrimSpace(os.Getenv("AUTH_OUTBOX_PATH")); v != "" {
		cfg.OutboxPath = v
	}
	current = normalize(cfg)
}

// ApplyFile loads a YAML config. Port and sqlite path are copied into the
// process environment so server startup picks them up. An explicit file wins
// over PORT and SQLITE_PATH already set.
func ApplyFile(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("load game config %s: %w", path, err)
	}
	cfg := Config{}
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return fmt.Errorf("parse game config %s: %w", path, err)
	}
	cfg = normalize(cfg)
	if cfg.Port != "" {
		if err := os.Setenv("PORT", cfg.Port); err != nil {
			return err
		}
	}
	if cfg.SQLitePath != "" {
		if err := os.Setenv("SQLITE_PATH", cfg.SQLitePath); err != nil {
			return err
		}
	}
	if secret := strings.TrimSpace(os.Getenv("SESSION_SECRET")); secret != "" {
		cfg.SessionSecret = secret
	}
	if outbox := strings.TrimSpace(os.Getenv("AUTH_OUTBOX_PATH")); outbox != "" {
		cfg.OutboxPath = outbox
	}
	current = cfg
	return nil
}

func normalize(cfg Config) Config {
	cfg.Presentation = strings.TrimSpace(cfg.Presentation)
	if cfg.Presentation == "" {
		cfg.Presentation = PresentationClassic
	}
	cfg.Auth = strings.TrimSpace(cfg.Auth)
	if cfg.Auth == "" {
		cfg.Auth = AuthExternal
	}
	cfg.Timezone = strings.TrimSpace(cfg.Timezone)
	if cfg.Timezone == "" {
		cfg.Timezone = "UTC"
	}
	cfg.Port = strings.TrimSpace(cfg.Port)
	cfg.SQLitePath = strings.TrimSpace(cfg.SQLitePath)
	cfg.WorldPack = strings.TrimSpace(cfg.WorldPack)
	cfg.Title = strings.TrimSpace(cfg.Title)
	cfg.SecretPath = strings.TrimSpace(cfg.SecretPath)
	if cfg.SecretPath == "" {
		cfg.SecretPath = "data/session.key"
	}
	cfg.OutboxPath = strings.TrimSpace(cfg.OutboxPath)
	if cfg.OutboxPath == "" {
		cfg.OutboxPath = "data/auth-outbox.log"
	}
	return cfg
}
