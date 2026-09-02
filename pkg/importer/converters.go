package importer

import (
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
	"github.com/talesmud/talesmud/pkg/scripts"
)

// ToEntity converts a YAMLRoom to a Room entity
func (y *YAMLRoom) ToEntity() *rooms.Room {
	room := &rooms.Room{
		Entity:          &entities.Entity{ID: y.ID},
		LookAt:          traits.LookAt{Detail: y.Detail},
		Name:            y.Name,
		Description:     y.Description,
		Area:            y.Area,
		Tags:            y.Tags,
		CanBind:         y.CanBind,
		OnEnterScriptID: y.OnEnter,
	}

	// Convert exits
	if len(y.Exits) > 0 {
		exits := make(rooms.Exits, len(y.Exits))
		for i, e := range y.Exits {
			exits[i] = rooms.Exit{
				Name:        e.Name,
				Target:      e.Target,
				Type:        rooms.RoomExitType(e.Type),
				Description: e.Description,
				Hidden:      e.Hidden,
				Instance:    e.Instance,
			}
		}
		room.Exits = &exits
	}

	// Convert actions
	if len(y.Actions) > 0 {
		actions := make(rooms.Actions, len(y.Actions))
		for i, a := range y.Actions {
			actions[i] = rooms.Action{
				Name:        a.Name,
				Type:        rooms.RoomActionType(a.Type),
				Description: a.Description,
				Response:    a.Response,
				ScriptId:    a.ScriptId,
				Params:      a.Params,
			}
		}
		room.Actions = &actions
	}

	// Convert room items (place item template IDs into the room)
	if len(y.Items) > 0 {
		itemIDs := make(rooms.Items, len(y.Items))
		for i, item := range y.Items {
			itemIDs[i] = item.ID
		}
		room.Items = &itemIDs
	}

	// Set meta if background is provided
	if y.Meta.Background != "" || y.Meta.Mood != "" {
		room.Meta = &struct {
			Mood       string `bson:"mood,omitempty" json:"mood,omitempty"`
			Background string `bson:"background,omitempty" json:"background,omitempty"`
		}{
			Mood:       y.Meta.Mood,
			Background: y.Meta.Background,
		}
	}

	// Set coords if provided
	if y.Coords != nil {
		room.Coords = &struct {
			X int32 `bson:"x" json:"x"`
			Y int32 `bson:"y" json:"y"`
			Z int32 `bson:"z" json:"z"`
		}{
			X: y.Coords.X,
			Y: y.Coords.Y,
			Z: y.Coords.Z,
		}
	}

	return room
}

// ToEntity converts a YAMLItem to an Item entity
func (y *YAMLItem) ToEntity() *items.Item {
	item := &items.Item{
		Entity:        &entities.Entity{ID: y.ID},
		LookAt:        traits.LookAt{Detail: y.Detail},
		IsTemplate:    true, // Imported items are always templates
		Name:          y.Name,
		Description:   y.Description,
		Type:          items.ItemType(y.Type),
		SubType:       items.ItemSubType(y.SubType),
		Slot:          items.ItemSlot(y.Slot),
		Quality:       items.ItemQuality(y.Quality),
		Level:         y.Level,
		BasePrice:     y.BasePrice,
		Stackable:     y.Stackable,
		MaxStack:      y.MaxStack,
		Consumable:    y.Consumable,
		CopyOnPickup:  y.CopyOnPickup,
		Tags:          y.Tags,
		Attributes:    y.Attributes,
		Properties:    y.Properties,
		OnUseScriptID: y.OnUseScript,
	}

	// Set meta if img is provided
	if y.Meta.Img != "" {
		item.Meta = &struct {
			Img string `bson:"img,omitempty" json:"img,omitempty"`
		}{
			Img: y.Meta.Img,
		}
	}

	return item
}

// ToEntity converts a YAMLNPC to an NPC entity
func (y *YAMLNPC) ToEntity() *npc.NPC {
	// Determine if this is a template:
	// - NPCs with spawnRoomId are unique (non-template) — they reside in their room
	// - NPCs without spawnRoomId are templates — they're spawned via spawners
	isTemplate := y.SpawnRoomId == ""

	n := &npc.NPC{
		Entity:           &entities.Entity{ID: y.ID},
		Name:             y.Name,
		Description:      y.Description,
		Level:            y.Level,
		MaxHitPoints:     y.MaxHitPoints,
		CurrentHitPoints: y.MaxHitPoints, // Start at full health
		DialogID:         y.DialogID,
		IsTemplate:       isTemplate,
		SpawnRoomID:      y.SpawnRoomId,
		State:            "idle",
	}

	// For unique NPCs (non-templates), set current room to spawn room
	if !isTemplate {
		n.CurrentRoomID = y.SpawnRoomId
	}

	// Set race if provided
	if y.Race.ID != "" {
		n.Race = characters.Race{
			ID:   y.Race.ID,
			Name: y.Race.Name,
		}
	}

	// Set class if provided
	if y.Class.ID != "" {
		n.Class = characters.Class{
			ID:   y.Class.ID,
			Name: y.Class.Name,
		}
	}

	// Convert enemy trait if present
	if y.EnemyTrait != nil {
		// Apply difficulty-based multipliers to base stats
		finalHP, finalAttack, finalDefense := balance.ApplyMultipliers(
			y.MaxHitPoints,
			y.EnemyTrait.AttackPower,
			y.EnemyTrait.Defense,
			y.EnemyTrait.Difficulty,
		)

		// Update NPC HP with multiplied value
		n.MaxHitPoints = finalHP
		n.CurrentHitPoints = finalHP

		n.EnemyTrait = &npc.EnemyTrait{
			CreatureType:  y.EnemyTrait.CreatureType,
			CombatStyle:   y.EnemyTrait.CombatStyle,
			Difficulty:    y.EnemyTrait.Difficulty,
			AttackPower:   finalAttack,
			Defense:       finalDefense,
			AttackSpeed:   y.EnemyTrait.AttackSpeed,
			AggroRadius:   y.EnemyTrait.AggroRadius,
			AggroOnSight:  y.EnemyTrait.AggroOnSight,
			CallForHelp:   y.EnemyTrait.CallForHelp,
			FleeThreshold: y.EnemyTrait.FleeThreshold,
			XPReward:      y.EnemyTrait.XPReward,
			LootTableID:   y.EnemyTrait.LootTableID,
			OnAggroScript: y.EnemyTrait.OnAggroScript,
			OnDeathScript: y.EnemyTrait.OnDeathScript,
			OnFleeScript:  y.EnemyTrait.OnFleeScript,
		}
	}

	// Convert merchant trait if present
	if y.MerchantTrait != nil {
		mt := &npc.MerchantTrait{
			MerchantType:   y.MerchantTrait.MerchantType,
			BuyMultiplier:  y.MerchantTrait.BuyMultiplier,
			SellMultiplier: y.MerchantTrait.SellMultiplier,
			AcceptedTypes:  append([]string{}, y.MerchantTrait.AcceptedTypes...),
			Inventory:      make([]npc.MerchantItem, 0),
		}
		// Convert inventory items
		for _, item := range y.MerchantTrait.Inventory {
			qty := item.Quantity
			if qty == 0 {
				qty = item.Stock
			}
			if qty == 0 {
				qty = -1 // Default to unlimited if not specified
			}
			maxQty := item.MaxQuantity
			if maxQty == 0 {
				maxQty = qty
			}
			mt.Inventory = append(mt.Inventory, npc.MerchantItem{
				ItemTemplateID: item.ItemTemplateID,
				BasePrice:      item.BasePrice,
				PriceOverride:  item.PriceOverride,
				Quantity:       qty,
				MaxQuantity:    maxQty,
			})
		}
		n.MerchantTrait = mt
	}

	return n
}

// ToEntity converts a YAMLScript to a Script entity
func (y *YAMLScript) ToEntity() *scripts.Script {
	return &scripts.Script{
		Entity:      &entities.Entity{ID: y.ID},
		Name:        y.Name,
		Description: y.Description,
		Type:        scripts.ScriptType(y.Type),
		Language:    scripts.ScriptLanguage(y.Language),
		Code:        y.Code,
	}
}

// ToEntity converts a YAMLDialog to a Dialog entity
// The YAML format uses a flat map of nodes, we need to convert to a tree structure
func (y *YAMLDialog) ToEntity() *dialogs.Dialog {
	dialog := &dialogs.Dialog{
		Entity: &entities.Entity{ID: y.ID},
		Name:   y.Name,
		NodeID: "root",
	}

	// Build the dialog tree from the flat map with visited tracking to prevent infinite recursion
	visited := make(map[string]bool)
	if rootNode, ok := y.Tree["root"]; ok {
		visited["root"] = true
		dialog.Text = rootNode.NPCText
		dialog.Options = convertDialogOptions(rootNode.Options, y.Tree, visited)
	}

	return dialog
}

// convertDialogOptions recursively converts dialog options with cycle detection
func convertDialogOptions(options []YAMLDialogOption, tree map[string]YAMLDialogNode, visited map[string]bool) []*dialogs.Dialog {
	if len(options) == 0 {
		return nil
	}

	result := make([]*dialogs.Dialog, 0, len(options))
	for _, opt := range options {
		optDialog := &dialogs.Dialog{
			NodeID:  opt.Next,
			Text:    opt.PlayerText,
			QuestID: opt.QuestID,
			Action:  opt.Action,
		}

		// If this option leads to another node and we haven't visited it, set up the answer
		if nextNode, ok := tree[opt.Next]; ok {
			if !visited[opt.Next] {
				// Mark as visited before recursing
				visited[opt.Next] = true
				optDialog.Answer = &dialogs.Dialog{
					NodeID:  opt.Next,
					Text:    nextNode.NPCText,
					Options: convertDialogOptions(nextNode.Options, tree, visited),
				}
				// Unmark after recursion to allow visiting from different paths
				delete(visited, opt.Next)

				// Mark as dialog exit if no options
				if len(nextNode.Options) == 0 {
					isExit := true
					optDialog.Answer.IsDialogExit = &isExit
				}
			} else {
				// Already visited - just reference the node without recursing
				optDialog.Answer = &dialogs.Dialog{
					NodeID: opt.Next,
					Text:   nextNode.NPCText,
				}
			}
		}

		result = append(result, optDialog)
	}

	return result
}

// ToEntity converts a YAMLLootTable to a LootTable entity
func (y *YAMLLootTable) ToEntity() *items.LootTable {
	lt := &items.LootTable{
		Entity:         &entities.Entity{ID: y.ID},
		Name:           y.Name,
		Description:    y.Description,
		Entries:        make([]items.LootEntry, 0),
		GoldMultiplier: 1.0,
		DropBonus:      0.0,
	}

	// Convert entries (new format: itemTemplateId, dropChance, minQuantity, maxQuantity, guaranteed)
	for _, e := range y.Entries {
		minQty := e.MinQuantity
		if minQty == 0 {
			minQty = 1
		}
		maxQty := e.MaxQuantity
		if maxQty == 0 {
			maxQty = 1
		}

		lt.Entries = append(lt.Entries, items.LootEntry{
			ItemTemplateID: e.ItemTemplateID,
			DropChance:     e.DropChance,
			MinQuantity:    minQty,
			MaxQuantity:    maxQty,
			Guaranteed:     e.Guaranteed,
		})
	}

	return lt
}

// parseDuration parses a duration string like "5m" or "30s"
func parseDuration(s string) time.Duration {
	if s == "" {
		return 5 * time.Minute // default
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 5 * time.Minute
	}
	return d
}

// ToEntity converts a YAMLSpawner to an NPCSpawner entity
func (y *YAMLSpawner) ToEntity() *npc.NPCSpawner {
	initialCount := y.InitialCount
	if initialCount == 0 {
		initialCount = 1
	}

	return &npc.NPCSpawner{
		Entity:        &entities.Entity{ID: y.ID},
		Name:          y.Name,
		TemplateID:    y.TemplateID,
		RoomID:        y.RoomID,
		MaxInstances:  y.MaxInstances,
		SpawnInterval: parseDuration(y.SpawnInterval),
		InitialCount:  initialCount,
		Created:       time.Now(),
	}
}

// ToEntity converts a YAMLQuest to a Quest entity
func (y *YAMLQuest) ToEntity() *quests.Quest {
	quest := &quests.Quest{
		Entity:             &entities.Entity{ID: y.ID},
		Name:               y.Name,
		Description:        y.Description,
		Category:           y.Category,
		Level:              y.Level,
		Repeatable:         y.Repeatable,
		RequiredQuestIDs:   y.RequiredQuestIDs,
		RequiredLevel:      y.RequiredLevel,
		AcceptDialogText:   y.AcceptDialogText,
		ProgressDialogText: y.ProgressDialogText,
		CompleteDialogText: y.CompleteDialogText,
		OnAcceptScriptID:   y.OnAcceptScriptID,
		OnCompleteScriptID: y.OnCompleteScriptID,
		TurnIn:             y.TurnIn,
		Created:            time.Now(),
		Updated:            time.Now(),
	}

	// Convert source
	quest.Source = quests.QuestSource{
		Type:   y.Source.Type,
		NPCID:  y.Source.NPCID,
		ItemID: y.Source.ItemID,
	}

	// Convert objectives
	if len(y.Objectives) > 0 {
		quest.Objectives = make([]quests.Objective, len(y.Objectives))
		for i, obj := range y.Objectives {
			quest.Objectives[i] = quests.Objective{
				ID:               obj.ID,
				Type:             quests.ObjectiveType(obj.Type),
				Description:      obj.Description,
				TargetID:         obj.TargetID,
				TargetName:       obj.TargetName,
				Amount:           obj.Amount,
				DeliverToNPCID:   obj.DeliverToNPCID,
				DeliverToNPCName: obj.DeliverToNPCName,
				DialogNodeID:     obj.DialogNodeID,
				CheckScriptID:    obj.CheckScriptID,
				Order:            obj.Order,
			}
		}
	}

	// Convert rewards
	quest.Rewards = quests.Reward{
		XP:              y.Rewards.XP,
		Gold:            y.Rewards.Gold,
		ItemTemplateIDs: y.Rewards.ItemTemplateIDs,
	}

	return quest
}

// ToEntity converts a YAMLSkill to a Skill entity
func (y *YAMLSkill) ToEntity() *skills.Skill {
	s := &skills.Skill{
		Entity:         &entities.Entity{ID: y.ID},
		Name:           y.Name,
		Description:    y.Description,
		ClassIDs:       y.ClassIDs,
		LevelRequired:  y.LevelRequired,
		ResourceType:   skills.ResourceType(y.ResourceType),
		ManaCost:       y.ManaCost,
		CooldownRounds: y.CooldownRounds,
		Target:         skills.TargetType(y.Target),
		Effect:         skills.EffectType(y.Effect),
		ScalingAttr:    y.ScalingAttr,
		BasePower:      y.BasePower,
		ScalingFactor:  y.ScalingFactor,
		Duration:       y.Duration,
		BuffStat:       y.BuffStat,
		BuffPercent:    y.BuffPercent,
		IgnoresDefense: y.IgnoresDefense,
		HitCount:       y.HitCount,
	}

	if y.SecondaryEffect != "" {
		s.SecondaryEffect = skills.EffectType(y.SecondaryEffect)
		s.SecondaryBasePower = y.SecondaryBasePower
		s.SecondaryScaling = y.SecondaryScaling
		s.SecondaryTarget = skills.TargetType(y.SecondaryTarget)
	}

	return s
}
