package importer

import (
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/scripts"
)

// ToEntity converts a YAMLRoom to a Room entity
func (y *YAMLRoom) ToEntity() *rooms.Room {
	room := &rooms.Room{
		Entity:  &entities.Entity{ID: y.ID},
		LookAt:  traits.LookAt{Detail: y.Detail},
		Name:    y.Name,
		Description: y.Description,
		Area:    y.Area,
		Tags:    y.Tags,
		CanBind: y.CanBind,
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
			}
		}
		room.Actions = &actions
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
		Entity:      &entities.Entity{ID: y.ID},
		LookAt:      traits.LookAt{Detail: y.Detail},
		IsTemplate:  true, // Imported items are always templates
		Name:        y.Name,
		Description: y.Description,
		Type:        items.ItemType(y.Type),
		SubType:     items.ItemSubType(y.SubType),
		Slot:        items.ItemSlot(y.Slot),
		Quality:     items.ItemQuality(y.Quality),
		Level:       y.Level,
		BasePrice:   y.BasePrice,
		Stackable:   y.Stackable,
		MaxStack:    y.MaxStack,
		Consumable:  y.Consumable,
		Tags:        y.Tags,
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
	n := &npc.NPC{
		Entity:           &entities.Entity{ID: y.ID},
		Name:             y.Name,
		Description:      y.Description,
		Level:            y.Level,
		MaxHitPoints:     y.MaxHitPoints,
		CurrentHitPoints: y.MaxHitPoints, // Start at full health
		DialogID:         y.DialogID,
		IsTemplate:       true, // Imported NPCs are templates
		State:            "idle",
	}

	// Convert enemy trait if present
	if y.EnemyTrait != nil {
		n.EnemyTrait = &npc.EnemyTrait{
			CreatureType:  y.EnemyTrait.CreatureType,
			CombatStyle:   y.EnemyTrait.CombatStyle,
			Difficulty:    y.EnemyTrait.Difficulty,
			AttackPower:   y.EnemyTrait.AttackPower,
			Defense:       y.EnemyTrait.Defense,
			AttackSpeed:   y.EnemyTrait.AttackSpeed,
			AggroRadius:   y.EnemyTrait.AggroRadius,
			AggroOnSight:  y.EnemyTrait.AggroOnSight,
			CallForHelp:   y.EnemyTrait.CallForHelp,
			FleeThreshold: y.EnemyTrait.FleeThreshold,
			XPReward:      y.EnemyTrait.XPReward,
			LootTableID:   y.EnemyTrait.LootTableID,
		}
	}

	// Convert merchant trait if present
	if y.MerchantTrait != nil {
		mt := &npc.MerchantTrait{
			BuyMultiplier:  y.MerchantTrait.BuyRate,
			SellMultiplier: y.MerchantTrait.SellRate,
			Inventory:      make([]npc.MerchantItem, 0),
		}
		// Convert inventory items
		for _, item := range y.MerchantTrait.Inventory {
			stock := item.Stock
			if stock == 0 {
				stock = -1 // Default to unlimited if not specified
			}
			mt.Inventory = append(mt.Inventory, npc.MerchantItem{
				ItemTemplateID: item.ItemTemplateID,
				Quantity:       stock,
				MaxQuantity:    stock,
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
			NodeID: opt.Next,
			Text:   opt.PlayerText,
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
		Entity:      &entities.Entity{ID: y.ID},
		Name:        y.Name,
		Description: y.Description,
		Entries:     make([]items.LootEntry, 0),
		GoldMultiplier: 1.0,
		DropBonus:      0.0,
	}

	// Convert entries
	for _, e := range y.Entries {
		// Convert weight (0-100) to drop chance (0.0-1.0)
		dropChance := float64(e.Weight) / 100.0
		if dropChance > 1.0 {
			dropChance = 1.0
		}

		minQty := e.MinCount
		if minQty == 0 {
			minQty = 1
		}
		maxQty := e.MaxCount
		if maxQty == 0 {
			maxQty = 1
		}

		lt.Entries = append(lt.Entries, items.LootEntry{
			ItemTemplateID: e.Item,
			DropChance:     dropChance,
			MinQuantity:    minQty,
			MaxQuantity:    maxQty,
			Guaranteed:     e.Weight >= 100,
		})
	}

	// Add guaranteed items
	for _, itemID := range y.Guaranteed {
		lt.Entries = append(lt.Entries, items.LootEntry{
			ItemTemplateID: itemID,
			DropChance:     1.0,
			MinQuantity:    1,
			MaxQuantity:    1,
			Guaranteed:     true,
		})
	}

	return lt
}
