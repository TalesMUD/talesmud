package commands

import (
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
)

// scriptDisplayName looks up a script's name for the "Special effect" fallback.
func scriptDisplayName(game def.GameCtrl) func(string) string {
	return func(id string) string {
		if game == nil || game.GetFacade() == nil || id == "" {
			return ""
		}
		script, err := game.GetFacade().ScriptsService().FindByID(id)
		if err != nil || script == nil {
			return ""
		}
		return script.Name
	}
}

// backfillItemEffects copies template effect text onto carried instances that
// were created before the template had it. It reports whether anything changed.
func backfillItemEffects(ch *characters.Character, template func(id string) *items.Item) bool {
	if ch == nil || template == nil {
		return false
	}
	changed := false
	fill := func(item *items.Item) {
		if item == nil || len(item.Effects) > 0 || item.TemplateID == "" {
			return
		}
		tmpl := template(item.TemplateID)
		if tmpl == nil || len(tmpl.Effects) == 0 {
			return
		}
		item.Effects = items.CopyEffects(tmpl.Effects)
		changed = true
	}
	for _, item := range ch.Inventory.Items {
		fill(item)
	}
	for _, item := range ch.EquippedItems {
		fill(item)
	}
	return changed
}
