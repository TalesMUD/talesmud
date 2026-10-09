package contenthealth

import (
	"fmt"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/scripts"
	"gopkg.in/yaml.v3"
)

// MarshalYAML renders one entity in the importer's YAML shape.
func MarshalYAML(entityType string, entity any) ([]byte, error) {
	body, err := yamlBody(entityType, entity)
	if err != nil {
		return nil, err
	}
	return yaml.Marshal(body)
}

func yamlBody(entityType string, entity any) (any, error) {
	switch normType(entityType) {
	case "room":
		room, ok := entity.(*rooms.Room)
		if !ok || room == nil {
			return nil, fmt.Errorf("room %s not found", entityType)
		}
		return roomYAML(room), nil
	case "npc":
		n, ok := entity.(*npc.NPC)
		if !ok || n == nil {
			return nil, fmt.Errorf("npc not found")
		}
		return npcYAML(n), nil
	case "item":
		item, ok := entity.(*items.Item)
		if !ok || item == nil {
			return nil, fmt.Errorf("item not found")
		}
		return itemYAML(item), nil
	case "loottable":
		table, ok := entity.(*items.LootTable)
		if !ok || table == nil {
			return nil, fmt.Errorf("loot table not found")
		}
		return lootYAML(table), nil
	case "spawner":
		spawner, ok := entity.(*npc.NPCSpawner)
		if !ok || spawner == nil {
			return nil, fmt.Errorf("spawner not found")
		}
		return spawnerYAML(spawner), nil
	case "dialog":
		dialog, ok := entity.(*dialogs.Dialog)
		if !ok || dialog == nil {
			return nil, fmt.Errorf("dialog not found")
		}
		return dialogYAML(dialog), nil
	case "quest":
		quest, ok := entity.(*quests.Quest)
		if !ok || quest == nil {
			return nil, fmt.Errorf("quest not found")
		}
		return questYAML(quest), nil
	case "script":
		script, ok := entity.(*scripts.Script)
		if !ok || script == nil {
			return nil, fmt.Errorf("script not found")
		}
		return scriptYAML(script), nil
	case "skill":
		skill, ok := entity.(*skills.Skill)
		if !ok || skill == nil {
			return nil, fmt.Errorf("skill not found")
		}
		return skillYAML(skill), nil
	case "charactertemplate":
		tmpl, ok := entity.(*characters.CharacterTemplate)
		if !ok || tmpl == nil {
			return nil, fmt.Errorf("character template not found")
		}
		return templateYAML(tmpl), nil
	default:
		return nil, fmt.Errorf("unsupported entity type %s", entityType)
	}
}

func put(m map[string]any, key string, value any) {
	switch v := value.(type) {
	case string:
		if v != "" {
			m[key] = v
		}
	case bool:
		if v {
			m[key] = v
		}
	case int:
		if v != 0 {
			m[key] = v
		}
	case int32:
		if v != 0 {
			m[key] = v
		}
	case int64:
		if v != 0 {
			m[key] = v
		}
	case float64:
		if v != 0 {
			m[key] = v
		}
	default:
		if value != nil {
			m[key] = value
		}
	}
}

func roomYAML(room *rooms.Room) map[string]any {
	out := map[string]any{"id": room.ID, "name": room.Name}
	put(out, "description", room.Description)
	put(out, "detail", room.Detail)
	put(out, "area", room.Area)
	if len(room.Tags) > 0 {
		out["tags"] = room.Tags
	}
	put(out, "canBind", room.CanBind)
	put(out, "onEnterScript", room.OnEnterScriptID)
	if room.Coords != nil {
		out["coords"] = map[string]any{"x": room.Coords.X, "y": room.Coords.Y, "z": room.Coords.Z}
	}
	if room.Meta != nil && (room.Meta.Background != "" || room.Meta.Mood != "") {
		meta := map[string]any{}
		put(meta, "background", room.Meta.Background)
		put(meta, "mood", room.Meta.Mood)
		out["meta"] = meta
	}
	if room.Exits != nil && len(*room.Exits) > 0 {
		exits := make([]map[string]any, 0, len(*room.Exits))
		for _, exit := range *room.Exits {
			row := map[string]any{}
			put(row, "name", exit.Name)
			put(row, "target", exit.Target)
			put(row, "type", string(exit.Type))
			put(row, "description", exit.Description)
			put(row, "hidden", exit.Hidden)
			put(row, "instance", exit.Instance)
			exits = append(exits, row)
		}
		out["exits"] = exits
	}
	if room.Actions != nil && len(*room.Actions) > 0 {
		actions := make([]map[string]any, 0, len(*room.Actions))
		for _, action := range *room.Actions {
			row := map[string]any{}
			put(row, "name", action.Name)
			put(row, "type", string(action.Type))
			put(row, "description", action.Description)
			put(row, "response", action.Response)
			put(row, "scriptId", action.ScriptId)
			if len(action.Params) > 0 {
				row["params"] = action.Params
			}
			actions = append(actions, row)
		}
		out["actions"] = actions
	}
	if room.Items != nil && len(*room.Items) > 0 {
		out["items"] = append([]string(nil), (*room.Items)...)
	}
	return out
}

func npcYAML(n *npc.NPC) map[string]any {
	out := map[string]any{"id": n.ID, "name": n.Name}
	put(out, "description", n.Description)
	put(out, "level", n.Level)
	put(out, "maxHitPoints", n.MaxHitPoints)
	put(out, "spawnRoomId", n.SpawnRoomID)
	put(out, "dialogID", n.DialogID)
	if n.RespawnTime > 0 {
		out["respawnTime"] = n.RespawnTime.String()
	}
	if n.Race.ID != "" || n.Race.Name != "" {
		out["race"] = map[string]any{"id": n.Race.ID, "name": n.Race.Name}
	}
	if n.Class.ID != "" || n.Class.Name != "" {
		out["class"] = map[string]any{"id": n.Class.ID, "name": n.Class.Name}
	}
	if n.EnemyTrait != nil {
		enemy := n.EnemyTrait
		row := map[string]any{}
		put(row, "creatureType", enemy.CreatureType)
		put(row, "combatStyle", enemy.CombatStyle)
		put(row, "difficulty", enemy.Difficulty)
		put(row, "attackPower", enemy.AttackPower)
		put(row, "defense", enemy.Defense)
		put(row, "attackSpeed", enemy.AttackSpeed)
		put(row, "aggroRadius", enemy.AggroRadius)
		put(row, "aggroOnSight", enemy.AggroOnSight)
		put(row, "callForHelp", enemy.CallForHelp)
		put(row, "fleeThreshold", enemy.FleeThreshold)
		put(row, "xpReward", enemy.XPReward)
		if enemy.GoldDrop.Min != 0 || enemy.GoldDrop.Max != 0 {
			row["goldDrop"] = map[string]any{"min": enemy.GoldDrop.Min, "max": enemy.GoldDrop.Max}
		}
		put(row, "lootTableId", enemy.LootTableID)
		if len(enemy.GuaranteedLoot) > 0 {
			row["guaranteedLoot"] = enemy.GuaranteedLoot
		}
		put(row, "onAggroScript", enemy.OnAggroScript)
		put(row, "onDeathScript", enemy.OnDeathScript)
		put(row, "onFleeScript", enemy.OnFleeScript)
		put(row, "onLowHealthScript", enemy.OnLowHealthScript)
		put(row, "lowHealthThreshold", enemy.LowHealthThreshold)
		out["enemyTrait"] = row
	}
	if n.MerchantTrait != nil {
		mt := n.MerchantTrait
		row := map[string]any{}
		put(row, "merchantType", mt.MerchantType)
		put(row, "buyMultiplier", mt.BuyMultiplier)
		put(row, "sellMultiplier", mt.SellMultiplier)
		put(row, "restockMinutes", mt.RestockMinutes)
		if len(mt.AcceptedTypes) > 0 {
			row["acceptedTypes"] = mt.AcceptedTypes
		}
		if len(mt.Inventory) > 0 {
			stock := make([]map[string]any, 0, len(mt.Inventory))
			for _, item := range mt.Inventory {
				entry := map[string]any{"itemTemplateId": item.ItemTemplateID}
				put(entry, "quantity", item.Quantity)
				put(entry, "maxQuantity", item.MaxQuantity)
				put(entry, "basePrice", item.BasePrice)
				put(entry, "priceOverride", item.PriceOverride)
				stock = append(stock, entry)
			}
			row["inventory"] = stock
		}
		out["merchantTrait"] = row
	}
	return out
}

func itemYAML(item *items.Item) map[string]any {
	out := map[string]any{"id": item.ID, "name": item.Name}
	put(out, "description", item.Description)
	put(out, "detail", item.Detail)
	put(out, "type", string(item.Type))
	put(out, "subType", string(item.SubType))
	put(out, "slot", string(item.Slot))
	put(out, "quality", string(item.Quality))
	put(out, "level", item.Level)
	put(out, "basePrice", item.BasePrice)
	put(out, "stackable", item.Stackable)
	put(out, "maxStack", item.MaxStack)
	put(out, "unique", item.Unique)
	put(out, "consumable", item.Consumable)
	put(out, "copyOnPickup", item.CopyOnPickup)
	if len(item.Tags) > 0 {
		out["tags"] = item.Tags
	}
	if len(item.Attributes) > 0 {
		out["attributes"] = item.Attributes
	}
	if len(item.Properties) > 0 {
		out["properties"] = item.Properties
	}
	put(out, "onUseScript", item.OnUseScriptID)
	put(out, "onHitScript", item.OnHitScriptID)
	if item.Meta != nil && item.Meta.Img != "" {
		out["meta"] = map[string]any{"img": item.Meta.Img}
	}
	if len(item.Effects) > 0 {
		effects := make([]map[string]any, 0, len(item.Effects))
		for _, effect := range item.Effects {
			row := map[string]any{}
			put(row, "trigger", effect.Trigger)
			put(row, "name", effect.Name)
			put(row, "text", effect.Text)
			effects = append(effects, row)
		}
		out["effects"] = effects
	}
	return out
}

func lootYAML(table *items.LootTable) map[string]any {
	out := map[string]any{"id": table.ID, "name": table.Name}
	put(out, "description", table.Description)
	entries := make([]map[string]any, 0, len(table.Entries))
	for _, entry := range table.Entries {
		row := map[string]any{"itemTemplateId": entry.ItemTemplateID}
		put(row, "dropChance", entry.DropChance)
		put(row, "minQuantity", entry.MinQuantity)
		put(row, "maxQuantity", entry.MaxQuantity)
		put(row, "guaranteed", entry.Guaranteed)
		put(row, "rarity", entry.Rarity)
		put(row, "bossOnly", entry.BossOnly)
		entries = append(entries, row)
	}
	out["entries"] = entries
	return out
}

func spawnerYAML(spawner *npc.NPCSpawner) map[string]any {
	out := map[string]any{
		"id":           spawner.ID,
		"templateId":   spawner.TemplateID,
		"roomId":       spawner.RoomID,
		"maxInstances": spawner.MaxInstances,
		"initialCount": spawner.InitialCount,
	}
	put(out, "name", spawner.Name)
	if spawner.SpawnInterval > 0 {
		out["spawnInterval"] = spawner.SpawnInterval.String()
	}
	return out
}

func dialogYAML(dialog *dialogs.Dialog) map[string]any {
	root := dialog.NodeID
	if root == "" {
		root = "root"
	}
	tree := map[string]any{}
	seen := map[string]bool{}
	writeDialogNode(tree, seen, root, dialog)
	out := map[string]any{"id": dialog.ID, "tree": tree}
	put(out, "name", dialog.Name)
	return out
}

func writeDialogNode(tree map[string]any, seen map[string]bool, id string, node *dialogs.Dialog) {
	if node == nil || id == "" || seen[id] {
		return
	}
	seen[id] = true
	options := make([]map[string]any, 0, len(node.Options))
	for _, option := range node.Options {
		if option == nil {
			continue
		}
		row := map[string]any{}
		put(row, "player_text", option.Text)
		put(row, "questId", option.QuestID)
		put(row, "action", option.Action)
		next := option.NodeID
		if option.Answer != nil && option.Answer.NodeID != "" {
			next = option.Answer.NodeID
		}
		put(row, "next", next)
		options = append(options, row)
		if option.Answer != nil && next != "" {
			writeDialogNode(tree, seen, next, option.Answer)
		}
	}
	tree[id] = map[string]any{"npc_text": node.Text, "options": options}
}

func questYAML(quest *quests.Quest) map[string]any {
	out := map[string]any{"id": quest.ID, "name": quest.Name}
	put(out, "description", quest.Description)
	put(out, "category", quest.Category)
	put(out, "level", quest.Level)
	put(out, "repeatable", quest.Repeatable)
	put(out, "turnIn", quest.TurnIn)
	put(out, "requiredLevel", quest.RequiredLevel)
	put(out, "acceptDialogText", quest.AcceptDialogText)
	put(out, "progressDialogText", quest.ProgressDialogText)
	put(out, "completeDialogText", quest.CompleteDialogText)
	put(out, "onAcceptScriptId", quest.OnAcceptScriptID)
	put(out, "onCompleteScriptId", quest.OnCompleteScriptID)
	if len(quest.RequiredQuestIDs) > 0 {
		out["requiredQuestIds"] = quest.RequiredQuestIDs
	}
	source := map[string]any{}
	put(source, "type", quest.Source.Type)
	put(source, "npcId", quest.Source.NPCID)
	put(source, "itemId", quest.Source.ItemID)
	if len(source) > 0 {
		out["source"] = source
	}
	if len(quest.Objectives) > 0 {
		objectives := make([]map[string]any, 0, len(quest.Objectives))
		for _, objective := range quest.Objectives {
			row := map[string]any{"type": string(objective.Type)}
			put(row, "id", objective.ID)
			put(row, "description", objective.Description)
			put(row, "targetId", objective.TargetID)
			put(row, "targetName", objective.TargetName)
			put(row, "amount", objective.Amount)
			put(row, "deliverToNpcId", objective.DeliverToNPCID)
			put(row, "deliverToNpcName", objective.DeliverToNPCName)
			put(row, "dialogNodeId", objective.DialogNodeID)
			put(row, "checkScriptId", objective.CheckScriptID)
			put(row, "order", objective.Order)
			objectives = append(objectives, row)
		}
		out["objectives"] = objectives
	}
	rewards := map[string]any{}
	put(rewards, "xp", quest.Rewards.XP)
	put(rewards, "gold", quest.Rewards.Gold)
	if len(quest.Rewards.ItemTemplateIDs) > 0 {
		rewards["itemTemplateIds"] = quest.Rewards.ItemTemplateIDs
	}
	if len(rewards) > 0 {
		out["rewards"] = rewards
	}
	return out
}

func scriptYAML(script *scripts.Script) map[string]any {
	out := map[string]any{"id": script.ID, "name": script.Name, "code": script.Code}
	put(out, "description", script.Description)
	put(out, "type", string(script.Type))
	put(out, "language", string(script.Language))
	return out
}

func skillYAML(skill *skills.Skill) map[string]any {
	out := map[string]any{"id": skill.ID, "name": skill.Name}
	put(out, "description", skill.Description)
	if len(skill.ClassIDs) > 0 {
		out["classIds"] = skill.ClassIDs
	}
	put(out, "levelRequired", skill.LevelRequired)
	put(out, "resourceType", string(skill.ResourceType))
	put(out, "manaCost", skill.ManaCost)
	put(out, "cooldownRounds", skill.CooldownRounds)
	put(out, "target", string(skill.Target))
	put(out, "effect", string(skill.Effect))
	put(out, "scalingAttr", skill.ScalingAttr)
	put(out, "basePower", skill.BasePower)
	put(out, "scalingFactor", skill.ScalingFactor)
	put(out, "duration", skill.Duration)
	put(out, "buffStat", skill.BuffStat)
	put(out, "buffPercent", skill.BuffPercent)
	put(out, "ignoresDefense", skill.IgnoresDefense)
	put(out, "hitCount", skill.HitCount)
	put(out, "secondaryEffect", string(skill.SecondaryEffect))
	put(out, "secondaryBasePower", skill.SecondaryBasePower)
	put(out, "secondaryScaling", skill.SecondaryScaling)
	put(out, "secondaryTarget", string(skill.SecondaryTarget))
	return out
}

func templateYAML(tmpl *characters.CharacterTemplate) map[string]any {
	out := map[string]any{"id": tmpl.ID, "name": tmpl.Name}
	put(out, "description", tmpl.Description)
	put(out, "archetype", tmpl.Archetype)
	if len(tmpl.StartingItems) > 0 {
		itemsOut := make([]map[string]any, 0, len(tmpl.StartingItems))
		for _, item := range tmpl.StartingItems {
			row := map[string]any{}
			put(row, "slot", string(item.Slot))
			put(row, "itemTemplateId", item.ItemTemplateID)
			itemsOut = append(itemsOut, row)
		}
		out["startingItems"] = itemsOut
	}
	if len(tmpl.DefaultSkills) > 0 {
		out["defaultSkills"] = tmpl.DefaultSkills
	}
	return out
}
