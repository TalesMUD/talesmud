package commands

import (
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// notifyLevelUp sends a level-up WS message when quest/combat XP produced a LevelUpResult.
// Callers still send characterUpdate separately so HUD stats refresh.
func notifyLevelUp(game def.GameCtrl, userID string, levelUp *leveling.LevelUpResult) {
	if game == nil || userID == "" || levelUp == nil || levelUp.LevelsGained <= 0 {
		return
	}
	game.SendMessage() <- messages.MessageResponse{
		Audience:   messages.MessageAudienceUser,
		AudienceID: userID,
		Type:       messages.MessageTypeLevelUp,
		Message:    levelUp.Message,
	}
}
