package importer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyAssetsFlattensSpritePortraits(t *testing.T) {
	root := t.TempDir()
	importPath := filepath.Join(root, "import", "mvp-rpg-1")
	mustWrite := func(rel string) {
		p := filepath.Join(importPath, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("png"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("assets/images/rooms/R0001.png")
	mustWrite("assets/images/sprites/npcs/NPC0001.png")
	mustWrite("assets/images/sprites/enemies/ENM0008.png")
	mustWrite("assets/images/sprites/_raw/npcs/NPC0001.png")
	mustWrite("assets/images/sprites/previews/NPC0001__R0203.png")
	mustWrite("assets/images/sprites/contact_sheet.png")
	mustWrite("assets/images/items/ITM0001.png")
	mustWrite("assets/images/items/generic-weapon.png")

	bg := filepath.Join(root, "uploads", "backgrounds")
	pt := filepath.Join(root, "uploads", "portraits")
	it := filepath.Join(root, "uploads", "items")
	t.Setenv("BACKGROUNDS_PATH", bg)
	t.Setenv("PORTRAITS_PATH", pt)
	t.Setenv("ITEM_ART_PATH", it)

	w := &WorldImporter{importPath: importPath}
	n, err := w.copyAssets()
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatalf("copied %d, want 5 (room + npc + enemy + 2 items)", n)
	}
	if _, err := os.Stat(filepath.Join(bg, "R0001.png")); err != nil {
		t.Fatal("room background missing")
	}
	if _, err := os.Stat(filepath.Join(pt, "NPC0001.png")); err != nil {
		t.Fatal("npc portrait missing")
	}
	if _, err := os.Stat(filepath.Join(pt, "ENM0008.png")); err != nil {
		t.Fatal("enemy portrait missing")
	}
	if _, err := os.Stat(filepath.Join(it, "ITM0001.png")); err != nil {
		t.Fatal("item art missing")
	}
	if _, err := os.Stat(filepath.Join(it, "generic-weapon.png")); err != nil {
		t.Fatal("generic item art missing")
	}
	if _, err := os.Stat(filepath.Join(pt, "NPC0001__R0203.png")); !os.IsNotExist(err) {
		t.Fatal("preview must not be imported as a portrait")
	}
	if _, err := os.Stat(filepath.Join(pt, "contact_sheet.png")); !os.IsNotExist(err) {
		t.Fatal("contact sheet must not be imported")
	}
}
