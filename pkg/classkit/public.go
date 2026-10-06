package classkit

// RaceInfo is one race row on GET /api/classes.
type RaceInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Blurb string `json:"blurb"`
}

// ClientSkill is the hotbar row the play client renders.
type ClientSkill struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	ClassIDs       []string `json:"classIds"`
	LevelRequired  int32    `json:"levelRequired"`
	Kit            string   `json:"kit"`
	KeepsSwing     bool     `json:"keepsSwing"`
	OncePerFight   bool     `json:"oncePerFight"`
	CooldownRounds int      `json:"cooldownRounds"`
	SwingMult      float64  `json:"swingMult"`
	Target         string   `json:"target"`
	ResourceType   string   `json:"resourceType"`
	ManaCost       int      `json:"manaCost"`
	Effect         string   `json:"effect"`
}

// ClientClass is one create-screen class.
type ClientClass struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	SkillClass    string        `json:"skillClass"`
	Aliases       []string      `json:"aliases"`
	Races         []string      `json:"races"`
	TemplateID    string        `json:"templateId"`
	PortraitClass string        `json:"portraitClass"`
	ArmorType     string        `json:"armorType"`
	CombatType    string        `json:"combatType"`
	HotbarCap     int           `json:"hotbarCap"`
	Caster        bool          `json:"caster"`
	Skills        []ClientSkill `json:"skills"`
}

// ClientPayload is the body of GET /api/classes.
type ClientPayload struct {
	Source  string        `json:"source"`
	Classes []ClientClass `json:"classes"`
	Races   []RaceInfo    `json:"races"`
}

// ClientView builds the public payload. base races are overwritten by pack blurbs.
func ClientView(base []RaceInfo) ClientPayload {
	src := "sample"
	if FromPack() {
		src = "pack"
	}
	overrides := RaceBlurbOverrides()
	races := make([]RaceInfo, 0, len(base))
	for _, r := range base {
		if b, ok := overrides[norm(r.ID)]; ok {
			r.Blurb = b
		}
		races = append(races, r)
	}
	classes := Playable()
	out := make([]ClientClass, 0, len(classes))
	for _, d := range classes {
		keys := map[string]bool{}
		add := func(s string) {
			s = norm(s)
			if s != "" {
				keys[s] = true
			}
		}
		add(d.ID)
		add(d.skillClass())
		add(d.Name)
		if d.Template != nil {
			add(d.Template.ID)
			add(d.Template.Archetype)
		}
		for _, a := range d.Aliases {
			add(a)
		}
		aliases := make([]string, 0, len(keys))
		for k := range keys {
			aliases = append(aliases, k)
		}
		sortStrings(aliases)
		skills := make([]ClientSkill, 0, len(d.Skills))
		for _, s := range d.Skills {
			target := s.Target
			if target == "" {
				target = "enemy"
			}
			skills = append(skills, ClientSkill{
				ID: s.ID, Name: s.Name, Description: s.Tooltip,
				ClassIDs: []string{d.skillClass()}, LevelRequired: s.Level,
				Kit: s.Effect, KeepsSwing: s.KeepsSwing, OncePerFight: s.OncePerFight,
				CooldownRounds: s.Cooldown, SwingMult: s.Multiplier, Target: target,
				ResourceType: "cooldown", Effect: "damage",
			})
		}
		templateID := ""
		if d.Template != nil {
			templateID = d.Template.ID
		}
		out = append(out, ClientClass{
			ID: d.ID, Name: d.Name, Description: d.Description,
			SkillClass: d.skillClass(), Aliases: aliases, Races: append([]string(nil), d.Races...),
			TemplateID: templateID, PortraitClass: d.PortraitClass,
			ArmorType: d.ArmorType, CombatType: d.CombatType,
			HotbarCap: d.Hotbar, Caster: d.Caster, Skills: skills,
		})
	}
	return ClientPayload{Source: src, Classes: out, Races: races}
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		j := i
		for j > 0 && s[j] < s[j-1] {
			s[j], s[j-1] = s[j-1], s[j]
			j--
		}
	}
}
