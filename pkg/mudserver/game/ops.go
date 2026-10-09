package game

import (
	"encoding/json"
	"fmt"

	"github.com/talesmud/talesmud/pkg/entities/audit"
)

// OpError is a live-ops refusal with an HTTP status.
type OpError struct {
	Status int
	Msg    string
}

func (e *OpError) Error() string {
	if e == nil {
		return ""
	}
	return e.Msg
}

func (e *OpError) StatusCode() int {
	if e == nil || e.Status == 0 {
		return 500
	}
	return e.Status
}

func opErr(status int, msg string) error {
	return &OpError{Status: status, Msg: msg}
}

// OpResult is one successful live op, including the inverse that undoes it.
type OpResult struct {
	Summary    string
	Undoable   bool
	Inverse    *audit.Inverse
	Before     json.RawMessage
	After      json.RawMessage
	EntityType string
	EntityID   string
	Detail     map[string]interface{}
}

func (r *OpResult) detail() map[string]interface{} {
	if r == nil || r.Detail == nil {
		return map[string]interface{}{}
	}
	return r.Detail
}

func jsonRaw(v interface{}) json.RawMessage {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

func inverseOf(action string, body interface{}) *audit.Inverse {
	return &audit.Inverse{Action: action, Body: jsonRaw(body)}
}

// RunOp decodes one admin action and runs it on the caller.
// Callers that serve HTTP must already be inside Game.Call.
func (g *Game) RunOp(action string, body json.RawMessage) (*OpResult, error) {
	if g == nil || g.Facade == nil {
		return nil, opErr(503, "game is not running")
	}
	switch action {
	case "teleport":
		var in teleportBody
		if err := decodeOp(body, &in); err != nil {
			return nil, err
		}
		return g.OpTeleport(in.CharacterID, in.RoomID, in.Force)
	case "give-item":
		var in giveBody
		if err := decodeOp(body, &in); err != nil {
			return nil, err
		}
		return g.OpGiveItem(in.CharacterID, in.ItemTemplateID, in.Quantity)
	case "take-item":
		var in takeBody
		if err := decodeOp(body, &in); err != nil {
			return nil, err
		}
		return g.OpTakeItem(in.CharacterID, in.ItemInstanceID, in.ItemTemplateID, in.Quantity)
	case "npc-heal":
		var in npcIDBody
		if err := decodeOp(body, &in); err != nil {
			return nil, err
		}
		return g.OpNPCHeal(in.NPCInstanceID)
	case "npc-respawn":
		var in respawnBody
		if err := decodeOp(body, &in); err != nil {
			return nil, err
		}
		return g.OpNPCRespawn(in.NPCTemplateID, in.RoomID)
	case "npc-despawn":
		var in npcIDBody
		if err := decodeOp(body, &in); err != nil {
			return nil, err
		}
		return g.OpNPCDespawn(in.NPCInstanceID)
	case "end-combat":
		var in combatBody
		if err := decodeOp(body, &in); err != nil {
			return nil, err
		}
		return g.OpEndCombat(in.CombatInstanceID, in.CharacterID, in.NPCInstanceID)
	case "instance-cleanup":
		var in cleanupBody
		if err := decodeOp(body, &in); err != nil {
			return nil, err
		}
		return g.OpInstanceCleanup(in.RoomCopyID, in.AllEmpty || in.AllEmptyDash)
	case "quest-step":
		var in questBody
		if err := decodeOp(body, &in); err != nil {
			return nil, err
		}
		return g.OpQuestStep(in.CharacterID, in.QuestID, in.ObjectiveID, in.Op)
	case "regrant-starter-kit":
		var in regrantBody
		if err := decodeOp(body, &in); err != nil {
			return nil, err
		}
		return g.OpRegrantStarterKit(in.CharacterID)
	default:
		return nil, opErr(404, "unknown op")
	}
}

func decodeOp(body json.RawMessage, dest interface{}) error {
	if len(body) == 0 {
		body = []byte("{}")
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return opErr(400, "invalid JSON")
	}
	return nil
}

type teleportBody struct {
	CharacterID string `json:"characterId"`
	RoomID      string `json:"roomId"`
	Force       bool   `json:"force"`
}

type giveBody struct {
	CharacterID    string `json:"characterId"`
	ItemTemplateID string `json:"itemTemplateId"`
	Quantity       int32  `json:"quantity"`
}

type takeBody struct {
	CharacterID    string `json:"characterId"`
	ItemInstanceID string `json:"itemInstanceId"`
	ItemTemplateID string `json:"itemTemplateId"`
	Quantity       int32  `json:"quantity"`
}

type npcIDBody struct {
	NPCInstanceID string `json:"npcInstanceId"`
}

type respawnBody struct {
	NPCTemplateID string `json:"npcTemplateId"`
	RoomID        string `json:"roomId"`
}

type combatBody struct {
	CombatInstanceID string `json:"combatInstanceId"`
	CharacterID      string `json:"characterId"`
	NPCInstanceID    string `json:"npcInstanceId"`
}

type cleanupBody struct {
	RoomCopyID   string `json:"roomCopyId"`
	AllEmpty     bool   `json:"allEmpty"`
	AllEmptyDash bool   `json:"all-empty"`
}

type questBody struct {
	CharacterID string `json:"characterId"`
	QuestID     string `json:"questId"`
	ObjectiveID string `json:"objectiveId"`
	Op          string `json:"op"`
}

type regrantBody struct {
	CharacterID string `json:"characterId"`
}

func qtyOrOne(n int32) (int32, error) {
	if n == 0 {
		return 1, nil
	}
	if n < 0 {
		return 0, opErr(400, "quantity must be at least 1")
	}
	return n, nil
}

func opSummary(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
