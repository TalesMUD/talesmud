package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/guide"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// GuideCommand handles "guide dismiss".
type GuideCommand struct{}

// Execute dismisses the next-step hint for this character, or refreshes it.
func (c *GuideCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message == nil || message.FromUser == nil || message.Character == nil {
		return true
	}
	parts := strings.Fields(strings.ToLower(message.Data))
	if len(parts) >= 2 && parts[1] == "dismiss" {
		message.Character.SetGuideDismissed()
		if facade := game.GetFacade(); facade != nil && facade.CharactersService() != nil && message.Character.ID != "" {
			_ = facade.CharactersService().Modify(message.Character.ID, func(ch *characters.Character) error {
				ch.SetGuideDismissed()
				return nil
			})
		}
		game.SetUserSessionCharacter(message.FromUser, message.Character)
	}
	PushGuideHint(game, message.FromUser.ID, message.Character)
	return true
}

// Key matches "guide" and "guide dismiss".
func (c *GuideCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// PushGuideHint sends the current next step, or an empty hint that clears the chip.
func PushGuideHint(game def.GameCtrl, userID string, char *characters.Character) {
	PushGuideHintAt(game, userID, char, "")
}

// PushGuideHintAt is PushGuideHint with the room the character just entered.
func PushGuideHintAt(game def.GameCtrl, userID string, char *characters.Character, roomID string) {
	if game == nil || userID == "" {
		return
	}
	text, name, step, command, label := "", "", "", "", ""
	if char != nil {
		world, ok := gatherGuideWorld(game, char, roomID)
		if !ok {
			return
		}
		if hint := guide.Build(world); hint != nil {
			text, name, step, command, label = hint.Text, hint.QuestName, hint.Step, hint.Command, hint.Label
		}
	}
	game.SendMessage() <- messages.NewGuideHintMessage(userID, text, name, step, command, label)
}

func gatherGuideWorld(game def.GameCtrl, char *characters.Character, roomID string) (guide.World, bool) {
	if char != nil && char.GuideDismissed() {
		return guide.World{Dismissed: true}, true
	}
	facade := game.GetFacade()
	if facade == nil || facade.QuestsService() == nil || facade.RoomsService() == nil {
		return guide.World{}, false
	}
	if roomID == "" && char != nil {
		roomID = char.CurrentRoomID
	}
	progress, err := facade.QuestsService().GetQuestLog(char.ID)
	if err != nil {
		return guide.World{}, false
	}
	for _, p := range progress {
		if p != nil && p.Status == quests.QuestStatusCompleted {
			return guide.World{Dismissed: true}, true
		}
	}
	defs, err := facade.QuestsService().FindAll()
	if err != nil {
		return guide.World{}, false
	}
	roomList, err := facade.RoomsService().FindAll()
	if err != nil {
		return guide.World{}, false
	}

	names := map[string]string{}
	var templates []*npc.NPC
	if facade.NPCsService() != nil {
		if all, nerr := facade.NPCsService().FindAll(); nerr == nil {
			templates = all
			for _, n := range all {
				if n == nil || n.Entity == nil {
					continue
				}
				names[n.ID] = n.Name
			}
		}
	}

	places := map[string][]string{}
	addPlace := func(id, room string) {
		id = strings.TrimSpace(id)
		room = strings.TrimSpace(room)
		if id == "" || room == "" {
			return
		}
		for _, have := range places[id] {
			if have == room {
				return
			}
		}
		places[id] = append(places[id], room)
	}

	if facade.NPCSpawnersService() != nil {
		if spawners, serr := facade.NPCSpawnersService().FindAll(); serr == nil {
			for _, sp := range spawners {
				if sp == nil {
					continue
				}
				addPlace(sp.TemplateID, sp.RoomID)
			}
		}
	}
	for _, n := range templates {
		if n == nil || n.Entity == nil || n.IsDead {
			continue
		}
		addPlace(n.ID, n.CurrentRoomID)
		addPlace(n.ID, n.SpawnRoomID)
		if n.TemplateID != "" {
			addPlace(n.TemplateID, n.CurrentRoomID)
			addPlace(n.TemplateID, n.SpawnRoomID)
		}
	}

	worldRooms := make([]guide.Room, 0, len(roomList))
	for _, room := range roomList {
		if room == nil || room.Entity == nil {
			continue
		}
		gr := guide.Room{ID: room.ID}
		if room.Exits != nil {
			for _, ex := range *room.Exits {
				gr.Exits = append(gr.Exits, guide.Exit{
					Name:   ex.Name,
					Target: ex.Target,
					Hidden: ex.Hidden,
					Locked: ex.Locked,
				})
			}
		}
		if room.NPCs != nil {
			for _, id := range *room.NPCs {
				addPlace(id, room.ID)
			}
		}
		worldRooms = append(worldRooms, gr)
	}

	if mgr := game.GetNPCInstanceManager(); mgr != nil {
		for _, inst := range mgr.GetAllInstances() {
			if inst == nil || inst.IsDead {
				continue
			}
			where := inst.CurrentRoomID
			if where == "" {
				where = inst.SpawnRoomID
			}
			id := inst.TemplateID
			if id == "" && inst.Entity != nil {
				id = inst.ID
			}
			addPlace(id, where)
		}
	}

	byQuest := map[string]*quests.QuestProgress{}
	for _, p := range progress {
		if p != nil {
			byQuest[p.QuestID] = p
		}
	}

	views := make([]guide.Quest, 0, len(defs))
	for _, q := range defs {
		if q == nil || q.Entity == nil {
			continue
		}
		view := guide.Quest{
			ID:            q.ID,
			Name:          q.Name,
			Level:         q.Level,
			Guide:         q.Guide,
			SourceType:    q.Source.Type,
			SourceNPCID:   q.Source.NPCID,
			SourceNPCName: names[q.Source.NPCID],
		}
		anywhere, npcID := q.ResolveTurnIn()
		view.TurnInAnywhere = anywhere
		view.TurnInNPCID = npcID
		view.TurnInNPCName = names[npcID]
		if p := byQuest[q.ID]; p != nil {
			view.Status = string(p.Status)
			view.Objectives = objectiveViews(q, p)
		} else {
			view.Objectives = objectiveViews(q, nil)
		}
		views = append(views, view)
	}

	world := guide.World{
		CurrentRoomID: roomID,
		Revealed:      char.RevealedExits,
		Quests:        views,
		Rooms:         worldRooms,
		Places:        places,
		Here:          actorsHere(game, roomID),
	}
	return world, true
}

func objectiveViews(q *quests.Quest, p *quests.QuestProgress) []guide.Objective {
	have := map[string]quests.ObjectiveProgress{}
	if p != nil {
		for _, op := range p.Objectives {
			have[op.ObjectiveID] = op
		}
	}
	out := make([]guide.Objective, 0, len(q.Objectives))
	for _, obj := range q.Objectives {
		ov := guide.Objective{
			Type:             string(obj.Type),
			Description:      obj.Description,
			TargetID:         obj.TargetID,
			TargetName:       obj.TargetName,
			Amount:           obj.Amount,
			Order:            obj.Order,
			DeliverToNPCID:   obj.DeliverToNPCID,
			DeliverToNPCName: obj.DeliverToNPCName,
		}
		if op, ok := have[obj.ID]; ok {
			ov.Current = op.Current
			ov.Completed = op.Completed
		}
		out = append(out, ov)
	}
	return out
}

func actorsHere(game def.GameCtrl, roomID string) []guide.Actor {
	if game == nil || roomID == "" {
		return nil
	}
	mgr := game.GetNPCInstanceManager()
	if mgr == nil {
		return nil
	}
	var out []guide.Actor
	for _, inst := range mgr.GetInstancesInRoom(roomID) {
		if inst == nil || inst.IsDead {
			continue
		}
		id := inst.TemplateID
		if id == "" && inst.Entity != nil {
			id = inst.ID
		}
		out = append(out, guide.Actor{TemplateID: id, Name: inst.GetDisplayName()})
	}
	return out
}
