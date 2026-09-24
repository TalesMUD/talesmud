package balance

import "testing"

func TestRewardMultiplierDefaults(t *testing.T) {
	cfg := &CombatBalanceConfig{RewardScale: defaultRewardScale()}
	if rewardMultiplier(cfg, ThreatGrey) != 0.15 {
		t.Fatalf("grey %v", rewardMultiplier(cfg, ThreatGrey))
	}
	if rewardMultiplier(cfg, ThreatGreen) != 0.60 {
		t.Fatalf("green %v", rewardMultiplier(cfg, ThreatGreen))
	}
	if rewardMultiplier(cfg, ThreatYellow) != 1 {
		t.Fatalf("yellow %v", rewardMultiplier(cfg, ThreatYellow))
	}
	if rewardMultiplier(cfg, ThreatSkull) != 2 {
		t.Fatalf("skull %v", rewardMultiplier(cfg, ThreatSkull))
	}
	if ScaleReward(20, 0.15) != 3 {
		t.Fatalf("20 * 0.15 = %d", ScaleReward(20, 0.15))
	}
	if ScaleReward(8, 0.15) != 1 {
		t.Fatalf("8 * 0.15 = %d", ScaleReward(8, 0.15))
	}
	if ScaleReward(20, 1) != 20 {
		t.Fatal("neutral scale changed the amount")
	}
	if ScaleReward(1, 0.01) != 1 {
		t.Fatal("positive trickle should stay at least 1")
	}
	if BonusReward(100, 0.5) != 50 {
		t.Fatalf("bonus %d", BonusReward(100, 0.5))
	}
	if BonusReward(0, 0.5) != 0 || BonusReward(10, 0) != 0 {
		t.Fatal("empty bonus should be 0")
	}
}

func TestRewardScaleConfigFileWired(t *testing.T) {
	if _, err := ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	if RewardMultiplier(ThreatOrange) != 1.25 {
		t.Fatalf("orange %v", RewardMultiplier(ThreatOrange))
	}
	if FirstKillBonusRate() != 0.5 {
		t.Fatalf("first kill %v", FirstKillBonusRate())
	}
}
