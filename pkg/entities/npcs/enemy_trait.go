package npc

// Range represents a min/max range for random values (e.g., gold drops)
type Range struct {
	Min int32 `json:"min"`
	Max int32 `json:"max"`
}

// CreatureType constants for NPC classification
const (
	CreatureTypeBeast     = "beast"     // Animals, insects, natural creatures
	CreatureTypeHumanoid  = "humanoid"  // Goblins, orcs, bandits - use Race/Class on NPC
	CreatureTypeUndead    = "undead"    // Skeletons, zombies, ghosts
	CreatureTypeElemental = "elemental" // Fire, water, earth, air beings
	CreatureTypeConstruct = "construct" // Golems, animated objects
	CreatureTypeDemon     = "demon"     // Demons, devils, otherworldly beings
	CreatureTypeDragon    = "dragon"    // Dragons and dragonkin
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

	// Combat Stats (used by future combat system)
	// AttackPower is the base damage dealt
	AttackPower int32 `json:"attackPower"`
	// Defense reduces incoming damage
	Defense int32 `json:"defense"`
	// AttackSpeed is attacks per second (e.g., 1.0 = one attack per second)
	AttackSpeed float64 `json:"attackSpeed"`

	// Behavior Configuration
	// AggroRadius is how many rooms away the NPC can detect players (0 = passive, must be attacked first)
	AggroRadius int `json:"aggroRadius"`
	// AggroOnSight if true, NPC will auto-attack players on sight within aggro radius
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
}
