package worldindex

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

// Index is the reference graph for one snapshot.
type Index struct {
	snap     Snapshot
	edges    []Edge
	inbound  map[string][]Edge
	outbound map[string][]Edge
	names    map[string]string
}

// Build indexes references in snap. The snapshot maps are not copied.
func Build(snap Snapshot) *Index {
	ix := &Index{
		snap:     snap,
		inbound:  map[string][]Edge{},
		outbound: map[string][]Edge{},
		names:    map[string]string{},
	}
	ix.collect()
	return ix
}

// Snapshot returns the world the index was built from.
func (ix *Index) Snapshot() Snapshot {
	if ix == nil {
		return NewSnapshot()
	}
	return ix.snap
}

// Inbound returns edges that point at (kind, id).
func (ix *Index) Inbound(kind Kind, id string) []Edge {
	if ix == nil || id == "" {
		return nil
	}
	return append([]Edge(nil), ix.inbound[key(kind, id)]...)
}

// Outbound returns edges that leave (kind, id).
func (ix *Index) Outbound(kind Kind, id string) []Edge {
	if ix == nil || id == "" {
		return nil
	}
	return append([]Edge(nil), ix.outbound[key(kind, id)]...)
}

// Name returns the display name for an entity, or the id when it has none.
func (ix *Index) Name(kind Kind, id string) string {
	if ix == nil {
		return id
	}
	if name := ix.names[key(kind, id)]; name != "" {
		return name
	}
	return id
}

// Has reports whether the snapshot contains the entity.
func (ix *Index) Has(kind Kind, id string) bool {
	if ix == nil || id == "" {
		return false
	}
	_, ok := ix.names[key(kind, id)]
	return ok
}

func (ix *Index) add(e Edge) {
	if e.FromID == "" || e.ToID == "" || e.FromType == "" || e.ToType == "" {
		return
	}
	ix.edges = append(ix.edges, e)
	ix.outbound[key(e.FromType, e.FromID)] = append(ix.outbound[key(e.FromType, e.FromID)], e)
	ix.inbound[key(e.ToType, e.ToID)] = append(ix.inbound[key(e.ToType, e.ToID)], e)
}

func (ix *Index) note(kind Kind, id, name string) {
	if id == "" {
		return
	}
	if name == "" {
		name = id
	}
	ix.names[key(kind, id)] = name
}

func key(kind Kind, id string) string {
	return string(kind) + "\x00" + id
}

func (ix *Index) collect() {
	for id, room := range ix.snap.Rooms {
		if room == nil {
			continue
		}
		ix.note(KindRoom, id, room.Name)
		ix.roomEdges(room)
	}
	for id, n := range ix.snap.NPCs {
		if n == nil {
			continue
		}
		ix.note(KindNPC, id, n.Name)
		ix.npcEdges(n)
	}
	for id, item := range ix.snap.Items {
		if item == nil {
			continue
		}
		ix.note(KindItem, id, item.Name)
		ix.itemEdges(item)
	}
	for id, table := range ix.snap.LootTables {
		if table == nil {
			continue
		}
		ix.note(KindLootTable, id, table.Name)
		ix.lootEdges(table)
	}
	for id, spawner := range ix.snap.Spawners {
		if spawner == nil {
			continue
		}
		name := spawner.Name
		if name == "" {
			name = spawner.ID
		}
		ix.note(KindSpawner, id, name)
		ix.add(Edge{FromType: KindSpawner, FromID: id, Field: "templateId", ToType: KindNPC, ToID: spawner.TemplateID, How: "spawner template"})
		ix.add(Edge{FromType: KindSpawner, FromID: id, Field: "roomId", ToType: KindRoom, ToID: spawner.RoomID, How: "spawner room"})
	}
	for id, dialog := range ix.snap.Dialogs {
		if dialog == nil {
			continue
		}
		ix.note(KindDialog, id, dialog.Name)
		ix.dialogEdges(id, dialog, "root")
	}
	for id, quest := range ix.snap.Quests {
		if quest == nil {
			continue
		}
		ix.note(KindQuest, id, quest.Name)
		ix.questEdges(quest)
	}
	for id, script := range ix.snap.Scripts {
		if script == nil {
			continue
		}
		ix.note(KindScript, id, script.Name)
		for _, edge := range ExtractLuaEdges(id, script.Code) {
			ix.add(edge)
		}
	}
	for id, skill := range ix.snap.Skills {
		if skill == nil {
			continue
		}
		ix.note(KindSkill, id, skill.Name)
		for i, classID := range skill.ClassIDs {
			ix.add(Edge{FromType: KindSkill, FromID: id, Field: fmt.Sprintf("classIds[%d]", i), ToType: KindClass, ToID: classID, How: "skill class"})
		}
	}
	for id, tmpl := range ix.snap.CharacterTemplates {
		if tmpl == nil {
			continue
		}
		ix.note(KindCharacterTemplate, id, tmpl.Name)
		for i, item := range tmpl.StartingItems {
			ix.add(Edge{FromType: KindCharacterTemplate, FromID: id, Field: fmt.Sprintf("startingItems[%d].itemTemplateId", i), ToType: KindItem, ToID: item.ItemTemplateID, How: "starting item"})
		}
		for i, skillID := range tmpl.DefaultSkills {
			ix.add(Edge{FromType: KindCharacterTemplate, FromID: id, Field: fmt.Sprintf("defaultSkills[%d]", i), ToType: KindSkill, ToID: skillID, How: "starting skill"})
		}
	}
	for _, class := range ix.snap.Classes {
		ix.note(KindClass, class.ID, class.Name)
		ix.add(Edge{FromType: KindClass, FromID: class.ID, Field: "templateId", ToType: KindCharacterTemplate, ToID: class.TemplateID, How: "class template"})
		for i, skillID := range class.SkillIDs {
			ix.add(Edge{FromType: KindClass, FromID: class.ID, Field: fmt.Sprintf("skills[%d]", i), ToType: KindSkill, ToID: skillID, How: "class skill"})
		}
	}
}

func (ix *Index) roomEdges(room *rooms.Room) {
	id := room.ID
	if room.OnEnterScriptID != "" {
		ix.add(Edge{FromType: KindRoom, FromID: id, Field: "onEnterScriptID", ToType: KindScript, ToID: room.OnEnterScriptID, How: "on-enter script"})
	}
	if room.Exits != nil {
		for i, exit := range *room.Exits {
			how := "exit " + exit.Name
			if exit.Hidden {
				how = "hidden exit " + exit.Name
			}
			if exit.Instance {
				how = "instance " + how
			}
			ix.add(Edge{FromType: KindRoom, FromID: id, Field: fmt.Sprintf("exits[%d].target", i), ToType: KindRoom, ToID: exit.Target, How: strings.TrimSpace(how)})
		}
	}
	if room.Actions != nil {
		for i, action := range *room.Actions {
			if action.ScriptId == "" {
				continue
			}
			how := "room action"
			if action.Name != "" {
				how = "room action " + action.Name
			}
			ix.add(Edge{FromType: KindRoom, FromID: id, Field: fmt.Sprintf("actions[%d].scriptId", i), ToType: KindScript, ToID: action.ScriptId, How: how})
		}
	}
	if room.Items != nil {
		for i, itemID := range *room.Items {
			ix.add(Edge{FromType: KindRoom, FromID: id, Field: fmt.Sprintf("items[%d]", i), ToType: KindItem, ToID: itemID, How: "room item"})
		}
	}
	if room.NPCs != nil {
		for i, npcID := range *room.NPCs {
			ix.add(Edge{FromType: KindRoom, FromID: id, Field: fmt.Sprintf("npcs[%d]", i), ToType: KindNPC, ToID: npcID, How: "room resident"})
		}
	}
}

func (ix *Index) npcEdges(n *npc.NPC) {
	id := n.ID
	ix.add(Edge{FromType: KindNPC, FromID: id, Field: "spawnRoomId", ToType: KindRoom, ToID: n.SpawnRoomID, How: "spawn room"})
	if n.CurrentRoomID != "" && n.CurrentRoomID != n.SpawnRoomID {
		ix.add(Edge{FromType: KindNPC, FromID: id, Field: "currentRoomID", ToType: KindRoom, ToID: n.CurrentRoomID, How: "current room"})
	}
	ix.add(Edge{FromType: KindNPC, FromID: id, Field: "dialogID", ToType: KindDialog, ToID: n.DialogID, How: "NPC dialog"})
	ix.add(Edge{FromType: KindNPC, FromID: id, Field: "idleDialogID", ToType: KindDialog, ToID: n.IdleDialogID, How: "idle dialog"})
	ix.add(Edge{FromType: KindNPC, FromID: id, Field: "templateId", ToType: KindNPC, ToID: n.TemplateID, How: "NPC template"})
	for i, roomID := range n.PatrolPath {
		ix.add(Edge{FromType: KindNPC, FromID: id, Field: fmt.Sprintf("patrolPath[%d]", i), ToType: KindRoom, ToID: roomID, How: "patrol"})
	}
	if et := n.EnemyTrait; et != nil {
		ix.add(Edge{FromType: KindNPC, FromID: id, Field: "enemyTrait.lootTableId", ToType: KindLootTable, ToID: et.LootTableID, How: "loot table"})
		for i, itemID := range et.GuaranteedLoot {
			ix.add(Edge{FromType: KindNPC, FromID: id, Field: fmt.Sprintf("enemyTrait.guaranteedLoot[%d]", i), ToType: KindItem, ToID: itemID, How: "guaranteed loot"})
		}
		hooks := []struct {
			field string
			id    string
			how   string
		}{
			{"enemyTrait.onAggroScript", et.OnAggroScript, "NPC hook onAggro"},
			{"enemyTrait.onDeathScript", et.OnDeathScript, "NPC hook onDeath"},
			{"enemyTrait.onFleeScript", et.OnFleeScript, "NPC hook onFlee"},
			{"enemyTrait.onLowHealthScript", et.OnLowHealthScript, "NPC hook onLowHealth"},
		}
		for _, hook := range hooks {
			ix.add(Edge{FromType: KindNPC, FromID: id, Field: hook.field, ToType: KindScript, ToID: hook.id, How: hook.how})
		}
	}
	if n.MerchantTrait != nil {
		for i, stock := range n.MerchantTrait.Inventory {
			ix.add(Edge{FromType: KindNPC, FromID: id, Field: fmt.Sprintf("merchantTrait.inventory[%d].itemTemplateId", i), ToType: KindItem, ToID: stock.ItemTemplateID, How: "merchant stock"})
		}
	}
}

func (ix *Index) itemEdges(item *items.Item) {
	id := item.ID
	ix.add(Edge{FromType: KindItem, FromID: id, Field: "onUseScriptId", ToType: KindScript, ToID: item.OnUseScriptID, How: "item on-use script"})
	ix.add(Edge{FromType: KindItem, FromID: id, Field: "onHitScriptId", ToType: KindScript, ToID: item.OnHitScriptID, How: "item on-hit script"})
	ix.add(Edge{FromType: KindItem, FromID: id, Field: "templateId", ToType: KindItem, ToID: item.TemplateID, How: "item template"})
}

func (ix *Index) lootEdges(table *items.LootTable) {
	id := table.ID
	for i, entry := range table.Entries {
		how := "loot entry"
		if entry.Guaranteed || entry.DropChance >= 1 {
			how = "guaranteed loot"
		} else if entry.DropChance > 0 {
			how = fmt.Sprintf("loot entry %g%%", entry.DropChance*100)
		}
		ix.add(Edge{FromType: KindLootTable, FromID: id, Field: fmt.Sprintf("entries[%d].itemTemplateId", i), ToType: KindItem, ToID: entry.ItemTemplateID, How: how})
	}
}

func (ix *Index) dialogEdges(rootID string, node *dialogs.Dialog, path string) {
	if node == nil {
		return
	}
	if node.QuestID != "" {
		how := "dialog quest link"
		switch strings.ToLower(node.Action) {
		case "accept":
			how = "dialog option accepts quest"
		case "complete":
			how = "dialog option completes quest"
		case "progress":
			how = "dialog option progresses quest"
		}
		ix.add(Edge{FromType: KindDialog, FromID: rootID, Field: path + ".questId", ToType: KindQuest, ToID: node.QuestID, How: how})
	}
	for i, option := range node.Options {
		ix.dialogEdges(rootID, option, fmt.Sprintf("%s.options[%d]", path, i))
	}
	if node.Answer != nil {
		ix.dialogEdges(rootID, node.Answer, path+".answer")
	}
}

func (ix *Index) questEdges(quest *quests.Quest) {
	id := quest.ID
	if quest.Source.NPCID != "" {
		ix.add(Edge{FromType: KindQuest, FromID: id, Field: "source.npcId", ToType: KindNPC, ToID: quest.Source.NPCID, How: "quest source"})
	}
	if quest.Source.ItemID != "" {
		ix.add(Edge{FromType: KindQuest, FromID: id, Field: "source.itemId", ToType: KindItem, ToID: quest.Source.ItemID, How: "quest source"})
	}
	ix.add(Edge{FromType: KindQuest, FromID: id, Field: "onAcceptScriptId", ToType: KindScript, ToID: quest.OnAcceptScriptID, How: "quest on-accept script"})
	ix.add(Edge{FromType: KindQuest, FromID: id, Field: "onCompleteScriptId", ToType: KindScript, ToID: quest.OnCompleteScriptID, How: "quest on-complete script"})
	for i, rewardID := range quest.Rewards.ItemTemplateIDs {
		ix.add(Edge{FromType: KindQuest, FromID: id, Field: fmt.Sprintf("rewards.itemTemplateIds[%d]", i), ToType: KindItem, ToID: rewardID, How: "quest reward"})
	}
	for i, reqID := range quest.RequiredQuestIDs {
		ix.add(Edge{FromType: KindQuest, FromID: id, Field: fmt.Sprintf("requiredQuestIds[%d]", i), ToType: KindQuest, ToID: reqID, How: "quest prerequisite"})
	}
	_, turnIn := quest.ResolveTurnIn()
	if turnIn != "" && turnIn != quest.Source.NPCID {
		ix.add(Edge{FromType: KindQuest, FromID: id, Field: "turnIn", ToType: KindNPC, ToID: turnIn, How: "quest turn-in"})
	}
	for i, objective := range quest.Objectives {
		prefix := fmt.Sprintf("objectives[%d]", i)
		ix.add(Edge{FromType: KindQuest, FromID: id, Field: prefix + ".checkScriptId", ToType: KindScript, ToID: objective.CheckScriptID, How: "quest objective script"})
		switch objective.Type {
		case quests.ObjectiveVisit:
			ix.add(Edge{FromType: KindQuest, FromID: id, Field: prefix + ".targetId", ToType: KindRoom, ToID: objective.TargetID, How: "quest objective visit"})
		case quests.ObjectiveKill:
			ix.add(Edge{FromType: KindQuest, FromID: id, Field: prefix + ".targetId", ToType: KindNPC, ToID: objective.TargetID, How: "quest objective kill"})
		case quests.ObjectiveCollect:
			ix.add(Edge{FromType: KindQuest, FromID: id, Field: prefix + ".targetId", ToType: KindItem, ToID: objective.TargetID, How: "quest objective collect"})
		case quests.ObjectiveTalk:
			ix.add(Edge{FromType: KindQuest, FromID: id, Field: prefix + ".targetId", ToType: KindNPC, ToID: objective.TargetID, How: "quest objective talk"})
		case quests.ObjectiveDeliver:
			ix.add(Edge{FromType: KindQuest, FromID: id, Field: prefix + ".targetId", ToType: KindItem, ToID: objective.TargetID, How: "quest objective deliver"})
			ix.add(Edge{FromType: KindQuest, FromID: id, Field: prefix + ".deliverToNpcId", ToType: KindNPC, ToID: objective.DeliverToNPCID, How: "quest objective deliver"})
		}
	}
}
