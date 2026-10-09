// Package inspect aggregates static content for the creator inspectors.
// Live rows stay on the existing live endpoints.
package inspect

import (
	"sort"
	"strings"
	"time"

	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
	"github.com/talesmud/talesmud/pkg/worldindex"
)

// Link is one related entity.
type Link struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Missing bool   `json:"missing,omitempty"`
}

// StatLine is one column of combat numbers.
type StatLine struct {
	MaxHitPoints int32 `json:"maxHitPoints"`
	AttackPower  int32 `json:"attackPower"`
	Defense      int32 `json:"defense"`
}

// StatFactors are the multipliers from the combat balance table.
type StatFactors struct {
	HP      float64 `json:"hp"`
	Attack  float64 `json:"attack"`
	Defense float64 `json:"defense"`
}

// NPCStats is the content base beside the numbers combat uses.
type NPCStats struct {
	Base        *StatLine    `json:"base,omitempty"`
	Effective   StatLine     `json:"effective"`
	AttackSpeed float64      `json:"attackSpeed"`
	UnknownTier bool         `json:"unknownTier"`
	Scaling     string       `json:"scaling"`
	Factors     *StatFactors `json:"factors,omitempty"`
}

// NPCBehavior is the enemy summary the inspector shows next to the stats.
type NPCBehavior struct {
	AggroOnSight  bool    `json:"aggroOnSight"`
	AggroRadius   int     `json:"aggroRadius"`
	FleeThreshold float64 `json:"fleeThreshold"`
	CallForHelp   bool    `json:"callForHelp"`
	XPReward      int64   `json:"xpReward"`
	GoldMin       int32   `json:"goldMin"`
	GoldMax       int32   `json:"goldMax"`
	CombatStyle   string  `json:"combatStyle,omitempty"`
	CreatureType  string  `json:"creatureType,omitempty"`
}

// ScriptHook is one Lua hook on the enemy trait.
type ScriptHook struct {
	Hook      string  `json:"hook"`
	ID        string  `json:"id,omitempty"`
	Name      string  `json:"name,omitempty"`
	Missing   bool    `json:"missing,omitempty"`
	Threshold float64 `json:"threshold,omitempty"`
}

// LootDrop is one loot-table row.
type LootDrop struct {
	ItemID         string  `json:"itemId"`
	ItemName       string  `json:"itemName,omitempty"`
	Missing        bool    `json:"missing,omitempty"`
	DropChance     float64 `json:"dropChance"`
	Guaranteed     bool    `json:"guaranteed,omitempty"`
	Rarity         string  `json:"rarity,omitempty"`
	BossOnly       bool    `json:"bossOnly,omitempty"`
	MinQuantity    int32   `json:"minQuantity,omitempty"`
	MaxQuantity    int32   `json:"maxQuantity,omitempty"`
	MinPlayerLevel int32   `json:"minPlayerLevel,omitempty"`
}

// LootView is the loot table and the guaranteed item list.
type LootView struct {
	TableID    string     `json:"tableId,omitempty"`
	TableName  string     `json:"tableName,omitempty"`
	Missing    bool       `json:"missing,omitempty"`
	Entries    []LootDrop `json:"entries"`
	Guaranteed []Link     `json:"guaranteed"`
}

// SpawnerView is a spawner that creates this NPC.
type SpawnerView struct {
	ID            string `json:"id"`
	Name          string `json:"name,omitempty"`
	RoomID        string `json:"roomId,omitempty"`
	RoomName      string `json:"roomName,omitempty"`
	RoomMissing   bool   `json:"roomMissing,omitempty"`
	MaxInstances  int    `json:"maxInstances"`
	RespawnTime   string `json:"respawnTime,omitempty"`
	RespawnSource string `json:"respawnSource,omitempty"`
	SpawnInterval string `json:"spawnInterval,omitempty"`
}

// QuestHit is a kill, talk, or deliver objective that names this NPC.
type QuestHit struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Role        string `json:"role"`
	ObjectiveID string `json:"objectiveId,omitempty"`
	Description string `json:"description,omitempty"`
}

// NPCView is the static half of the NPC inspector.
// RequestedID is set when the caller asked for a running instance and the view is its template.
type NPCView struct {
	ID              string        `json:"id"`
	RequestedID     string        `json:"requestedId,omitempty"`
	Name            string        `json:"name"`
	Description     string        `json:"description,omitempty"`
	Kinds           []string      `json:"kinds"`
	Level           int32         `json:"level"`
	Difficulty      string        `json:"difficulty,omitempty"`
	IsTemplate      bool          `json:"isTemplate"`
	MaxHitPoints    int32         `json:"maxHitPoints"`
	TemplateRespawn string        `json:"templateRespawn,omitempty"`
	Stats           *NPCStats     `json:"stats,omitempty"`
	Behavior        *NPCBehavior  `json:"behavior,omitempty"`
	Dialog          *Link         `json:"dialog,omitempty"`
	IdleDialog      *Link         `json:"idleDialog,omitempty"`
	Scripts         []ScriptHook  `json:"scripts"`
	Loot            *LootView     `json:"loot,omitempty"`
	Spawners        []SpawnerView `json:"spawners"`
	SpawnRoom       *Link         `json:"spawnRoom,omitempty"`
	Quests          []QuestHit    `json:"quests"`
}

// NPC builds the static inspector for one template or unique NPC.
// id must be the content id. Running instances are not in the index.
func NPC(ix *worldindex.Index, id string) (NPCView, bool) {
	id = strings.TrimSpace(id)
	if ix == nil || id == "" || !ix.Has(worldindex.KindNPC, id) {
		return NPCView{}, false
	}
	n := ix.Snapshot().NPCs[id]
	if n == nil {
		return NPCView{}, false
	}
	view := NPCView{
		ID:              n.ID,
		Name:            n.Name,
		Description:     n.Description,
		Kinds:           npcKinds(n),
		Level:           n.Level,
		IsTemplate:      n.IsTemplate,
		MaxHitPoints:    n.MaxHitPoints,
		TemplateRespawn: durationString(n.RespawnTime),
		Dialog:          link(ix, worldindex.KindDialog, n.DialogID),
		IdleDialog:      link(ix, worldindex.KindDialog, n.IdleDialogID),
		Scripts:         []ScriptHook{},
		Spawners:        spawnersFor(ix, n),
		SpawnRoom:       link(ix, worldindex.KindRoom, n.SpawnRoomID),
		Quests:          questsFor(ix, id),
	}
	if n.ID == "" {
		view.ID = id
	}
	if et := n.EnemyTrait; et != nil {
		view.Difficulty = strings.TrimSpace(et.Difficulty)
		view.Stats = statsFor(n)
		view.Behavior = &NPCBehavior{
			AggroOnSight:  et.AggroOnSight,
			AggroRadius:   et.AggroRadius,
			FleeThreshold: et.FleeThreshold,
			CallForHelp:   et.CallForHelp,
			XPReward:      et.XPReward,
			GoldMin:       et.GoldDrop.Min,
			GoldMax:       et.GoldDrop.Max,
			CombatStyle:   et.CombatStyle,
			CreatureType:  et.CreatureType,
		}
		view.Scripts = hooksFor(ix, et)
		view.Loot = lootFor(ix, et)
	}
	return view, true
}

func npcKinds(n *npc.NPC) []string {
	kinds := []string{}
	if n.IsEnemy() {
		kinds = append(kinds, "enemy")
	}
	if n.IsMerchant() {
		kinds = append(kinds, "merchant")
	}
	if len(kinds) == 0 {
		kinds = append(kinds, "npc")
	}
	return kinds
}

func statsFor(n *npc.NPC) *NPCStats {
	et := n.EnemyTrait
	factors, source, ok := balance.EnemyFactors(et.Difficulty, n.Name)
	stats := &NPCStats{
		AttackSpeed: et.AttackSpeed,
		UnknownTier: strings.TrimSpace(et.Difficulty) != "" && !ok,
	}
	if et.BaseStats != nil {
		stats.Base = &StatLine{
			MaxHitPoints: et.BaseStats.MaxHitPoints,
			AttackPower:  et.BaseStats.AttackPower,
			Defense:      et.BaseStats.Defense,
		}
		hp, attack, defense := balance.ApplyEnemyMultipliers(
			et.BaseStats.MaxHitPoints,
			et.BaseStats.AttackPower,
			et.BaseStats.Defense,
			et.Difficulty,
			n.Name,
		)
		stats.Effective = StatLine{MaxHitPoints: hp, AttackPower: attack, Defense: defense}
		if ok {
			stats.Scaling = source
			stats.Factors = &StatFactors{HP: factors.HP, Attack: factors.Attack, Defense: factors.Defense}
		} else {
			stats.Scaling = "unknown"
		}
		return stats
	}
	stats.Effective = StatLine{
		MaxHitPoints: n.MaxHitPoints,
		AttackPower:  et.AttackPower,
		Defense:      et.Defense,
	}
	stats.Scaling = "stored"
	return stats
}

func hooksFor(ix *worldindex.Index, et *npc.EnemyTrait) []ScriptHook {
	rows := []struct {
		hook string
		id   string
	}{
		{"onAggro", et.OnAggroScript},
		{"onDeath", et.OnDeathScript},
		{"onFlee", et.OnFleeScript},
		{"onLowHealth", et.OnLowHealthScript},
	}
	out := make([]ScriptHook, 0, len(rows))
	for _, row := range rows {
		hook := ScriptHook{Hook: row.hook}
		if link := link(ix, worldindex.KindScript, row.id); link != nil {
			hook.ID = link.ID
			hook.Name = link.Name
			hook.Missing = link.Missing
		}
		if row.hook == "onLowHealth" {
			hook.Threshold = npc.NormalizeLowHealthThreshold(et.LowHealthThreshold)
		}
		out = append(out, hook)
	}
	return out
}

func lootFor(ix *worldindex.Index, et *npc.EnemyTrait) *LootView {
	view := &LootView{Entries: []LootDrop{}, Guaranteed: []Link{}}
	if tableID := strings.TrimSpace(et.LootTableID); tableID != "" {
		view.TableID = tableID
		if table := ix.Snapshot().LootTables[tableID]; table != nil {
			if name := ix.Name(worldindex.KindLootTable, tableID); name != tableID {
				view.TableName = name
			}
			view.Entries = lootEntries(ix, table)
		} else {
			view.Missing = true
		}
	}
	for _, itemID := range et.GuaranteedLoot {
		if item := link(ix, worldindex.KindItem, itemID); item != nil {
			view.Guaranteed = append(view.Guaranteed, *item)
		}
	}
	return view
}

func lootEntries(ix *worldindex.Index, table *items.LootTable) []LootDrop {
	out := make([]LootDrop, 0, len(table.Entries))
	for _, entry := range table.Entries {
		drop := LootDrop{
			ItemID:         entry.ItemTemplateID,
			DropChance:     entry.DropChance,
			Guaranteed:     entry.Guaranteed,
			Rarity:         entry.Rarity,
			BossOnly:       entry.BossOnly,
			MinQuantity:    entry.MinQuantity,
			MaxQuantity:    entry.MaxQuantity,
			MinPlayerLevel: entry.MinPlayerLevel,
		}
		if item := link(ix, worldindex.KindItem, entry.ItemTemplateID); item != nil {
			drop.ItemName = item.Name
			drop.Missing = item.Missing
		} else {
			drop.Missing = true
		}
		out = append(out, drop)
	}
	return out
}

func spawnersFor(ix *worldindex.Index, n *npc.NPC) []SpawnerView {
	out := []SpawnerView{}
	for id, spawner := range ix.Snapshot().Spawners {
		if spawner == nil || strings.TrimSpace(spawner.TemplateID) != n.ID {
			continue
		}
		if spawner.ID != "" {
			id = spawner.ID
		}
		row := SpawnerView{
			ID:            id,
			Name:          spawner.Name,
			RoomID:        spawner.RoomID,
			MaxInstances:  spawner.MaxInstances,
			SpawnInterval: durationString(spawner.SpawnInterval),
		}
		if room := link(ix, worldindex.KindRoom, spawner.RoomID); room != nil {
			row.RoomName = room.Name
			row.RoomMissing = room.Missing
		}
		row.RespawnTime, row.RespawnSource = spawnerRespawn(spawner, n)
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].RoomID != out[j].RoomID {
			return out[i].RoomID < out[j].RoomID
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func spawnerRespawn(spawner *npc.NPCSpawner, n *npc.NPC) (string, string) {
	if spawner.RespawnTimeOverride != nil && *spawner.RespawnTimeOverride > 0 {
		return spawner.RespawnTimeOverride.String(), "override"
	}
	if spawner.RespawnDelay > 0 {
		return spawner.RespawnDelay.String(), "delay"
	}
	if n != nil && n.RespawnTime > 0 {
		return n.RespawnTime.String(), "template"
	}
	return "", "none"
}

func questsFor(ix *worldindex.Index, npcID string) []QuestHit {
	out := []QuestHit{}
	for _, quest := range ix.Snapshot().Quests {
		if quest == nil || quest.Entity == nil {
			continue
		}
		name := ix.Name(worldindex.KindQuest, quest.ID)
		if name == quest.ID {
			name = quest.Name
		}
		for _, objective := range quest.Objectives {
			role := questRole(objective, npcID)
			if role == "" {
				continue
			}
			out = append(out, QuestHit{
				ID:          quest.ID,
				Name:        name,
				Role:        role,
				ObjectiveID: objective.ID,
				Description: objective.Description,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ID != out[j].ID {
			return out[i].ID < out[j].ID
		}
		if out[i].ObjectiveID != out[j].ObjectiveID {
			return out[i].ObjectiveID < out[j].ObjectiveID
		}
		return out[i].Role < out[j].Role
	})
	return out
}

func questRole(objective quests.Objective, npcID string) string {
	target := strings.TrimSpace(objective.TargetID)
	deliver := strings.TrimSpace(objective.DeliverToNPCID)
	switch objective.Type {
	case quests.ObjectiveKill:
		if target == npcID {
			return "kill"
		}
	case quests.ObjectiveTalk:
		if target == npcID {
			return "talk"
		}
	case quests.ObjectiveDeliver:
		if deliver == npcID || target == npcID {
			return "deliver"
		}
	}
	return ""
}

func link(ix *worldindex.Index, kind worldindex.Kind, id string) *Link {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}
	item := &Link{ID: id}
	if !ix.Has(kind, id) {
		item.Missing = true
		return item
	}
	if name := ix.Name(kind, id); name != id {
		item.Name = name
	}
	return item
}

func durationString(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	return d.String()
}
