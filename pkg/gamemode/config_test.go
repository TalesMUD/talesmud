package gamemode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/talesmud/talesmud/pkg/ruleset"
)

func TestDefaultIsClassic(t *testing.T) {
	current = normalize(Config{})
	if ANSI() || LocalAuth() {
		t.Fatalf("zero config must stay classic, got %+v", Current())
	}
	if Current().Presentation != PresentationClassic || Current().Auth != AuthExternal {
		t.Fatalf("defaults = %+v", Current())
	}
	if Current().Timezone != "UTC" {
		t.Fatalf("tz = %s", Current().Timezone)
	}
}

func TestApplyFileSetsPortAndDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mode.yaml")
	body := []byte("presentation: door_tui\nauth: local\nport: \"8030\"\nsqlite_path: data/smoke.db\nworld_pack: worlds/sample\ntimezone: Europe/Berlin\ntitle: Sample\n")
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PORT", "8010")
	t.Setenv("SQLITE_PATH", "talesmud.db")
	t.Setenv("SESSION_SECRET", "")
	if err := ApplyFile(path); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { current = normalize(Config{}) })
	if os.Getenv("PORT") != "8030" || os.Getenv("SQLITE_PATH") != "data/smoke.db" {
		t.Fatalf("PORT=%s SQLITE=%s", os.Getenv("PORT"), os.Getenv("SQLITE_PATH"))
	}
	cfg := Current()
	if !ANSI() || !LocalAuth() || cfg.WorldPack != "worlds/sample" || cfg.Title != "Sample" {
		t.Fatalf("%+v", cfg)
	}
	if cfg.Location().String() != "Europe/Berlin" {
		t.Fatalf("tz %s", cfg.Location())
	}
}

func TestApplyFileAlsoLoadsRulesetSections(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(func() {
		rulesetFromConfig = false
		ruleset.Reset()
		current = normalize(Config{})
	})
	path := filepath.Join(t.TempDir(), "mode.yaml")
	body := []byte("presentation: door_tui\nauth: local\nport: \"8030\"\nprogression:\n  level_cap: 12\n  level_up_mode: trainer\ncombat:\n  pacing: turn_based\nresources:\n  forest_walks:\n    allowance: 25\n    reset: calendar\n    timezone: Europe/Berlin\n")
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ApplyFile(path); err != nil {
		t.Fatal(err)
	}
	if !RulesetFromConfig() || ruleset.LevelCap() != 12 || ruleset.LevelUpMode() != ruleset.ModeTrainer || ruleset.Pacing() != ruleset.PacingTurnBased {
		t.Fatalf("cap=%d mode=%s pacing=%s from=%v", ruleset.LevelCap(), ruleset.LevelUpMode(), ruleset.Pacing(), RulesetFromConfig())
	}
	allow := ruleset.ResourceAllowances()
	if len(allow) != 1 || allow[0].Key != "forest_walks" || allow[0].Amount != 25 {
		t.Fatalf("%+v", allow)
	}
}

func TestTrustedProxiesDefaultToLoopback(t *testing.T) {
	current = normalize(Config{})
	t.Cleanup(func() { current = normalize(Config{}) })
	t.Setenv("TRUSTED_PROXIES", "")
	got := TrustedProxies()
	if len(got) != 2 || got[0] != "127.0.0.1" || got[1] != "::1" {
		t.Fatalf("default proxies %v", got)
	}
	t.Setenv("TRUSTED_PROXIES", " 10.0.0.1 , 10.0.0.2 ")
	got = TrustedProxies()
	if len(got) != 2 || got[0] != "10.0.0.1" || got[1] != "10.0.0.2" {
		t.Fatalf("env proxies %v", got)
	}
	current.TrustedProxies = []string{"192.0.2.1"}
	got = TrustedProxies()
	if len(got) != 1 || got[0] != "192.0.2.1" {
		t.Fatalf("config proxies %v", got)
	}
}

func TestClientPageDefaultsAreGeneric(t *testing.T) {
	current = normalize(Config{})
	t.Cleanup(func() { current = normalize(Config{}) })
	title, subtitle, key := ClientPage()
	if title != "TalesMUD Door" || subtitle != "A text client on TalesMUD" || key != "talesmudDoorToken" {
		t.Fatalf("title=%q subtitle=%q key=%q", title, subtitle, key)
	}
	current.Title = "Sample"
	current.Subtitle = "A place"
	current.TokenKey = "sample_token"
	title, subtitle, key = ClientPage()
	if title != "Sample" || subtitle != "A place" || key != "sample_token" {
		t.Fatalf("configured title=%q subtitle=%q key=%q", title, subtitle, key)
	}
	current.TokenKey = "bad key"
	if _, _, key = ClientPage(); key != "talesmudDoorToken" {
		t.Fatalf("unsafe key accepted: %s", key)
	}
}

func TestApplyEnvLeavesClassicWhenUnset(t *testing.T) {
	current = normalize(Config{})
	t.Setenv("PRESENTATION", "")
	t.Setenv("AUTH_MODE", "")
	ApplyEnv()
	if ANSI() || LocalAuth() {
		t.Fatalf("empty env enabled a mode: %+v", Current())
	}
}
