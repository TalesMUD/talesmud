package commands

import (
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
)

// NotifyQuestItemPickup notifies the quest tracker of an item pickup
func NotifyQuestItemPickup(game def.GameCtrl, characterID, userID string, item *items.Item) {
	if qt := game.GetQuestTracker(); qt != nil {
		qt.OnItemPickup(characterID, userID, item)
	}
}

// NotifyQuestRoomEnter notifies the quest tracker of a room entry
func NotifyQuestRoomEnter(game def.GameCtrl, characterID, userID string, room *rooms.Room) {
	if room != nil && room.Area != "" && game.GetFacade() != nil {
		game.GetFacade().QuestsService().GrantAutoQuests(characterID, room.Area)
	}
	if qt := game.GetQuestTracker(); qt != nil {
		qt.OnRoomEnter(characterID, userID, room)
	}
}

// NotifyQuestDialogNode notifies the quest tracker of a dialog node reached
func NotifyQuestDialogNode(game def.GameCtrl, characterID, userID, npcID, dialogID, nodeID string) {
	if qt := game.GetQuestTracker(); qt != nil {
		qt.OnDialogNode(characterID, userID, npcID, dialogID, nodeID)
	}
}

// NotifyQuestTalkToNPC notifies the quest tracker of talking to an NPC
func NotifyQuestTalkToNPC(game def.GameCtrl, characterID, userID string, talkNPC *npc.NPC) {
	if qt := game.GetQuestTracker(); qt != nil {
		qt.OnTalkToNPC(characterID, userID, talkNPC)
	}
}
