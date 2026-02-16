package commands

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	m "github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
)

// SelectCharacterCommand ... select a character
type SelectCharacterCommand struct {
}

// Key ...
func (command *SelectCharacterCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute ... executes scream command
func (command *SelectCharacterCommand) Execute(game def.GameCtrl, message *messages.Message) bool {

	parts := strings.Fields(message.Data)
	characterName := strings.Join(parts[1:], " ")

	if characters, err := game.GetFacade().CharactersService().FindByName(characterName); err == nil {

		for _, character := range characters {
			if character.Name == characterName && character.BelongsUserID == message.FromUser.ID {
				// found character to select
				handleCharacterSelected(game, message.FromUser, character)
				return true
			}
		}
	}
	game.SendMessage() <- message.Reply("Could not select character: " + characterName)
	return true
}

func handleCharacterSelected(game def.GameCtrl, user *entities.User, character *characters.Character) {

	// Normalize attribute short names to uppercase (migration for pre-fix characters)
	character.NormalizeAttributeShorts()

	// Ensure mana is initialized for caster classes (migration for pre-mana characters)
	expectedMaxMana := character.CalculateMaxMana()
	if expectedMaxMana > 0 && character.MaxMana == 0 {
		character.MaxMana = expectedMaxMana
		character.CurrentMana = expectedMaxMana
		game.GetFacade().CharactersService().Update(character.ID, character)
	}

	// Retroactive migration: grant distributable attribute points for existing leveled characters
	if character.Level > 1 && character.UnspentAttributePoints == 0 && len(character.SpentAttributePoints) == 0 {
		retroactivePoints := (character.Level - 1) * int32(leveling.PointsPerLevel)
		character.UnspentAttributePoints = retroactivePoints
		game.GetFacade().CharactersService().Update(character.ID, character)
	}

	// handle Character deselection
	if user.LastCharacter != "" && user.LastCharacter != character.ID {
		if character, err := game.GetFacade().CharactersService().FindByID(user.LastCharacter); err == nil {
			if room, err := game.GetFacade().RoomsService().FindByID(character.CurrentRoomID); err == nil {

				// remove character from current room
				// send all players a left room message
				game.SendMessage() <- messages.CharacterLeftRoom{
					MessageResponse: messages.MessageResponse{
						Audience:   m.MessageAudienceRoomWithoutOrigin,
						AudienceID: room.ID,
						OriginID:   character.ID,
						Message:    character.Name + " left.",
					},
				}

				room.RemoveCharacter(character.ID)
				game.GetFacade().RoomsService().Update(room.ID, room)
			}
		}
	}

	// update player
	user.LastCharacter = character.ID
	game.GetFacade().UsersService().Update(user.RefID, user)

	characterSelected := &messages.CharacterSelected{
		MessageResponse: messages.MessageResponse{
			Audience:   messages.MessageAudienceOrigin,
			AudienceID: user.ID,
			Type:       messages.MessageTypeCharacterSelected,
			Message:    fmt.Sprintf("You are now playing as [%v]", character.Name),
		},
		Character:      character,
		XPForNextLevel: leveling.GetXPRequired(character.Level + 1),
	}

	game.SendMessage() <- characterSelected

	var currentRoom *rooms.Room
	var err error

	if character.CurrentRoomID != "" {
		if currentRoom, err = game.GetFacade().RoomsService().FindByID(character.CurrentRoomID); err != nil {
			log.WithField("room", character.CurrentRoomID).Warn("CurrentRoomID for player not found (room might have been deleted or temporary)")
			// set to ""
			character.CurrentRoomID = ""
		}
	}

	// new character or not part of a room?
	if character.CurrentRoomID == "" {
		// find a random room to start in or get starting room
		rooms, _ := game.GetFacade().RoomsService().FindAll()

		if len(rooms) > 0 {
			// TOOD make this random or select a starting room
			currentRoom = rooms[0]

			//TODO: send this as message
			character.CurrentRoomID = currentRoom.ID
			game.GetFacade().CharactersService().Update(character.ID, character)

		}
	}

	// update room // send these state change messages via channel
	currentRoom.AddCharacter(character.ID)
	game.GetFacade().RoomsService().Update(currentRoom.ID, currentRoom)

	enterRoom := m.NewEnterRoomMessage(util.RoomWithCharacterReveals(currentRoom, character), user, game, character)
	enterRoom.AudienceID = user.ID
	game.SendMessage() <- enterRoom

	game.SendMessage() <- messages.CharacterJoinedRoom{
		MessageResponse: messages.MessageResponse{
			Audience:   m.MessageAudienceRoomWithoutOrigin,
			AudienceID: currentRoom.ID,
			OriginID:   character.ID,
			Message:    character.Name + " entered.",
		},
	}

	// Send initial character stats (HP, mana, XP, etc.) to the client
	game.SendMessage() <- m.NewCharacterUpdateMessage(user.ID, character)

	// Send initial inventory state to the client
	game.SendMessage() <- &m.InventoryUpdateMessage{
		MessageResponse: m.MessageResponse{
			Audience:   m.MessageAudienceOrigin,
			AudienceID: user.ID,
			Type:       m.MessageTypeInventoryUpdate,
		},
		Inventory:     character.Inventory,
		EquippedItems: character.EquippedItems,
		Gold:          character.Gold,
	}

	// Send initial quest log to the client
	sendQuestLogToPlayer(game, user.ID, character.ID)

}

// sendQuestLogToPlayer sends the full quest log with enriched quest details
func sendQuestLogToPlayer(game def.GameCtrl, userID, characterID string) {
	progressList, err := game.GetFacade().QuestsService().GetQuestLog(characterID)
	if err != nil {
		return
	}

	// Build enriched quest log entries
	entries := make([]m.QuestLogEntry, 0, len(progressList))
	for _, progress := range progressList {
		quest, err := game.GetFacade().QuestsService().FindByID(progress.QuestID)
		if err != nil || quest == nil {
			continue
		}

		// Build objective progress list
		objectives := make([]m.QuestObjectiveProgress, len(progress.Objectives))
		for i, op := range progress.Objectives {
			// Find matching objective definition to get description
			objDesc := ""
			objRequired := op.Required
			for _, questObj := range quest.Objectives {
				if questObj.ID == op.ObjectiveID {
					objDesc = questObj.Description
					if questObj.Amount > 0 {
						objRequired = questObj.Amount
					}
					break
				}
			}

			objectives[i] = m.QuestObjectiveProgress{
				ObjectiveID: op.ObjectiveID,
				Description: objDesc,
				Current:     op.Current,
				Required:    objRequired,
				Completed:   op.Completed,
			}
		}

		entry := m.QuestLogEntry{
			QuestID:     progress.QuestID,
			QuestName:   quest.Name,
			Status:      string(progress.Status),
			Description: quest.Description,
			Category:    quest.Category,
			Level:       quest.Level,
			Objectives:  objectives,
			Rewards: &m.QuestReward{
				XP:              quest.Rewards.XP,
				Gold:            quest.Rewards.Gold,
				ItemTemplateIDs: quest.Rewards.ItemTemplateIDs,
			},
		}

		if !progress.AcceptedAt.IsZero() {
			entry.AcceptedAt = progress.AcceptedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		if !progress.CompletedAt.IsZero() {
			entry.CompletedAt = progress.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		}

		entries = append(entries, entry)
	}

	// Send quest log message
	game.SendMessage() <- m.NewQuestLogMessage(userID, entries)
}
