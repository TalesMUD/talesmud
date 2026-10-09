package npc

import (
	"math"
	"strings"
)

// DefaultLowHealthFraction is the onLowHealth line when the threshold is unset or <= 0.
const DefaultLowHealthFraction = 0.30

// Range represents a min/max range for random values (e.g., gold drops)
type Range struct {
	Min int32 `json:"min"`
	Max int32 `json:"max"`
}

// CreatureType constants for NPC classification
const (
	CreatureTypeBeast      = "beast"      // Animals, insects, natural creatures
	CreatureTypeHumanoid   = "humanoid"   // Goblins, orcs, bandits - use Race/Class on NPC
	CreatureTypeUndead     = "undead"     // Skeletons, zombies, ghosts
	CreatureTypeElemental  = "elemental"  // Fire, water, earth, air beings
	CreatureTypeConstruct  = "construct"  // Golems, animated objects
	CreatureTypeDemon      = "demon"      // Demons, devils, otherworldly beings
	CreatureTypeDragon     = "dragon"     // Dragons and dragonkin
	CreatureTypeAberration = "aberration" // Unnatural, eldritch creatures
)

// CombatStyle constants for enemy fighting approach
const (
	CombatStyleMelee  = "melee"  // Close-range physical attacks
	CombatStyleRanged = "ranged" // Bows, thrown weapons, spitting
	CombatStyleMagic  = "magic"  // Spells and magical attacks
	CombatStyleSwarm  = "swarm"  // Overwhelm with numbers (rats, insects)
	CombatStyleBrute  = "brute"  // Heavy, slow, powerful attacks
	CombatStyleAgile  = "agile"  // Fast, evasive, hit-and-run
)

// EnemyTrait contains enemy-specific configuration for NPCs
type EnemyTrait struct {
	// Classification
	// CreatureType categorizes what the enemy fundamentally is (e.g., "beast", "undead", "humanoid")
	// For humanoids, the NPC's Race/Class fields should also be set
	CreatureType string `json:"creatureType"`
	// CombatStyle describes how the enemy fights (e.g., "melee", "ranged", "magic", "swarm")
	CombatStyle string `json:"combatStyle"`
	// Difficulty indicates combat challenge level: "trivial", "easy", "normal", "hard", "boss"
	Difficulty string `json:"difficulty"`
	// ResetOnDisengage restores full HP when a fight ends with this NPC still
	// alive (players fled, died, disconnected or timed out). Nil uses the default:
	// on for difficulty "boss", off otherwise.
	ResetOnDisengage *bool `json:"resetOnDisengage,omitempty"`

	// Combat Stats (used by future combat system)
	// AttackPower is the base damage dealt
	AttackPower int32 `json:"attackPower"`
	// Defense reduces incoming damage
	Defense int32 `json:"defense"`
	// AttackSpeed is attacks per round. 0 or omitted is one swing and never holds.
	// Positive values clamp to 0.25–3. 1.0 is also one swing. 2.0 is two swings
	// on that attack action. Below 1, the enemy swings then holds for
	// round(1/speed) attack actions (period clamped to 2–4). The turn beat
	// does not change. See docs/COMBAT-BALANCE.md.
	AttackSpeed float64 `json:"attackSpeed"`

	// Behavior Configuration
	// AggroRadius is how many rooms away the NPC can detect players (0 = passive, must be attacked first)
	AggroRadius int `json:"aggroRadius"`
	// AggroOnSight starts a fight when a player enters this NPC's room, or when this
	// NPC arrives in a room that already has players. Sight is that room only.
	// AggroRadius is not a leash. Ruleset combat.aggro_on_sight can disable it.
	AggroOnSight bool `json:"aggroOnSight"`
	// CallForHelp if true, NPC will alert nearby enemies when attacked
	CallForHelp bool `json:"callForHelp"`
	// FleeThreshold is HP percentage at which NPC attempts to flee (0 = never flee)
	FleeThreshold float64 `json:"fleeThreshold"`

	// Rewards
	// XPReward is experience points granted on kill
	XPReward int64 `json:"xpReward"`
	// GoldDrop is the min/max gold dropped on death
	GoldDrop Range `json:"goldDrop"`
	// LootTableID references a loot table for item drops
	LootTableID string `json:"lootTableId,omitempty"`
	// GuaranteedLoot contains item template IDs that always drop on death
	GuaranteedLoot []string `json:"guaranteedLoot,omitempty"`
	// MaxDrops limits the number of items from the loot table (0 = unlimited)
	MaxDrops int32 `json:"maxDrops,omitempty"`

	// Scripts (Lua script IDs to execute on events)
	// OnAggroScript runs when NPC enters combat
	OnAggroScript string `json:"onAggroScript,omitempty"`
	// OnDeathScript runs when NPC is killed
	OnDeathScript string `json:"onDeathScript,omitempty"`
	// OnFleeScript runs when NPC starts fleeing
	OnFleeScript string `json:"onFleeScript,omitempty"`
	// OnLowHealthScript runs once when HP first drops below LowHealthThreshold
	// while the NPC is still alive. A hit that kills from above the line does not run it.
	OnLowHealthScript string `json:"onLowHealthScript,omitempty"`
	// LowHealthThreshold is a fraction of max HP in (0,1). 0 or unset uses
	// DefaultLowHealthFraction. Other out-of-range values clamp into (0,1).
	LowHealthThreshold float64 `json:"lowHealthThreshold,omitempty"`
}

// NormalizeLowHealthThreshold maps an authored fraction onto (0,1).
// <= 0 (including unset) is the default 0.30. >= 1 clamps to just under 1.
func NormalizeLowHealthThreshold(v float64) float64 {
	if v <= 0 || math.IsNaN(v) {
		return DefaultLowHealthFraction
	}
	if v >= 1 {
		return math.Nextafter(1, 0)
	}
	return v
}

// ResetsOnDisengage reports whether this NPC returns to full HP after a fight
// it survives. Content can set EnemyTrait.ResetOnDisengage; otherwise bosses do.
func (n *NPC) ResetsOnDisengage() bool {
	if n == nil || n.EnemyTrait == nil {
		return false
	}
	if n.EnemyTrait.ResetOnDisengage != nil {
		return *n.EnemyTrait.ResetOnDisengage
	}
	return strings.EqualFold(strings.TrimSpace(n.EnemyTrait.Difficulty), "boss")
}
