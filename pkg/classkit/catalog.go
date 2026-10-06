// Package classkit is the in-memory class catalog.
// Mechanics stay in the engine. Names, blurbs, race lists, and numbers come
// from a world pack. With no pack, the catalog is the generic sample trio.
package classkit

import "strings"

// KnownEffects are the kit primitives the combat engine implements.
var KnownEffects = map[string]bool{
	"brace": true, "slam": true, "stand": true,
	"slip": true, "nick": true, "smoke": true,
	"inscribe": true, "sear": true, "glyph": true,
	"guard": true, "bolt": true, "rig": true, "overload": true,
}

// Row is one class's damage row. 0 on a multiplier means "use 1".
type Row struct {
	DamageDealt float64 `yaml:"damage_dealt" json:"-"`
	DamageTaken float64 `yaml:"damage_taken" json:"-"`
	BehindDealt float64 `yaml:"behind_dealt" json:"-"`
	Swings      int     `yaml:"swings" json:"-"`
}

// Grit is the soak primitive. Nil means the class does not use it.
type Grit struct {
	Cap               int     `yaml:"cap" json:"-"`
	Opening           int     `yaml:"opening" json:"-"`
	StarterSwing      int32   `yaml:"starter_swing" json:"-"`
	SlamPerStack      float64 `yaml:"slam_per_stack" json:"-"`
	SlamCap           float64 `yaml:"slam_cap" json:"-"`
	RetaliatePerStack float64 `yaml:"retaliate_per_stack" json:"-"`
	SelfGuardGain     int     `yaml:"self_guard_gain" json:"-"`
}

// SkillSpec is one hotbar skill. Effect is a KnownEffects id.
type SkillSpec struct {
	ID           string  `yaml:"id" json:"id"`
	Name         string  `yaml:"name" json:"name"`
	Effect       string  `yaml:"effect" json:"effect"`
	Level        int32   `yaml:"level" json:"levelRequired"`
	OncePerFight bool    `yaml:"once_per_fight" json:"oncePerFight"`
	KeepsSwing   bool    `yaml:"keeps_swing" json:"keepsSwing"`
	Cooldown     int     `yaml:"cooldown" json:"cooldownRounds"`
	Multiplier   float64 `yaml:"multiplier" json:"swingMult"`
	Tooltip      string  `yaml:"tooltip" json:"description"`
	Target       string  `yaml:"target" json:"target"`
	ClassID      string  `yaml:"-" json:"-"`
}

// StartItem is a template starter piece named the way the importer names items.
type StartItem struct {
	Slot string `yaml:"slot" json:"-"`
	Name string `yaml:"name" json:"-"`
}

// Template is the create-screen card for a class.
type Template struct {
	ID            string      `yaml:"id" json:"-"`
	Backstory     string      `yaml:"backstory" json:"-"`
	Origin        string      `yaml:"origin_area" json:"-"`
	Archetype     string      `yaml:"archetype" json:"-"`
	Race          string      `yaml:"race" json:"-"`
	Str           int32       `yaml:"str" json:"-"`
	Dex           int32       `yaml:"dex" json:"-"`
	Int           int32       `yaml:"int" json:"-"`
	Wis           int32       `yaml:"wis" json:"-"`
	Sta           int32       `yaml:"sta" json:"-"`
	Mana          int32       `yaml:"mana" json:"-"`
	Items         []StartItem `yaml:"starting_items" json:"-"`
	DefaultSkills []string    `yaml:"default_skills" json:"-"`
}

// Def is one class. Also lists ids that share this row (ranger, hunter)
// without becoming this class. Aliases are this class (hitch, mage, template ids).
type Def struct {
	ID            string            `yaml:"id" json:"id"`
	Name          string            `yaml:"name" json:"name"`
	Description   string            `yaml:"description" json:"description"`
	ArmorType     string            `yaml:"armor_type" json:"armorType"`
	CombatType    string            `yaml:"combat_type" json:"combatType"`
	Primary       string            `yaml:"primary" json:"-"`
	SkillClass    string            `yaml:"skill_class" json:"skillClass"`
	Aliases       []string          `yaml:"aliases" json:"aliases"`
	Also          []string          `yaml:"also" json:"-"`
	Races         []string          `yaml:"races" json:"races"`
	PortraitClass string            `yaml:"portrait_class" json:"portraitClass"`
	HPMultiplier  float64           `yaml:"hp_multiplier" json:"-"`
	Caster        bool              `yaml:"caster" json:"caster"`
	Hotbar        int               `yaml:"hotbar_cap" json:"hotbarCap"`
	BraceCharges  int               `yaml:"brace_charges" json:"-"`
	SlipCharges   int               `yaml:"slip_charges" json:"-"`
	BoltCharges   int               `yaml:"bolt_charges" json:"-"`
	RigCharges    int               `yaml:"rig_charges" json:"-"`
	InscribeBasic bool              `yaml:"inscribe_basic" json:"-"`
	Order         int               `yaml:"order" json:"-"`
	Balance       Row               `yaml:"balance" json:"-"`
	Grit          *Grit             `yaml:"grit" json:"-"`
	RaceBlurbs    map[string]string `yaml:"race_blurbs" json:"-"`
	Template      *Template         `yaml:"template" json:"-"`
	Skills        []SkillSpec       `yaml:"skills" json:"skills"`
}

func norm(id string) string {
	return strings.ToLower(strings.TrimSpace(id))
}

func (d *Def) skillClass() string {
	if d == nil {
		return ""
	}
	if s := norm(d.SkillClass); s != "" {
		return s
	}
	return norm(d.ID)
}

func (d *Def) copy() *Def {
	if d == nil {
		return nil
	}
	cp := *d
	if d.Grit != nil {
		g := *d.Grit
		cp.Grit = &g
	}
	if d.Template != nil {
		t := *d.Template
		t.Items = append([]StartItem(nil), d.Template.Items...)
		t.DefaultSkills = append([]string(nil), d.Template.DefaultSkills...)
		cp.Template = &t
	}
	cp.Aliases = append([]string(nil), d.Aliases...)
	cp.Also = append([]string(nil), d.Also...)
	cp.Races = append([]string(nil), d.Races...)
	cp.Skills = append([]SkillSpec(nil), d.Skills...)
	if d.RaceBlurbs != nil {
		cp.RaceBlurbs = map[string]string{}
		for k, v := range d.RaceBlurbs {
			cp.RaceBlurbs[k] = v
		}
	}
	return &cp
}
