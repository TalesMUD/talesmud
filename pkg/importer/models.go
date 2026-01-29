package importer

// YAML model definitions for importing world data
// These match the structure of YAML files in the import folder

// YAMLRoom represents a room in YAML format
type YAMLRoom struct {
	ID          string       `yaml:"id"`
	Name        string       `yaml:"name"`
	Description string       `yaml:"description"`
	Detail      string       `yaml:"detail"`
	Area        string       `yaml:"area"`
	Tags        []string     `yaml:"tags"`
	CanBind     bool         `yaml:"canBind"`
	Coords      *YAMLCoords  `yaml:"coords"`
	Exits       []YAMLExit   `yaml:"exits"`
	Actions     []YAMLAction `yaml:"actions"`
	Meta        YAMLRoomMeta `yaml:"meta"`
	OnEnter     string       `yaml:"onEnterScript"`
}

// YAMLExit represents a room exit
type YAMLExit struct {
	Name        string `yaml:"name"`
	Target      string `yaml:"target"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	Hidden      bool   `yaml:"hidden"`
}

// YAMLAction represents a room action
type YAMLAction struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	Response    string `yaml:"response"`
}

// YAMLRoomMeta contains room metadata
type YAMLRoomMeta struct {
	Background string `yaml:"background"`
	Mood       string `yaml:"mood"`
}

// YAMLCoords represents room coordinates - supports both list [x,y,z] and object {x,y,z} formats
type YAMLCoords struct {
	X int32
	Y int32
	Z int32
}

// UnmarshalYAML handles both list and object formats for coordinates
func (c *YAMLCoords) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try list format first: [x, y, z]
	var list []int32
	if err := unmarshal(&list); err == nil && len(list) >= 2 {
		c.X = list[0]
		c.Y = list[1]
		if len(list) >= 3 {
			c.Z = list[2]
		}
		return nil
	}

	// Try object format: {x: 1, y: 2, z: 3}
	var obj struct {
		X int32 `yaml:"x"`
		Y int32 `yaml:"y"`
		Z int32 `yaml:"z"`
	}
	if err := unmarshal(&obj); err == nil {
		c.X = obj.X
		c.Y = obj.Y
		c.Z = obj.Z
		return nil
	}

	return nil // Return nil to allow missing coords
}

// YAMLItem represents an item in YAML format
type YAMLItem struct {
	ID          string       `yaml:"id"`
	Name        string       `yaml:"name"`
	Description string       `yaml:"description"`
	Detail      string       `yaml:"detail"`
	Type        string       `yaml:"type"`
	SubType     string       `yaml:"subType"`
	Slot        string       `yaml:"slot"`
	Quality     string       `yaml:"quality"`
	Level       int32        `yaml:"level"`
	BasePrice   int64        `yaml:"basePrice"`
	Stackable   bool         `yaml:"stackable"`
	MaxStack    int32        `yaml:"maxStack"`
	Consumable  bool         `yaml:"consumable"`
	Tags        []string     `yaml:"tags"`
	Meta        YAMLItemMeta `yaml:"meta"`
	OnUseScript string       `yaml:"onUseScript"`
}

// YAMLItemMeta contains item metadata
type YAMLItemMeta struct {
	Img string `yaml:"img"`
}

// YAMLNPC represents an NPC in YAML format
type YAMLNPC struct {
	ID            string         `yaml:"id"`
	Name          string         `yaml:"name"`
	Description   string         `yaml:"description"`
	Detail        string         `yaml:"detail"`
	Type          string         `yaml:"type"`
	Tags          []string       `yaml:"tags"`
	Level         int32          `yaml:"level"`
	MaxHitPoints  int32          `yaml:"maxHitPoints"`
	DialogID      string         `yaml:"dialogID"`
	EnemyTrait    *YAMLEnemyTrait `yaml:"enemyTrait"`
	MerchantTrait *YAMLMerchantTrait `yaml:"merchantTrait"`
	Meta          YAMLNPCMeta    `yaml:"meta"`
}

// YAMLEnemyTrait contains enemy-specific configuration
type YAMLEnemyTrait struct {
	CreatureType  string  `yaml:"creatureType"`
	CombatStyle   string  `yaml:"combatStyle"`
	Difficulty    string  `yaml:"difficulty"`
	AttackPower   int32   `yaml:"attackPower"`
	Defense       int32   `yaml:"defense"`
	AttackSpeed   float64 `yaml:"attackSpeed"`
	AggroRadius   int     `yaml:"aggroRadius"`
	AggroOnSight  bool    `yaml:"aggroOnSight"`
	CallForHelp   bool    `yaml:"callForHelp"`
	FleeThreshold float64 `yaml:"fleeThreshold"`
	XPReward      int64   `yaml:"xpReward"`
	LootTableID   string  `yaml:"lootTableID"`
}

// YAMLMerchantTrait contains merchant-specific configuration
type YAMLMerchantTrait struct {
	ShopName       string                      `yaml:"shopName"`
	BuyRate        float64                     `yaml:"buyRate"`
	SellRate       float64                     `yaml:"sellRate"`
	Inventory      []YAMLMerchantInventoryItem `yaml:"inventory"`
}

// YAMLMerchantInventoryItem represents an item in a merchant's inventory
type YAMLMerchantInventoryItem struct {
	ItemTemplateID string `yaml:"itemTemplateID"`
	Stock          int32  `yaml:"stock"`
}

// YAMLNPCMeta contains NPC metadata
type YAMLNPCMeta struct {
	Img string `yaml:"img"`
}

// YAMLScript represents a script in YAML format
type YAMLScript struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
	Language    string `yaml:"language"`
	Code        string `yaml:"code"`
}

// YAMLDialog represents a dialog tree in YAML format
type YAMLDialog struct {
	ID          string                   `yaml:"id"`
	Name        string                   `yaml:"name"`
	Type        string                   `yaml:"type"`
	NPCRef      string                   `yaml:"npc_ref"`
	Description string                   `yaml:"description"`
	Barks       []YAMLBark               `yaml:"barks"`
	Tree        map[string]YAMLDialogNode `yaml:"tree"`
	Tags        []string                 `yaml:"tags"`
}

// YAMLBark represents an idle/bark line
type YAMLBark struct {
	Text       string   `yaml:"text"`
	Conditions []string `yaml:"conditions"`
	Weight     int      `yaml:"weight"`
}

// YAMLDialogNode represents a node in the dialog tree
type YAMLDialogNode struct {
	NPCText string             `yaml:"npc_text"`
	Options []YAMLDialogOption `yaml:"options"`
}

// YAMLDialogOption represents a player's dialog choice
type YAMLDialogOption struct {
	PlayerText string   `yaml:"player_text"`
	Next       string   `yaml:"next"`
	Conditions []string `yaml:"conditions"`
}

// YAMLLootTable represents a loot table in YAML format
type YAMLLootTable struct {
	ID          string           `yaml:"id"`
	Name        string           `yaml:"name"`
	Type        string           `yaml:"type"`
	Description string           `yaml:"description"`
	Entries     []YAMLLootEntry  `yaml:"entries"`
	Guaranteed  []string         `yaml:"guaranteed"`
	GoldRange   []int32          `yaml:"gold_range"`
	Tags        []string         `yaml:"tags"`
	Flags       YAMLLootFlags    `yaml:"flags"`
}

// YAMLLootEntry represents a loot table entry
type YAMLLootEntry struct {
	Item       string   `yaml:"item"`
	Weight     int      `yaml:"weight"`
	MinCount   int32    `yaml:"min_count"`
	MaxCount   int32    `yaml:"max_count"`
	Conditions []string `yaml:"conditions"`
}

// YAMLLootFlags contains loot table flags
type YAMLLootFlags struct {
	ScalesWithLevel bool `yaml:"scales_with_level"`
	BossLoot        bool `yaml:"boss_loot"`
}

// ImportConfig contains configuration for the import process
type ImportConfig struct {
	StartRoomID string `yaml:"startRoomID"`
}
