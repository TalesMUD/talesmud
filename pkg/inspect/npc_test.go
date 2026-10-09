package inspect

import (
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/scripts"
	"github.com/talesmud/talesmud/pkg/worldindex"
)

func TestNPCInspectNamedOverrideLootSpawnersAndQuests(t *testing.T) {
	delay := 30 * time.Minute
	override := 10 * time.Minute
	snap := worldindex.NewSnapshot()
	snap.Rooms["R0228"] = &rooms.Room{Entity: &entities.Entity{ID: "R0228"}, Name: "Construct Vault"}
	snap.Rooms["R0002"] = &rooms.Room{Entity: &entities.Entity{ID: "R0002"}, Name: "Alley"}
	snap.Dialogs["DLG1"] = &dialogs.Dialog{Entity: &entities.Entity{ID: "DLG1"}, Name: "Vault Greeting"}
	snap.Dialogs["DLG2"] = &dialogs.Dialog{Entity: &entities.Entity{ID: "DLG2"}, Name: "Idle Clang"}
	snap.Scripts["SCR0207"] = &scripts.Script{Entity: &entities.Entity{ID: "SCR0207"}, Name: "Summon rats"}
	snap.Items["ITM0020"] = &items.Item{Entity: &entities.Entity{ID: "ITM0020"}, Name: "Hollow Knight Shard", IsTemplate: true}
	snap.Items["ITM0265"] = &items.Item{Entity: &entities.Entity{ID: "ITM0265"}, Name: "Unmarked Vigil Blade", IsTemplate: true}
	snap.Items["ITM0017"] = &items.Item{Entity: &entities.Entity{ID: "ITM0017"}, Name: "Strange Gear", IsTemplate: true}
	snap.LootTables["LT0204"] = &items.LootTable{
		Entity: &entities.Entity{ID: "LT0204"},
		Name:   "Hollow Knight Loot",
		Entries: []items.LootEntry{
			{ItemTemplateID: "ITM0020", DropChance: 1, Guaranteed: true, MinQuantity: 1, MaxQuantity: 1},
			{ItemTemplateID: "ITM0265", DropChance: 0.1, Rarity: "unique", BossOnly: true, MinQuantity: 1, MaxQuantity: 1},
			{ItemTemplateID: "ITM0099", DropChance: 0.08},
		},
	}
	snap.NPCs["ENM0009"] = &npc.NPC{
		Entity:       &entities.Entity{ID: "ENM0009"},
		Name:         "The Hollow Knight",
		Description:  "A sealed construct.",
		Level:        6,
		IsTemplate:   true,
		MaxHitPoints: 150,
		SpawnRoomID:  "R0228",
		DialogID:     "DLG1",
		IdleDialogID: "DLG2",
		EnemyTrait: &npc.EnemyTrait{
			CreatureType:   "construct",
			CombatStyle:    "melee",
			Difficulty:     "boss",
			AttackPower:    11,
			Defense:        5,
			AttackSpeed:    0,
			FleeThreshold:  0,
			XPReward:       135,
			GoldDrop:       npc.Range{Min: 6, Max: 15},
			LootTableID:    "LT0204",
			GuaranteedLoot: []string{"ITM0017"},
			OnAggroScript:  "SCR0207",
			BaseStats:      &npc.BaseStats{MaxHitPoints: 150, AttackPower: 13, Defense: 6},
		},
	}
	snap.NPCs["ENM0008"] = &npc.NPC{Entity: &entities.Entity{ID: "ENM0008"}, Name: "Sewer Rat", IsTemplate: true}
	snap.Spawners["SPW1"] = &npc.NPCSpawner{
		Entity: &entities.Entity{ID: "SPW1"}, Name: "Vault knight", TemplateID: "ENM0009", RoomID: "R0228",
		MaxInstances: 1, SpawnInterval: time.Minute, RespawnDelay: delay,
	}
	snap.Spawners["SPW2"] = &npc.NPCSpawner{
		Entity: &entities.Entity{ID: "SPW2"}, TemplateID: "ENM0009", RoomID: "R0002",
		MaxInstances: 2, RespawnTimeOverride: &override,
	}
	snap.Spawners["SPW-other"] = &npc.NPCSpawner{
		Entity: &entities.Entity{ID: "SPW-other"}, TemplateID: "ENM0008", RoomID: "R0002", MaxInstances: 4,
	}
	snap.Quests["QST-kill"] = &quests.Quest{
		Entity: &entities.Entity{ID: "QST-kill"}, Name: "The Truth Below",
		Objectives: []quests.Objective{{ID: "slay", Type: quests.ObjectiveKill, TargetID: "ENM0009", Description: "Slay the knight"}},
	}
	snap.Quests["QST-talk"] = &quests.Quest{
		Entity: &entities.Entity{ID: "QST-talk"}, Name: "Ask the knight",
		Objectives: []quests.Objective{{ID: "ask", Type: quests.ObjectiveTalk, TargetID: "ENM0009"}},
	}
	snap.Quests["QST-give"] = &quests.Quest{
		Entity: &entities.Entity{ID: "QST-give"}, Name: "Return the gear",
		Objectives: []quests.Objective{{
			ID: "hand", Type: quests.ObjectiveDeliver, TargetID: "ITM0017", DeliverToNPCID: "ENM0009",
		}},
	}
	snap.Quests["QST-other"] = &quests.Quest{
		Entity: &entities.Entity{ID: "QST-other"}, Name: "Rats",
		Source: quests.QuestSource{Type: "npc", NPCID: "ENM0009"},
		Objectives: []quests.Objective{
			{ID: "rats", Type: quests.ObjectiveKill, TargetID: "ENM0008"},
			{ID: "gear", Type: quests.ObjectiveCollect, TargetID: "ITM0017"},
		},
	}

	view, ok := NPC(worldindex.Build(snap), "ENM0009")
	if !ok {
		t.Fatal("missing npc")
	}
	if view.Name != "The Hollow Knight" || view.Level != 6 || view.Difficulty != "boss" || !view.IsTemplate {
		t.Fatalf("summary %+v", view)
	}
	if len(view.Kinds) != 1 || view.Kinds[0] != "enemy" {
		t.Fatalf("kinds %v", view.Kinds)
	}
	if view.Stats == nil || view.Stats.Scaling != "named" || view.Stats.UnknownTier {
		t.Fatalf("scaling %+v", view.Stats)
	}
	if view.Stats.Base == nil || view.Stats.Base.AttackPower != 13 || view.Stats.Effective.AttackPower != 11 || view.Stats.Effective.Defense != 5 || view.Stats.Effective.MaxHitPoints != 150 {
		t.Fatalf("stats %+v", view.Stats)
	}
	if view.Stats.Factors == nil || view.Stats.Factors.Attack != 0.85 || view.Stats.Factors.HP != 1 {
		t.Fatalf("factors %+v", view.Stats.Factors)
	}
	if view.Dialog == nil || view.Dialog.Name != "Vault Greeting" || view.IdleDialog == nil || view.IdleDialog.ID != "DLG2" {
		t.Fatalf("dialogs %+v %+v", view.Dialog, view.IdleDialog)
	}
	if len(view.Scripts) != 4 || view.Scripts[0].ID != "SCR0207" || view.Scripts[0].Name != "Summon rats" || view.Scripts[1].ID != "" {
		t.Fatalf("scripts %+v", view.Scripts)
	}
	if view.Scripts[3].Hook != "onLowHealth" || view.Scripts[3].Threshold != npc.DefaultLowHealthFraction {
		t.Fatalf("low health %+v", view.Scripts[3])
	}
	if view.Loot == nil || view.Loot.TableName != "Hollow Knight Loot" || len(view.Loot.Entries) != 3 {
		t.Fatalf("loot %+v", view.Loot)
	}
	if !view.Loot.Entries[0].Guaranteed || view.Loot.Entries[1].DropChance != 0.1 || view.Loot.Entries[1].Rarity != "unique" || !view.Loot.Entries[1].BossOnly {
		t.Fatalf("drops %+v", view.Loot.Entries)
	}
	if view.Loot.Entries[2].ItemID != "ITM0099" || !view.Loot.Entries[2].Missing {
		t.Fatalf("missing drop %+v", view.Loot.Entries[2])
	}
	if len(view.Loot.Guaranteed) != 1 || view.Loot.Guaranteed[0].ID != "ITM0017" {
		t.Fatalf("guaranteed %+v", view.Loot.Guaranteed)
	}
	if len(view.Spawners) != 2 {
		t.Fatalf("spawners %+v", view.Spawners)
	}
	if view.Spawners[0].RoomID != "R0002" || view.Spawners[0].MaxInstances != 2 || view.Spawners[0].RespawnTime != "10m0s" || view.Spawners[0].RespawnSource != "override" {
		t.Fatalf("alley spawner %+v", view.Spawners[0])
	}
	if view.Spawners[1].RoomName != "Construct Vault" || view.Spawners[1].RespawnSource != "delay" || view.Spawners[1].RespawnTime != "30m0s" || view.Spawners[1].MaxInstances != 1 {
		t.Fatalf("vault spawner %+v", view.Spawners[1])
	}
	if view.SpawnRoom == nil || view.SpawnRoom.Name != "Construct Vault" {
		t.Fatalf("spawn room %+v", view.SpawnRoom)
	}
	if len(view.Quests) != 3 {
		t.Fatalf("quests %+v", view.Quests)
	}
	got := map[string]string{}
	for _, hit := range view.Quests {
		got[hit.ID] = hit.Role
	}
	if got["QST-give"] != "deliver" || got["QST-kill"] != "kill" || got["QST-talk"] != "talk" {
		t.Fatalf("roles %v", got)
	}
	if _, noise := got["QST-other"]; noise {
		t.Fatal("source and other targets are not inspector hits")
	}
}

func TestNPCInspectUnknownTierKeepsBase(t *testing.T) {
	snap := worldindex.NewSnapshot()
	snap.NPCs["ENM"] = &npc.NPC{
		Entity: &entities.Entity{ID: "ENM"}, Name: "Odd Beast", Level: 2,
		EnemyTrait: &npc.EnemyTrait{
			Difficulty: "elite",
			BaseStats:  &npc.BaseStats{MaxHitPoints: 10, AttackPower: 4, Defense: 2},
		},
	}
	view, ok := NPC(worldindex.Build(snap), "ENM")
	if !ok {
		t.Fatal("missing")
	}
	if !view.Stats.UnknownTier || view.Stats.Scaling != "unknown" || view.Stats.Factors != nil {
		t.Fatalf("stats %+v", view.Stats)
	}
	if view.Stats.Effective.MaxHitPoints != 10 || view.Stats.Effective.AttackPower != 4 || view.Stats.Effective.Defense != 2 {
		t.Fatalf("effective %+v", view.Stats.Effective)
	}
	if len(view.Spawners) != 0 || len(view.Quests) != 0 {
		t.Fatalf("empty lists spawners=%d quests=%d", len(view.Spawners), len(view.Quests))
	}
}

func TestNPCInspectStoredStatsAndKinds(t *testing.T) {
	snap := worldindex.NewSnapshot()
	snap.NPCs["SHOP"] = &npc.NPC{
		Entity: &entities.Entity{ID: "SHOP"}, Name: "Mira", Level: 3, MaxHitPoints: 20,
		MerchantTrait: &npc.MerchantTrait{},
		EnemyTrait: &npc.EnemyTrait{
			Difficulty: "normal", AttackPower: 5, Defense: 1,
		},
	}
	snap.NPCs["CIV"] = &npc.NPC{Entity: &entities.Entity{ID: "CIV"}, Name: "Clerk", Level: 1, MaxHitPoints: 8}
	ix := worldindex.Build(snap)
	shop, ok := NPC(ix, "SHOP")
	if !ok || len(shop.Kinds) != 2 || shop.Kinds[0] != "enemy" || shop.Kinds[1] != "merchant" {
		t.Fatalf("kinds %+v", shop.Kinds)
	}
	if shop.Stats.Scaling != "stored" || shop.Stats.Base != nil || shop.Stats.UnknownTier || shop.Stats.Effective.AttackPower != 5 {
		t.Fatalf("stored %+v", shop.Stats)
	}
	civ, ok := NPC(ix, "CIV")
	if !ok || len(civ.Kinds) != 1 || civ.Kinds[0] != "npc" || civ.Stats != nil || civ.Loot != nil {
		t.Fatalf("civilian %+v", civ)
	}
	if _, ok := NPC(ix, "missing"); ok {
		t.Fatal("missing npc should not inspect")
	}
	if _, ok := NPC(nil, "CIV"); ok {
		t.Fatal("nil index")
	}
}

func TestNPCInspectTierScaling(t *testing.T) {
	snap := worldindex.NewSnapshot()
	snap.NPCs["WOLF"] = &npc.NPC{
		Entity: &entities.Entity{ID: "WOLF"}, Name: "Meadow Wolf",
		EnemyTrait: &npc.EnemyTrait{
			Difficulty: "normal",
			BaseStats:  &npc.BaseStats{MaxHitPoints: 18, AttackPower: 4, Defense: 1},
		},
	}
	view, ok := NPC(worldindex.Build(snap), "WOLF")
	if !ok {
		t.Fatal("missing")
	}
	if view.Stats.Scaling != "tier" || view.Stats.UnknownTier || view.Stats.Factors == nil || view.Stats.Factors.HP != 2.7 {
		t.Fatalf("tier %+v", view.Stats)
	}
	if view.Stats.Effective.MaxHitPoints != 48 || view.Stats.Effective.AttackPower != 5 {
		t.Fatalf("effective %+v", view.Stats.Effective)
	}
}
