package importer

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/scripts"
	"github.com/talesmud/talesmud/pkg/service/validation"
)

// validateData runs all consistency checks on loaded YAML data.
// All issues are recorded via w.addError(). Returns the number of warnings found.
func (w *WorldImporter) validateData(
	yamlRooms []*YAMLRoom,
	yamlItems []*YAMLItem,
	yamlNPCs []*YAMLNPC,
	yamlScripts []*YAMLScript,
	yamlDialogs []*YAMLDialog,
	yamlLootTables []*YAMLLootTable,
	yamlSpawners []*YAMLSpawner,
	yamlQuests []*YAMLQuest,
	yamlSkills []*YAMLSkill,
) int {
	warnings := 0

	// Build ID lookup sets
	roomIDs := make(map[string]bool, len(yamlRooms))
	for _, r := range yamlRooms {
		roomIDs[r.ID] = true
	}
	itemIDs := make(map[string]bool, len(yamlItems))
	for _, i := range yamlItems {
		itemIDs[i.ID] = true
	}
	npcIDs := make(map[string]bool, len(yamlNPCs))
	for _, n := range yamlNPCs {
		npcIDs[n.ID] = true
	}
	scriptIDs := make(map[string]bool, len(yamlScripts))
	scriptMap := make(map[string]*YAMLScript, len(yamlScripts))
	for _, s := range yamlScripts {
		scriptIDs[s.ID] = true
		scriptMap[s.ID] = s
	}
	dialogIDs := make(map[string]bool, len(yamlDialogs))
	for _, d := range yamlDialogs {
		dialogIDs[d.ID] = true
	}
	lootTableIDs := make(map[string]bool, len(yamlLootTables))
	for _, lt := range yamlLootTables {
		lootTableIDs[lt.ID] = true
	}
	questIDs := make(map[string]bool, len(yamlQuests))
	for _, q := range yamlQuests {
		questIDs[q.ID] = true
	}

	// 1. Duplicate ID checks
	warnings += checkDuplicateIDs(w, "room", yamlRooms, func(r *YAMLRoom) string { return r.ID })
	warnings += checkDuplicateIDs(w, "item", yamlItems, func(i *YAMLItem) string { return i.ID })
	warnings += checkDuplicateIDs(w, "NPC", yamlNPCs, func(n *YAMLNPC) string { return n.ID })
	warnings += checkDuplicateIDs(w, "script", yamlScripts, func(s *YAMLScript) string { return s.ID })
	warnings += checkDuplicateIDs(w, "dialog", yamlDialogs, func(d *YAMLDialog) string { return d.ID })
	warnings += checkDuplicateIDs(w, "loot table", yamlLootTables, func(lt *YAMLLootTable) string { return lt.ID })
	warnings += checkDuplicateIDs(w, "spawner", yamlSpawners, func(s *YAMLSpawner) string { return s.ID })
	warnings += checkDuplicateIDs(w, "quest", yamlQuests, func(q *YAMLQuest) string { return q.ID })
	warnings += checkDuplicateIDs(w, "skill", yamlSkills, func(s *YAMLSkill) string { return s.ID })

	// 2. Room cross-reference checks
	for _, r := range yamlRooms {
		// Exit targets
		for _, exit := range r.Exits {
			if exit.Target != "" && !roomIDs[exit.Target] {
				w.addValidation("Room %s (%s): exit '%s' targets unknown room %s", r.ID, r.Name, exit.Name, exit.Target)
				warnings++
			}
		}
		// OnEnter script
		if r.OnEnter != "" && !scriptIDs[r.OnEnter] {
			w.addValidation("Room %s (%s): onEnterScript references unknown script %s", r.ID, r.Name, r.OnEnter)
			warnings++
		}
		// Action scripts
		for _, action := range r.Actions {
			if action.Type == "script" && action.ScriptId != "" && !scriptIDs[action.ScriptId] {
				w.addValidation("Room %s (%s): action '%s' references unknown script %s", r.ID, r.Name, action.Name, action.ScriptId)
				warnings++
			}
		}
		// Room items
		for _, item := range r.Items {
			if item.ID != "" && !itemIDs[item.ID] {
				w.addValidation("Room %s (%s): references unknown item %s", r.ID, r.Name, item.ID)
				warnings++
			}
		}
	}

	// 3. NPC cross-reference checks
	for _, n := range yamlNPCs {
		if n.SpawnRoomId != "" && !roomIDs[n.SpawnRoomId] {
			w.addValidation("NPC %s (%s): spawnRoomId references unknown room %s", n.ID, n.Name, n.SpawnRoomId)
			warnings++
		}
		if n.DialogID != "" && !dialogIDs[n.DialogID] {
			w.addValidation("NPC %s (%s): dialogID references unknown dialog %s", n.ID, n.Name, n.DialogID)
			warnings++
		}
		if n.EnemyTrait != nil {
			if n.EnemyTrait.LootTableID != "" && !lootTableIDs[n.EnemyTrait.LootTableID] {
				w.addValidation("NPC %s (%s): enemyTrait.lootTableId references unknown loot table %s", n.ID, n.Name, n.EnemyTrait.LootTableID)
				warnings++
			}
			for _, field := range []struct{ name, val string }{
				{"onAggroScript", n.EnemyTrait.OnAggroScript},
				{"onDeathScript", n.EnemyTrait.OnDeathScript},
				{"onFleeScript", n.EnemyTrait.OnFleeScript},
			} {
				if field.val != "" && !scriptIDs[field.val] {
					w.addValidation("NPC %s (%s): enemyTrait.%s references unknown script %s", n.ID, n.Name, field.name, field.val)
					warnings++
				}
			}
		}
		if n.MerchantTrait != nil {
			for _, mi := range n.MerchantTrait.Inventory {
				if mi.ItemTemplateID != "" && !itemIDs[mi.ItemTemplateID] {
					w.addValidation("NPC %s (%s): merchant inventory references unknown item %s", n.ID, n.Name, mi.ItemTemplateID)
					warnings++
				}
			}
		}
	}

	// 4. Item cross-reference checks
	for _, i := range yamlItems {
		if i.OnUseScript != "" && !scriptIDs[i.OnUseScript] {
			w.addValidation("Item %s (%s): onUseScript references unknown script %s", i.ID, i.Name, i.OnUseScript)
			warnings++
		}
	}

	// 5. Loot table cross-reference checks
	for _, lt := range yamlLootTables {
		for _, entry := range lt.Entries {
			if entry.ItemTemplateID != "" && !itemIDs[entry.ItemTemplateID] {
				w.addValidation("Loot table %s (%s): entry references unknown item %s", lt.ID, lt.Name, entry.ItemTemplateID)
				warnings++
			}
		}
	}

	// 6. Spawner cross-reference checks
	for _, s := range yamlSpawners {
		if s.TemplateID != "" && !npcIDs[s.TemplateID] {
			w.addValidation("Spawner %s (%s): templateId references unknown NPC %s", s.ID, s.Name, s.TemplateID)
			warnings++
		}
		if s.RoomID != "" && !roomIDs[s.RoomID] {
			w.addValidation("Spawner %s (%s): roomId references unknown room %s", s.ID, s.Name, s.RoomID)
			warnings++
		}
	}

	// 7. Quest cross-reference checks
	for _, q := range yamlQuests {
		if q.Source.NPCID != "" && !npcIDs[q.Source.NPCID] {
			w.addValidation("Quest %s (%s): source.npcId references unknown NPC %s", q.ID, q.Name, q.Source.NPCID)
			warnings++
		}
		if q.Source.ItemID != "" && !itemIDs[q.Source.ItemID] {
			w.addValidation("Quest %s (%s): source.itemId references unknown item %s", q.ID, q.Name, q.Source.ItemID)
			warnings++
		}
		if q.OnCompleteScriptID != "" && !scriptIDs[q.OnCompleteScriptID] {
			w.addValidation("Quest %s (%s): onCompleteScriptId references unknown script %s", q.ID, q.Name, q.OnCompleteScriptID)
			warnings++
		}
		for _, reqID := range q.RequiredQuestIDs {
			if !questIDs[reqID] {
				w.addValidation("Quest %s (%s): requiredQuestIds references unknown quest %s", q.ID, q.Name, reqID)
				warnings++
			}
		}
		for _, obj := range q.Objectives {
			if obj.CheckScriptID != "" && !scriptIDs[obj.CheckScriptID] {
				w.addValidation("Quest %s (%s): objective '%s' checkScriptId references unknown script %s", q.ID, q.Name, obj.Description, obj.CheckScriptID)
				warnings++
			}
			if obj.DeliverToNPCID != "" && !npcIDs[obj.DeliverToNPCID] {
				w.addValidation("Quest %s (%s): objective '%s' deliverToNpcId references unknown NPC %s", q.ID, q.Name, obj.Description, obj.DeliverToNPCID)
				warnings++
			}
		}
		for _, rewardItemID := range q.Rewards.ItemTemplateIDs {
			if !itemIDs[rewardItemID] {
				w.addValidation("Quest %s (%s): reward references unknown item %s", q.ID, q.Name, rewardItemID)
				warnings++
			}
		}
	}

	// 8. Dialog internal consistency checks
	for _, d := range yamlDialogs {
		if d.Tree == nil {
			continue
		}
		for nodeKey, node := range d.Tree {
			for _, opt := range node.Options {
				if opt.Next != "" {
					if _, exists := d.Tree[opt.Next]; !exists {
						w.addValidation("Dialog %s (%s): node '%s' option '%s' references unknown node '%s'", d.ID, d.Name, nodeKey, opt.PlayerText, opt.Next)
						warnings++
					}
				}
			}
		}
	}

	// 9. Lua script code validation
	// Build a map of which scripts are used as onEnterScript (type: room context)
	roomOnEnterScripts := make(map[string]bool)
	for _, r := range yamlRooms {
		if r.OnEnter != "" {
			roomOnEnterScripts[r.OnEnter] = true
		}
	}

	for _, s := range yamlScripts {
		if s.Language != "lua" || s.Code == "" {
			continue
		}
		warnings += w.validateScriptCode(s.ID, s.Name, s.Type, s.Code, roomOnEnterScripts[s.ID])
	}

	return warnings
}

// addValidation logs a validation warning and appends it to the errors list.
func (w *WorldImporter) addValidation(format string, args ...interface{}) {
	msg := fmt.Sprintf("[VALIDATE] "+format, args...)
	w.errors = append(w.errors, msg)
	log.Warn(msg)
}

// checkDuplicateIDs detects duplicate IDs within a single entity type.
func checkDuplicateIDs[T any](w *WorldImporter, entityType string, entities []*T, getID func(*T) string) int {
	seen := make(map[string]int, len(entities))
	warnings := 0
	for _, e := range entities {
		id := getID(e)
		seen[id]++
		if seen[id] == 2 { // Report only on the second occurrence
			w.addValidation("Duplicate %s ID: %s", entityType, id)
			warnings++
		}
	}
	return warnings
}

// validateScriptCode checks Lua script code for common issues:
// - Unknown tales.game.* function calls
// - Wrong argument counts for known functions
// - Context variable misuse (ctx.roomID in onEnter scripts)
func (w *WorldImporter) validateScriptCode(scriptID, scriptName, scriptType, code string, isOnEnterScript bool) int {
	result := validation.ValidateLuaScript(&scripts.Script{
		Entity:      nil,
		Name:        scriptName,
		Code:        code,
		Type:        scripts.ScriptType(scriptType),
		Language:    scripts.ScriptLanguageLua,
	}, isOnEnterScript)
	warnings := 0
	for _, issue := range result.Issues {
		if issue.Severity == validation.SeverityWarning {
			w.addValidation("%s", issue.Message)
			warnings++
		}
	}
	return warnings
}
