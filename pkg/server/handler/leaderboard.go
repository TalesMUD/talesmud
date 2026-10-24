package handler

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/ranking"
	"github.com/talesmud/talesmud/pkg/ruleset"
)

const (
	leaderboardDefaultLimit = 25
	leaderboardMaxLimit     = 100
)

type leaderboardCharacters interface {
	FindAll() ([]*characters.Character, error)
}

type leaderboardUsers interface {
	FindAll() ([]*entities.User, error)
}

// LeaderboardHandler serves a public, cached character board.
type LeaderboardHandler struct {
	Characters leaderboardCharacters
	Users      leaderboardUsers
	Now        func() time.Time

	mu   sync.Mutex
	at   time.Time
	rows []ranking.Placed
}

type leaderboardRow struct {
	Rank      int    `json:"rank"`
	Name      string `json:"name"`
	Class     string `json:"class"`
	Race      string `json:"race"`
	Level     int    `json:"level"`
	ItemScore int    `json:"itemScore"`
}

type leaderboardResponse struct {
	Limit int              `json:"limit"`
	Rows  []leaderboardRow `json:"rows"`
	You   *leaderboardRow  `json:"you"`
}

// Get handles GET /api/leaderboard?limit=&self=
func (h *LeaderboardHandler) Get(c *gin.Context) {
	if h == nil || h.Characters == nil || h.Users == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "leaderboard unavailable"})
		return
	}
	rows, err := h.snapshot()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "leaderboard unavailable"})
		return
	}
	limit := clampLeaderboardLimit(c.Query("limit"))
	if limit > len(rows) {
		limit = len(rows)
	}
	public := make([]leaderboardRow, 0, limit)
	for i := 0; i < limit; i++ {
		public = append(public, toLeaderboardRow(rows[i]))
	}
	var you *leaderboardRow
	if found := ranking.Find(rows, c.Query("self")); found != nil {
		row := toLeaderboardRow(*found)
		you = &row
	}
	c.JSON(http.StatusOK, leaderboardResponse{
		Limit: clampLeaderboardLimit(c.Query("limit")),
		Rows:  public,
		You:   you,
	})
}

func (h *LeaderboardHandler) snapshot() ([]ranking.Placed, error) {
	now := time.Now()
	if h.Now != nil {
		now = h.Now()
	}
	ttl := ruleset.RankingCache()
	h.mu.Lock()
	defer h.mu.Unlock()
	if ttl > 0 && !h.at.IsZero() && now.Sub(h.at) < ttl {
		return append([]ranking.Placed(nil), h.rows...), nil
	}
	chars, err := h.Characters.FindAll()
	if err != nil {
		return nil, err
	}
	users, err := h.Users.FindAll()
	if err != nil {
		return nil, err
	}
	rows := ranking.Collect(chars, users, ruleset.ItemScoreWeights())
	h.rows = rows
	h.at = now
	return append([]ranking.Placed(nil), rows...), nil
}

func toLeaderboardRow(row ranking.Placed) leaderboardRow {
	return leaderboardRow{
		Rank:      row.Rank,
		Name:      row.Name,
		Class:     row.Class,
		Race:      row.Race,
		Level:     row.Level,
		ItemScore: row.ItemScore,
	}
}

func clampLeaderboardLimit(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return leaderboardDefaultLimit
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return leaderboardDefaultLimit
	}
	if n < 1 {
		return 1
	}
	if n > leaderboardMaxLimit {
		return leaderboardMaxLimit
	}
	return n
}
