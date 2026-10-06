package characters

import (
	"github.com/talesmud/talesmud/pkg/classkit"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
)

// Playable roster. HP uses balance.ScaleClassHP once off a shared base of 25.
const classHPBase int32 = 25

func presetEntity(id string) *entities.Entity {
	return &entities.Entity{ID: id}
}

// SystemCharacterTemplatePresets is the create roster from the class catalog.
func SystemCharacterTemplatePresets() []*CharacterTemplate {
	out := make([]*CharacterTemplate, 0)
	for _, d := range classkit.Playable() {
		if d == nil || d.Template == nil || d.Template.ID == "" {
			continue
		}
		t := d.Template
		hp := balance.ScaleClassHP(d.ID, classHPBase)
		race, ok := RaceByID(t.Race)
		if !ok {
			race = RaceHuman
		}
		starting := make([]StartingItem, 0, len(t.Items))
		for _, it := range t.Items {
			starting = append(starting, StartingItem{
				Slot:             items.ItemSlot(it.Slot),
				ItemTemplateName: it.Name,
			})
		}
		tpl := &CharacterTemplate{
			Entity:           presetEntity(t.ID),
			Name:             d.Name,
			Description:      d.Description,
			Backstory:        t.Backstory,
			OriginArea:       t.Origin,
			Archetype:        t.Archetype,
			Race:             race,
			Class:            classFromDef(d),
			Level:            1,
			CurrentHitPoints: hp,
			MaxHitPoints:     hp,
			Attributes:       createBaseAttributes(t.Str, t.Dex, t.Int, t.Wis, t.Sta),
			StartingItems:    starting,
			DefaultSkills:    append([]string(nil), t.DefaultSkills...),
			Source:           "system",
		}
		if t.Mana > 0 {
			tpl.CurrentMana = t.Mana
			tpl.MaxMana = t.Mana
		}
		out = append(out, tpl)
	}
	return out
}

// PresetByID finds a signed create template by its stable id.
func PresetByID(id string) *CharacterTemplate {
	if id == "" {
		return nil
	}
	for _, preset := range SystemCharacterTemplatePresets() {
		if preset != nil && preset.Entity != nil && preset.Entity.ID == id {
			return preset
		}
	}
	return nil
}
