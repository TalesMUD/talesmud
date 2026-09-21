// Package gamemode holds the thin process flags that select a TalesMUD
// presentation and ruleset. The default is classic Veilspan play.
package gamemode

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	PresentationClassic = "classic"
	PresentationDoorTUI = "door_tui"

	RulesetClassic = "classic_mud"
	RulesetDaily   = "daily_menu_rpg"

	AuthAuth0 = "auth0"
	AuthLocal = "local"

	DefaultTimezone    = "Europe/Berlin"
	DefaultDailyFights = 15
	DefaultHealCost    = 20
)

// Config is the process-wide mode. Zero value matches an unset classic server.
type Config struct {
	Presentation string `yaml:"presentation"`
	Ruleset      string `yaml:"ruleset"`
	Auth         string `yaml:"auth"`
	Port         string `yaml:"port"`
	SQLitePath   string `yaml:"sqlite_path"`
	WorldPack    string `yaml:"world_pack"`
	Timezone     string `yaml:"timezone"`
	DailyFights  int    `yaml:"daily_fights"`
	// SessionSecret is optional. Prefer DOOR_SESSION_SECRET. Never required in yaml.
	SessionSecret string `yaml:"session_secret"`
	SecretPath    string `yaml:"secret_path"`
	OutboxPath    string `yaml:"outbox_path"`
}

var current = Config{
	Presentation: PresentationClassic,
	Ruleset:      RulesetClassic,
	Auth:         AuthAuth0,
	Timezone:     DefaultTimezone,
	DailyFights:  DefaultDailyFights,
}

// Current returns the loaded mode. Safe to call before Apply; defaults are classic.
func Current() Config {
	return current
}

// DoorTUI reports whether this process paints the ANSI door client.
func DoorTUI() bool {
	return current.Presentation == PresentationDoorTUI
}

// DailyMenu reports whether daily menu-RPG rules are active.
func DailyMenu() bool {
	return current.Ruleset == RulesetDaily
}

// LocalAuth reports whether username/password sessions are enabled.
func LocalAuth() bool {
	return current.Auth == AuthLocal
}

// Location returns the daily-reset timezone. Invalid names fall back to UTC.
func (c Config) Location() *time.Location {
	name := strings.TrimSpace(c.Timezone)
	if name == "" {
		name = DefaultTimezone
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}

// ApplyEnv reads mode flags from the environment without overriding PORT or SQLITE_PATH.
// Unset variables leave the classic defaults in place.
func ApplyEnv() {
	cfg := current
	if v := strings.TrimSpace(os.Getenv("PRESENTATION")); v != "" {
		cfg.Presentation = v
	}
	if v := strings.TrimSpace(os.Getenv("RULESET")); v != "" {
		cfg.Ruleset = v
	}
	if v := strings.TrimSpace(os.Getenv("AUTH_MODE")); v != "" {
		cfg.Auth = v
	}
	if v := strings.TrimSpace(os.Getenv("WORLD_PACK")); v != "" {
		cfg.WorldPack = v
	}
	if v := strings.TrimSpace(os.Getenv("DOOR_TIMEZONE")); v != "" {
		cfg.Timezone = v
	}
	if v := strings.TrimSpace(os.Getenv("DOOR_DAILY_FIGHTS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.DailyFights = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("DOOR_SESSION_SECRET")); v != "" {
		cfg.SessionSecret = v
	}
	if v := strings.TrimSpace(os.Getenv("DOOR_OUTBOX_PATH")); v != "" {
		cfg.OutboxPath = v
	}
	current = normalize(cfg)
}

// ApplyFile loads a YAML config and applies port and sqlite path to the process
// environment so the existing server startup picks them up. An explicit -config
// wins over PORT and SQLITE_PATH already loaded from .env.
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
	if secret := strings.TrimSpace(os.Getenv("DOOR_SESSION_SECRET")); secret != "" {
		cfg.SessionSecret = secret
	}
	if outbox := strings.TrimSpace(os.Getenv("DOOR_OUTBOX_PATH")); outbox != "" {
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
	cfg.Ruleset = strings.TrimSpace(cfg.Ruleset)
	if cfg.Ruleset == "" {
		cfg.Ruleset = RulesetClassic
	}
	cfg.Auth = strings.TrimSpace(cfg.Auth)
	if cfg.Auth == "" {
		cfg.Auth = AuthAuth0
	}
	cfg.Timezone = strings.TrimSpace(cfg.Timezone)
	if cfg.Timezone == "" {
		cfg.Timezone = DefaultTimezone
	}
	if cfg.DailyFights <= 0 {
		cfg.DailyFights = DefaultDailyFights
	}
	cfg.Port = strings.TrimSpace(cfg.Port)
	cfg.SQLitePath = strings.TrimSpace(cfg.SQLitePath)
	cfg.WorldPack = strings.TrimSpace(cfg.WorldPack)
	cfg.SecretPath = strings.TrimSpace(cfg.SecretPath)
	if cfg.SecretPath == "" {
		cfg.SecretPath = "data/door.session.key"
	}
	cfg.OutboxPath = strings.TrimSpace(cfg.OutboxPath)
	if cfg.OutboxPath == "" {
		cfg.OutboxPath = "data/door-outbox.log"
	}
	return cfg
}
