package commands

import (
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
)

func clearStaleCombatState(game def.GameCtrl, char *characters.Character) {
	if char == nil || (!char.InCombat && char.CombatInstanceID == "") {
		return
	}

	char.InCombat = false
	char.CombatInstanceID = ""
	game.GetFacade().CharactersService().Update(char.ID, char)
}

func isInActiveCombat(game def.GameCtrl, char *characters.Character, combatEngine def.CombatEngineCtrl) bool {
	if char == nil {
		return false
	}
	if combatEngine == nil {
		clearStaleCombatState(game, char)
		return false
	}
	if combatEngine.IsPlayerInCombat(char.ID) {
		return true
	}

	clearStaleCombatState(game, char)
	return false
}
