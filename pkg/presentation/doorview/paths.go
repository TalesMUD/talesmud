package doorview

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/gamemode"
)

type pathChoice struct {
	Name  string
	Blurb string
	Class characters.Class
	Race  characters.Race
	HP    int32
	Mana  int32
	Gold  int64
	Attrs characters.Attributes
}

type pathFile struct {
	Paths []struct {
		Name      string `yaml:"name"`
		Blurb     string `yaml:"blurb"`
		Class     string `yaml:"class"`
		Race      string `yaml:"race"`
		HitPoints int32  `yaml:"hit_points"`
		Mana      int32  `yaml:"mana"`
		Gold      int64  `yaml:"gold"`
	} `yaml:"paths"`
}

func loadPaths() []pathChoice {
	root := strings.TrimSpace(gamemode.Current().WorldPack)
	if root == "" {
		return presetPaths()
	}
	raw, err := os.ReadFile(filepath.Join(root, "character_paths.yaml"))
	if err != nil {
		return presetPaths()
	}
	var file pathFile
	if err := yaml.Unmarshal(raw, &file); err != nil || len(file.Paths) == 0 {
		return presetPaths()
	}
	out := make([]pathChoice, 0, len(file.Paths))
	for _, row := range file.Paths {
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}
		class, ok := classByID(row.Class)
		if !ok {
			class = characters.ClassWarrior
		}
		choice := pathChoice{
			Name:  name,
			Blurb: strings.TrimSpace(row.Blurb),
			Class: class,
			Race:  raceByID(row.Race),
			HP:    row.HitPoints,
			Mana:  row.Mana,
			Gold:  row.Gold,
		}
		if preset := presetForClass(class.ID); preset != nil {
			if choice.HP <= 0 {
				choice.HP = preset.MaxHitPoints
			}
			if choice.Mana <= 0 {
				choice.Mana = preset.MaxMana
			}
			choice.Attrs = append(characters.Attributes(nil), preset.Attributes...)
			if row.Race == "" {
				choice.Race = preset.Race
			}
		}
		if choice.HP <= 0 {
			choice.HP = 20
		}
		out = append(out, choice)
	}
	if len(out) == 0 {
		return presetPaths()
	}
	return out
}

func presetPaths() []pathChoice {
	presets := characters.SystemCharacterTemplatePresets()
	out := make([]pathChoice, 0, len(presets))
	for _, preset := range presets {
		if preset == nil || strings.TrimSpace(preset.Name) == "" {
			continue
		}
		hp := preset.MaxHitPoints
		if hp <= 0 {
			hp = preset.CurrentHitPoints
		}
		if hp <= 0 {
			hp = 20
		}
		out = append(out, pathChoice{
			Name:  preset.Name,
			Blurb: preset.Description,
			Class: preset.Class,
			Race:  preset.Race,
			HP:    hp,
			Mana:  preset.MaxMana,
			Attrs: append(characters.Attributes(nil), preset.Attributes...),
		})
	}
	return out
}

func presetForClass(id string) *characters.CharacterTemplate {
	id = strings.ToLower(strings.TrimSpace(id))
	for _, preset := range characters.SystemCharacterTemplatePresets() {
		if preset != nil && strings.EqualFold(preset.Class.ID, id) {
			return preset
		}
	}
	return nil
}

func classByID(id string) (characters.Class, bool) {
	switch strings.ToLower(strings.TrimSpace(id)) {
	case "warrior":
		return characters.ClassWarrior, true
	case "rogue":
		return characters.ClassRogue, true
	case "wizard", "mage":
		return characters.ClassWizard, true
	case "ranger":
		return characters.ClassRanger, true
	case "hunter":
		return characters.ClassHunter, true
	default:
		return characters.Class{}, false
	}
}

func raceByID(id string) characters.Race {
	switch strings.ToLower(strings.TrimSpace(id)) {
	case "dwarf":
		return characters.RaceDwarf
	case "elve", "elf":
		return characters.RaceElve
	default:
		return characters.RaceHuman
	}
}

func sexWord(key string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "m":
		return "man", true
	case "f":
		return "woman", true
	case "x":
		return "neither", true
	default:
		return "", false
	}
}
