package game

import (
	"github.com/talesmud/talesmud/pkg/classkit"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
)

type kitPiece struct {
	Name string
	Slot items.ItemSlot
}

// OpRegrantStarterKit adds missing class-kit pieces. It does not replace gear
// the character already holds. When the class has no starter list, the op is
// skipped and is not undoable.
func (g *Game) OpRegrantStarterKit(characterID string) (*OpResult, error) {
	if characterID == "" {
		return nil, opErr(400, "characterId is required")
	}
	char, err := g.Facade.CharactersService().FindByID(characterID)
	if err != nil || char == nil {
		return nil, opErr(404, "character not found")
	}
	pieces := starterPieces(g, char)
	if len(pieces) == 0 {
		return &OpResult{
			Summary:    "No starter kit is defined for this class.",
			Undoable:   false,
			EntityType: "character",
			EntityID:   characterID,
			Detail: map[string]interface{}{
				"characterId": characterID,
				"skipped":     true,
				"note":        "No class kit or character-template starter list matched this character.",
			},
		}, nil
	}
	var created []*items.Item
	var skipped []string
	for _, piece := range pieces {
		if holdsName(char, piece.Name) {
			skipped = append(skipped, piece.Name)
			continue
		}
		inst, ierr := g.instantiateNamed(piece.Name)
		if ierr != nil || inst == nil {
			skipped = append(skipped, piece.Name)
			continue
		}
		if inst.Unique && char.CountOfTemplate(items.TemplateKey(inst)) >= 1 {
			skipped = append(skipped, piece.Name)
			continue
		}
		inst.Quantity = 1
		created = append(created, inst)
		_ = piece
	}
	if len(created) == 0 {
		return &OpResult{
			Summary:    "Starter kit is already on the character.",
			Undoable:   false,
			EntityType: "character",
			EntityID:   characterID,
			Detail: map[string]interface{}{
				"characterId":  characterID,
				"skipped":      true,
				"note":         "Every starter piece is already held.",
				"skippedNames": skipped,
			},
		}, nil
	}
	// Pair created items with the pieces that were not already held, in order.
	pending := make([]kitPiece, 0, len(created))
	made := 0
	for _, piece := range pieces {
		if holdsName(char, piece.Name) {
			continue
		}
		if made >= len(created) {
			break
		}
		pending = append(pending, piece)
		made++
	}
	var added []string
	var deltas []stackDelta
	err = g.Facade.CharactersService().Modify(characterID, func(ch *characters.Character) error {
		before := itemQtyByID(ch)
		for i, inst := range created {
			slot := items.ItemSlotInventory
			if i < len(pending) {
				slot = pending[i].Slot
			}
			if inst.Unique {
				key := items.TemplateKey(inst)
				if key != "" && ch.CountOfTemplate(key) >= 1 {
					continue
				}
			}
			if slot != "" && slot != items.ItemSlotInventory && (ch.EquippedItems == nil || ch.EquippedItems[slot] == nil) {
				if ch.EquippedItems == nil {
					ch.EquippedItems = map[items.ItemSlot]*items.Item{}
				}
				ch.EquippedItems[slot] = inst
				continue
			}
			if addErr := ch.Inventory.AddItem(inst); addErr != nil {
				return opErr(400, addErr.Error())
			}
		}
		added, deltas = inventoryDelta(ch, before)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(added) == 0 && len(deltas) == 0 {
		return &OpResult{
			Summary:    "Starter kit is already on the character.",
			Undoable:   false,
			EntityType: "character",
			EntityID:   characterID,
			Detail: map[string]interface{}{
				"skipped": true,
				"note":    "Every starter piece is already held.",
			},
		}, nil
	}
	fresh, _ := g.Facade.CharactersService().FindByID(characterID)
	g.notifyInventory(fresh, "Your starter kit was re-granted.")
	return &OpResult{
		Summary:    opSummary("Re-granted %d starter piece(s) to %s.", len(added)+len(deltas), characterName(fresh)),
		Undoable:   true,
		EntityType: "character",
		EntityID:   characterID,
		After:      jsonRaw(map[string]interface{}{"instanceIds": added}),
		Inverse: inverseOf("take-added", map[string]interface{}{
			"characterId": characterID,
			"instanceIds": added,
			"stackDeltas": deltas,
		}),
		Detail: map[string]interface{}{
			"characterId": characterID,
			"instanceIds": added,
			"skipped":     skipped,
		},
	}, nil
}

func starterPieces(g *Game, ch *characters.Character) []kitPiece {
	if ch == nil {
		return nil
	}
	if pieces := piecesFromKit(classkit.Lookup(ch.Class.ID)); len(pieces) > 0 {
		return pieces
	}
	if pieces := piecesFromKit(classkit.Lookup(ch.Class.Name)); len(pieces) > 0 {
		return pieces
	}
	if g == nil || g.Facade == nil || g.Facade.CharacterTemplatesRepo() == nil {
		return nil
	}
	templates, err := g.Facade.CharacterTemplatesRepo().FindAll()
	if err != nil || len(templates) == 0 {
		templates = characters.SystemCharacterTemplatePresets()
	}
	for _, tpl := range templates {
		if tpl == nil || !templateMatchesClass(tpl, ch) {
			continue
		}
		var pieces []kitPiece
		for _, item := range tpl.StartingItems {
			name := item.ItemTemplateName
			if name == "" {
				continue
			}
			pieces = append(pieces, kitPiece{Name: name, Slot: item.Slot})
		}
		if len(pieces) > 0 {
			return pieces
		}
	}
	return nil
}

func piecesFromKit(def *classkit.Def) []kitPiece {
	if def == nil || def.Template == nil {
		return nil
	}
	var pieces []kitPiece
	for _, item := range def.Template.Items {
		if item.Name == "" {
			continue
		}
		pieces = append(pieces, kitPiece{Name: item.Name, Slot: items.ItemSlot(item.Slot)})
	}
	return pieces
}

func templateMatchesClass(tpl *characters.CharacterTemplate, ch *characters.Character) bool {
	if tpl.Class.ID != "" && tpl.Class.ID == ch.Class.ID {
		return true
	}
	if tpl.Class.Name != "" && tpl.Class.Name == ch.Class.Name {
		return true
	}
	if tpl.Name != "" && (tpl.Name == ch.Class.Name || tpl.Name == ch.Class.ID) {
		return true
	}
	return false
}

func holdsName(ch *characters.Character, name string) bool {
	if ch == nil || name == "" {
		return false
	}
	found := false
	walkItems(ch.Inventory.Items, func(item *items.Item) {
		if item.Name == name {
			found = true
		}
	})
	for _, item := range ch.EquippedItems {
		if item != nil && item.Name == name {
			found = true
		}
	}
	return found
}

func (g *Game) instantiateNamed(name string) (*items.Item, error) {
	if g.Facade.ItemsService() != nil {
		found, err := g.Facade.ItemsService().FindByName(name)
		if err == nil {
			for _, item := range found {
				if item != nil && item.Name == name && item.IsTemplate {
					return g.Facade.ItemsService().CreateInstanceFromTemplate(item.ID)
				}
			}
		}
	}
	preset := items.StarterItemTemplateByName(name)
	if preset == nil {
		return nil, opErr(404, "no starter item named "+name)
	}
	return cloneStarterItem(preset), nil
}
