package door

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Pack is the fat content for one door title. The engine stays thin.
type Pack struct {
	ID          string             `yaml:"id"`
	Title       string             `yaml:"title"`
	Setting     string             `yaml:"setting"`
	Town        string             `yaml:"town"`
	Forest      string             `yaml:"forest"`
	Healer      string             `yaml:"healer"`
	Bank        string             `yaml:"bank"`
	Banker      string             `yaml:"banker"`
	Inn         string             `yaml:"inn"`
	Innkeeper   string             `yaml:"innkeeper"`
	Armory      string             `yaml:"armory"`
	Smith       string             `yaml:"smith"`
	HealCost     int                `yaml:"heal_cost"`
	DailyFights  int                `yaml:"daily_fights"`
	HealOnNewDay bool               `yaml:"-"` // from daily.heal_on_new_day; default true
	Intro       []string           `yaml:"intro"`
	News        []string           `yaml:"news"`
	Tracks      []Track            `yaml:"tracks"`
	Monsters    []Monster          `yaml:"monsters"`
	Weapons     []Weapon           `yaml:"weapons"`
	Armor       []Armor            `yaml:"armor"`
	Screens     map[string]*Screen `yaml:"-"`
	Timezone    string             `yaml:"-"`
}

// Screen is one ANSI page from screens/<id>/screen.yaml plus art.ans.
type Screen struct {
	ID      string            `yaml:"id"`
	Title   string            `yaml:"title"`
	Prompt  string            `yaml:"prompt"`
	Art     string            `yaml:"-"`
	Hotkeys map[string]Hotkey `yaml:"hotkeys"`
}

// Hotkey is a single-key action declared by the pack.
type Hotkey struct {
	Action string `yaml:"action"`
	Screen string `yaml:"screen"`
}

// Track is a starting skill path. Names are original to this pack.
type Track struct {
	ID    string `yaml:"id"`
	Name  string `yaml:"name"`
	Blurb string `yaml:"blurb"`
	STR   int    `yaml:"str"`
	DEX   int    `yaml:"dex"`
	CON   int    `yaml:"con"`
	INT   int    `yaml:"int"`
	WIS   int    `yaml:"wis"`
	Class string `yaml:"class"`
}

// Monster is a forest encounter template.
type Monster struct {
	ID     string `yaml:"id"`
	Name   string `yaml:"name"`
	HP     int    `yaml:"hp"`
	Attack int    `yaml:"attack"`
	AC     int    `yaml:"ac"`
	XP     int    `yaml:"xp"`
	Gold   int    `yaml:"gold"`
	Weight int    `yaml:"weight"`
}

// Weapon is a single carried weapon.
type Weapon struct {
	ID     string `yaml:"id"`
	Name   string `yaml:"name"`
	Attack int    `yaml:"attack"`
	Price  int    `yaml:"price"`
}

// Armor is a single carried armor piece.
type Armor struct {
	ID      string `yaml:"id"`
	Name    string `yaml:"name"`
	Defense int    `yaml:"defense"`
	Price   int    `yaml:"price"`
}

// Load reads a world pack. A directory uses world.yaml plus monsters, weapons,
// armor, and screens/<id>/. A single YAML file is the flat fallback.
func Load(path string) (*Pack, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("read door pack %s: %w", path, err)
	}
	if info.IsDir() {
		return loadDir(path)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read door pack %s: %w", path, err)
	}
	pack, err := parseWorld(raw)
	if err != nil {
		return nil, fmt.Errorf("parse door pack %s: %w", path, err)
	}
	pack.normalize()
	if len(pack.Monsters) == 0 {
		return nil, fmt.Errorf("door pack %s is missing monsters", path)
	}
	return pack, nil
}

func loadDir(dir string) (*Pack, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "world.yaml"))
	if err != nil {
		return nil, err
	}
	pack, err := parseWorld(raw)
	if err != nil {
		return nil, err
	}
	if monsters, err := os.ReadFile(filepath.Join(dir, "monsters.yaml")); err == nil {
		var doc struct {
			Forest []Monster `yaml:"forest"`
		}
		if err := yaml.Unmarshal(monsters, &doc); err != nil {
			return nil, err
		}
		if len(doc.Forest) > 0 {
			pack.Monsters = doc.Forest
		}
	}
	if weapons, err := os.ReadFile(filepath.Join(dir, "weapons.yaml")); err == nil {
		var doc struct {
			Weapons []weaponFile `yaml:"weapons"`
		}
		if err := yaml.Unmarshal(weapons, &doc); err != nil {
			return nil, err
		}
		if len(doc.Weapons) > 0 {
			pack.Weapons = nil
			for _, w := range doc.Weapons {
				pack.Weapons = append(pack.Weapons, w.weapon())
			}
		}
	}
	if armor, err := os.ReadFile(filepath.Join(dir, "armor.yaml")); err == nil {
		var doc struct {
			Armor []armorFile `yaml:"armor"`
		}
		if err := yaml.Unmarshal(armor, &doc); err != nil {
			return nil, err
		}
		if len(doc.Armor) > 0 {
			pack.Armor = nil
			for _, a := range doc.Armor {
				pack.Armor = append(pack.Armor, a.armor())
			}
		}
	}
	screens, err := loadScreens(filepath.Join(dir, "screens"))
	if err != nil {
		return nil, err
	}
	pack.Screens = screens
	pack.normalize()
	if len(pack.Monsters) == 0 {
		return nil, fmt.Errorf("door pack %s is missing monsters", dir)
	}
	return pack, nil
}

type weaponFile struct {
	ID     string `yaml:"id"`
	Name   string `yaml:"name"`
	Damage int    `yaml:"damage"`
	Attack int    `yaml:"attack"`
	Cost   int    `yaml:"cost"`
	Price  int    `yaml:"price"`
}

func (w weaponFile) weapon() Weapon {
	atk := w.Attack
	if atk == 0 {
		atk = w.Damage
	}
	price := w.Price
	if price == 0 {
		price = w.Cost
	}
	return Weapon{ID: w.ID, Name: w.Name, Attack: atk, Price: price}
}

type armorFile struct {
	ID      string `yaml:"id"`
	Name    string `yaml:"name"`
	Defense int    `yaml:"defense"`
	Cost    int    `yaml:"cost"`
	Price   int    `yaml:"price"`
}

func (a armorFile) armor() Armor {
	price := a.Price
	if price == 0 {
		price = a.Cost
	}
	return Armor{ID: a.ID, Name: a.Name, Defense: a.Defense, Price: price}
}

func parseWorld(raw []byte) (*Pack, error) {
	var doc struct {
		Pack  `yaml:",inline"`
		Daily struct {
			Timezone     string `yaml:"timezone"`
			HealOnNewDay *bool  `yaml:"heal_on_new_day"`
			Resources    map[string]struct {
				PerDay int `yaml:"per_day"`
			} `yaml:"resources"`
		} `yaml:"daily"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	pack := doc.Pack
	pack.Timezone = doc.Daily.Timezone
	if doc.Daily.HealOnNewDay != nil {
		pack.HealOnNewDay = *doc.Daily.HealOnNewDay
	} else {
		pack.HealOnNewDay = true
	}
	if res, ok := doc.Daily.Resources["forest_fights"]; ok && res.PerDay > 0 {
		pack.DailyFights = res.PerDay
	}
	return &pack, nil
}

func loadScreens(dir string) (map[string]*Screen, error) {
	out := map[string]*Screen{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		base := filepath.Join(dir, entry.Name())
		raw, err := os.ReadFile(filepath.Join(base, "screen.yaml"))
		if err != nil {
			continue
		}
		var screen Screen
		if err := yaml.Unmarshal(raw, &screen); err != nil {
			return nil, err
		}
		if screen.ID == "" {
			screen.ID = entry.Name()
		}
		artName := "art.ans"
		if art, err := os.ReadFile(filepath.Join(base, artName)); err == nil {
			screen.Art = strings.ReplaceAll(string(art), "\r\n", "\n")
		}
		out[screen.ID] = &screen
	}
	return out, nil
}

// DefaultPack is a small original fallback used when the pack file cannot be read.
func DefaultPack() *Pack {
	pack := &Pack{
		ID:          "aethermoor-door",
		Title:       "Aethermoor Door",
		Setting:     "Veilspan",
		Town:        "Ashmarket Square",
		Forest:      "The Bramblewood",
		Healer:      "Sister Callewyn",
		Bank:        "Ashmarket Vault",
		Banker:      "Ilen Vos",
		Inn:         "The Emberloft",
		Innkeeper:   "Maela Quinn",
		Armory:      "Soot and Steel",
		Smith:       "Harl Quill",
		HealCost:     20,
		DailyFights:  15,
		HealOnNewDay: true,
		Intro: []string{
			"Soot-warm cobbles and the Bramblewood gate standing open.",
			"Each day allows a fixed number of walks into the trees.",
		},
		News: []string{
			"The Ashwyrm has not been seen above the cinder ridge.",
		},
		Tracks: []Track{{
			ID: "gravebound", Name: "Gravebound", Blurb: "Oath-bound strikers.",
			STR: 16, DEX: 11, CON: 14, INT: 8, WIS: 10, Class: "warrior",
		}},
		Monsters: []Monster{{
			ID: "bramble_wolf", Name: "Bramble Wolf", HP: 14, Attack: 4, AC: 12, XP: 18, Gold: 7, Weight: 1,
		}},
		Weapons: []Weapon{
			{ID: "walking_stick", Name: "Walking Stick", Attack: 1, Price: 0},
			{ID: "hedge_knife", Name: "Hedge Knife", Attack: 3, Price: 40},
		},
		Armor: []Armor{
			{ID: "travelers_coat", Name: "Traveler's Coat", Defense: 0, Price: 0},
			{ID: "quilted_jack", Name: "Quilted Jack", Defense: 2, Price: 50},
		},
	}
	pack.normalize()
	return pack
}

func (p *Pack) normalize() {
	if p.Title == "" {
		p.Title = "Aethermoor Door"
	}
	if p.Town == "" {
		p.Town = "Ashmarket Square"
	}
	if p.Forest == "" {
		p.Forest = "The Bramblewood"
	}
	if p.HealCost <= 0 {
		p.HealCost = 20
	}
	if p.DailyFights <= 0 {
		p.DailyFights = 15
	}
	if len(p.Tracks) == 0 {
		p.Tracks = []Track{
			{ID: "gravebound", Name: "Gravebound", Blurb: "Oath-bound strikers who borrow weight from old barrows.", STR: 16, DEX: 11, CON: 14, INT: 8, WIS: 10, Class: "warrior"},
			{ID: "veilcaller", Name: "Veilcaller", Blurb: "Readers of the thin places, sharp when the mist is up.", STR: 9, DEX: 12, CON: 11, INT: 16, WIS: 14, Class: "wizard"},
			{ID: "cutpurse", Name: "Cutpurse", Blurb: "Market hands. Quick steel, quicker exits.", STR: 11, DEX: 16, CON: 12, INT: 10, WIS: 9, Class: "rogue"},
		}
	}
	if p.Setting == "" {
		p.Setting = "Veilspan"
	}
}

// WeaponByID returns a carried weapon. An empty id is bare hands.
func (p *Pack) WeaponByID(id string) Weapon {
	if id == "" {
		return Weapon{Name: "Bare hands", Attack: 0}
	}
	for _, w := range p.Weapons {
		if w.ID == id {
			return w
		}
	}
	return Weapon{ID: id, Name: "Bare hands", Attack: 0}
}

// ArmorByID returns carried armor. An empty id is no armor.
func (p *Pack) ArmorByID(id string) Armor {
	if id == "" {
		return Armor{Name: "None", Defense: 0}
	}
	for _, a := range p.Armor {
		if a.ID == id {
			return a
		}
	}
	return Armor{ID: id, Name: "None"}
}

// TrackByIndex is 1-based, matching the create-warrior hotkeys.
func (p *Pack) TrackByIndex(n int) (Track, bool) {
	if n < 1 || n > len(p.Tracks) {
		return Track{}, false
	}
	return p.Tracks[n-1], true
}

// AttackAttr is the attribute short name used for this track's strikes.
func (t Track) AttackAttr() string {
	switch t.Class {
	case "wizard":
		return "INT"
	case "rogue":
		return "DEX"
	default:
		return "STR"
	}
}
