package contenthealth

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
	"github.com/talesmud/talesmud/pkg/service"
	"github.com/talesmud/talesmud/pkg/service/validation"
	"github.com/talesmud/talesmud/pkg/worldindex"
)

var codeTitles = map[string]string{
	"duplicate_dialog_node":             "Dialog contains a duplicate node",
	"unknown_lua_game_function":         "Script calls an unknown Lua function",
	"wrong_lua_game_function_arg_count": "Script calls a Lua function with the wrong arguments",
	"missing_script":                    "References a missing script",
	"missing_npc":                       "References a missing NPC",
	"missing_item":                      "References a missing item",
	"missing_room":                      "References a missing room",
	"missing_dialog":                    "References a missing dialog",
	"missing_quest":                     "References a missing quest",
	"missing_loot_table":                "References a missing loot table",
	"non_lua_script":                    "Script is not Lua",
	"invalid_stock_quantity":            "Merchant stock quantity is invalid",
}

type catalogEntry struct {
	id       string
	title    string
	severity string
	hint     string
	always   bool
}

func builtinCatalog() []catalogEntry {
	return []catalogEntry{
		{RuleUnreachable, "Rooms unreachable from start", "error", "Add the hidden exit the reveal script names, or wire the reveal script to a reachable action.", true},
		{RuleRevealMissing, "Reveal script targets a missing exit", "error", "Add that hidden exit, or change the script so it reveals an exit that exists.", true},
		{RuleQuest, "Quest objective can never complete", "error", "Make the target reachable, or change the objective.", true},
		{RuleMissingItem, "Room item references a missing item", "error", "Add the item or remove it from the room.", true},
		{RuleDangling, "Exit targets exist", "error", "Point the exit at a room that exists, or remove the exit.", true},
		{RuleBoss, "Unique boss without a spawner", "warning", "Add a spawner for this boss so it can return.", true},
		{RuleUnknownTier, "Unknown difficulty tier", "warning", "Use a tier from the balance table.", true},
		{RuleUnreferenced, "Script not referenced anywhere", "warning", "Wire the script to a room, item, or NPC, or remove it.", true},
		{RuleHiddenNoReveal, "Hidden exit with no revealer", "warning", "Add revealExit from a script the player can run.", true},
	}
}

// Run checks world and returns the health report.
func Run(world World, opt Options) Report {
	snap := world.indexSnapshot()
	ix := worldindex.Build(snap)
	reach := ix.Reachability()
	bal := opt.Balance
	if bal == nil {
		loaded := defaultBalance()
		bal = &loaded
	}
	b := &builder{
		hits:  map[string][]Hit{},
		seen:  map[string]bool{},
		index: ix,
		snap:  snap,
	}
	b.addGraph(reach, bal)
	b.addValidation(world.validationSnapshot(snap))
	b.addWorldValidation(snap)
	b.addPack(opt.Rules, snap)
	b.addLive(opt.Live)
	drift := compareDrift(opt.Baseline, baselineEntries(world))
	return b.report(world.ContentCommit, opt.Rules, opt.Muted, drift)
}

func defaultBalance() Balance {
	cfg := balance.GetConfig()
	out := Balance{Tiers: map[string]bool{}, BossTiers: map[string]bool{}}
	if cfg != nil {
		for key := range cfg.DifficultyMultipliers {
			out.Tiers[strings.ToLower(key)] = true
		}
		for _, tier := range cfg.BossMechanics.PhaseTiers {
			out.BossTiers[strings.ToLower(tier)] = true
		}
		for _, tier := range cfg.BossMechanics.EnrageTiers {
			out.BossTiers[strings.ToLower(tier)] = true
		}
	}
	if len(out.BossTiers) == 0 {
		out.BossTiers["boss"] = true
	}
	return out
}

type builder struct {
	hits      map[string][]Hit
	seen      map[string]bool
	index     *worldindex.Index
	snap      worldindex.Snapshot
	anomalies []Anomaly
}

func (b *builder) add(hit Hit) {
	if hit.RuleID == "" {
		return
	}
	if hit.EntityName == "" && b.index != nil {
		hit.EntityName = b.index.Name(kindOf(hit.EntityType), hit.EntityID)
	}
	if hit.EntityType == "spawner" && b.snap.Spawners != nil {
		if spawner := b.snap.Spawners[hit.EntityID]; spawner != nil && spawner.RoomID != "" && !hasRelated(hit, "room", spawner.RoomID) {
			hit.Related = append(hit.Related, Related{Type: "room", ID: spawner.RoomID})
		}
	}
	key := dedupeKey(hit)
	if b.seen[key] {
		return
	}
	b.seen[key] = true
	b.hits[hit.RuleID] = append(b.hits[hit.RuleID], hit)
}

func dedupeKey(hit Hit) string {
	return normType(hit.EntityType) + "\x00" + hit.EntityID + "\x00" + hit.Field + "\x00" + hit.Severity
}

func normType(entityType string) string {
	entityType = strings.ToLower(entityType)
	entityType = strings.ReplaceAll(entityType, "_", "")
	return entityType
}

func kindOf(entityType string) worldindex.Kind {
	switch normType(entityType) {
	case "room":
		return worldindex.KindRoom
	case "npc":
		return worldindex.KindNPC
	case "item":
		return worldindex.KindItem
	case "loottable":
		return worldindex.KindLootTable
	case "spawner":
		return worldindex.KindSpawner
	case "dialog":
		return worldindex.KindDialog
	case "quest":
		return worldindex.KindQuest
	case "script":
		return worldindex.KindScript
	case "skill":
		return worldindex.KindSkill
	case "charactertemplate":
		return worldindex.KindCharacterTemplate
	default:
		return worldindex.Kind(entityType)
	}
}

func (b *builder) addGraph(reach worldindex.Reachability, bal *Balance) {
	hint := hintFor(RuleUnreachable)
	for _, island := range reach.Islands {
		reason := island.Reason
		if reason == "" {
			reason = "no inbound exit"
		}
		for _, roomID := range island.RoomIDs {
			b.add(Hit{
				RuleID:     RuleUnreachable,
				Severity:   "error",
				EntityType: "room",
				EntityID:   roomID,
				Field:      "exits",
				Message:    roomID + ": " + reason,
				FixHint:    hint,
				Group:      reason,
			})
		}
	}
	hint = hintFor(RuleRevealMissing)
	for _, scriptID := range sortedIDs(b.snap.Scripts) {
		for _, edge := range b.index.Outbound(worldindex.KindScript, scriptID) {
			if edge.How != "revealExit" || edge.ExitName == "" {
				continue
			}
			for _, roomID := range b.index.RevealTargets(edge) {
				room := b.snap.Rooms[roomID]
				if room == nil {
					continue
				}
				if _, ok := room.GetExit(edge.ExitName); ok {
					continue
				}
				b.add(Hit{
					RuleID:     RuleRevealMissing,
					Severity:   "error",
					EntityType: "script",
					EntityID:   scriptID,
					Field:      "code",
					Message:    fmt.Sprintf("revealExit target '%s' missing on %s (%s)", edge.ExitName, roomID, scriptID),
					Related:    []Related{{Type: "room", ID: roomID}},
					FixHint:    hint,
				})
			}
		}
	}
	hint = hintFor(RuleHiddenNoReveal)
	for _, roomID := range sortedIDs(b.snap.Rooms) {
		room := b.snap.Rooms[roomID]
		if room == nil || room.Exits == nil {
			continue
		}
		for i, exit := range *room.Exits {
			if !exit.Hidden || exit.Name == "" {
				continue
			}
			if b.index.ExitHasRevealer(roomID, exit.Name) {
				continue
			}
			b.add(Hit{
				RuleID:     RuleHiddenNoReveal,
				Severity:   "warning",
				EntityType: "room",
				EntityID:   roomID,
				Field:      fmt.Sprintf("exits[%d].hidden", i),
				Message:    fmt.Sprintf("hidden exit '%s' on %s has no revealer", exit.Name, roomID),
				FixHint:    hint,
			})
		}
	}
	hint = hintFor(RuleUnreferenced)
	for _, scriptID := range sortedIDs(b.snap.Scripts) {
		if b.index.ScriptReferenced(scriptID) {
			continue
		}
		b.add(Hit{
			RuleID:     RuleUnreferenced,
			Severity:   "warning",
			EntityType: "script",
			EntityID:   scriptID,
			Field:      "id",
			Message:    scriptID + " is not referenced by a room, item, NPC, quest, or dialog",
			FixHint:    hint,
		})
	}
	b.addQuests(reach)
	b.addBosses(bal)
	b.addUnknownTiers(bal)
}

func (b *builder) addQuests(reach worldindex.Reachability) {
	hint := hintFor(RuleQuest)
	report := evaluateQuests(b.index, reach)
	for _, questID := range sortedIDs(b.snap.Quests) {
		ev := report.ByID[questID]
		if ev == nil {
			continue
		}
		for _, objective := range ev.Objectives {
			if objective.Verdict != verdictRed || len(objective.Problems) == 0 {
				continue
			}
			b.add(Hit{
				RuleID:     RuleQuest,
				Severity:   "error",
				EntityType: "quest",
				EntityID:   questID,
				Field:      fmt.Sprintf("objectives[%d]", objective.Index),
				Message:    questID + " cannot complete: " + strings.Join(objective.Problems, "; "),
				FixHint:    hint,
			})
		}
	}
}

func (b *builder) addBosses(bal *Balance) {
	hint := hintFor(RuleBoss)
	spawned := map[string]bool{}
	for _, spawner := range b.snap.Spawners {
		if spawner != nil && spawner.TemplateID != "" {
			spawned[spawner.TemplateID] = true
		}
	}
	for _, npcID := range sortedIDs(b.snap.NPCs) {
		n := b.snap.NPCs[npcID]
		if n == nil || n.EnemyTrait == nil || isRuntimeNPC(n) {
			continue
		}
		tier := strings.ToLower(strings.TrimSpace(n.EnemyTrait.Difficulty))
		if tier == "" || !bal.BossTiers[tier] {
			continue
		}
		if spawned[npcID] {
			continue
		}
		b.add(Hit{
			RuleID:     RuleBoss,
			Severity:   "warning",
			EntityType: "npc",
			EntityID:   npcID,
			Field:      "enemyTrait.difficulty",
			Message:    fmt.Sprintf("%s is a %s-tier enemy with no spawner", npcID, tier),
			FixHint:    hint,
		})
	}
}

func (b *builder) addUnknownTiers(bal *Balance) {
	hint := hintFor(RuleUnknownTier)
	for _, npcID := range sortedIDs(b.snap.NPCs) {
		n := b.snap.NPCs[npcID]
		if n == nil || n.EnemyTrait == nil || isRuntimeNPC(n) {
			continue
		}
		tier := strings.ToLower(strings.TrimSpace(n.EnemyTrait.Difficulty))
		if tier == "" || bal.Tiers[tier] {
			continue
		}
		b.add(Hit{
			RuleID:     RuleUnknownTier,
			Severity:   "warning",
			EntityType: "npc",
			EntityID:   npcID,
			Field:      "enemyTrait.difficulty",
			Message:    fmt.Sprintf("%s uses unknown difficulty tier %q", npcID, n.EnemyTrait.Difficulty),
			FixHint:    hint,
		})
	}
}

func (b *builder) addValidation(snapshot validation.WorldSnapshot) {
	result := validation.ValidateWorld(snapshot)
	for _, issue := range result.Issues {
		code := routeRuleID(issue.Code, issue.EntityType, issue.Field)
		hit := Hit{
			RuleID:     code,
			Severity:   string(issue.Severity),
			EntityType: issue.EntityType,
			EntityID:   issue.EntityID,
			Field:      issue.Field,
			Message:    issue.Message,
			FixHint:    hintFor(code),
		}
		if issue.RefID != "" {
			hit.Related = []Related{{Type: issue.RefType, ID: issue.RefID}}
			if !strings.Contains(hit.Message, issue.RefID) {
				hit.Message = strings.TrimSpace(hit.Message + " " + issue.RefID)
			}
		}
		b.add(hit)
	}
}

func (b *builder) addWorldValidation(snap worldindex.Snapshot) {
	report := service.ValidateContents(service.WorldContents{
		Rooms:              mapRooms(snap),
		NPCs:               mapNPCs(snap),
		Spawners:           mapSpawners(snap),
		Dialogs:            idSet(sortedIDs(snap.Dialogs)),
		Scripts:            idSet(sortedIDs(snap.Scripts)),
		Items:              mapItems(snap),
		LootTables:         mapLoot(snap),
		Quests:             mapQuests(snap),
		CharacterTemplates: mapTemplates(snap),
	})
	if report == nil {
		return
	}
	for _, issue := range report.Issues {
		code := routeRuleID(worldCode(issue.Message), issue.EntityType, issue.Field)
		b.add(Hit{
			RuleID:     code,
			Severity:   issue.Severity,
			EntityType: issue.EntityType,
			EntityID:   issue.EntityID,
			Field:      issue.Field,
			Message:    issue.Message,
			FixHint:    hintFor(code),
		})
	}
}

func (b *builder) addPack(rules []PackRule, snap worldindex.Snapshot) {
	for _, rule := range rules {
		entityType := rule.Entity
		if entityType == "" {
			entityType = "npc"
		}
		for _, ref := range entitiesOf(snap, entityType) {
			hit, ok := packHit(rule, ref.entity, entityType, ref.id, ref.name)
			if ok {
				hit.FixHint = rule.Message
				b.add(hit)
			}
		}
	}
}

type entityRef struct {
	id     string
	name   string
	entity any
}

func entitiesOf(snap worldindex.Snapshot, entityType string) []entityRef {
	var out []entityRef
	switch normType(entityType) {
	case "room":
		for _, id := range sortedIDs(snap.Rooms) {
			room := snap.Rooms[id]
			out = append(out, entityRef{id, room.Name, room})
		}
	case "npc":
		for _, id := range sortedIDs(snap.NPCs) {
			n := snap.NPCs[id]
			out = append(out, entityRef{id, n.Name, n})
		}
	case "item":
		for _, id := range sortedIDs(snap.Items) {
			item := snap.Items[id]
			out = append(out, entityRef{id, item.Name, item})
		}
	case "loottable":
		for _, id := range sortedIDs(snap.LootTables) {
			table := snap.LootTables[id]
			out = append(out, entityRef{id, table.Name, table})
		}
	case "spawner":
		for _, id := range sortedIDs(snap.Spawners) {
			spawner := snap.Spawners[id]
			out = append(out, entityRef{id, spawner.Name, spawner})
		}
	case "dialog":
		for _, id := range sortedIDs(snap.Dialogs) {
			dialog := snap.Dialogs[id]
			out = append(out, entityRef{id, dialog.Name, dialog})
		}
	case "quest":
		for _, id := range sortedIDs(snap.Quests) {
			quest := snap.Quests[id]
			out = append(out, entityRef{id, quest.Name, quest})
		}
	case "script":
		for _, id := range sortedIDs(snap.Scripts) {
			script := snap.Scripts[id]
			out = append(out, entityRef{id, script.Name, script})
		}
	case "skill":
		for _, id := range sortedIDs(snap.Skills) {
			skill := snap.Skills[id]
			out = append(out, entityRef{id, skill.Name, skill})
		}
	case "charactertemplate":
		for _, id := range sortedIDs(snap.CharacterTemplates) {
			tmpl := snap.CharacterTemplates[id]
			out = append(out, entityRef{id, tmpl.Name, tmpl})
		}
	}
	return out
}

func (b *builder) addLive(live *LiveView) {
	if live == nil {
		return
	}
	rooms := live.RoomIDs
	if rooms == nil {
		rooms = map[string]bool{}
	}
	occupied := map[string]bool{}
	for _, character := range live.Characters {
		if character.CurrentRoomID == "" {
			continue
		}
		occupied[character.CurrentRoomID] = true
		if rooms[character.CurrentRoomID] {
			continue
		}
		b.anomalies = append(b.anomalies, Anomaly{
			Kind:       "missing-room",
			EntityType: "character",
			EntityID:   character.ID,
			Name:       character.Name,
			Message:    fmt.Sprintf("%s is in missing room %s", character.Name, character.CurrentRoomID),
		})
	}
	for _, roomID := range live.InstanceRoomIDs {
		if occupied[roomID] {
			continue
		}
		b.anomalies = append(b.anomalies, Anomaly{
			Kind:       "leaked-instance",
			EntityType: "room",
			EntityID:   roomID,
			Message:    fmt.Sprintf("instance room %s has no character", roomID),
		})
	}
}

func (b *builder) report(commit string, pack []PackRule, muted []string, drift []DriftRow) Report {
	mutedSet := map[string]bool{}
	for _, id := range muted {
		if id != "" {
			mutedSet[id] = true
		}
	}
	meta := map[string]catalogEntry{}
	for _, entry := range builtinCatalog() {
		meta[entry.id] = entry
	}
	for _, rule := range pack {
		title := rule.Title
		if title == "" {
			title = rule.Message
		}
		if title == "" {
			title = rule.ID
		}
		severity := rule.Severity
		if severity == "" {
			severity = "error"
		}
		if _, exists := meta[rule.ID]; !exists {
			meta[rule.ID] = catalogEntry{id: rule.ID, title: title, severity: severity, always: true}
		}
	}
	ids := map[string]bool{}
	for id := range meta {
		if meta[id].always {
			ids[id] = true
		}
	}
	for id, hits := range b.hits {
		if len(hits) > 0 {
			ids[id] = true
		}
	}
	var rules []Rule
	for id := range ids {
		hits := append([]Hit(nil), b.hits[id]...)
		if hits == nil {
			hits = []Hit{}
		}
		sort.Slice(hits, func(i, j int) bool {
			if hits[i].EntityID != hits[j].EntityID {
				return hits[i].EntityID < hits[j].EntityID
			}
			if hits[i].Field != hits[j].Field {
				return hits[i].Field < hits[j].Field
			}
			return hits[i].Message < hits[j].Message
		})
		entry, known := meta[id]
		title := entry.title
		severity := entry.severity
		if !known {
			title = codeTitles[id]
			if title == "" {
				title = humanize(id)
			}
			severity = worstSeverity(hits)
		} else if len(hits) > 0 {
			severity = worstSeverity(hits)
			if entry.severity == "error" && severity == "" {
				severity = entry.severity
			}
		}
		if severity == "" {
			severity = "error"
		}
		rules = append(rules, Rule{
			ID:       id,
			Title:    title,
			Severity: severity,
			Muted:    mutedSet[id],
			Summary:  summarize(id, hits),
			Hits:     hits,
		})
	}
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].Muted != rules[j].Muted {
			return !rules[i].Muted
		}
		if sevRank(rules[i].Severity) != sevRank(rules[j].Severity) {
			return sevRank(rules[i].Severity) < sevRank(rules[j].Severity)
		}
		if rules[i].Title != rules[j].Title {
			return rules[i].Title < rules[j].Title
		}
		return rules[i].ID < rules[j].ID
	})
	report := Report{
		GeneratedAt:   time.Now().UTC(),
		ContentCommit: commit,
		Rules:         rules,
		LiveAnomalies: b.anomalies,
		Drift:         drift,
	}
	if report.LiveAnomalies == nil {
		report.LiveAnomalies = []Anomaly{}
	}
	if report.Rules == nil {
		report.Rules = []Rule{}
	}
	for _, rule := range rules {
		if rule.Muted {
			report.Summary.Muted += len(rule.Hits)
			continue
		}
		for _, hit := range rule.Hits {
			switch hit.Severity {
			case "warning":
				report.Summary.Warnings++
			case "info":
				report.Summary.Info++
			default:
				report.Summary.Errors++
			}
		}
	}
	report.Summary.Drift = len(drift)
	report.Summary.LiveAnomalies = len(report.LiveAnomalies)
	return report
}

func worstSeverity(hits []Hit) string {
	rank := 3
	severity := "info"
	for _, hit := range hits {
		if sevRank(hit.Severity) < rank {
			rank = sevRank(hit.Severity)
			severity = hit.Severity
		}
	}
	if len(hits) == 0 {
		return ""
	}
	return severity
}

func sevRank(severity string) int {
	switch severity {
	case "error":
		return 0
	case "warning":
		return 1
	case "info":
		return 2
	default:
		return 3
	}
}

func summarize(id string, hits []Hit) string {
	if len(hits) == 0 {
		return "No hits"
	}
	if id == RuleUnreachable {
		groups := map[string]bool{}
		for _, hit := range hits {
			groups[hit.Group] = true
		}
		return fmt.Sprintf("%d rooms in %d islands", len(hits), len(groups))
	}
	var labels []string
	seen := map[string]bool{}
	for _, hit := range hits {
		label := hit.EntityID
		if hit.EntityName != "" {
			label = hit.EntityID + " " + hit.EntityName
		}
		if label == "" || seen[label] {
			continue
		}
		seen[label] = true
		labels = append(labels, label)
		if len(labels) == 4 {
			break
		}
	}
	return strings.Join(labels, ", ")
}

func humanize(id string) string {
	id = strings.ReplaceAll(id, "_", " ")
	id = strings.ReplaceAll(id, "-", " ")
	id = strings.TrimSpace(id)
	if id == "" {
		return id
	}
	return strings.ToUpper(id[:1]) + id[1:]
}

func hintFor(id string) string {
	for _, entry := range builtinCatalog() {
		if entry.id == id {
			return entry.hint
		}
	}
	return ""
}

func routeRuleID(code, entityType, field string) string {
	entityType = normType(entityType)
	if code == "missing_room" && entityType == "room" && strings.Contains(field, "exits[") {
		return RuleDangling
	}
	if code == "missing_item" && entityType == "room" && strings.HasPrefix(field, "items[") {
		return RuleMissingItem
	}
	if code == "" {
		return "world-check"
	}
	return code
}

func worldCode(message string) string {
	switch {
	case strings.Contains(message, "no target room"), strings.Contains(message, "missing room"):
		return "missing_room"
	case strings.Contains(message, "missing NPC"):
		return "missing_npc"
	case strings.Contains(message, "missing dialog"):
		return "missing_dialog"
	case strings.Contains(message, "missing script"), strings.Contains(message, "missing check script"):
		return "missing_script"
	case strings.Contains(message, "missing item"):
		return "missing_item"
	case strings.Contains(message, "missing loot table"):
		return "missing_loot_table"
	case strings.Contains(message, "missing prerequisite"):
		return "missing_quest"
	case strings.Contains(message, "quantity should be -1"):
		return "invalid_stock_quantity"
	default:
		return slug(message)
	}
}

func slug(message string) string {
	var parts []string
	for _, field := range strings.Fields(message) {
		trimmed := strings.Trim(field, ".,;:\"'")
		if trimmed == "" || looksLikeID(trimmed) {
			continue
		}
		parts = append(parts, strings.ToLower(trimmed))
	}
	out := strings.Join(parts, "-")
	out = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
			return r
		}
		return '-'
	}, out)
	for strings.Contains(out, "--") {
		out = strings.ReplaceAll(out, "--", "-")
	}
	out = strings.Trim(out, "-")
	if out == "" {
		return "world-check"
	}
	return out
}

func looksLikeID(token string) bool {
	if len(token) < 2 {
		return false
	}
	digits := 0
	for _, r := range token {
		if r >= '0' && r <= '9' {
			digits++
		}
	}
	return digits >= 3
}

func hasRelated(hit Hit, entityType, id string) bool {
	for _, related := range hit.Related {
		if related.Type == entityType && related.ID == id {
			return true
		}
	}
	return false
}

func idSet(ids []string) map[string]bool {
	out := make(map[string]bool, len(ids))
	for _, id := range ids {
		out[id] = true
	}
	return out
}

func mapRooms(snap worldindex.Snapshot) []*rooms.Room {
	out := make([]*rooms.Room, 0, len(snap.Rooms))
	for _, id := range sortedIDs(snap.Rooms) {
		out = append(out, snap.Rooms[id])
	}
	return out
}

func mapNPCs(snap worldindex.Snapshot) []*npc.NPC {
	out := make([]*npc.NPC, 0, len(snap.NPCs))
	for _, id := range sortedIDs(snap.NPCs) {
		out = append(out, snap.NPCs[id])
	}
	return out
}

func mapItems(snap worldindex.Snapshot) []*items.Item {
	out := make([]*items.Item, 0, len(snap.Items))
	for _, id := range sortedIDs(snap.Items) {
		out = append(out, snap.Items[id])
	}
	return out
}

func mapLoot(snap worldindex.Snapshot) []*items.LootTable {
	out := make([]*items.LootTable, 0, len(snap.LootTables))
	for _, id := range sortedIDs(snap.LootTables) {
		out = append(out, snap.LootTables[id])
	}
	return out
}

func mapSpawners(snap worldindex.Snapshot) []*npc.NPCSpawner {
	out := make([]*npc.NPCSpawner, 0, len(snap.Spawners))
	for _, id := range sortedIDs(snap.Spawners) {
		out = append(out, snap.Spawners[id])
	}
	return out
}

func mapQuests(snap worldindex.Snapshot) []*quests.Quest {
	out := make([]*quests.Quest, 0, len(snap.Quests))
	for _, id := range sortedIDs(snap.Quests) {
		out = append(out, snap.Quests[id])
	}
	return out
}

func mapTemplates(snap worldindex.Snapshot) []*characters.CharacterTemplate {
	out := make([]*characters.CharacterTemplate, 0, len(snap.CharacterTemplates))
	for _, id := range sortedIDs(snap.CharacterTemplates) {
		out = append(out, snap.CharacterTemplates[id])
	}
	return out
}

func baselineEntries(world World) []BaselineEntry {
	snap := world.indexSnapshot()
	var entries []BaselineEntry
	add := func(entityType, id, name string, entity any) {
		if id == "" || entity == nil {
			return
		}
		hash, canonical, err := Hash(entity)
		if err != nil {
			return
		}
		entries = append(entries, BaselineEntry{
			Type:      entityType,
			ID:        id,
			Name:      name,
			Hash:      hash,
			Canonical: canonical,
		})
	}
	for _, id := range sortedIDs(snap.Rooms) {
		add("room", id, snap.Rooms[id].Name, snap.Rooms[id])
	}
	for _, id := range sortedIDs(snap.NPCs) {
		add("npc", id, snap.NPCs[id].Name, snap.NPCs[id])
	}
	for _, id := range sortedIDs(snap.Items) {
		add("item", id, snap.Items[id].Name, snap.Items[id])
	}
	for _, id := range sortedIDs(snap.LootTables) {
		add("lootTable", id, snap.LootTables[id].Name, snap.LootTables[id])
	}
	for _, id := range sortedIDs(snap.Spawners) {
		name := snap.Spawners[id].Name
		add("spawner", id, name, snap.Spawners[id])
	}
	for _, id := range sortedIDs(snap.Dialogs) {
		add("dialog", id, snap.Dialogs[id].Name, snap.Dialogs[id])
	}
	for _, id := range sortedIDs(snap.Quests) {
		add("quest", id, snap.Quests[id].Name, snap.Quests[id])
	}
	for _, id := range sortedIDs(snap.Scripts) {
		add("script", id, snap.Scripts[id].Name, snap.Scripts[id])
	}
	for _, id := range sortedIDs(snap.Skills) {
		add("skill", id, snap.Skills[id].Name, snap.Skills[id])
	}
	for _, id := range sortedIDs(snap.CharacterTemplates) {
		add("characterTemplate", id, snap.CharacterTemplates[id].Name, snap.CharacterTemplates[id])
	}
	return entries
}
