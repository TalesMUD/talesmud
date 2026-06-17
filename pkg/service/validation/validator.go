package validation

import (
	"fmt"

	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/scripts"
)

func ValidateRoom(room *rooms.Room, snapshot WorldSnapshot) Result {
	result := NewResult()
	if room == nil {
		return result
	}

	if room.Exits != nil {
		for i, exit := range *room.Exits {
			if exit.Target != "" && !snapshot.HasRoom(exit.Target) {
				result.Add(Error("missing_room", "room", room.ID, fmt.Sprintf("exits[%d].target", i), "room", exit.Target, "Exit target references a missing room."))
			}
			if exit.Target != "" && exit.Name == "" {
				result.Add(Warning("exit_without_name", "room", room.ID, fmt.Sprintf("exits[%d].name", i), "", "", "Exit has a target but no name."))
			}
			if exit.Target == room.ID && exit.Type != rooms.RoomExitTypeTeleport {
				result.Add(Warning("self_referential_exit", "room", room.ID, fmt.Sprintf("exits[%d].target", i), "room", exit.Target, "Exit points back to the same room."))
			}
		}
	}

	if room.OnEnterScriptID != "" && !snapshot.HasScript(room.OnEnterScriptID) {
		result.Add(Error("missing_script", "room", room.ID, "onEnterScriptID", "script", room.OnEnterScriptID, "On-enter script references a missing script."))
	}

	if room.Actions != nil {
		for i, action := range *room.Actions {
			if action.Type == rooms.RoomActionTypeScript && action.ScriptId == "" {
				result.Add(Warning("script_action_without_script", "room", room.ID, fmt.Sprintf("actions[%d].scriptId", i), "script", "", "Script action has no script ID."))
			}
			if action.Type == rooms.RoomActionTypeScript && action.ScriptId != "" && !snapshot.HasScript(action.ScriptId) {
				result.Add(Error("missing_script", "room", room.ID, fmt.Sprintf("actions[%d].scriptId", i), "script", action.ScriptId, "Room action references a missing script."))
			}
		}
	}

	if room.Items != nil {
		for i, itemID := range *room.Items {
			if itemID != "" && !snapshot.HasItem(itemID) {
				result.Add(Error("missing_item", "room", room.ID, fmt.Sprintf("items[%d]", i), "item", itemID, "Room item references a missing item."))
			}
		}
	}

	if room.NPCs != nil {
		for i, npcID := range *room.NPCs {
			if npcID != "" && !snapshot.HasNPC(npcID) {
				result.Add(Error("missing_npc", "room", room.ID, fmt.Sprintf("npcs[%d]", i), "npc", npcID, "Room resident references a missing NPC."))
			}
		}
	}

	return result
}

func ValidateQuest(quest *quests.Quest, snapshot WorldSnapshot) Result {
	result := NewResult()
	if quest == nil {
		return result
	}

	switch quest.Source.Type {
	case "", "npc", "item", "auto", "script":
	default:
		result.Add(Warning("unknown_quest_source_type", "quest", quest.ID, "source.type", "", "", "Quest source type is unknown."))
	}

	if quest.Source.NPCID != "" && !snapshot.HasNPC(quest.Source.NPCID) {
		result.Add(Error("missing_npc", "quest", quest.ID, "source.npcId", "npc", quest.Source.NPCID, "Quest source NPC references a missing NPC."))
	}
	if quest.Source.ItemID != "" && !snapshot.HasItem(quest.Source.ItemID) {
		result.Add(Error("missing_item", "quest", quest.ID, "source.itemId", "item", quest.Source.ItemID, "Quest source item references a missing item."))
	}
	for i, rewardID := range quest.Rewards.ItemTemplateIDs {
		if rewardID != "" && !snapshot.HasItem(rewardID) {
			result.Add(Error("missing_item", "quest", quest.ID, fmt.Sprintf("rewards.itemTemplateIds[%d]", i), "item", rewardID, "Quest reward references a missing item template."))
		}
	}
	for i, reqID := range quest.RequiredQuestIDs {
		if reqID != "" && !snapshot.HasQuest(reqID) {
			result.Add(Error("missing_quest", "quest", quest.ID, fmt.Sprintf("requiredQuestIds[%d]", i), "quest", reqID, "Quest prerequisite references a missing quest."))
		}
	}
	if quest.OnCompleteScriptID != "" && !snapshot.HasScript(quest.OnCompleteScriptID) {
		result.Add(Error("missing_script", "quest", quest.ID, "onCompleteScriptId", "script", quest.OnCompleteScriptID, "Quest completion script references a missing script."))
	}

	for i, objective := range quest.Objectives {
		prefix := fmt.Sprintf("objectives[%d]", i)
		if objective.ID == "" {
			result.Add(Warning("objective_without_id", "quest", quest.ID, prefix+".id", "", "", "Quest objective has no stable ID."))
		}
		if objective.Amount < 1 && objective.Type != quests.ObjectiveVisit && objective.Type != quests.ObjectiveTalk && objective.Type != quests.ObjectiveCustom {
			result.Add(Warning("objective_amount_below_one", "quest", quest.ID, prefix+".amount", "", "", "Quest objective amount is below one."))
		}
		if objective.CheckScriptID != "" && !snapshot.HasScript(objective.CheckScriptID) {
			result.Add(Error("missing_script", "quest", quest.ID, prefix+".checkScriptId", "script", objective.CheckScriptID, "Quest objective check script references a missing script."))
		}
		validateQuestObjectiveTarget(&result, quest.ID, prefix, objective, snapshot)
	}

	return result
}

func validateQuestObjectiveTarget(result *Result, questID, prefix string, objective quests.Objective, snapshot WorldSnapshot) {
	switch objective.Type {
	case quests.ObjectiveKill:
		if objective.TargetID != "" && !snapshot.HasNPC(objective.TargetID) {
			result.Add(Error("missing_npc", "quest", questID, prefix+".targetId", "npc", objective.TargetID, "Kill objective target references a missing NPC."))
		}
	case quests.ObjectiveCollect:
		if objective.TargetID != "" && !snapshot.HasItem(objective.TargetID) {
			result.Add(Error("missing_item", "quest", questID, prefix+".targetId", "item", objective.TargetID, "Collect objective target references a missing item."))
		}
	case quests.ObjectiveDeliver:
		if objective.TargetID != "" && !snapshot.HasItem(objective.TargetID) {
			result.Add(Error("missing_item", "quest", questID, prefix+".targetId", "item", objective.TargetID, "Deliver objective target references a missing item."))
		}
		if objective.DeliverToNPCID != "" && !snapshot.HasNPC(objective.DeliverToNPCID) {
			result.Add(Error("missing_npc", "quest", questID, prefix+".deliverToNpcId", "npc", objective.DeliverToNPCID, "Deliver objective NPC references a missing NPC."))
		}
	case quests.ObjectiveVisit:
		if objective.TargetID != "" && !snapshot.HasRoom(objective.TargetID) {
			result.Add(Error("missing_room", "quest", questID, prefix+".targetId", "room", objective.TargetID, "Visit objective target references a missing room."))
		}
	case quests.ObjectiveTalk:
		if objective.TargetID != "" && !snapshot.HasNPC(objective.TargetID) {
			result.Add(Error("missing_npc", "quest", questID, prefix+".targetId", "npc", objective.TargetID, "Talk objective target references a missing NPC."))
		}
	}
}

func ValidateNPC(n *npc.NPC, snapshot WorldSnapshot) Result {
	result := NewResult()
	if n == nil {
		return result
	}

	if n.SpawnRoomID != "" && !snapshot.HasRoom(n.SpawnRoomID) {
		result.Add(Error("missing_room", "npc", n.ID, "spawnRoomId", "room", n.SpawnRoomID, "NPC spawn room references a missing room."))
	}
	if n.CurrentRoomID != "" && !snapshot.HasRoom(n.CurrentRoomID) {
		result.Add(Error("missing_room", "npc", n.ID, "currentRoomID", "room", n.CurrentRoomID, "NPC current room references a missing room."))
	}
	if n.DialogID != "" && !snapshot.HasDialog(n.DialogID) {
		result.Add(Error("missing_dialog", "npc", n.ID, "dialogID", "dialog", n.DialogID, "NPC dialog references a missing dialog."))
	}
	if n.IdleDialogID != "" && !snapshot.HasDialog(n.IdleDialogID) {
		result.Add(Error("missing_dialog", "npc", n.ID, "idleDialogID", "dialog", n.IdleDialogID, "NPC idle dialog references a missing dialog."))
	}
	if n.TemplateID != "" && !snapshot.HasNPC(n.TemplateID) {
		result.Add(Error("missing_npc", "npc", n.ID, "templateId", "npc", n.TemplateID, "NPC instance references a missing NPC template."))
	}
	for i, roomID := range n.PatrolPath {
		if roomID != "" && !snapshot.HasRoom(roomID) {
			result.Add(Error("missing_room", "npc", n.ID, fmt.Sprintf("patrolPath[%d]", i), "room", roomID, "NPC patrol path references a missing room."))
		}
	}

	if n.EnemyTrait != nil {
		if n.EnemyTrait.LootTableID != "" && !snapshot.HasLootTable(n.EnemyTrait.LootTableID) {
			result.Add(Error("missing_loot_table", "npc", n.ID, "enemyTrait.lootTableId", "loottable", n.EnemyTrait.LootTableID, "Enemy loot table references a missing loot table."))
		}
		for _, scriptField := range []struct {
			field string
			id    string
		}{
			{"enemyTrait.onAggroScript", n.EnemyTrait.OnAggroScript},
			{"enemyTrait.onDeathScript", n.EnemyTrait.OnDeathScript},
			{"enemyTrait.onFleeScript", n.EnemyTrait.OnFleeScript},
		} {
			if scriptField.id != "" && !snapshot.HasScript(scriptField.id) {
				result.Add(Error("missing_script", "npc", n.ID, scriptField.field, "script", scriptField.id, "Enemy event script references a missing script."))
			}
		}
		for i, itemID := range n.EnemyTrait.GuaranteedLoot {
			if itemID != "" && !snapshot.HasItem(itemID) {
				result.Add(Error("missing_item", "npc", n.ID, fmt.Sprintf("enemyTrait.guaranteedLoot[%d]", i), "item", itemID, "Enemy guaranteed loot references a missing item template."))
			}
		}
	}

	if n.MerchantTrait != nil {
		for i, stock := range n.MerchantTrait.Inventory {
			prefix := fmt.Sprintf("merchantTrait.inventory[%d]", i)
			if stock.ItemTemplateID != "" && !snapshot.HasItem(stock.ItemTemplateID) {
				result.Add(Error("missing_item", "npc", n.ID, prefix+".itemTemplateId", "item", stock.ItemTemplateID, "Merchant stock references a missing item template."))
			}
			if stock.Quantity < -1 {
				result.Add(Warning("invalid_stock_quantity", "npc", n.ID, prefix+".quantity", "", "", "Merchant stock quantity is invalid."))
			}
			if stock.Quantity >= 0 && stock.MaxQuantity > 0 && stock.MaxQuantity < stock.Quantity {
				result.Add(Warning("stock_max_below_quantity", "npc", n.ID, prefix+".maxQuantity", "", "", "Merchant max quantity is below current quantity."))
			}
		}
	}

	return result
}

func ValidateDialog(dialog *dialogs.Dialog, snapshot WorldSnapshot) Result {
	result := NewResult()
	if dialog == nil {
		return result
	}

	seen := map[string]bool{}
	var walk func(node *dialogs.Dialog, path string)
	walk = func(node *dialogs.Dialog, path string) {
		if node == nil {
			return
		}
		if node.NodeID != "" {
			if seen[node.NodeID] {
				result.Add(Error("duplicate_dialog_node", "dialog", dialog.ID, path+".nodeId", "dialog_node", node.NodeID, "Dialog contains a duplicate node ID."))
			}
			seen[node.NodeID] = true
		}
		if node.QuestID != "" && !snapshot.HasQuest(node.QuestID) {
			result.Add(Error("missing_quest", "dialog", dialog.ID, path+".questId", "quest", node.QuestID, "Dialog quest link references a missing quest."))
		}
		for i, option := range node.Options {
			if option != nil && option.Text == "" && option.Answer == nil && len(option.Options) == 0 {
				result.Add(Warning("empty_dialog_option", "dialog", dialog.ID, fmt.Sprintf("%s.options[%d]", path, i), "", "", "Dialog option has no text or child response."))
			}
			walk(option, fmt.Sprintf("%s.options[%d]", path, i))
		}
		if node.Answer != nil {
			walk(node.Answer, path+".answer")
		}
	}
	walk(dialog, "root")
	if dialog.NodeID == "" {
		result.Add(Warning("dialog_without_root_node", "dialog", dialog.ID, "nodeId", "", "", "Dialog root has no node ID."))
	}

	return result
}

func ValidateLootTable(table *items.LootTable, snapshot WorldSnapshot) Result {
	result := NewResult()
	if table == nil {
		return result
	}

	for i, entry := range table.Entries {
		prefix := fmt.Sprintf("entries[%d]", i)
		if entry.ItemTemplateID != "" && !snapshot.HasItem(entry.ItemTemplateID) {
			result.Add(Error("missing_item", "loottable", table.ID, prefix+".itemTemplateId", "item", entry.ItemTemplateID, "Loot entry references a missing item template."))
		}
		if entry.DropChance < 0 || entry.DropChance > 1 {
			result.Add(Error("invalid_drop_chance", "loottable", table.ID, prefix+".dropChance", "", "", "Loot entry drop chance must be between 0.0 and 1.0."))
		}
		if entry.MinQuantity > entry.MaxQuantity {
			result.Add(Error("invalid_quantity_range", "loottable", table.ID, prefix+".minQuantity", "", "", "Loot entry minimum quantity is greater than maximum quantity."))
		}
		if !entry.Guaranteed && entry.DropChance == 0 {
			result.Add(Warning("zero_drop_chance", "loottable", table.ID, prefix+".dropChance", "", "", "Non-guaranteed loot entry has a zero drop chance."))
		}
	}

	return result
}

func ValidateItem(item *items.Item, snapshot WorldSnapshot) Result {
	result := NewResult()
	if item == nil {
		return result
	}

	if item.OnUseScriptID != "" && !snapshot.HasScript(item.OnUseScriptID) {
		result.Add(Error("missing_script", "item", item.ID, "onUseScriptId", "script", item.OnUseScriptID, "Item use script references a missing script."))
	}
	if item.TemplateID != "" && !snapshot.HasItem(item.TemplateID) {
		result.Add(Error("missing_item", "item", item.ID, "templateId", "item", item.TemplateID, "Item instance references a missing template."))
	}
	for i, child := range item.Items {
		if child != nil && child.ID != "" && !snapshot.HasItem(child.ID) {
			result.Add(Error("missing_item", "item", item.ID, fmt.Sprintf("items[%d].id", i), "item", child.ID, "Container child references a missing item."))
		}
	}
	if item.Stackable && item.MaxStack < 1 {
		result.Add(Warning("invalid_max_stack", "item", item.ID, "maxStack", "", "", "Stackable item has an invalid maximum stack size."))
	}
	if item.Stackable && item.MaxStack > 0 && item.Quantity > item.MaxStack {
		result.Add(Warning("quantity_above_max_stack", "item", item.ID, "quantity", "", "", "Stackable item quantity is above max stack."))
	}
	if item.Consumable && item.OnUseScriptID == "" && len(item.Attributes) == 0 {
		result.Add(Warning("consumable_without_effect", "item", item.ID, "onUseScriptId", "", "", "Consumable item has no use script or built-in effect."))
	}

	return result
}

func ValidateSpawner(spawner *npc.NPCSpawner, snapshot WorldSnapshot) Result {
	result := NewResult()
	if spawner == nil {
		return result
	}

	if spawner.TemplateID != "" && !snapshot.HasNPC(spawner.TemplateID) {
		result.Add(Error("missing_npc", "spawner", spawner.ID, "templateId", "npc", spawner.TemplateID, "Spawner template references a missing NPC."))
	}
	if spawner.RoomID != "" && !snapshot.HasRoom(spawner.RoomID) {
		result.Add(Error("missing_room", "spawner", spawner.ID, "roomId", "room", spawner.RoomID, "Spawner room references a missing room."))
	}
	if template := snapshot.NPCs[spawner.TemplateID]; template != nil && !template.IsTemplate {
		result.Add(Warning("spawner_uses_non_template", "spawner", spawner.ID, "templateId", "npc", spawner.TemplateID, "Spawner uses a non-template NPC as its template."))
	}

	return result
}

func ValidateScript(script *scripts.Script, snapshot WorldSnapshot, roomOnEnter bool) Result {
	return ValidateLuaScript(script, roomOnEnter)
}
