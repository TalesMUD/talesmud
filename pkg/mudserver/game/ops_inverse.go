package game

import (
	"bytes"
	"encoding/json"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/audit"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
)

// ApplyInverse runs the stored inverse of an ops audit row.
// The current world must still match the snapshot taken after the original op.
func (g *Game) ApplyInverse(inv *audit.Inverse) (*OpResult, error) {
	if g == nil || g.Facade == nil {
		return nil, opErr(503, "game is not running")
	}
	if inv == nil || inv.Action == "" {
		return nil, opErr(400, "this change cannot be undone")
	}
	switch inv.Action {
	case "teleport":
		return g.inverseTeleport(inv.Body)
	case "take-added":
		return g.inverseTakeAdded(inv.Body)
	case "restore-items":
		return g.inverseRestoreItems(inv.Body)
	case "restore-hp":
		return g.inverseRestoreHP(inv.Body)
	case "npc-despawn":
		return g.inverseDespawn(inv.Body)
	case "restore-npc":
		return g.inverseRestoreNPC(inv.Body)
	case "restore-progress":
		return g.inverseRestoreProgress(inv.Body)
	default:
		return nil, opErr(400, "this change cannot be undone")
	}
}

func (g *Game) inverseTeleport(body json.RawMessage) (*OpResult, error) {
	var in struct {
		CharacterID  string `json:"characterId"`
		RoomID       string `json:"roomId"`
		ExpectRoomID string `json:"expectRoomId"`
	}
	if err := decodeOp(body, &in); err != nil {
		return nil, err
	}
	char, err := g.Facade.CharactersService().FindByID(in.CharacterID)
	if err != nil || char == nil {
		return nil, opErr(404, "character not found")
	}
	if in.ExpectRoomID != "" && char.CurrentRoomID != in.ExpectRoomID {
		return nil, opErr(409, "the entity changed since this change; undo was refused")
	}
	if g.characterInCombat(char) {
		return nil, opErr(409, "the entity changed since this change; undo was refused")
	}
	return g.OpTeleport(in.CharacterID, in.RoomID, false)
}

func (g *Game) inverseTakeAdded(body json.RawMessage) (*OpResult, error) {
	var in struct {
		CharacterID string       `json:"characterId"`
		InstanceIDs []string     `json:"instanceIds"`
		StackDeltas []stackDelta `json:"stackDeltas"`
	}
	if err := decodeOp(body, &in); err != nil {
		return nil, err
	}
	char, err := g.Facade.CharactersService().FindByID(in.CharacterID)
	if err != nil || char == nil {
		return nil, opErr(404, "character not found")
	}
	for _, id := range in.InstanceIDs {
		if !holdsInstance(char, id) {
			return nil, opErr(409, "the entity changed since this change; undo was refused")
		}
	}
	qty := itemQtyByID(char)
	for _, delta := range in.StackDeltas {
		if qty[delta.ItemID] < delta.Quantity {
			return nil, opErr(409, "the entity changed since this change; undo was refused")
		}
	}
	var removed []takenPiece
	err = g.Facade.CharactersService().Modify(in.CharacterID, func(ch *characters.Character) error {
		for _, id := range in.InstanceIDs {
			piece, takeErr := takeInstance(ch, id, 0)
			if takeErr != nil {
				return opErr(409, "the entity changed since this change; undo was refused")
			}
			removed = append(removed, piece...)
		}
		for _, delta := range in.StackDeltas {
			piece, takeErr := takeInstance(ch, delta.ItemID, delta.Quantity)
			if takeErr != nil {
				return opErr(409, "the entity changed since this change; undo was refused")
			}
			removed = append(removed, piece...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	fresh, _ := g.Facade.CharactersService().FindByID(in.CharacterID)
	g.notifyInventory(fresh, "An operator undid a gift.")
	return &OpResult{
		Summary:    opSummary("Removed the items given to %s.", characterName(fresh)),
		Undoable:   true,
		EntityType: "character",
		EntityID:   in.CharacterID,
		Before:     jsonRaw(map[string]interface{}{"instanceIds": in.InstanceIDs}),
		After:      jsonRaw(map[string]interface{}{"removed": len(removed)}),
		Inverse: inverseOf("restore-items", map[string]interface{}{
			"characterId": in.CharacterID,
			"items":       removed,
		}),
		Detail: map[string]interface{}{"characterId": in.CharacterID},
	}, nil
}

func (g *Game) inverseRestoreItems(body json.RawMessage) (*OpResult, error) {
	var in struct {
		CharacterID string       `json:"characterId"`
		Items       []takenPiece `json:"items"`
	}
	if err := decodeOp(body, &in); err != nil {
		return nil, err
	}
	char, err := g.Facade.CharactersService().FindByID(in.CharacterID)
	if err != nil || char == nil {
		return nil, opErr(404, "character not found")
	}
	for _, piece := range in.Items {
		if piece.Item == nil {
			continue
		}
		if piece.Partial {
			if !holdsInstance(char, piece.Item.ID) {
				return nil, opErr(409, "the entity changed since this change; undo was refused")
			}
			continue
		}
		if holdsInstance(char, piece.Item.ID) {
			return nil, opErr(409, "the entity changed since this change; undo was refused")
		}
		if piece.Item.Unique && char.CountOfTemplate(items.TemplateKey(piece.Item)) >= 1 {
			return nil, opErr(409, "the entity changed since this change; undo was refused")
		}
	}
	var added []string
	var deltas []stackDelta
	err = g.Facade.CharactersService().Modify(in.CharacterID, func(ch *characters.Character) error {
		before := itemQtyByID(ch)
		for _, piece := range in.Items {
			if piece.Item == nil {
				continue
			}
			if piece.Partial {
				if !addQuantity(ch, piece.Item.ID, piece.Item.Quantity) {
					return opErr(409, "the entity changed since this change; undo was refused")
				}
				continue
			}
			if piece.EquippedSlot != "" && (ch.EquippedItems == nil || ch.EquippedItems[piece.EquippedSlot] == nil) {
				if ch.EquippedItems == nil {
					ch.EquippedItems = map[items.ItemSlot]*items.Item{}
				}
				ch.EquippedItems[piece.EquippedSlot] = piece.Item
				continue
			}
			if addErr := ch.Inventory.AddItem(piece.Item); addErr != nil {
				return opErr(400, addErr.Error())
			}
		}
		added, deltas = inventoryDelta(ch, before)
		return nil
	})
	if err != nil {
		return nil, err
	}
	fresh, _ := g.Facade.CharactersService().FindByID(in.CharacterID)
	g.notifyInventory(fresh, "An operator restored an item.")
	return &OpResult{
		Summary:    opSummary("Restored items to %s.", characterName(fresh)),
		Undoable:   true,
		EntityType: "character",
		EntityID:   in.CharacterID,
		Inverse: inverseOf("take-added", map[string]interface{}{
			"characterId": in.CharacterID,
			"instanceIds": added,
			"stackDeltas": deltas,
		}),
		Detail: map[string]interface{}{"characterId": in.CharacterID, "instanceIds": added},
	}, nil
}

func addQuantity(ch *characters.Character, id string, qty int32) bool {
	if qty < 1 {
		qty = 1
	}
	var found bool
	walkItems(ch.Inventory.Items, func(item *items.Item) {
		if item.ID == id {
			item.Quantity = itemQty(item) + qty
			found = true
		}
	})
	for _, item := range ch.EquippedItems {
		if item != nil && item.ID == id {
			item.Quantity = itemQty(item) + qty
			found = true
		}
	}
	return found
}

func (g *Game) inverseRestoreHP(body json.RawMessage) (*OpResult, error) {
	var in struct {
		NPCInstanceID string `json:"npcInstanceId"`
		HP            int32  `json:"hp"`
		ExpectHP      int32  `json:"expectHp"`
	}
	if err := decodeOp(body, &in); err != nil {
		return nil, err
	}
	inst := g.NPCManager.GetInstance(in.NPCInstanceID)
	if inst == nil || inst.IsDead {
		return nil, opErr(409, "the entity changed since this change; undo was refused")
	}
	if g.CombatController != nil && g.CombatController.IsNPCInCombat(in.NPCInstanceID) {
		return nil, opErr(409, "the entity changed since this change; undo was refused")
	}
	if inst.CurrentHitPoints != in.ExpectHP {
		return nil, opErr(409, "the entity changed since this change; undo was refused")
	}
	g.NPCManager.UpdateInstance(in.NPCInstanceID, func(n *npc.NPC) {
		n.CurrentHitPoints = in.HP
	})
	fresh := g.NPCManager.GetInstance(in.NPCInstanceID)
	g.persistNPCIfStored(fresh)
	g.NPCManager.Note(in.NPCInstanceID, "restored previous hit points")
	return &OpResult{
		Summary:    opSummary("Restored %s to %d hit points.", npcName(fresh), in.HP),
		Undoable:   true,
		EntityType: "npc",
		EntityID:   in.NPCInstanceID,
		Inverse: inverseOf("restore-hp", map[string]interface{}{
			"npcInstanceId": in.NPCInstanceID,
			"hp":            in.ExpectHP,
			"expectHp":      in.HP,
		}),
		Detail: map[string]interface{}{"npcInstanceId": in.NPCInstanceID, "hp": in.HP},
	}, nil
}

func (g *Game) inverseDespawn(body json.RawMessage) (*OpResult, error) {
	var in struct {
		NPCInstanceID string `json:"npcInstanceId"`
		ExpectRoomID  string `json:"expectRoomId"`
		ExpectHP      int32  `json:"expectHp"`
		ExpectAlive   bool   `json:"expectAlive"`
	}
	if err := decodeOp(body, &in); err != nil {
		return nil, err
	}
	inst := g.NPCManager.GetInstance(in.NPCInstanceID)
	if inst == nil || inst.IsDead != !in.ExpectAlive || inst.CurrentRoomID != in.ExpectRoomID || inst.CurrentHitPoints != in.ExpectHP {
		return nil, opErr(409, "the entity changed since this change; undo was refused")
	}
	if g.CombatController != nil && g.CombatController.IsNPCInCombat(in.NPCInstanceID) {
		return nil, opErr(409, "the entity changed since this change; undo was refused")
	}
	return g.OpNPCDespawn(in.NPCInstanceID)
}

func (g *Game) inverseRestoreNPC(body json.RawMessage) (*OpResult, error) {
	var in struct {
		NPC           *npc.NPC `json:"npc"`
		NPCInstanceID string   `json:"npcInstanceId"`
		ExpectRoomID  string   `json:"expectRoomId"`
		ExpectHP      int32    `json:"expectHp"`
		ExpectAlive   bool     `json:"expectAlive"`
		ExpectMissing bool     `json:"expectMissing"`
	}
	if err := decodeOp(body, &in); err != nil {
		return nil, err
	}
	if in.NPC == nil || in.NPC.Entity == nil {
		return nil, opErr(400, "this change cannot be undone")
	}
	current := g.NPCManager.GetInstance(in.NPC.ID)
	if in.ExpectMissing {
		if current != nil {
			return nil, opErr(409, "the entity changed since this change; undo was refused")
		}
	} else if current == nil || current.IsDead != !in.ExpectAlive || current.CurrentRoomID != in.ExpectRoomID || current.CurrentHitPoints != in.ExpectHP {
		return nil, opErr(409, "the entity changed since this change; undo was refused")
	}
	if current != nil && g.CombatController != nil && g.CombatController.IsNPCInCombat(current.ID) {
		return nil, opErr(409, "the entity changed since this change; undo was refused")
	}
	g.NPCManager.RestoreInstance(in.NPC)
	g.persistNPCIfStored(in.NPC)
	g.NPCManager.Note(in.NPC.ID, "restored")
	alive := !in.NPC.IsDead
	return &OpResult{
		Summary:    opSummary("Restored %s.", npcName(in.NPC)),
		Undoable:   true,
		EntityType: "npc",
		EntityID:   in.NPC.ID,
		Inverse: inverseOf("npc-despawn", map[string]interface{}{
			"npcInstanceId": in.NPC.ID,
			"expectRoomId":  in.NPC.CurrentRoomID,
			"expectHp":      in.NPC.CurrentHitPoints,
			"expectAlive":   alive,
		}),
		Detail: map[string]interface{}{"npcInstanceId": in.NPC.ID},
	}, nil
}

func (g *Game) inverseRestoreProgress(body json.RawMessage) (*OpResult, error) {
	var in struct {
		CharacterID string          `json:"characterId"`
		QuestID     string          `json:"questId"`
		Before      json.RawMessage `json:"before"`
		Expect      json.RawMessage `json:"expect"`
	}
	if err := decodeOp(body, &in); err != nil {
		return nil, err
	}
	current, _ := g.Facade.QuestsService().GetProgress(in.CharacterID, in.QuestID)
	if canonicalJSON(jsonRaw(current)) != canonicalJSON(in.Expect) {
		return nil, opErr(409, "the entity changed since this change; undo was refused")
	}
	var restored *quests.QuestProgress
	if len(bytes.TrimSpace(in.Before)) > 0 && string(bytes.TrimSpace(in.Before)) != "null" {
		restored = &quests.QuestProgress{Entity: &entities.Entity{}}
		if err := json.Unmarshal(in.Before, restored); err != nil {
			return nil, opErr(400, "this change cannot be undone")
		}
	}
	if err := g.Facade.QuestsService().ReplaceProgress(in.CharacterID, in.QuestID, restored); err != nil {
		return nil, opErr(500, err.Error())
	}
	after, _ := g.Facade.QuestsService().GetProgress(in.CharacterID, in.QuestID)
	return &OpResult{
		Summary:    opSummary("Restored quest progress for %s.", in.QuestID),
		Undoable:   true,
		EntityType: "quest-progress",
		EntityID:   in.CharacterID + ":" + in.QuestID,
		Before:     jsonRaw(current),
		After:      jsonRaw(after),
		Inverse: inverseOf("restore-progress", map[string]interface{}{
			"characterId": in.CharacterID,
			"questId":     in.QuestID,
			"before":      json.RawMessage(jsonRaw(current)),
			"expect":      json.RawMessage(jsonRaw(after)),
		}),
		Detail: map[string]interface{}{
			"characterId": in.CharacterID,
			"questId":     in.QuestID,
			"note":        "Quest rewards already granted are not taken back.",
		},
	}, nil
}

func canonicalJSON(raw json.RawMessage) string {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return ""
	}
	var v interface{}
	if err := json.Unmarshal(trimmed, &v); err != nil {
		return string(trimmed)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return string(trimmed)
	}
	return string(b)
}
