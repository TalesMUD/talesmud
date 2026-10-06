package game

import (
	"path/filepath"
	"testing"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/ruleset"
	"github.com/talesmud/talesmud/pkg/scripts"
	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
	"github.com/talesmud/talesmud/pkg/scripts/runner/lua/modules"
	"github.com/talesmud/talesmud/pkg/service"
)

func TestSparDefeatSkipsDeathAndRunsAfterScript(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	ruleset.SetDeath(ruleset.DeathPolicy{
		XPLossPercent:    10,
		GoldLossPercent:  100,
		Respawn:          ruleset.RespawnNextReset,
		RespawnHPPercent: 0,
		DamageArmor:      true,
	})

	client, err := sqlite.Open(filepath.Join(t.TempDir(), "spar.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	runner := luarunner.NewLuaRunner()
	t.Cleanup(runner.Shutdown)
	facade := service.NewFacade(repository.NewSQLiteFactory(client), runner)
	modules.RegisterAllModules(runner)
	runner.SetServices(facade, nil)
	g := New(facade)

	storeTestRoom(t, facade, "yard", nil)
	storeTestRoom(t, facade, "town", nil)
	if _, err := facade.ScriptsService().Import(&scripts.Script{
		Entity:   &entities.Entity{ID: "SCR_TEST_END"},
		Name:     "end",
		Language: scripts.ScriptLanguageLua,
		Code: `
local v = tales.game.getFlag("char-spar", "last_combat")
if v == "defeat" then
  tales.game.setFlag("char-spar", "echo", "defeat")
else
  tales.game.setFlag("char-spar", "echo", "other")
end
`,
	}); err != nil {
		t.Fatal(err)
	}

	jerkin := &items.Item{
		Entity:        &entities.Entity{ID: "armor-spar"},
		Name:          "Coat",
		Type:          items.ItemTypeArmor,
		Slot:          items.ItemSlotChest,
		Durability:    4,
		MaxDurability: 4,
		Attributes:    map[string]interface{}{"defense": float64(1)},
	}
	character, err := facade.CharactersService().Import(&characters.Character{
		Entity:           &entities.Entity{ID: "char-spar"},
		Name:             "Pupil",
		BelongsUser:      *traits.BelongsToUser("user-spar"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "yard"},
		BoundRoomID:      "town",
		MaxHitPoints:     20,
		CurrentHitPoints: 1,
		XP:               100,
		Gold:             50,
		EquippedItems:    map[items.ItemSlot]*items.Item{items.ItemSlotChest: jerkin},
		Flags: map[string]interface{}{
			"spar":         true,
			"after_combat": "SCR_TEST_END",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	g.CombatController.cleanupCombatInstance(&combat.CombatInstance{
		ID:           "bout",
		OriginRoomID: "yard",
		State:        combat.CombatStateDefeat,
		Players: []combat.CombatantRef{
			{ID: character.ID, Name: character.Name, CurrentHP: 0, MaxHP: 20, IsAlive: false},
		},
	}, combat.CombatStateDefeat)

	stored, err := facade.CharactersService().FindByID(character.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.CurrentHitPoints != 20 || stored.MaxHitPoints != 20 {
		t.Fatalf("hp %d/%d", stored.CurrentHitPoints, stored.MaxHitPoints)
	}
	if stored.Gold != 50 || stored.XP != 100 {
		t.Fatalf("gold %d xp %d", stored.Gold, stored.XP)
	}
	if stored.CurrentRoomID != "yard" || stored.AwaitingReset {
		t.Fatalf("room %s reset %v", stored.CurrentRoomID, stored.AwaitingReset)
	}
	if stored.EquippedItems[items.ItemSlotChest] == nil || stored.EquippedItems[items.ItemSlotChest].Durability != 4 {
		t.Fatalf("armor %+v", stored.EquippedItems[items.ItemSlotChest])
	}
	if _, ok := stored.Flags["spar"]; ok {
		t.Fatal("spar flag should be cleared")
	}
	if _, ok := stored.Flags["after_combat"]; ok {
		t.Fatal("after_combat flag should be cleared before the script runs")
	}
	if stored.Flags["last_combat"] != "defeat" || stored.Flags["echo"] != "defeat" {
		t.Fatalf("flags %#v", stored.Flags)
	}
}

func TestSparFlagDropsWhenTheFightIsNotADefeat(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "yard", nil)
	character, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "char-flee"},
		Name:             "Runner",
		BelongsUser:      *traits.BelongsToUser("user-flee"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "yard"},
		MaxHitPoints:     20,
		CurrentHitPoints: 7,
		XP:               40,
		Gold:             9,
		Flags:            map[string]interface{}{"spar": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	g.CombatController.cleanupCombatInstance(&combat.CombatInstance{
		ID:           "flee-bout",
		OriginRoomID: "yard",
		State:        combat.CombatStateFled,
		Players: []combat.CombatantRef{
			{ID: character.ID, Name: character.Name, CurrentHP: 7, MaxHP: 20, IsAlive: true},
		},
	}, combat.CombatStateFled)

	stored, err := facade.CharactersService().FindByID(character.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := stored.Flags["spar"]; ok {
		t.Fatal("spar should not survive a flee")
	}
	if stored.CurrentHitPoints != 7 || stored.Gold != 9 || stored.XP != 40 || stored.CurrentRoomID != "yard" {
		t.Fatalf("hp %d gold %d xp %d room %s", stored.CurrentHitPoints, stored.Gold, stored.XP, stored.CurrentRoomID)
	}
}
