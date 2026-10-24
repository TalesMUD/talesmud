package ranking

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
)

// Entry is one character before a rank is assigned.
// Created is the character's creation time. There is no level-reached timestamp,
// so an earlier Created wins a tie. A zero Created sorts after any real time.
type Entry struct {
	ID        string
	Name      string
	Class     string
	Race      string
	Level     int
	ItemScore int
	Created   time.Time
}

// Placed is an Entry with a 1-based rank.
type Placed struct {
	Entry
	Rank int
}

// Eligible reports whether a character's account may appear.
// Guests and admin accounts are left off. A missing account is left off.
// Creator accounts stay. There is no separate test-character flag.
func Eligible(found, guest bool, role string) bool {
	if !found || guest {
		return false
	}
	return !strings.EqualFold(strings.TrimSpace(role), entities.RoleAdmin)
}

// Place orders entries by level, item score, earliest Created, then name.
func Place(entries []Entry) []Placed {
	cp := append([]Entry(nil), entries...)
	sort.SliceStable(cp, func(i, j int) bool {
		return before(cp[i], cp[j])
	})
	out := make([]Placed, len(cp))
	for i, entry := range cp {
		out[i] = Placed{Entry: entry, Rank: i + 1}
	}
	return out
}

func before(a, b Entry) bool {
	if a.Level != b.Level {
		return a.Level > b.Level
	}
	if a.ItemScore != b.ItemScore {
		return a.ItemScore > b.ItemScore
	}
	aZero, bZero := a.Created.IsZero(), b.Created.IsZero()
	if aZero != bZero {
		return !aZero
	}
	if !aZero && !a.Created.Equal(b.Created) {
		return a.Created.Before(b.Created)
	}
	an := strings.ToLower(a.Name)
	bn := strings.ToLower(b.Name)
	if an != bn {
		return an < bn
	}
	return a.ID < b.ID
}

// Collect scores eligible characters. Offline characters are included.
// Callers pass every persisted character and user. The result is ranked.
func Collect(chars []*characters.Character, users []*entities.User, w Weights) []Placed {
	byID := map[string]*entities.User{}
	for _, user := range users {
		if user == nil || user.Entity == nil || user.ID == "" {
			continue
		}
		byID[user.ID] = user
	}
	entries := make([]Entry, 0, len(chars))
	for _, character := range chars {
		if character == nil || character.Entity == nil || character.ID == "" {
			continue
		}
		user := byID[character.BelongsUserID]
		found := user != nil
		guest := found && user.IsGuest
		role := ""
		if found {
			role = user.Role
		}
		if !Eligible(found, guest, role) {
			continue
		}
		className := strings.TrimSpace(character.Class.Name)
		if className == "" {
			className = character.Class.ID
		}
		raceName := strings.TrimSpace(character.Race.Name)
		if raceName == "" {
			raceName = character.Race.ID
		}
		entries = append(entries, Entry{
			ID:        character.ID,
			Name:      character.Name,
			Class:     className,
			Race:      raceName,
			Level:     int(character.Level),
			ItemScore: ScorePieces(PiecesFromEquipped(character.EquippedItems), w),
			Created:   character.Created,
		})
	}
	return Place(entries)
}

// Find returns the placed row for a character id, then an exact name.
// Name matching is case-insensitive and keeps the better rank.
func Find(rows []Placed, self string) *Placed {
	self = strings.TrimSpace(self)
	if self == "" {
		return nil
	}
	for i := range rows {
		if rows[i].ID == self {
			copy := rows[i]
			return &copy
		}
	}
	var found *Placed
	for i := range rows {
		if strings.EqualFold(rows[i].Name, self) {
			copy := rows[i]
			found = &copy
			break
		}
	}
	return found
}

// Format renders the top rows plus the caller's rank for a text client.
func Format(rows []Placed, you *Placed, top int) string {
	if top < 1 {
		top = 10
	}
	if top > len(rows) {
		top = len(rows)
	}
	var b strings.Builder
	b.WriteString("Leaderboard\n")
	if top == 0 {
		b.WriteString("No ranked characters yet.\n")
	}
	for i := 0; i < top; i++ {
		row := rows[i]
		fmt.Fprintf(&b, "%2d  %s  %s %s  L%d  gear %d\n", row.Rank, row.Name, row.Class, row.Race, row.Level, row.ItemScore)
	}
	if you == nil {
		b.WriteString("Your character is not ranked.\n")
	} else {
		fmt.Fprintf(&b, "Your rank: %d  L%d  gear %d\n", you.Rank, you.Level, you.ItemScore)
	}
	return b.String()
}
