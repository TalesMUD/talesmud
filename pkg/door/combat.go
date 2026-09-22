package door

import (
	"fmt"

	"github.com/talesmud/talesmud/pkg/entities/characters"
)

// fight kinds
const (
	fightForest = "forest"
	fightMaster = "master"
)

// fight is a lightweight 1v1. It uses the character's HP, attributes, and
// gear, with the same d20-plus-modifier against AC shape as room combat,
// without opening a room tactical instance.
type fight struct {
	Name    string
	Kind    string // fightForest or fightMaster
	MaxHP   int
	HP      int
	Attack  int
	AC      int
	XP      int32
	Gold    int64
	Log     []string
	Done    bool
	Outcome string
}

type roundLog struct {
	Lines []string
}

func (h *Hub) spawn(level int32) fight {
	m := h.pickMonster()
	if level < 1 {
		level = 1
	}
	bonus := int(level - 1)
	hp := m.HP + bonus*2
	if hp < 1 {
		hp = 1
	}
	atk := m.Attack
	if atk <= 0 {
		atk = m.HP / 4
		if atk < 2 {
			atk = 2
		}
	}
	atk += bonus / 3
	ac := m.AC
	if ac <= 0 {
		ac = 11
	}
	return fight{
		Name:   m.Name,
		Kind:   fightForest,
		MaxHP:  hp,
		HP:     hp,
		Attack: atk,
		AC:     ac,
		XP:     int32(m.XP + bonus),
		Gold:   int64(m.Gold + bonus),
	}
}

// spawnMaster builds the Ashmarket Master duel scaled to the warrior's level.
// Original name — no licensed LORD trainers.
func (h *Hub) spawnMaster(level int32) fight {
	if level < 1 {
		level = 1
	}
	hp := 18 + int(level)*5
	if hp < 16 {
		hp = 16
	}
	atk := 5 + int(level)
	ac := 12 + int(level)/2
	if ac > 18 {
		ac = 18
	}
	return fight{
		Name:   "Ashmarket Master",
		Kind:   fightMaster,
		MaxHP:  hp,
		HP:     hp,
		Attack: atk,
		AC:     ac,
		XP:     0,
		Gold:   0,
	}
}

func (h *Hub) pickMonster() Monster {
	total := 0
	for _, m := range h.pack.Monsters {
		w := m.Weight
		if w <= 0 {
			w = 1
		}
		total += w
	}
	if total <= 0 || len(h.pack.Monsters) == 0 {
		return Monster{Name: "Bramble Wolf", HP: 14, Attack: 4, AC: 12, XP: 18, Gold: 7, Weight: 1}
	}
	n := h.roll.Intn(total)
	for _, m := range h.pack.Monsters {
		w := m.Weight
		if w <= 0 {
			w = 1
		}
		if n < w {
			return m
		}
		n -= w
	}
	return h.pack.Monsters[0]
}

func (h *Hub) playerAttack(ch *characters.Character, f *fight) roundLog {
	var out roundLog
	if f == nil || f.Done {
		return out
	}
	weapon := h.pack.WeaponByID(weaponID(ch))
	attr := "STR"
	if ch.Door != nil {
		if tr, ok := trackByID(h.pack, ch.Door.Track); ok {
			attr = tr.AttackAttr()
		}
	}
	mod := int(ch.GetAttribute(attr)-10) / 2
	roll := h.roll.Intn(20) + 1
	toHit := roll + mod
	armor := h.pack.ArmorByID(armorID(ch))
	switch {
	case roll == 1:
		out.Lines = append(out.Lines, fmt.Sprintf("You swing wide. The %s does not even flinch.", f.Name))
	case roll == 20 || toHit >= f.AC:
		dmg := weapon.Attack + mod + h.roll.Intn(4)
		if dmg < 1 {
			dmg = 1
		}
		if roll == 20 {
			dmg *= 2
		}
		f.HP -= dmg
		if f.HP < 0 {
			f.HP = 0
		}
		if roll == 20 {
			out.Lines = append(out.Lines, fmt.Sprintf("A clean opening. You hit the %s for %d.", f.Name, dmg))
		} else {
			out.Lines = append(out.Lines, fmt.Sprintf("You strike the %s for %d.", f.Name, dmg))
		}
	default:
		out.Lines = append(out.Lines, fmt.Sprintf("The %s turns your %s aside.", f.Name, weapon.Name))
	}
	if f.HP <= 0 {
		return out
	}
	out.Lines = append(out.Lines, h.enemyAttack(ch, f, armor)...)
	return out
}

func (h *Hub) enemyAttack(ch *characters.Character, f *fight, armor Armor) []string {
	conMod := int(ch.GetAttribute("CON")-10) / 2
	ac := 10 + armor.Defense + conMod
	if ac < 8 {
		ac = 8
	}
	roll := h.roll.Intn(20) + 1
	if roll != 20 && roll+f.Attack/2 < ac {
		return []string{fmt.Sprintf("The %s misses you.", f.Name)}
	}
	dmg := f.Attack/2 + h.roll.Intn(4) - armor.Defense/2
	if roll == 20 {
		dmg *= 2
	}
	if dmg < 1 {
		dmg = 1
	}
	ch.CurrentHitPoints -= int32(dmg)
	if ch.CurrentHitPoints < 0 {
		ch.CurrentHitPoints = 0
	}
	return []string{fmt.Sprintf("The %s hits you for %d.", f.Name, dmg)}
}

func (h *Hub) tryFlee(ch *characters.Character, f *fight) (bool, []string) {
	if h.roll.Intn(100) < 65 {
		return true, []string{"You break for the gate. The trees let you go."}
	}
	armor := h.pack.ArmorByID(armorID(ch))
	lines := []string{"You turn to run. It is not fast enough."}
	lines = append(lines, h.enemyAttack(ch, f, armor)...)
	return false, lines
}

func weaponID(ch *characters.Character) string {
	if ch != nil && ch.Door != nil {
		return ch.Door.WeaponID
	}
	return ""
}

func armorID(ch *characters.Character) string {
	if ch != nil && ch.Door != nil {
		return ch.Door.ArmorID
	}
	return ""
}

func trackByID(pack *Pack, id string) (Track, bool) {
	for _, t := range pack.Tracks {
		if t.ID == id {
			return t, true
		}
	}
	return Track{}, false
}
