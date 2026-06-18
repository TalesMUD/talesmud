package service

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/quests"
)

type questValidationError struct {
	errs []string
}

func (e *questValidationError) Error() string {
	return "invalid quest definition: " + strings.Join(e.errs, "; ")
}

func validateQuestDefinition(quest *quests.Quest) error {
	if quest == nil {
		return &questValidationError{errs: []string{"quest is required"}}
	}

	var errs []string
	add := func(format string, args ...interface{}) {
		errs = append(errs, fmt.Sprintf(format, args...))
	}

	if strings.TrimSpace(quest.Name) == "" {
		add("name is required")
	}
	if strings.TrimSpace(quest.Description) == "" {
		add("description is required")
	}

	switch quest.Source.Type {
	case "npc":
		if strings.TrimSpace(quest.Source.NPCID) == "" {
			add("source.npcId is required for npc-sourced quests")
		}
	case "item":
		if strings.TrimSpace(quest.Source.ItemID) == "" {
			add("source.itemId is required for item-sourced quests")
		}
	case "auto", "script":
	case "":
		add("source.type is required")
	default:
		add("source.type %q is invalid", quest.Source.Type)
	}

	if len(quest.Objectives) == 0 {
		add("objectives must include at least one objective")
	}
	if quest.Rewards.XP < 0 {
		add("rewards.xp cannot be negative")
	}
	if quest.Rewards.Gold < 0 {
		add("rewards.gold cannot be negative")
	}
	for i, id := range quest.Rewards.ItemTemplateIDs {
		if strings.TrimSpace(id) == "" {
			add("rewards.itemTemplateIds[%d] cannot be empty", i)
		}
	}
	for i, id := range quest.RequiredQuestIDs {
		if strings.TrimSpace(id) == "" {
			add("requiredQuestIds[%d] cannot be empty", i)
			continue
		}
		if quest.Entity != nil && quest.ID != "" && id == quest.ID {
			add("requiredQuestIds[%d] cannot reference this quest", i)
		}
	}

	seenObjectives := map[string]int{}
	for i, obj := range quest.Objectives {
		path := fmt.Sprintf("objectives[%d]", i)
		if strings.TrimSpace(obj.ID) == "" {
			add("%s.id is required", path)
		} else if first, exists := seenObjectives[obj.ID]; exists {
			add("%s.id duplicates objectives[%d].id", path, first)
		} else {
			seenObjectives[obj.ID] = i
		}

		if strings.TrimSpace(obj.Description) == "" {
			add("%s.description is required", path)
		}
		if obj.Amount < 0 {
			add("%s.amount cannot be negative", path)
		}

		switch obj.Type {
		case quests.ObjectiveKill, quests.ObjectiveCollect, quests.ObjectiveVisit:
			if strings.TrimSpace(obj.TargetID) == "" {
				add("%s.targetId is required for %s objectives", path, obj.Type)
			}
		case quests.ObjectiveDeliver:
			if strings.TrimSpace(obj.TargetID) == "" {
				add("%s.targetId is required for deliver objectives", path)
			}
			if strings.TrimSpace(obj.DeliverToNPCID) == "" {
				add("%s.deliverToNpcId is required for deliver objectives", path)
			}
		case quests.ObjectiveTalk:
			if strings.TrimSpace(obj.TargetID) == "" && strings.TrimSpace(obj.DialogNodeID) == "" {
				add("%s.targetId or %s.dialogNodeId is required for talk objectives", path, path)
			}
		case quests.ObjectiveCustom:
			if strings.TrimSpace(obj.CheckScriptID) == "" {
				add("%s.checkScriptId is required for custom objectives", path)
			}
		case "":
			add("%s.type is required", path)
		default:
			add("%s.type %q is invalid", path, obj.Type)
		}
	}

	if len(errs) > 0 {
		return &questValidationError{errs: errs}
	}
	return nil
}

func validateQuestReferences(quest *quests.Quest, facade Facade) error {
	if facade == nil || quest == nil {
		return nil
	}

	var errs []string
	addMissing := func(field, kind, id string) {
		errs = append(errs, fmt.Sprintf("%s references missing %s %q", field, kind, id))
	}

	if quest.Source.Type == "npc" && quest.Source.NPCID != "" {
		if found, err := facade.NPCsService().FindByID(quest.Source.NPCID); err != nil || found == nil {
			addMissing("source.npcId", "NPC", quest.Source.NPCID)
		}
	}
	if quest.Source.Type == "item" && quest.Source.ItemID != "" {
		if found, err := facade.ItemsService().FindByID(quest.Source.ItemID); err != nil || found == nil {
			addMissing("source.itemId", "item", quest.Source.ItemID)
		}
	}
	if quest.OnCompleteScriptID != "" {
		if found, err := facade.ScriptsService().FindByID(quest.OnCompleteScriptID); err != nil || found == nil {
			addMissing("onCompleteScriptId", "script", quest.OnCompleteScriptID)
		}
	}
	for i, id := range quest.Rewards.ItemTemplateIDs {
		if id == "" {
			continue
		}
		if found, err := facade.ItemsService().FindByID(id); err != nil || found == nil {
			addMissing(fmt.Sprintf("rewards.itemTemplateIds[%d]", i), "item", id)
		}
	}
	for i, id := range quest.RequiredQuestIDs {
		if id == "" {
			continue
		}
		if found, err := facade.QuestsService().FindByID(id); err != nil || found == nil {
			addMissing(fmt.Sprintf("requiredQuestIds[%d]", i), "quest", id)
		}
	}
	for i, obj := range quest.Objectives {
		path := fmt.Sprintf("objectives[%d]", i)
		switch obj.Type {
		case quests.ObjectiveKill, quests.ObjectiveTalk:
			if obj.TargetID != "" {
				if found, err := facade.NPCsService().FindByID(obj.TargetID); err != nil || found == nil {
					addMissing(path+".targetId", "NPC", obj.TargetID)
				}
			}
		case quests.ObjectiveCollect:
			if obj.TargetID != "" {
				if found, err := facade.ItemsService().FindByID(obj.TargetID); err != nil || found == nil {
					addMissing(path+".targetId", "item", obj.TargetID)
				}
			}
		case quests.ObjectiveDeliver:
			if obj.TargetID != "" {
				if found, err := facade.ItemsService().FindByID(obj.TargetID); err != nil || found == nil {
					addMissing(path+".targetId", "item", obj.TargetID)
				}
			}
			if obj.DeliverToNPCID != "" {
				if found, err := facade.NPCsService().FindByID(obj.DeliverToNPCID); err != nil || found == nil {
					addMissing(path+".deliverToNpcId", "NPC", obj.DeliverToNPCID)
				}
			}
		case quests.ObjectiveVisit:
			if obj.TargetID != "" {
				if found, err := facade.RoomsService().FindByID(obj.TargetID); err != nil || found == nil {
					addMissing(path+".targetId", "room", obj.TargetID)
				}
			}
		case quests.ObjectiveCustom:
			if obj.CheckScriptID != "" {
				if found, err := facade.ScriptsService().FindByID(obj.CheckScriptID); err != nil || found == nil {
					addMissing(path+".checkScriptId", "script", obj.CheckScriptID)
				}
			}
		}
	}

	if len(errs) > 0 {
		return &questValidationError{errs: errs}
	}
	return nil
}
