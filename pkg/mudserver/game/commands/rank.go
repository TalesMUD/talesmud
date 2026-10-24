package commands

import (
	"strings"
	"sync"
	"time"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/ranking"
	"github.com/talesmud/talesmud/pkg/ruleset"
)

// RankCommand prints the top of the public board and the caller's rank.
type RankCommand struct {
	mu   sync.Mutex
	at   time.Time
	rows []ranking.Placed
}

type rankKey struct{}

func (rankKey) Matches(key, cmd string) bool {
	return strings.EqualFold(strings.TrimSpace(cmd), key)
}

// Key matches the bare words rank and leaderboard.
func (command *RankCommand) Key() CommandKey { return rankKey{} }

// Execute loads a short-lived snapshot and replies to the caller.
func (command *RankCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	userID := ""
	if message != nil && message.FromUser != nil {
		userID = message.FromUser.ID
	}
	rows, err := command.snapshot(game)
	if err != nil {
		game.SendMessage() <- messages.Reply(userID, "The leaderboard is unavailable.")
		return true
	}
	var you *ranking.Placed
	if message != nil && message.Character != nil {
		you = ranking.Find(rows, message.Character.ID)
	}
	game.SendMessage() <- messages.Reply(userID, ranking.Format(rows, you, 10))
	return true
}

func (command *RankCommand) snapshot(game def.GameCtrl) ([]ranking.Placed, error) {
	now := time.Now()
	ttl := ruleset.RankingCache()
	command.mu.Lock()
	defer command.mu.Unlock()
	if ttl > 0 && !command.at.IsZero() && now.Sub(command.at) < ttl {
		return append([]ranking.Placed(nil), command.rows...), nil
	}
	facade := game.GetFacade()
	chars, err := facade.CharactersService().FindAll()
	if err != nil {
		return nil, err
	}
	users, err := facade.UsersService().FindAll()
	if err != nil {
		return nil, err
	}
	rows := ranking.Collect(chars, users, ruleset.ItemScoreWeights())
	command.rows = rows
	command.at = now
	return append([]ranking.Placed(nil), rows...), nil
}
