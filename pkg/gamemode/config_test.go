package gamemode

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultIsClassic(t *testing.T) {
	current = Config{}
	current = normalize(current)
	if DoorTUI() || DailyMenu() || LocalAuth() {
		t.Fatalf("zero config must stay classic, got %+v", Current())
	}
	if Current().Presentation != PresentationClassic || Current().Ruleset != RulesetClassic {
		t.Fatalf("defaults = %+v", Current())
	}
}

func TestApplyFileSetsPortAndIsolatedDB(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "door.yaml")
	body := []byte("presentation: door_tui\nruleset: daily_menu_rpg\nauth: local\nport: \"8020\"\nsqlite_path: data/aethermoor-door.db\nworld_pack: worlds/aethermoor-door\ntimezone: Europe/Berlin\ndaily_fights: 12\n")
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PORT", "8010")
	t.Setenv("SQLITE_PATH", "talesmud.db")
	t.Setenv("DOOR_SESSION_SECRET", "")
	if err := ApplyFile(path); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("PORT") != "8020" {
		t.Fatalf("PORT = %q", os.Getenv("PORT"))
	}
	if os.Getenv("SQLITE_PATH") != "data/aethermoor-door.db" {
		t.Fatalf("SQLITE_PATH = %q", os.Getenv("SQLITE_PATH"))
	}
	cfg := Current()
	if !DoorTUI() || !DailyMenu() || !LocalAuth() {
		t.Fatalf("flags not applied: %+v", cfg)
	}
	if cfg.WorldPack != "worlds/aethermoor-door" || cfg.DailyFights != 12 {
		t.Fatalf("pack config = %+v", cfg)
	}
	if cfg.Location().String() != "Europe/Berlin" {
		t.Fatalf("tz = %s", cfg.Location())
	}
}

func TestApplyEnvLeavesClassicWhenUnset(t *testing.T) {
	current = normalize(Config{})
	t.Setenv("PRESENTATION", "")
	t.Setenv("RULESET", "")
	t.Setenv("AUTH_MODE", "")
	ApplyEnv()
	if DoorTUI() || LocalAuth() {
		t.Fatalf("empty env must not enable door mode: %+v", Current())
	}
}
