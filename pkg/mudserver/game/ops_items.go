package game

import (
	"encoding/json"
	"fmt"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

type takenPiece struct {
	Item         *items.Item    `json:"item"`
	EquippedSlot items.ItemSlot `json:"equippedSlot,omitempty"`
	Partial      bool           `json:"partial,omitempty"`
}

type stackDelta struct {
	ItemID   string `json:"itemId"`
	Quantity int32  `json:"quantity"`
}

// OpGiveItem creates instances from a template and puts them on the character.
// A unique template refuses when one is already held, and refuses quantity above 1.
func (g *Game) OpGiveItem(characterID, templateID string, quantity int32) (*OpResult, error) {
	qty, err := qtyOrOne(quantity)
	if err != nil {
		return nil, err
	}
	if characterID == "" || templateID == "" {
		return nil, opErr(400, "characterId and itemTemplateId are required")
	}
	tpl, err := g.Facade.ItemsService().FindByID(templateID)
	if err != nil || tpl == nil {
		return nil, opErr(404, "item template not found")
	}
	if tpl.Unique && qty > 1 {
		return nil, opErr(400, "a unique item can only be given once")
	}
	char, err := g.Facade.CharactersService().FindByID(characterID)
	if err != nil || char == nil {
		return nil, opErr(404, "character not found")
	}
	if tpl.Unique && char.CountOfTemplate(templateID) >= 1 {
		return nil, opErr(409, "character already holds that unique item")
	}
	created := make([]*items.Item, 0, qty)
	for i := int32(0); i < qty; i++ {
		inst, cerr := g.Facade.ItemsService().CreateInstanceFromTemplate(templateID)
		if cerr != nil || inst == nil {
			return nil, opErr(400, "could not create the item")
		}
		created = append(created, inst)
	}
	var added []string
	var deltas []stackDelta
	err = g.Facade.CharactersService().Modify(characterID, func(ch *characters.Character) error {
		if tpl.Unique && ch.CountOfTemplate(templateID) >= 1 {
			return opErr(409, "character already holds that unique item")
		}
		before := itemQtyByID(ch)
		for _, inst := range created {
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
	fresh, _ := g.Facade.CharactersService().FindByID(characterID)
	g.notifyInventory(fresh, fmt.Sprintf("You receive %s.", itemLabel(tpl, qty)))
	if g.QuestTracker != nil && fresh != nil {
		for _, inst := range created {
			g.QuestTracker.OnItemPickup(fresh.ID, fresh.BelongsUserID, inst)
		}
	}
	name := characterName(fresh)
	return &OpResult{
		Summary:    opSummary("Gave %s ×%d to %s.", tpl.Name, qty, name),
		Undoable:   true,
		EntityType: "character",
		EntityID:   characterID,
		Before:     jsonRaw(map[string]interface{}{"instanceIds": []string{}}),
		After:      jsonRaw(map[string]interface{}{"instanceIds": added, "stackDeltas": deltas}),
		Inverse: inverseOf("take-added", map[string]interface{}{
			"characterId": characterID,
			"instanceIds": added,
			"stackDeltas": deltas,
		}),
		Detail: map[string]interface{}{
			"characterId": characterID,
			"templateId":  templateID,
			"quantity":    qty,
			"instanceIds": added,
		},
	}, nil
}

// OpTakeItem removes an instance, or a quantity of a template, from the bag and then equipment.
func (g *Game) OpTakeItem(characterID, instanceID, templateID string, quantity int32) (*OpResult, error) {
	qty, err := qtyOrOne(quantity)
	if err != nil {
		return nil, err
	}
	if characterID == "" {
		return nil, opErr(400, "characterId is required")
	}
	if instanceID == "" && templateID == "" {
		return nil, opErr(400, "itemInstanceId or itemTemplateId is required")
	}
	char, err := g.Facade.CharactersService().FindByID(characterID)
	if err != nil || char == nil {
		return nil, opErr(404, "character not found")
	}
	var taken []takenPiece
	err = g.Facade.CharactersService().Modify(characterID, func(ch *characters.Character) error {
		var takeErr error
		if instanceID != "" {
			taken, takeErr = takeInstance(ch, instanceID, qty)
		} else {
			taken, takeErr = takeTemplate(ch, templateID, qty)
		}
		return takeErr
	})
	if err != nil {
		return nil, err
	}
	if len(taken) == 0 {
		return nil, opErr(404, "item not found on the character")
	}
	fresh, _ := g.Facade.CharactersService().FindByID(characterID)
	label := taken[0].Item.Name
	g.notifyInventory(fresh, fmt.Sprintf("%s was taken.", label))
	return &OpResult{
		Summary:    opSummary("Took %s from %s.", label, characterName(fresh)),
		Undoable:   true,
		EntityType: "character",
		EntityID:   characterID,
		Before:     jsonRaw(map[string]interface{}{"taken": taken}),
		After:      jsonRaw(map[string]interface{}{"removed": len(taken)}),
		Inverse: inverseOf("restore-items", map[string]interface{}{
			"characterId": characterID,
			"items":       taken,
		}),
		Detail: map[string]interface{}{
			"characterId": characterID,
			"taken":       len(taken),
		},
	}, nil
}

func (g *Game) notifyInventory(ch *characters.Character, text string) {
	if g == nil || ch == nil || ch.BelongsUserID == "" || !g.characterOnline(ch.ID) {
		return
	}
	if g.sendMessage == nil {
		return
	}
	reply := messages.Reply(ch.BelongsUserID, text)
	g.sendMessage <- reply
	g.sendMessage <- &messages.InventoryUpdateMessage{
		MessageResponse: messages.MessageResponse{
			Audience:   messages.MessageAudienceOrigin,
			AudienceID: ch.BelongsUserID,
			Type:       messages.MessageTypeInventoryUpdate,
		},
		Inventory:     ch.Inventory,
		EquippedItems: ch.EquippedItems,
		Gold:          ch.Gold,
	}
	if update := messages.NewCharacterUpdateMessage(ch.BelongsUserID, ch); update != nil {
		g.sendMessage <- update
	}
}

func (g *Game) characterOnline(characterID string) bool {
	if g == nil || g.Sessions == nil || characterID == "" {
		return false
	}
	for _, player := range g.Sessions.all() {
		if player.CharacterID == characterID {
			return true
		}
	}
	return false
}

func characterName(ch *characters.Character) string {
	if ch == nil {
		return ""
	}
	if ch.Name != "" {
		return ch.Name
	}
	return ch.ID
}

func itemLabel(item *items.Item, qty int32) string {
	name := "an item"
	if item != nil && item.Name != "" {
		name = item.Name
	}
	if qty > 1 {
		return fmt.Sprintf("%s ×%d", name, qty)
	}
	return name
}

func itemQty(item *items.Item) int32 {
	if item == nil {
		return 0
	}
	if item.Quantity > 0 {
		return item.Quantity
	}
	return 1
}

func itemQtyByID(ch *characters.Character) map[string]int32 {
	out := map[string]int32{}
	if ch == nil {
		return out
	}
	walkItems(ch.Inventory.Items, func(item *items.Item) {
		if item != nil && item.ID != "" {
			out[item.ID] = itemQty(item)
		}
	})
	for _, item := range ch.EquippedItems {
		if item != nil && item.ID != "" {
			out[item.ID] = itemQty(item)
		}
	}
	return out
}

func inventoryDelta(ch *characters.Character, before map[string]int32) ([]string, []stackDelta) {
	var added []string
	var deltas []stackDelta
	after := itemQtyByID(ch)
	for id, qty := range after {
		prev, ok := before[id]
		if !ok {
			added = append(added, id)
			continue
		}
		if qty > prev {
			deltas = append(deltas, stackDelta{ItemID: id, Quantity: qty - prev})
		}
	}
	return added, deltas
}

func walkItems(list []*items.Item, fn func(*items.Item)) {
	for _, item := range list {
		if item == nil {
			continue
		}
		fn(item)
		if len(item.Items) > 0 {
			walkItems([]*items.Item(item.Items), fn)
		}
	}
}

func holdsInstance(ch *characters.Character, id string) bool {
	if ch == nil || id == "" {
		return false
	}
	found := false
	walkItems(ch.Inventory.Items, func(item *items.Item) {
		if item.ID == id {
			found = true
		}
	})
	for _, item := range ch.EquippedItems {
		if item != nil && item.ID == id {
			found = true
		}
	}
	return found
}

func takeInstance(ch *characters.Character, id string, qty int32) ([]takenPiece, error) {
	if piece, ok := takeFromList(&ch.Inventory.Items, id, qty, ""); ok {
		return []takenPiece{piece}, nil
	}
	for slot, item := range ch.EquippedItems {
		if item == nil || item.ID != id {
			continue
		}
		copyItem := cloneItem(item)
		delete(ch.EquippedItems, slot)
		copyItem.Quantity = itemQty(item)
		return []takenPiece{{Item: copyItem, EquippedSlot: slot}}, nil
	}
	return nil, opErr(404, "item not found on the character")
}

func takeFromList(list *[]*items.Item, id string, qty int32, parent string) (takenPiece, bool) {
	if list == nil {
		return takenPiece{}, false
	}
	for i, item := range *list {
		if item == nil {
			continue
		}
		if item.ID == id {
			have := itemQty(item)
			if qty <= 0 {
				qty = have
			}
			if item.Stackable && qty < have {
				item.Quantity = have - qty
				partial := cloneItem(item)
				partial.Quantity = qty
				return takenPiece{Item: partial, Partial: true}, true
			}
			*list = append((*list)[:i], (*list)[i+1:]...)
			return takenPiece{Item: item}, true
		}
		if len(item.Items) > 0 {
			children := []*items.Item(item.Items)
			if piece, ok := takeFromList(&children, id, qty, item.ID); ok {
				item.Items = items.Items(children)
				return piece, true
			}
		}
	}
	return takenPiece{}, false
}

func takeTemplate(ch *characters.Character, templateID string, qty int32) ([]takenPiece, error) {
	var taken []takenPiece
	left := qty
	left, taken = consumeList(&ch.Inventory.Items, templateID, left, taken)
	if left > 0 && ch.EquippedItems != nil {
		for slot, item := range ch.EquippedItems {
			if left <= 0 {
				break
			}
			if item == nil || !item.MatchesTemplate(templateID) {
				continue
			}
			copyItem := cloneItem(item)
			delete(ch.EquippedItems, slot)
			taken = append(taken, takenPiece{Item: copyItem, EquippedSlot: slot})
			left -= itemQty(item)
		}
	}
	if left > 0 {
		return nil, opErr(404, "not enough matching items on the character")
	}
	return taken, nil
}

func consumeList(list *[]*items.Item, templateID string, left int32, taken []takenPiece) (int32, []takenPiece) {
	if list == nil || left <= 0 {
		return left, taken
	}
	kept := make([]*items.Item, 0, len(*list))
	for _, item := range *list {
		if item == nil || left <= 0 || !item.MatchesTemplate(templateID) {
			if item != nil && len(item.Items) > 0 && left > 0 {
				children := []*items.Item(item.Items)
				left, taken = consumeList(&children, templateID, left, taken)
				item.Items = items.Items(children)
			}
			kept = append(kept, item)
			continue
		}
		have := itemQty(item)
		if item.Stackable && have > left {
			item.Quantity = have - left
			partial := cloneItem(item)
			partial.Quantity = left
			taken = append(taken, takenPiece{Item: partial, Partial: true})
			left = 0
			kept = append(kept, item)
			continue
		}
		taken = append(taken, takenPiece{Item: item})
		left -= have
	}
	*list = kept
	return left, taken
}

func cloneItem(item *items.Item) *items.Item {
	if item == nil {
		return nil
	}
	raw, err := json.Marshal(item)
	if err != nil {
		cp := *item
		return &cp
	}
	var out items.Item
	if err := json.Unmarshal(raw, &out); err != nil {
		cp := *item
		return &cp
	}
	return &out
}

func cloneStarterItem(tpl *items.Item) *items.Item {
	if tpl == nil {
		return nil
	}
	cp := cloneItem(tpl)
	if cp == nil {
		return nil
	}
	templateID := ""
	if tpl.Entity != nil {
		templateID = tpl.ID
	}
	cp.Entity = entities.NewEntity()
	cp.IsTemplate = false
	if templateID != "" {
		cp.TemplateID = templateID
	}
	if cp.Quantity < 1 {
		cp.Quantity = 1
	}
	return cp
}
