package game

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func TestGreyRewardUsesHighestPartyLevel(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	g, facade := newNPCTestGame(t)
	_, high := sharePlayer(t, facade, "user-high", "ref-high", "char-high", "High", "room-grey")
	_, low := sharePlayer(t, facade, "user-low", "ref-low", "char-low", "Low", "room-grey")
	high.Level = 10
	low.Level = 1
	if err := facade.CharactersService().Update(high.ID, high); err != nil {
		t.Fatal(err)
	}
	if err := facade.CharactersService().Update(low.ID, low); err != nil {
		t.Fatal(err)
	}
	shareParty(t, facade, "Greys", high.ID, high.ID, low.ID)

	const baseXP int64 = 100
	runShareVictory(t, g, facade, "room-grey", "rat-grey", baseXP, 0, high, low)

	wantScaled := balance.ScaleReward(baseXP, balance.RewardMultiplier(balance.ThreatGrey))
	gotHigh, _ := facade.CharactersService().FindByID(high.ID)
	gotLow, _ := facade.CharactersService().FindByID(low.ID)
	sum := int64(gotHigh.XP) + int64(gotLow.XP)
	if sum != wantScaled {
		t.Fatalf("party XP sum %d, want scaled %d (high %d low %d)", sum, wantScaled, gotHigh.XP, gotLow.XP)
	}
	if int64(gotHigh.XP) >= baseXP || int64(gotLow.XP) >= baseXP {
		t.Fatalf("highest-in-split should grey-trickle both, high %d low %d", gotHigh.XP, gotLow.XP)
	}
}

func TestBossFirstKillBonusOnce(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	g, facade := newNPCTestGame(t)
	_, hero := sharePlayer(t, facade, "user-boss", "ref-boss", "char-boss", "Hero", "room-boss")
	hero.Level = 5
	if err := facade.CharactersService().Update(hero.ID, hero); err != nil {
		t.Fatal(err)
	}

	const baseXP int64 = 100
	const gold int32 = 10
	kill := func(id string) {
		t.Helper()
		enemy := &npc.NPC{
			Entity:           &entities.Entity{ID: id},
			Name:             "Sample Warden",
			TemplateID:       "ENM-WARDEN",
			Level:            5,
			IsDead:           true,
			MaxHitPoints:     100,
			CurrentHitPoints: 0,
			EnemyTrait: &npc.EnemyTrait{
				XPReward:   baseXP,
				GoldDrop:   npc.Range{Min: gold, Max: gold},
				Difficulty: "boss",
			},
		}
		g.NPCManager.RegisterExistingNPC(enemy, "room-boss")
		g.CombatController.processCombatVictory(&combat.CombatInstance{
			ID:           "combat-" + id,
			OriginRoomID: "room-boss",
			State:        combat.CombatStateVictory,
			Players: []combat.CombatantRef{{
				ID: hero.ID, Name: hero.Name, IsAlive: true, CurrentHP: 40, MaxHP: 40, Level: 5,
			}},
			Enemies: []combat.CombatantRef{{
				ID: enemy.ID, Name: enemy.Name, IsAlive: false, Type: combat.CombatantTypeNPC, Level: 5,
			}},
		})
	}

	storeTestRoom(t, facade, "room-boss", nil)
	_ = drainGameMessages(g.SendMessage())
	kill("warden-1")

	first, err := facade.CharactersService().FindByID(hero.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantFirstXP := baseXP + balance.BonusReward(baseXP, balance.FirstKillBonusRate())
	wantFirstGold := int64(gold) + balance.BonusReward(int64(gold), balance.FirstKillBonusRate())
	if int64(first.XP) != wantFirstXP || first.Gold != wantFirstGold {
		t.Fatalf("first kill xp %d gold %d, want xp %d gold %d", first.XP, first.Gold, wantFirstXP, wantFirstGold)
	}
	if len(first.FirstBossKills) != 1 || first.FirstBossKills[0] != "tpl:ENM-WARDEN" {
		t.Fatalf("first-kill flag = %#v", first.FirstBossKills)
	}

	var saw *messages.RewardBreakdown
	for _, msg := range drainGameMessages(g.SendMessage()) {
		if end, ok := msg.(*messages.CombatEndMessage); ok && end.Rewards != nil {
			saw = end.Rewards
		}
	}
	if saw == nil {
		t.Fatal("victory payload missing reward breakdown")
	}
	if saw.BaseXP != baseXP || saw.LevelModXP != 0 || saw.FirstKillXP != wantFirstXP-baseXP || saw.PartySize != 1 || saw.XP != wantFirstXP {
		t.Fatalf("breakdown %+v", saw)
	}

	kill("warden-2")
	second, err := facade.CharactersService().FindByID(hero.ID)
	if err != nil {
		t.Fatal(err)
	}
	if int64(second.XP) != wantFirstXP+baseXP || second.Gold != wantFirstGold+int64(gold) {
		t.Fatalf("second kill xp %d gold %d, want xp %d gold %d", second.XP, second.Gold, wantFirstXP+baseXP, wantFirstGold+int64(gold))
	}
	if len(second.FirstBossKills) != 1 {
		t.Fatalf("first-kill flag should stay one entry, got %#v", second.FirstBossKills)
	}
}

func TestVictoryPayloadCarriesLootAndLevelUp(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "room-loot", nil)
	_, hero := sharePlayer(t, facade, "user-loot", "ref-loot", "char-loot", "Looter", "room-loot")
	hero.Level = 1
	hero.XP = 0
	if err := facade.CharactersService().Update(hero.ID, hero); err != nil {
		t.Fatal(err)
	}
	if _, err := facade.ItemsService().Import(&items.Item{
		Entity:     &entities.Entity{ID: "tpl-fang"},
		Name:       "Rat Fang",
		IsTemplate: true,
		Quality:    items.ItemQualityRare,
		Quantity:   1,
	}); err != nil {
		t.Fatal(err)
	}

	const baseXP int64 = 100
	enemy := &npc.NPC{
		Entity:           &entities.Entity{ID: "rat-loot"},
		Name:             "Cellar Rat",
		Level:            1,
		IsDead:           true,
		MaxHitPoints:     8,
		CurrentHitPoints: 0,
		EnemyTrait: &npc.EnemyTrait{
			XPReward:       baseXP,
			GoldDrop:       npc.Range{Min: 4, Max: 4},
			Difficulty:     "easy",
			GuaranteedLoot: []string{"tpl-fang"},
		},
	}
	g.NPCManager.RegisterExistingNPC(enemy, "room-loot")
	_ = drainGameMessages(g.SendMessage())
	g.CombatController.processCombatVictory(&combat.CombatInstance{
		ID:           "combat-loot",
		OriginRoomID: "room-loot",
		State:        combat.CombatStateVictory,
		Players: []combat.CombatantRef{{
			ID: hero.ID, Name: hero.Name, IsAlive: true, CurrentHP: 20, MaxHP: 30, Level: 1,
		}},
		Enemies: []combat.CombatantRef{{
			ID: enemy.ID, Name: enemy.Name, IsAlive: false, Type: combat.CombatantTypeNPC, Level: 1,
		}},
	})

	var end *messages.CombatEndMessage
	for _, msg := range drainGameMessages(g.SendMessage()) {
		got, ok := msg.(*messages.CombatEndMessage)
		if ok {
			end = got
		}
	}
	if end == nil {
		t.Fatal("missing combat end")
	}
	if end.Outcome != "victory" || end.Message == "" {
		t.Fatalf("old clients need outcome and text, got %#v", end.Outcome)
	}
	if end.Rewards == nil || end.Rewards.BaseXP != baseXP || end.Rewards.BaseGold != 4 {
		t.Fatalf("breakdown %+v", end.Rewards)
	}
	if end.Rewards.LevelModXP != 0 || end.Rewards.FirstKillXP != 0 || end.Rewards.XP != baseXP {
		t.Fatalf("even fight should not add a gap or first-kill bonus: %+v", end.Rewards)
	}
	if len(end.Loot) != 1 || end.Loot[0].Name != "Rat Fang" || end.Loot[0].Quality != "rare" || end.Loot[0].Quantity != 1 {
		t.Fatalf("loot %+v", end.Loot)
	}
	if end.LevelUp == nil || end.LevelUp.NewLevel <= end.LevelUp.OldLevel || end.LevelUp.OldLevel != 1 {
		t.Fatalf("level-up callout %+v", end.LevelUp)
	}
	if end.Defeat != nil {
		t.Fatalf("victory must not carry a defeat summary: %+v", end.Defeat)
	}
}
