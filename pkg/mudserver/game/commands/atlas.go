package commands

import (
	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/worldmap"
)

// PushAtlas sends the character's discovered-world atlas over the websocket
// so the client map does not depend on a REST token lookup.
func PushAtlas(game def.GameCtrl, userID string, character *characters.Character) {
	if game == nil || character == nil || userID == "" {
		return
	}
	rooms, err := game.GetFacade().RoomsService().FindAll()
	if err != nil {
		log.WithError(err).Warn("PushAtlas: failed to load rooms")
		return
	}
	atlas := worldmap.Reveal(worldmap.Compile(rooms), character)
	game.SendMessage() <- messages.NewAtlasMessage(userID, atlas)
}
