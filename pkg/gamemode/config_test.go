package gamemode

import (
	"os"
	"path/filepath"
	"testing"
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

func TestApplyEnvLeavesClassicWhenUnset(t *testing.T) {
	current = normalize(Config{})
	t.Setenv("PRESENTATION", "")
	t.Setenv("AUTH_MODE", "")
	ApplyEnv()
	if ANSI() || LocalAuth() {
		t.Fatalf("empty env enabled a mode: %+v", Current())
	}
}
