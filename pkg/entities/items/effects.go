package items

import "strings"

// Effect triggers. Content may use others; clients show unknown ones as written.
const (
	EffectTriggerOnHit   = "onHit"
	EffectTriggerOnEquip = "onEquip"
	EffectTriggerOnUse   = "onUse"
)

// ItemEffect is a player-facing line for something an item does beyond its stats.
// Content writes the text; the engine only carries it to clients.
type ItemEffect struct {
	Trigger string `bson:"trigger,omitempty" json:"trigger,omitempty" yaml:"trigger"`
	Name    string `bson:"name,omitempty" json:"name,omitempty" yaml:"name"`
	Text    string `bson:"text,omitempty" json:"text,omitempty" yaml:"text"`
}

// SpecialEffectText is the fallback when an item has a script but no effect text.
const SpecialEffectText = "Special effect"

// DisplayEffects returns the effect lines a player should see.
// Authored effects win. An attached script without authored text still gets a
// neutral "Special effect" line, named after the script when scriptName knows it.
func (item *Item) DisplayEffects(scriptName func(id string) string) []ItemEffect {
	if item == nil {
		return nil
	}
	out := make([]ItemEffect, 0, len(item.Effects)+2)
	covered := map[string]bool{}
	for _, eff := range item.Effects {
		if strings.TrimSpace(eff.Text) == "" && strings.TrimSpace(eff.Name) == "" {
			continue
		}
		out = append(out, eff)
		covered[eff.Trigger] = true
	}
	fallback := func(trigger, scriptID string) {
		if scriptID == "" || covered[trigger] {
			return
		}
		name := ""
		if scriptName != nil {
			name = strings.TrimSpace(scriptName(scriptID))
		}
		out = append(out, ItemEffect{Trigger: trigger, Name: name, Text: SpecialEffectText})
	}
	fallback(EffectTriggerOnHit, item.OnHitScriptID)
	fallback(EffectTriggerOnUse, item.OnUseScriptID)
	if len(out) == 0 {
		return nil
	}
	return out
}

// EffectTriggerLabel is the short player label for a trigger.
func EffectTriggerLabel(trigger string) string {
	switch trigger {
	case EffectTriggerOnHit:
		return "On hit"
	case EffectTriggerOnEquip:
		return "While equipped"
	case EffectTriggerOnUse:
		return "On use"
	case "":
		return "Effect"
	default:
		return trigger
	}
}

// FormatEffect renders one effect as a single line, e.g.
// "On hit — Vigil Burn: 2 damage each turn for 3 turns."
func FormatEffect(eff ItemEffect) string {
	label := EffectTriggerLabel(eff.Trigger)
	name := strings.TrimSpace(eff.Name)
	text := strings.TrimSpace(eff.Text)
	switch {
	case name != "" && text != "":
		return label + " — " + name + ": " + text
	case name != "":
		return label + " — " + name
	default:
		return label + " — " + text
	}
}

// CopyEffects returns an independent copy of an effect list.
func CopyEffects(src []ItemEffect) []ItemEffect {
	if len(src) == 0 {
		return nil
	}
	return append([]ItemEffect(nil), src...)
}
