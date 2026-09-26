package ruleset

import "github.com/talesmud/talesmud/pkg/entities/characters"

// DeathPolicy is applied only from combat defeat. It does not award XP or gold.
type DeathPolicy struct {
	XPLossPercent    float64
	GoldLossFlat     int64
	GoldLossPercent  float64
	Respawn          string
	RespawnHPPercent float64
	DamageArmor      bool
}

// DeathOutcome is the penalty already written onto the character.
// RespawnRoomID is empty when the character stays in the death room.
type DeathOutcome struct {
	XPLost        int32
	GoldLost      int64
	RespawnRoomID string
	DamageArmor   bool
	AwaitingReset bool
}

// ApplyDeath writes the configured penalty onto char.
// Relocation and armor durability stay with the caller.
func ApplyDeath(char *characters.Character) DeathOutcome {
	if char == nil {
		return DeathOutcome{}
	}
	mu.RLock()
	policy := current.death
	mu.RUnlock()

	out := DeathOutcome{DamageArmor: policy.DamageArmor}
	if policy.XPLossPercent > 0 && char.XP > 0 {
		var lost int32
		if policy.XPLossPercent == 10 {
			lost = int32(float64(char.XP) * 0.10)
		} else {
			lost = int32(float64(char.XP) * policy.XPLossPercent / 100)
		}
		if lost > char.XP {
			lost = char.XP
		}
		char.XP -= lost
		out.XPLost = lost
	}

	if policy.GoldLossPercent > 0 && char.Gold > 0 {
		lost := int64(float64(char.Gold) * policy.GoldLossPercent / 100)
		if lost > char.Gold {
			lost = char.Gold
		}
		char.Gold -= lost
		out.GoldLost = lost
	} else if policy.GoldLossFlat > 0 && char.Gold > 0 {
		lost := policy.GoldLossFlat
		if lost > char.Gold {
			lost = char.Gold
		}
		char.Gold -= lost
		out.GoldLost = lost
	}

	char.CurrentHitPoints = respawnHP(char.MaxHitPoints, policy.RespawnHPPercent)

	switch policy.Respawn {
	case RespawnNextReset:
		char.AwaitingReset = true
		out.AwaitingReset = true
	default:
		char.AwaitingReset = false
		if char.BoundRoomID != "" {
			out.RespawnRoomID = char.BoundRoomID
		}
	}
	return out
}

func respawnHP(max int32, percent float64) int32 {
	if max <= 0 || percent <= 0 {
		return 0
	}
	if percent == 50 {
		hp := max / 2
		if hp < 1 {
			hp = 1
		}
		return hp
	}
	hp := int32(float64(max) * percent / 100)
	if hp < 1 {
		hp = 1
	}
	if hp > max {
		hp = max
	}
	return hp
}
