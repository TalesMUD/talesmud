package service

import (
	"fmt"
	"sort"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/repository"
)

const (
	ValidationSeverityError   = "error"
	ValidationSeverityWarning = "warning"
)

// ValidationIssue describes a content problem found by the world validator.
type ValidationIssue struct {
	Severity   string `json:"severity"`
	System     string `json:"system"`
	EntityType string `json:"entityType"`
	EntityID   string `json:"entityId,omitempty"`
	Field      string `json:"field,omitempty"`
	Message    string `json:"message"`
}

// ValidationReport is the response returned by world validation.
type ValidationReport struct {
	Issues       []ValidationIssue `json:"issues"`
	ErrorCount   int               `json:"errorCount"`
	WarningCount int               `json:"warningCount"`
}

// WorldValidationService validates cross-entity world references.
type WorldValidationService interface {
	Validate() (*ValidationReport, error)
}

type worldValidationService struct {
	facade Facade
}

func NewWorldValidationService(facade Facade) WorldValidationService {
	return &worldValidationService{facade: facade}
}

func (s *worldValidationService) Validate() (*ValidationReport, error) {
	ctx := &validationContext{service: s}
	if err := ctx.load(); err != nil {
		return nil, err
	}

	ctx.validateRooms()
	ctx.validateNPCs()
	ctx.validateSpawners()
	ctx.validateLootTables()
	ctx.validateQuests()
	ctx.validateItems()
	ctx.validateCharacterTemplates()

	sort.SliceStable(ctx.issues, func(i, j int) bool {
		a := ctx.issues[i]
		b := ctx.issues[j]
		if a.Severity != b.Severity {
			return a.Severity < b.Severity
		}
		if a.System != b.System {
			return a.System < b.System
		}
		if a.EntityType != b.EntityType {
			return a.EntityType < b.EntityType
		}
		if a.EntityID != b.EntityID {
			return a.EntityID < b.EntityID
		}
		if a.Field != b.Field {
			return a.Field < b.Field
		}
		return a.Message < b.Message
	})

	report := &ValidationReport{Issues: ctx.issues}
	for _, issue := range ctx.issues {
		switch issue.Severity {
		case ValidationSeverityError:
			report.ErrorCount++
		case ValidationSeverityWarning:
			report.WarningCount++
		}
	}
	return report, nil
}

type validationContext struct {
	service *worldValidationService

	rooms              []*rooms.Room
	npcs               []*npc.NPC
	spawners           []*npc.NPCSpawner
	dialogs            map[string]bool
	scripts            map[string]bool
	items              []*items.Item
	itemIDs            map[string]bool
	templates          map[string]bool
	lootTables         []*items.LootTable
	lootIDs            map[string]bool
	quests             []*quests.Quest
	questIDs           map[string]bool
	characterTemplates []*characters.CharacterTemplate

	roomIDs map[string]bool
	npcIDs  map[string]bool

	issues []ValidationIssue
}

func (ctx *validationContext) load() error {
	var err error
	ctx.rooms, err = ctx.service.facade.RoomsService().FindAll()
	if err != nil {
		return fmt.Errorf("load rooms: %w", err)
	}
	ctx.npcs, err = ctx.service.facade.NPCsService().FindAll()
	if err != nil {
		return fmt.Errorf("load npcs: %w", err)
	}
	ctx.spawners, err = ctx.service.facade.NPCSpawnersService().FindAll()
	if err != nil {
		return fmt.Errorf("load spawners: %w", err)
	}
	dialogs, err := ctx.service.facade.DialogsService().FindAll()
	if err != nil {
		return fmt.Errorf("load dialogs: %w", err)
	}
	scripts, err := ctx.service.facade.ScriptsService().FindAll()
	if err != nil {
		return fmt.Errorf("load scripts: %w", err)
	}
	ctx.items, err = ctx.service.facade.ItemsService().FindAll(repository.ItemsQuery{})
	if err != nil {
		return fmt.Errorf("load items: %w", err)
	}
	ctx.lootTables, err = ctx.service.facade.LootTablesService().FindAll()
	if err != nil {
		return fmt.Errorf("load loot tables: %w", err)
	}
	ctx.quests, err = ctx.service.facade.QuestsService().FindAll()
	if err != nil {
		return fmt.Errorf("load quests: %w", err)
	}
	ctx.characterTemplates, err = ctx.service.facade.CharacterTemplatesRepo().FindAll()
	if err != nil {
		return fmt.Errorf("load character templates: %w", err)
	}

	ctx.roomIDs = make(map[string]bool, len(ctx.rooms))
	for _, room := range ctx.rooms {
		if room != nil && room.ID != "" {
			ctx.roomIDs[room.ID] = true
		}
	}
	ctx.npcIDs = make(map[string]bool, len(ctx.npcs))
	for _, n := range ctx.npcs {
		if n != nil && n.ID != "" {
			ctx.npcIDs[n.ID] = true
		}
	}
	ctx.dialogs = make(map[string]bool, len(dialogs))
	for _, dialog := range dialogs {
		if dialog != nil && dialog.ID != "" {
			ctx.dialogs[dialog.ID] = true
		}
	}
	ctx.scripts = make(map[string]bool, len(scripts))
	for _, script := range scripts {
		if script != nil && script.ID != "" {
			ctx.scripts[script.ID] = true
		}
	}
	ctx.itemIDs = make(map[string]bool, len(ctx.items))
	ctx.templates = make(map[string]bool, len(ctx.items))
	for _, item := range ctx.items {
		if item == nil || item.ID == "" {
			continue
		}
		ctx.itemIDs[item.ID] = true
		if item.IsTemplate {
			ctx.templates[item.ID] = true
		}
	}
	ctx.lootIDs = make(map[string]bool, len(ctx.lootTables))
	for _, table := range ctx.lootTables {
		if table != nil && table.ID != "" {
			ctx.lootIDs[table.ID] = true
		}
	}
	ctx.questIDs = make(map[string]bool, len(ctx.quests))
	for _, quest := range ctx.quests {
		if quest != nil && quest.ID != "" {
			ctx.questIDs[quest.ID] = true
		}
	}
	return nil
}

func (ctx *validationContext) add(severity, system, entityType, entityID, field, message string) {
	ctx.issues = append(ctx.issues, ValidationIssue{
		Severity:   severity,
		System:     system,
		EntityType: entityType,
		EntityID:   entityID,
		Field:      field,
		Message:    message,
	})
}

func (ctx *validationContext) requireRoom(system, entityType, entityID, field, roomID string) {
	if roomID != "" && !ctx.roomIDs[roomID] {
		ctx.add(ValidationSeverityError, system, entityType, entityID, field, "references missing room "+roomID)
	}
}

func (ctx *validationContext) requireNPC(system, entityType, entityID, field, npcID string) {
	if npcID != "" && !ctx.npcIDs[npcID] {
		ctx.add(ValidationSeverityError, system, entityType, entityID, field, "references missing NPC "+npcID)
	}
}

func (ctx *validationContext) requireDialog(system, entityType, entityID, field, dialogID string) {
	if dialogID != "" && !ctx.dialogs[dialogID] {
		ctx.add(ValidationSeverityError, system, entityType, entityID, field, "references missing dialog "+dialogID)
	}
}

func (ctx *validationContext) requireScript(system, entityType, entityID, field, scriptID string) {
	if scriptID != "" && !ctx.scripts[scriptID] {
		ctx.add(ValidationSeverityError, system, entityType, entityID, field, "references missing script "+scriptID)
	}
}

func (ctx *validationContext) requireItem(system, entityType, entityID, field, itemID string) {
	if itemID != "" && !ctx.itemIDs[itemID] {
		ctx.add(ValidationSeverityError, system, entityType, entityID, field, "references missing item "+itemID)
	}
}

func (ctx *validationContext) requireTemplate(system, entityType, entityID, field, templateID string) {
	if templateID != "" && !ctx.templates[templateID] {
		ctx.add(ValidationSeverityError, system, entityType, entityID, field, "references missing item template "+templateID)
	}
}

func (ctx *validationContext) requireLootTable(system, entityType, entityID, field, lootTableID string) {
	if lootTableID != "" && !ctx.lootIDs[lootTableID] {
		ctx.add(ValidationSeverityError, system, entityType, entityID, field, "references missing loot table "+lootTableID)
	}
}

func (ctx *validationContext) validateRooms() {
	for _, room := range ctx.rooms {
		if room == nil {
			continue
		}
		if room.Name == "" {
			ctx.add(ValidationSeverityWarning, "creator", "room", room.ID, "name", "room has no name")
		}
		ctx.requireScript("itemization", "room", room.ID, "onEnterScriptID", room.OnEnterScriptID)
		if room.Exits != nil {
			for i, exit := range *room.Exits {
				field := fmt.Sprintf("exits[%d].target", i)
				if exit.Target == "" {
					ctx.add(ValidationSeverityWarning, "player-session", "room", room.ID, field, "exit has no target room")
					continue
				}
				ctx.requireRoom("player-session", "room", room.ID, field, exit.Target)
			}
		}
		if room.Items != nil {
			for i, itemID := range *room.Items {
				ctx.requireItem("item", "room", room.ID, fmt.Sprintf("items[%d]", i), itemID)
			}
		}
		if room.NPCs != nil {
			for i, npcID := range *room.NPCs {
				ctx.requireNPC("npc", "room", room.ID, fmt.Sprintf("npcs[%d]", i), npcID)
			}
		}
		if room.Actions != nil {
			for i, action := range *room.Actions {
				ctx.requireScript("creator", "room", room.ID, fmt.Sprintf("actions[%d].scriptId", i), action.ScriptId)
			}
		}
	}
}

func (ctx *validationContext) validateNPCs() {
	for _, n := range ctx.npcs {
		if n == nil {
			continue
		}
		if n.Name == "" {
			ctx.add(ValidationSeverityWarning, "npc", "npc", n.ID, "name", "NPC has no name")
		}
		ctx.requireRoom("npc", "npc", n.ID, "currentRoomID", n.CurrentRoomID)
		ctx.requireRoom("npc", "npc", n.ID, "spawnRoomId", n.SpawnRoomID)
		ctx.requireDialog("npc", "npc", n.ID, "dialogID", n.DialogID)
		ctx.requireDialog("npc", "npc", n.ID, "idleDialogID", n.IdleDialogID)
		for i, roomID := range n.PatrolPath {
			ctx.requireRoom("npc", "npc", n.ID, fmt.Sprintf("patrolPath[%d]", i), roomID)
		}
		if n.MerchantTrait != nil {
			ctx.validateMerchant(n)
		}
		if n.EnemyTrait != nil {
			ctx.validateEnemy(n)
		}
	}
}

func (ctx *validationContext) validateMerchant(n *npc.NPC) {
	mt := n.MerchantTrait
	if mt.BuyMultiplier <= 0 {
		ctx.add(ValidationSeverityWarning, "item", "npc", n.ID, "merchantTrait.buyMultiplier", "buy multiplier should be greater than 0")
	}
	if mt.SellMultiplier < 0 {
		ctx.add(ValidationSeverityWarning, "item", "npc", n.ID, "merchantTrait.sellMultiplier", "sell multiplier should not be negative")
	}
	for i, entry := range mt.Inventory {
		prefix := fmt.Sprintf("merchantTrait.inventory[%d]", i)
		if entry.ItemTemplateID == "" {
			ctx.add(ValidationSeverityError, "item", "npc", n.ID, prefix+".itemTemplateId", "merchant stock item is missing an item template ID")
		} else {
			ctx.requireTemplate("item", "npc", n.ID, prefix+".itemTemplateId", entry.ItemTemplateID)
		}
		if entry.Quantity < -1 {
			ctx.add(ValidationSeverityWarning, "item", "npc", n.ID, prefix+".quantity", "quantity should be -1 for unlimited or 0+")
		}
		if entry.MaxQuantity < 0 {
			ctx.add(ValidationSeverityWarning, "item", "npc", n.ID, prefix+".maxQuantity", "max quantity should not be negative")
		}
		if entry.RequiredLevel < 0 {
			ctx.add(ValidationSeverityWarning, "item", "npc", n.ID, prefix+".requiredLevel", "required level should not be negative")
		}
	}
}

func (ctx *validationContext) validateEnemy(n *npc.NPC) {
	enemy := n.EnemyTrait
	if n.MaxHitPoints <= 0 {
		ctx.add(ValidationSeverityWarning, "combat", "npc", n.ID, "maxHitPoints", "enemy has no max hit points")
	}
	if enemy.AttackPower < 0 {
		ctx.add(ValidationSeverityWarning, "combat", "npc", n.ID, "enemyTrait.attackPower", "attack power should not be negative")
	}
	if enemy.Defense < 0 {
		ctx.add(ValidationSeverityWarning, "combat", "npc", n.ID, "enemyTrait.defense", "defense should not be negative")
	}
	if enemy.XPReward < 0 {
		ctx.add(ValidationSeverityWarning, "combat", "npc", n.ID, "enemyTrait.xpReward", "XP reward should not be negative")
	}
	if enemy.GoldDrop.Min < 0 || enemy.GoldDrop.Max < 0 || enemy.GoldDrop.Min > enemy.GoldDrop.Max {
		ctx.add(ValidationSeverityWarning, "combat", "npc", n.ID, "enemyTrait.goldDrop", "gold drop range should be non-negative and min <= max")
	}
	ctx.requireLootTable("item", "npc", n.ID, "enemyTrait.lootTableId", enemy.LootTableID)
	for i, templateID := range enemy.GuaranteedLoot {
		ctx.requireTemplate("item", "npc", n.ID, fmt.Sprintf("enemyTrait.guaranteedLoot[%d]", i), templateID)
	}
	ctx.requireScript("combat", "npc", n.ID, "enemyTrait.onAggroScript", enemy.OnAggroScript)
	ctx.requireScript("combat", "npc", n.ID, "enemyTrait.onDeathScript", enemy.OnDeathScript)
	ctx.requireScript("combat", "npc", n.ID, "enemyTrait.onFleeScript", enemy.OnFleeScript)
}

func (ctx *validationContext) validateSpawners() {
	for _, spawner := range ctx.spawners {
		if spawner == nil {
			continue
		}
		ctx.requireRoom("npc", "spawner", spawner.ID, "roomId", spawner.RoomID)
		if spawner.TemplateID == "" {
			ctx.add(ValidationSeverityError, "npc", "spawner", spawner.ID, "templateId", "spawner is missing an NPC template ID")
		} else if !ctx.npcIDs[spawner.TemplateID] {
			ctx.add(ValidationSeverityError, "npc", "spawner", spawner.ID, "templateId", "references missing NPC template "+spawner.TemplateID)
		}
		if spawner.MaxInstances < 0 {
			ctx.add(ValidationSeverityWarning, "npc", "spawner", spawner.ID, "maxInstances", "max instances should not be negative")
		}
		if spawner.InitialCount < 0 {
			ctx.add(ValidationSeverityWarning, "npc", "spawner", spawner.ID, "initialCount", "initial count should not be negative")
		}
	}
}

func (ctx *validationContext) validateLootTables() {
	for _, table := range ctx.lootTables {
		if table == nil {
			continue
		}
		if table.Name == "" {
			ctx.add(ValidationSeverityWarning, "item", "lootTable", table.ID, "name", "loot table has no name")
		}
		for i, entry := range table.Entries {
			prefix := fmt.Sprintf("entries[%d]", i)
			ctx.requireTemplate("item", "lootTable", table.ID, prefix+".itemTemplateId", entry.ItemTemplateID)
			if entry.DropChance < 0 || entry.DropChance > 1 {
				ctx.add(ValidationSeverityWarning, "item", "lootTable", table.ID, prefix+".dropChance", "drop chance should be between 0 and 1")
			}
			if entry.MinQuantity < 0 || entry.MaxQuantity < 0 || entry.MinQuantity > entry.MaxQuantity {
				ctx.add(ValidationSeverityWarning, "item", "lootTable", table.ID, prefix+".quantity", "quantity range should be non-negative and min <= max")
			}
		}
	}
}

func (ctx *validationContext) validateQuests() {
	for _, quest := range ctx.quests {
		if quest == nil {
			continue
		}
		if quest.Name == "" {
			ctx.add(ValidationSeverityWarning, "quest", "quest", quest.ID, "name", "quest has no name")
		}
		switch quest.Source.Type {
		case "npc":
			ctx.requireNPC("quest", "quest", quest.ID, "source.npcId", quest.Source.NPCID)
		case "item":
			ctx.requireTemplate("quest", "quest", quest.ID, "source.itemId", quest.Source.ItemID)
		}
		for i, objective := range quest.Objectives {
			ctx.validateQuestObjective(quest, objective, i)
		}
		for i, templateID := range quest.Rewards.ItemTemplateIDs {
			ctx.requireTemplate("quest", "quest", quest.ID, fmt.Sprintf("rewards.itemTemplateIds[%d]", i), templateID)
		}
		for i, requiredQuestID := range quest.RequiredQuestIDs {
			if requiredQuestID != "" && !ctx.questIDs[requiredQuestID] {
				ctx.add(ValidationSeverityError, "quest", "quest", quest.ID, fmt.Sprintf("requiredQuestIds[%d]", i), "references missing prerequisite quest "+requiredQuestID)
			}
		}
		ctx.requireScript("quest", "quest", quest.ID, "onCompleteScriptId", quest.OnCompleteScriptID)
	}
}

func (ctx *validationContext) validateQuestObjective(quest *quests.Quest, objective quests.Objective, index int) {
	prefix := fmt.Sprintf("objectives[%d]", index)
	if objective.ID == "" {
		ctx.add(ValidationSeverityWarning, "quest", "quest", quest.ID, prefix+".id", "objective has no ID")
	}
	if objective.Amount < 0 {
		ctx.add(ValidationSeverityWarning, "quest", "quest", quest.ID, prefix+".amount", "objective amount should not be negative")
	}
	switch objective.Type {
	case quests.ObjectiveKill:
		ctx.requireNPC("quest", "quest", quest.ID, prefix+".targetId", objective.TargetID)
	case quests.ObjectiveCollect:
		ctx.requireTemplate("quest", "quest", quest.ID, prefix+".targetId", objective.TargetID)
	case quests.ObjectiveDeliver:
		ctx.requireTemplate("quest", "quest", quest.ID, prefix+".targetId", objective.TargetID)
		ctx.requireNPC("quest", "quest", quest.ID, prefix+".deliverToNpcId", objective.DeliverToNPCID)
	case quests.ObjectiveVisit:
		ctx.requireRoom("quest", "quest", quest.ID, prefix+".targetId", objective.TargetID)
	case quests.ObjectiveTalk:
		ctx.requireNPC("quest", "quest", quest.ID, prefix+".targetId", objective.TargetID)
	case quests.ObjectiveCustom:
		if objective.CheckScriptID == "" {
			ctx.add(ValidationSeverityWarning, "quest", "quest", quest.ID, prefix+".checkScriptId", "custom objective has no check script")
		}
	}
	ctx.requireScript("quest", "quest", quest.ID, prefix+".checkScriptId", objective.CheckScriptID)
}

func (ctx *validationContext) validateItems() {
	for _, item := range ctx.items {
		if item == nil {
			continue
		}
		if item.Name == "" {
			ctx.add(ValidationSeverityWarning, "item", "item", item.ID, "name", "item has no name")
		}
		if item.TemplateID != "" {
			ctx.requireTemplate("item", "item", item.ID, "templateId", item.TemplateID)
		}
		for i, contained := range item.Items {
			if contained != nil {
				ctx.requireItem("item", "item", item.ID, fmt.Sprintf("items[%d]", i), contained.ID)
			}
		}
		if item.Stackable {
			if item.MaxStack < 0 || item.Quantity < 0 {
				ctx.add(ValidationSeverityWarning, "item", "item", item.ID, "stack", "stack quantity and max stack should not be negative")
			}
			if item.MaxStack > 0 && item.Quantity > item.MaxStack {
				ctx.add(ValidationSeverityWarning, "item", "item", item.ID, "quantity", "quantity exceeds max stack")
			}
		}
		ctx.requireScript("item", "item", item.ID, "onUseScriptId", item.OnUseScriptID)
	}
}

func (ctx *validationContext) validateCharacterTemplates() {
	for _, template := range ctx.characterTemplates {
		if template == nil {
			continue
		}
		if template.Name == "" {
			ctx.add(ValidationSeverityWarning, "creator", "characterTemplate", template.ID, "name", "character template has no name")
		}
		if template.MaxHitPoints < 0 || template.CurrentHitPoints < 0 {
			ctx.add(ValidationSeverityWarning, "creator", "characterTemplate", template.ID, "hitPoints", "hit points should not be negative")
		}
		if template.MaxMana < 0 || template.CurrentMana < 0 {
			ctx.add(ValidationSeverityWarning, "creator", "characterTemplate", template.ID, "mana", "mana should not be negative")
		}
		for i, startingItem := range template.StartingItems {
			field := fmt.Sprintf("startingItems[%d].itemTemplateId", i)
			if startingItem.ItemTemplateID == "" {
				ctx.add(ValidationSeverityError, "creator", "characterTemplate", template.ID, field, "starting item is missing an item template ID")
				continue
			}
			ctx.requireTemplate("creator", "characterTemplate", template.ID, field, startingItem.ItemTemplateID)
		}
	}
}
