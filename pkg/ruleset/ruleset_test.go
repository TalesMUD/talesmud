package ruleset_test

import (
	"strings"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
	"github.com/talesmud/talesmud/pkg/ruleset"
)

func TestShippedFileMatchesCurrentPlay(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	if ruleset.LevelCap() != 50 || ruleset.LevelUpMode() != ruleset.ModeAuto || ruleset.Pacing() != ruleset.PacingAuto {
		t.Fatalf("cap=%d mode=%s pacing=%s", ruleset.LevelCap(), ruleset.LevelUpMode(), ruleset.Pacing())
	}
	if len(ruleset.ResourceAllowances()) != 0 {
		t.Fatalf("allowances = %+v", ruleset.ResourceAllowances())
	}
	char := &characters.Character{
		Entity:           &entities.Entity{ID: "c"},
		XP:               100,
		Gold:             5,
		MaxHitPoints:     20,
		CurrentHitPoints: 1,
		BoundRoomID:      "R0203",
	}
	char.CurrentRoomID = "R0106"
	out := ruleset.ApplyDeath(char)
	if out.XPLost != 10 || char.XP != 90 || out.GoldLost != 1 || char.Gold != 4 {
		t.Fatalf("penalty xp=%d gold=%d char=%+v", out.XPLost, out.GoldLost, char)
	}
	if char.CurrentHitPoints != 10 || out.RespawnRoomID != "R0203" || out.AwaitingReset || !out.DamageArmor {
		t.Fatalf("outcome %+v hp=%d", out, char.CurrentHitPoints)
	}
	healed := &characters.Character{CurrentHitPoints: 1, MaxHitPoints: 8, LastResetDay: ""}
	if ruleset.ApplyNewDay(healed, time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)) {
		t.Fatal("default new day should not heal")
	}
	if healed.CurrentHitPoints != 1 || healed.LastResetDay != "" {
		t.Fatalf("new day wrote %+v", healed)
	}
}

func TestRejectsCombatBalanceKeys(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	err := ruleset.LoadBytes([]byte("progression:\n  level_cap: 12\nlevel_gap:\n  max_levels: 6\n"))
	if err == nil || !strings.Contains(err.Error(), "level_gap") {
		t.Fatalf("err = %v", err)
	}
	err = ruleset.LoadBytes([]byte("death:\n  reward_scale:\n    grey: 1\n"))
	if err == nil || !strings.Contains(err.Error(), "reward_scale") {
		t.Fatalf("nested err = %v", err)
	}
	if ruleset.LevelCap() != 50 {
		t.Fatalf("rejected document changed cap to %d", ruleset.LevelCap())
	}
}

func TestXPTableIsBaseForRewardScale(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	ruleset.SetBaseEnemyXP(map[int32]int64{4: 80})
	base := leveling.ResolveEnemyBaseXP(4, 0)
	if base != 80 {
		t.Fatalf("table base = %d", base)
	}
	if authored := leveling.ResolveEnemyBaseXP(4, 15); authored != 15 {
		t.Fatalf("authored base = %d", authored)
	}
	if fallback := leveling.ResolveEnemyBaseXP(3, 0); fallback != leveling.CalculateEnemyXPReward(3) {
		t.Fatalf("missing level fallback = %d", fallback)
	}
	grey := balance.ScaleReward(base, balance.RewardMultiplier(balance.ThreatGrey))
	yellow := balance.ScaleReward(base, balance.RewardMultiplier(balance.ThreatYellow))
	if grey != 12 || yellow != 80 {
		t.Fatalf("scaled grey=%d yellow=%d", grey, yellow)
	}
}

func TestTrainerBanksUntilApply(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	ruleset.SetLevelUpMode(ruleset.ModeTrainer)
	need := leveling.GetXPRequired(2)
	char := &characters.Character{
		Entity:           &entities.Entity{ID: "trainee"},
		Level:            1,
		XP:               need,
		MaxHitPoints:     20,
		CurrentHitPoints: 20,
		Class:            characters.ClassWarrior,
		Attributes: []characters.Attribute{
			{Short: "STR", Value: 10},
			{Short: "DEX", Value: 10},
			{Short: "INT", Value: 10},
			{Short: "WIS", Value: 10},
			{Short: "STA", Value: 10},
		},
	}
	if leveling.MaybeLevelUp(char) != nil || char.Level != 1 {
		t.Fatalf("trainer mode applied a level: %+v", char.Level)
	}
	got := leveling.ApplyPendingLevels(char)
	if got == nil || got.NewLevel != 2 || char.Level != 2 {
		t.Fatalf("pending = %+v level=%d", got, char.Level)
	}
}

func TestCustomCurveAndCap(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	ruleset.SetXPRequired(map[int32]int32{1: 0, 2: 40, 3: 90})
	ruleset.SetLevelCap(3)
	if leveling.GetXPRequired(2) != 40 || leveling.GetXPRequired(10) != leveling.CalculateXPRequired(10) {
		t.Fatalf("curve 2=%d 10=%d formula=%d", leveling.GetXPRequired(2), leveling.GetXPRequired(10), leveling.CalculateXPRequired(10))
	}
	char := &characters.Character{Entity: &entities.Entity{ID: "c"}, Level: 2, XP: 90, Class: characters.ClassWarrior}
	levels, next := leveling.CheckLevelUp(char)
	if levels != 1 || next != 3 {
		t.Fatalf("levels=%d next=%d", levels, next)
	}
	char.Level = 3
	char.XP = 500
	levels, next = leveling.CheckLevelUp(char)
	if levels != 0 || next != 3 {
		t.Fatalf("at cap levels=%d next=%d", levels, next)
	}
}

func TestNextResetAndNewDay(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	ruleset.SetDeath(ruleset.DeathPolicy{
		XPLossPercent:    25,
		GoldLossPercent:  50,
		Respawn:          ruleset.RespawnNextReset,
		RespawnHPPercent: 0,
		DamageArmor:      false,
	})
	char := &characters.Character{
		Entity:           &entities.Entity{ID: "c"},
		XP:               100,
		Gold:             10,
		MaxHitPoints:     20,
		CurrentHitPoints: 4,
		BoundRoomID:      "town",
	}
	char.CurrentRoomID = "wild"
	out := ruleset.ApplyDeath(char)
	if out.RespawnRoomID != "" || !char.AwaitingReset || char.CurrentHitPoints != 0 {
		t.Fatalf("reset outcome %+v hp=%d room=%s", out, char.CurrentHitPoints, char.CurrentRoomID)
	}
	if char.XP != 75 || char.Gold != 5 {
		t.Fatalf("xp=%d gold=%d", char.XP, char.Gold)
	}
	ruleset.SetNewDay(true, "Europe/Berlin")
	// 2026-09-26 23:30 UTC is 2026-09-27 in Berlin.
	now := time.Date(2026, 9, 26, 23, 30, 0, 0, time.UTC)
	if !ruleset.ApplyNewDay(char, now) || char.LastResetDay != "2026-09-27" || char.AwaitingReset {
		t.Fatalf("dawn %+v", char)
	}
	if char.CurrentHitPoints != 20 {
		t.Fatalf("hp = %d", char.CurrentHitPoints)
	}
	if ruleset.ApplyNewDay(char, now) {
		t.Fatal("same day healed twice")
	}
}

func TestResourceCatalog(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	err := ruleset.LoadBytes([]byte(`
resources:
  gatherings:
    allowance: 4
    reset: interval
    interval: 2h
`))
	if err != nil {
		t.Fatal(err)
	}
	list := ruleset.ResourceAllowances()
	if len(list) != 1 || list[0].Key != "gatherings" || list[0].Amount != 4 || list[0].Interval != 2*time.Hour {
		t.Fatalf("%+v", list)
	}
}
