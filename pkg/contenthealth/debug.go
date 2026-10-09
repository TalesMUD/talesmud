package contenthealth

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/worldindex"
)

// QuestStepOpPath is the live-ops route that resets a quest step.
// This slice does not implement that route.
const QuestStepOpPath = "/api/ops/quest-step"

// DebugCharacterInput is one stored quest-progress row joined to its character.
type DebugCharacterInput struct {
	ID            string
	Name          string
	Level         int32
	Guest         bool
	CurrentRoomID string
	Progress      *quests.QuestProgress
}

// QuestDebug is the static plus live view of one quest.
type QuestDebug struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Category      string           `json:"category,omitempty"`
	Area          string           `json:"area,omitempty"`
	Level         int32            `json:"level,omitempty"`
	Repeatable    bool             `json:"repeatable,omitempty"`
	Verdict       string           `json:"verdict"`
	Reason        string           `json:"reason"`
	Offer         DebugParty       `json:"offer"`
	TurnIn        DebugParty       `json:"turnIn"`
	Rewards       DebugRewards     `json:"rewards"`
	Prerequisites []QuestLink      `json:"prerequisites"`
	Next          []QuestLink      `json:"next"`
	Steps         []DebugStep      `json:"steps"`
	Characters    []DebugCharacter `json:"characters"`
	Ops           DebugOps         `json:"ops"`
}

// DebugParty is who offers a quest or who turns it in.
type DebugParty struct {
	Kind     string  `json:"kind"`
	NPCID    string  `json:"npcId,omitempty"`
	ItemID   string  `json:"itemId,omitempty"`
	Anywhere bool    `json:"anywhere,omitempty"`
	Verdict  string  `json:"verdict"`
	Reason   string  `json:"reason,omitempty"`
	Places   []Place `json:"places"`
}

// DebugRewards is what the quest grants.
type DebugRewards struct {
	XP    int32   `json:"xp"`
	Gold  int64   `json:"gold"`
	Items []Place `json:"items"`
}

// QuestLink is a prerequisite or a follow-on quest.
type QuestLink struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Verdict string `json:"verdict,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

// DebugStep is one row in the quest flow.
type DebugStep struct {
	Kind        string  `json:"kind"`
	Label       string  `json:"label"`
	Verdict     string  `json:"verdict"`
	Reason      string  `json:"reason,omitempty"`
	ObjectiveID string  `json:"objectiveId,omitempty"`
	Type        string  `json:"type,omitempty"`
	TargetID    string  `json:"targetId,omitempty"`
	Places      []Place `json:"places"`
}

// DebugCharacter is one character's progress on this quest.
type DebugCharacter struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Level         int32      `json:"level"`
	Guest         bool       `json:"guest"`
	Status        string     `json:"status"`
	Step          string     `json:"step"`
	CurrentRoomID string     `json:"currentRoomId,omitempty"`
	AcceptedAt    *time.Time `json:"acceptedAt,omitempty"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
}

// DebugOps names live operations the panel can call later.
type DebugOps struct {
	QuestStep DebugOp `json:"questStep"`
}

// DebugOp is a hook for a live operation this slice does not perform.
type DebugOp struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Implemented bool   `json:"implemented"`
}

// DebugQuest builds the debugger view for one quest.
// The bool is false when the quest is not in the world snapshot.
func DebugQuest(world World, questID string, characters []DebugCharacterInput) (QuestDebug, bool) {
	questID = strings.TrimSpace(questID)
	snap := world.indexSnapshot()
	quest := snap.Quests[questID]
	if quest == nil || questID == "" {
		return QuestDebug{}, false
	}
	ix := worldindex.Build(snap)
	reach := ix.Reachability()
	report := evaluateQuests(ix, reach)
	ev := report.ByID[questID]
	objVerdict, objReason := verdictGreen, "can complete"
	if ev != nil {
		objVerdict, objReason = ev.verdict()
	}
	offer := offerParty(ix, reach, report, quest)
	turnIn := turnInParty(ix, reach, quest)
	acceptVerdict, acceptReason, acceptPlaces := acceptStep(ix, quest, offer.Verdict)
	verdict, reason := objVerdict, objReason
	verdict, reason = foldVerdict(verdict, reason, offer.Verdict, offer.Reason)
	verdict, reason = foldVerdict(verdict, reason, acceptVerdict, acceptReason)
	verdict, reason = foldVerdict(verdict, reason, turnIn.Verdict, turnIn.Reason)

	out := QuestDebug{
		ID:            questID,
		Name:          quest.Name,
		Category:      quest.Category,
		Area:          quest.DisplayArea(),
		Level:         quest.Level,
		Repeatable:    quest.Repeatable,
		Verdict:       verdict,
		Reason:        reason,
		Offer:         offer,
		TurnIn:        turnIn,
		Rewards:       rewardView(ix, quest),
		Prerequisites: prereqLinks(ix, report, quest),
		Next:          nextLinks(ix, report, questID),
		Steps:         []DebugStep{},
		Characters:    []DebugCharacter{},
		Ops: DebugOps{QuestStep: DebugOp{
			Method:      "POST",
			Path:        QuestStepOpPath,
			Implemented: false,
		}},
	}
	out.Steps = append(out.Steps, DebugStep{
		Kind:    "offer",
		Label:   "Offer",
		Verdict: offer.Verdict,
		Reason:  offer.Reason,
		Places:  orPlaces(offer.Places),
	})
	out.Steps = append(out.Steps, DebugStep{
		Kind:    "accept",
		Label:   "On accept",
		Verdict: acceptVerdict,
		Reason:  acceptReason,
		Places:  acceptPlaces,
	})
	if ev != nil {
		for _, objective := range ev.Objectives {
			target := ""
			if quest != nil && objective.Index >= 0 && objective.Index < len(quest.Objectives) {
				target = quest.Objectives[objective.Index].TargetID
			}
			label := objective.Type
			if label == "" {
				label = "objective"
			}
			if target != "" {
				label = label + " " + target
			}
			out.Steps = append(out.Steps, DebugStep{
				Kind:        "objective",
				Label:       label,
				Verdict:     objective.Verdict,
				Reason:      objective.Reason,
				ObjectiveID: objective.ID,
				Type:        objective.Type,
				TargetID:    target,
				Places:      orPlaces(objective.Places),
			})
		}
	}
	out.Steps = append(out.Steps, DebugStep{
		Kind:    "complete",
		Label:   "Complete",
		Verdict: turnIn.Verdict,
		Reason:  turnIn.Reason + " · " + rewardSummary(quest),
		Places:  orPlaces(turnIn.Places),
	})
	out.Characters = debugCharacters(quest, characters)
	return out, true
}

func orPlaces(places []Place) []Place {
	if places == nil {
		return []Place{}
	}
	return places
}

func foldVerdict(verdict, reason, extra, extraReason string) (string, string) {
	if verdictRank(extra) > verdictRank(verdict) {
		if extraReason != "" {
			return extra, extraReason
		}
		return extra, reason
	}
	if extra == verdict && verdict != verdictGreen && extraReason != "" && extraReason != reason && !strings.Contains(reason, extraReason) {
		if reason == "" || reason == "can complete" {
			return verdict, extraReason
		}
		return verdict, reason + "; " + extraReason
	}
	return verdict, reason
}

func offerParty(ix *worldindex.Index, reach worldindex.Reachability, report questReport, quest *quests.Quest) DebugParty {
	party := DebugParty{Places: []Place{}}
	if quest == nil {
		party.Verdict = verdictAmber
		party.Reason = "quest is missing"
		return party
	}
	kind := strings.ToLower(strings.TrimSpace(quest.Source.Type))
	party.Kind = kind
	party.NPCID = strings.TrimSpace(quest.Source.NPCID)
	party.ItemID = strings.TrimSpace(quest.Source.ItemID)
	if party.Kind == "" {
		party.Kind = "auto"
	}
	snap := ix.Snapshot()
	switch {
	case party.NPCID != "":
		n := snap.NPCs[party.NPCID]
		party.Places = append(party.Places, npcPlaces(ix, reach, party.NPCID)...)
		if n == nil {
			party.Verdict = verdictAmber
			party.Reason = "offer NPC " + party.NPCID + " is not in the world"
		} else if !ix.CanMeet(n, reach) {
			party.Verdict = verdictRed
			party.Reason = "offer NPC " + party.NPCID + " is not reachable"
		} else {
			party.Verdict = verdictGreen
			party.Reason = "offered by " + ix.Name(worldindex.KindNPC, party.NPCID)
		}
	case party.ItemID != "":
		party.Places = append(party.Places, itemPlaces(ix, reach, report, quest.ID, party.ItemID)...)
		if snap.Items[party.ItemID] == nil {
			party.Verdict = verdictAmber
			party.Reason = "offer item " + party.ItemID + " is not in the world"
		} else if !report.Obtainable[party.ItemID] {
			party.Verdict = verdictRed
			party.Reason = "offer item " + party.ItemID + " is not obtainable"
		} else {
			party.Verdict = verdictGreen
			party.Reason = "offered by item " + ix.Name(worldindex.KindItem, party.ItemID)
		}
	default:
		party.Verdict = verdictGreen
		if party.Kind == "script" {
			party.Reason = "offered by a script"
		} else {
			party.Reason = "offered automatically"
		}
	}
	party.Places = append(party.Places, questDialogPlaces(ix, reach, quest.ID, "accept")...)
	sortPlaces(party.Places)
	if party.Places == nil {
		party.Places = []Place{}
	}
	return party
}

func turnInParty(ix *worldindex.Index, reach worldindex.Reachability, quest *quests.Quest) DebugParty {
	party := DebugParty{Places: []Place{}}
	if quest == nil {
		party.Verdict = verdictAmber
		party.Reason = "quest is missing"
		return party
	}
	anywhere, npcID := quest.ResolveTurnIn()
	party.NPCID = npcID
	if anywhere {
		party.Kind = "anywhere"
		party.Anywhere = true
		party.Verdict = verdictGreen
		party.Reason = "turns in anywhere"
	} else if npcID == "" {
		party.Kind = "npc"
		party.Verdict = verdictAmber
		party.Reason = "turn-in NPC is not set"
	} else {
		party.Kind = "npc"
		n := ix.Snapshot().NPCs[npcID]
		party.Places = append(party.Places, npcPlaces(ix, reach, npcID)...)
		if n == nil {
			party.Verdict = verdictAmber
			party.Reason = "turn-in NPC " + npcID + " is not in the world"
		} else if !ix.CanMeet(n, reach) {
			party.Verdict = verdictRed
			party.Reason = "turn-in NPC " + npcID + " is not reachable"
		} else {
			party.Verdict = verdictGreen
			party.Reason = "turns in at " + ix.Name(worldindex.KindNPC, npcID)
		}
	}
	party.Places = append(party.Places, questDialogPlaces(ix, reach, quest.ID, "complete")...)
	sortPlaces(party.Places)
	if party.Places == nil {
		party.Places = []Place{}
	}
	return party
}

func acceptStep(ix *worldindex.Index, quest *quests.Quest, offerVerdict string) (string, string, []Place) {
	if quest == nil || strings.TrimSpace(quest.OnAcceptScriptID) == "" {
		return verdictGreen, "none", []Place{}
	}
	id := strings.TrimSpace(quest.OnAcceptScriptID)
	exists := ix.Snapshot().Scripts[id] != nil
	place := Place{
		Type:      "script",
		ID:        id,
		Name:      ix.Name(worldindex.KindScript, id),
		How:       "on-accept script",
		Reachable: exists && offerVerdict != verdictRed,
	}
	if !exists {
		return verdictAmber, "on-accept script " + id + " is not in the world", []Place{place}
	}
	return verdictGreen, "runs " + id, []Place{place}
}

func rewardView(ix *worldindex.Index, quest *quests.Quest) DebugRewards {
	out := DebugRewards{Items: []Place{}}
	if quest == nil {
		return out
	}
	out.XP = quest.Rewards.XP
	out.Gold = quest.Rewards.Gold
	for _, id := range quest.Rewards.ItemTemplateIDs {
		if id == "" {
			continue
		}
		out.Items = append(out.Items, Place{
			Type:      "item",
			ID:        id,
			Name:      ix.Name(worldindex.KindItem, id),
			How:       "quest reward",
			Reachable: ix.Snapshot().Items[id] != nil,
		})
	}
	return out
}

func rewardSummary(quest *quests.Quest) string {
	if quest == nil {
		return "no rewards"
	}
	var parts []string
	if quest.Rewards.XP != 0 {
		parts = append(parts, fmt.Sprintf("%d XP", quest.Rewards.XP))
	}
	if quest.Rewards.Gold != 0 {
		parts = append(parts, fmt.Sprintf("%d gold", quest.Rewards.Gold))
	}
	for _, id := range quest.Rewards.ItemTemplateIDs {
		if id != "" {
			parts = append(parts, id)
		}
	}
	if len(parts) == 0 {
		return "no rewards"
	}
	return strings.Join(parts, ", ")
}

func prereqLinks(ix *worldindex.Index, report questReport, quest *quests.Quest) []QuestLink {
	out := []QuestLink{}
	if quest == nil {
		return out
	}
	for _, id := range quest.RequiredQuestIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		out = append(out, questLink(ix, report, id))
	}
	return out
}

func nextLinks(ix *worldindex.Index, report questReport, questID string) []QuestLink {
	out := []QuestLink{}
	snap := ix.Snapshot()
	for _, id := range sortedIDs(snap.Quests) {
		quest := snap.Quests[id]
		if quest == nil || id == questID {
			continue
		}
		for _, req := range quest.RequiredQuestIDs {
			if strings.TrimSpace(req) == questID {
				out = append(out, questLink(ix, report, id))
				break
			}
		}
	}
	return out
}

func questLink(ix *worldindex.Index, report questReport, id string) QuestLink {
	link := QuestLink{ID: id, Name: ix.Name(worldindex.KindQuest, id)}
	ev := report.ByID[id]
	if ev == nil {
		link.Verdict = verdictAmber
		link.Reason = "not in the world"
		return link
	}
	link.Verdict, link.Reason = ev.verdict()
	return link
}

func debugCharacters(quest *quests.Quest, characters []DebugCharacterInput) []DebugCharacter {
	out := make([]DebugCharacter, 0, len(characters))
	for _, in := range characters {
		if in.Progress != nil && quest != nil && in.Progress.QuestID != "" && quest.ID != "" && in.Progress.QuestID != quest.ID {
			continue
		}
		row := DebugCharacter{
			ID:            in.ID,
			Name:          in.Name,
			Level:         in.Level,
			Guest:         in.Guest,
			CurrentRoomID: in.CurrentRoomID,
		}
		if in.Progress != nil {
			if row.ID == "" {
				row.ID = in.Progress.CharacterID
			}
			row.Status = string(in.Progress.Status)
			row.Step = progressStep(quest, in.Progress)
			if !in.Progress.AcceptedAt.IsZero() {
				accepted := in.Progress.AcceptedAt
				row.AcceptedAt = &accepted
			}
			if !in.Progress.CompletedAt.IsZero() {
				completed := in.Progress.CompletedAt
				row.CompletedAt = &completed
			}
		}
		if row.Name == "" {
			row.Name = row.ID
		}
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func progressStep(quest *quests.Quest, progress *quests.QuestProgress) string {
	if progress == nil {
		return ""
	}
	if progress.Status == quests.QuestStatusCompleted {
		return "completed"
	}
	if quest == nil {
		return string(progress.Status)
	}
	byID := map[string]quests.ObjectiveProgress{}
	for _, objective := range progress.Objectives {
		byID[objective.ObjectiveID] = objective
	}
	for _, objective := range quest.Objectives {
		op, ok := byID[objective.ID]
		required := objective.Amount
		if required <= 0 {
			required = 1
		}
		current := int32(0)
		if ok {
			current = op.Current
			if op.Required > 0 {
				required = op.Required
			}
			if op.Completed || current >= required {
				continue
			}
		}
		return fmt.Sprintf("%s %d/%d", objective.Type, current, required)
	}
	if progress.Status == quests.QuestStatusActive {
		return "ready"
	}
	return string(progress.Status)
}
