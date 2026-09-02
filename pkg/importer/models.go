package importer

// YAML model definitions for importing world data
// These match the structure of YAML files in the import folder

// YAMLRoom represents a room in YAML format
type YAMLRoom struct {
	ID          string         `yaml:"id"`
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Detail      string         `yaml:"detail"`
	Area        string         `yaml:"area"`
	Tags        []string       `yaml:"tags"`
	CanBind     bool           `yaml:"canBind"`
	Coords      *YAMLCoords    `yaml:"coords"`
	Exits       []YAMLExit     `yaml:"exits"`
	Actions     []YAMLAction   `yaml:"actions"`
	Items       []YAMLRoomItem `yaml:"items"`
	Meta        YAMLRoomMeta   `yaml:"meta"`
	OnEnter     string         `yaml:"onEnterScript"`
}

// YAMLRoomItem represents an item placed in a room.
// Supports both string format ("ITM0004") and object format ({id: "ITM0004", location: "on table"}).
type YAMLRoomItem struct {
	ID       string `yaml:"id"`
	Location string `yaml:"location"`
}

// UnmarshalYAML handles both string and object formats for room items
func (ri *YAMLRoomItem) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try plain string first: "ITM0004"
	var s string
	if err := unmarshal(&s); err == nil {
		ri.ID = s
		return nil
	}

	// Try object format: {id: "ITM0004", location: "on table"}
	var obj struct {
		ID       string `yaml:"id"`
		Location string `yaml:"location"`
	}
	if err := unmarshal(&obj); err != nil {
		return err
	}
	ri.ID = obj.ID
	ri.Location = obj.Location
	return nil
}

// YAMLExit represents a room exit
type YAMLExit struct {
	Name        string `yaml:"name"`
	Target      string `yaml:"target"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	Hidden      bool   `yaml:"hidden"`
	Instance    bool   `yaml:"instance"`
}

// YAMLAction represents a room action
type YAMLAction struct {
	Name        string                 `yaml:"name"`
	Type        string                 `yaml:"type"`
	Description string                 `yaml:"description"`
	Response    string                 `yaml:"response"`
	ScriptId    string                 `yaml:"scriptId"`
	Params      map[string]interface{} `yaml:"params"`
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
	ID           string                 `yaml:"id"`
	Name         string                 `yaml:"name"`
	Description  string                 `yaml:"description"`
	Detail       string                 `yaml:"detail"`
	Type         string                 `yaml:"type"`
	SubType      string                 `yaml:"subType"`
	Slot         string                 `yaml:"slot"`
	Quality      string                 `yaml:"quality"`
	Level        int32                  `yaml:"level"`
	BasePrice    int64                  `yaml:"basePrice"`
	Stackable    bool                   `yaml:"stackable"`
	MaxStack     int32                  `yaml:"maxStack"`
	Consumable   bool                   `yaml:"consumable"`
	CopyOnPickup bool                   `yaml:"copyOnPickup"`
	Tags         []string               `yaml:"tags"`
	Attributes   map[string]interface{} `yaml:"attributes"`
	Properties   map[string]interface{} `yaml:"properties"`
	Meta         YAMLItemMeta           `yaml:"meta"`
	OnUseScript  string                 `yaml:"onUseScript"`
}

// YAMLItemMeta contains item metadata
type YAMLItemMeta struct {
	Img string `yaml:"img"`
}

// YAMLRace represents NPC race info
type YAMLRace struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

// YAMLClass represents NPC class info
type YAMLClass struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

// YAMLNPC represents an NPC in YAML format
type YAMLNPC struct {
	ID            string             `yaml:"id"`
	Name          string             `yaml:"name"`
	Description   string             `yaml:"description"`
	Detail        string             `yaml:"detail"`
	Type          string             `yaml:"type"`
	Tags          []string           `yaml:"tags"`
	Level         int32              `yaml:"level"`
	MaxHitPoints  int32              `yaml:"maxHitPoints"`
	Race          YAMLRace           `yaml:"race"`
	Class         YAMLClass          `yaml:"class"`
	SpawnRoomId   string             `yaml:"spawnRoomId"`
	DialogID      string             `yaml:"dialogID"`
	RespawnTime   string             `yaml:"respawnTime"`
	EnemyTrait    *YAMLEnemyTrait    `yaml:"enemyTrait"`
	MerchantTrait *YAMLMerchantTrait `yaml:"merchantTrait"`
	Meta          YAMLNPCMeta        `yaml:"meta"`
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
	LootTableID   string  `yaml:"lootTableId"`
	OnAggroScript string  `yaml:"onAggroScript"`
	OnDeathScript string  `yaml:"onDeathScript"`
	OnFleeScript  string  `yaml:"onFleeScript"`
}

// YAMLMerchantTrait contains merchant-specific configuration
type YAMLMerchantTrait struct {
	MerchantType   string                      `yaml:"merchantType"`
	BuyMultiplier  float64                     `yaml:"buyMultiplier"`
	SellMultiplier float64                     `yaml:"sellMultiplier"`
	AcceptedTypes  []string                    `yaml:"acceptedTypes"`
	Inventory      []YAMLMerchantInventoryItem `yaml:"inventory"`
}

// YAMLMerchantInventoryItem represents an item in a merchant's inventory
type YAMLMerchantInventoryItem struct {
	ItemTemplateID string `yaml:"itemTemplateId"`
	Stock          int32  `yaml:"stock"`
	BasePrice      int64  `yaml:"basePrice"`
	PriceOverride  int64  `yaml:"priceOverride"`
	Quantity       int32  `yaml:"quantity"`
	MaxQuantity    int32  `yaml:"maxQuantity"`
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
	ID          string                    `yaml:"id"`
	Name        string                    `yaml:"name"`
	Type        string                    `yaml:"type"`
	NPCRef      string                    `yaml:"npc_ref"`
	Description string                    `yaml:"description"`
	Barks       []YAMLBark                `yaml:"barks"`
	Tree        map[string]YAMLDialogNode `yaml:"tree"`
	Tags        []string                  `yaml:"tags"`
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
	QuestID    string   `yaml:"questId"`
	Action     string   `yaml:"action"`
}

// YAMLLootTable represents a loot table in YAML format
type YAMLLootTable struct {
	ID          string          `yaml:"id"`
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Entries     []YAMLLootEntry `yaml:"entries"`
}

// YAMLLootEntry represents a loot table entry (new format)
type YAMLLootEntry struct {
	ItemTemplateID string  `yaml:"itemTemplateId"`
	DropChance     float64 `yaml:"dropChance"`
	MinQuantity    int32   `yaml:"minQuantity"`
	MaxQuantity    int32   `yaml:"maxQuantity"`
	Guaranteed     bool    `yaml:"guaranteed"`
}

// YAMLSpawner represents an NPC spawner in YAML format
type YAMLSpawner struct {
	ID            string `yaml:"id"`
	Name          string `yaml:"name"`
	TemplateID    string `yaml:"templateId"`
	RoomID        string `yaml:"roomId"`
	MaxInstances  int    `yaml:"maxInstances"`
	SpawnInterval string `yaml:"spawnInterval"`
	InitialCount  int    `yaml:"initialCount"`
}

// YAMLQuest represents a quest in YAML format
type YAMLQuest struct {
	ID                 string               `yaml:"id"`
	Name               string               `yaml:"name"`
	Description        string               `yaml:"description"`
	Category           string               `yaml:"category"`
	Level              int32                `yaml:"level"`
	Repeatable         bool                 `yaml:"repeatable"`
	Source             YAMLQuestSource      `yaml:"source"`
	TurnIn             string               `yaml:"turnIn"`
	Objectives         []YAMLQuestObjective `yaml:"objectives"`
	Rewards            YAMLQuestRewards     `yaml:"rewards"`
	RequiredQuestIDs   []string             `yaml:"requiredQuestIds"`
	RequiredLevel      int32                `yaml:"requiredLevel"`
	AcceptDialogText   string               `yaml:"acceptDialogText"`
	ProgressDialogText string               `yaml:"progressDialogText"`
	CompleteDialogText string               `yaml:"completeDialogText"`
	OnAcceptScriptID   string               `yaml:"onAcceptScriptId"`
	OnCompleteScriptID string               `yaml:"onCompleteScriptId"`
}

// YAMLQuestSource represents how a quest is obtained
type YAMLQuestSource struct {
	Type   string `yaml:"type"`
	NPCID  string `yaml:"npcId"`
	ItemID string `yaml:"itemId"`
}

// YAMLQuestObjective represents a quest objective
type YAMLQuestObjective struct {
	ID               string `yaml:"id"`
	Type             string `yaml:"type"`
	Description      string `yaml:"description"`
	TargetID         string `yaml:"targetId"`
	TargetName       string `yaml:"targetName"`
	Amount           int32  `yaml:"amount"`
	DeliverToNPCID   string `yaml:"deliverToNpcId"`
	DeliverToNPCName string `yaml:"deliverToNpcName"`
	DialogNodeID     string `yaml:"dialogNodeId"`
	CheckScriptID    string `yaml:"checkScriptId"`
	Order            int32  `yaml:"order"`
}

// YAMLQuestRewards represents quest completion rewards
type YAMLQuestRewards struct {
	XP              int32    `yaml:"xp"`
	Gold            int64    `yaml:"gold"`
	ItemTemplateIDs []string `yaml:"itemTemplateIds"`
}

// YAMLSkill represents a skill in YAML format
type YAMLSkill struct {
	ID             string   `yaml:"id"`
	Name           string   `yaml:"name"`
	Description    string   `yaml:"description"`
	ClassIDs       []string `yaml:"classIds"`
	LevelRequired  int32    `yaml:"levelRequired"`
	ResourceType   string   `yaml:"resourceType"`
	ManaCost       int32    `yaml:"manaCost"`
	CooldownRounds int      `yaml:"cooldownRounds"`
	Target         string   `yaml:"target"`
	Effect         string   `yaml:"effect"`
	ScalingAttr    string   `yaml:"scalingAttr"`
	BasePower      int32    `yaml:"basePower"`
	ScalingFactor  float64  `yaml:"scalingFactor"`
	Duration       int      `yaml:"duration"`
	BuffStat       string   `yaml:"buffStat"`
	BuffPercent    float64  `yaml:"buffPercent"`
	IgnoresDefense bool     `yaml:"ignoresDefense"`
	HitCount       int      `yaml:"hitCount"`

	SecondaryEffect    string  `yaml:"secondaryEffect,omitempty"`
	SecondaryBasePower int32   `yaml:"secondaryBasePower,omitempty"`
	SecondaryScaling   float64 `yaml:"secondaryScaling,omitempty"`
	SecondaryTarget    string  `yaml:"secondaryTarget,omitempty"`
}

// ImportConfig contains configuration for the import process
type ImportConfig struct {
	StartRoomID string `yaml:"startRoomID"`
}
