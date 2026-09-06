package commands

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/scripts"
)

// UseCommand handles using consumable/usable items
type UseCommand struct {
}

// Key returns the command key matcher
func (command *UseCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute handles the use command
func (command *UseCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	// Parse: "use potion" | "use flint on torch" | "use flint on dusty torch"
	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- message.Reply("Use what? Usage: use <item> [on <target>]")
		return true
	}

	itemName, targetName := parseUseArgs(parts[1:])
	if itemName == "" {
		game.SendMessage() <- message.Reply("Use what? Usage: use <item> [on <target>]")
		return true
	}

	item := message.Character.Inventory.FindItemByName(itemName)
	if item == nil {
		item = message.Character.Inventory.FindItemByTargetName(itemName)
	}
	if item == nil {
		game.SendMessage() <- message.Reply("You don't have a '" + itemName + "' in your inventory.")
		return true
	}

	var room *rooms.Room
	if message.Character.CurrentRoomID != "" {
		room, _ = game.GetFacade().RoomsService().FindByID(message.Character.CurrentRoomID)
	}

	var targetItem *items.Item
	if targetName != "" {
		targetItem = findUseTarget(message, game, room, targetName)
		if targetItem == nil {
			game.SendMessage() <- message.Reply("You don't see a '" + targetName + "' to use that on.")
			return true
		}
	}

	// Must have a real effect path — never treat consumable:true alone as usable.
	if !isUsable(item) && !canLightTarget(item, targetItem) {
		game.SendMessage() <- message.Reply("You can't use " + item.Name + ".")
		return true
	}

	effectApplied := applyBuiltInEffects(game, message, item)
	if !effectApplied {
		effectApplied = applyLightSourceUse(game, message, item, targetItem)
	}

	scriptExecuted := false
	if item.OnUseScriptID != "" {
		scriptExecuted = executeItemScript(game, message, item, room, targetItem, targetName)
	}

	if !effectApplied && !scriptExecuted {
		// Refuse no-op: do NOT print "You use X" and do NOT consume.
		if targetName != "" {
			game.SendMessage() <- message.Reply("Nothing happens when you use " + item.Name + " on " + targetItem.Name + ".")
		} else {
			game.SendMessage() <- message.Reply("Nothing happens when you use " + item.Name + ".")
		}
		return true
	}

	if item.Consumable {
		consumeItem(game, message, item)
	}

	err := game.GetFacade().CharactersService().Update(message.Character.ID, message.Character)
	if err != nil {
		log.WithError(err).Error("Failed to update character after item use")
	}

	if inv := messages.NewInventoryUpdateMessage(message); inv != nil {
		game.SendMessage() <- inv
	}

	return true
}

// parseUseArgs splits "flint on torch" into item + optional target.
func parseUseArgs(args []string) (itemName, targetName string) {
	onIdx := -1
	for i, p := range args {
		if strings.EqualFold(p, "on") {
			onIdx = i
			break
		}
	}
	if onIdx < 0 {
		return strings.Join(args, " "), ""
	}
	if onIdx == 0 || onIdx == len(args)-1 {
		return strings.Join(args, " "), ""
	}
	return strings.Join(args[:onIdx], " "), strings.Join(args[onIdx+1:], " ")
}

func findUseTarget(message *messages.Message, game def.GameCtrl, room *rooms.Room, targetName string) *items.Item {
	if message.Character != nil {
		if inv := message.Character.Inventory.FindItemByName(targetName); inv != nil {
			return inv
		}
		if inv := message.Character.Inventory.FindItemByTargetName(targetName); inv != nil {
			return inv
		}
		for _, eq := range message.Character.EquippedItems {
			if eq == nil {
				continue
			}
			if strings.EqualFold(eq.Name, targetName) || strings.EqualFold(eq.GetTargetName(), targetName) ||
				strings.HasPrefix(strings.ToLower(eq.Name), strings.ToLower(targetName)) {
				return eq
			}
		}
	}
	if room != nil {
		return findItemInRoom(room, game, targetName, message.Character)
	}
	return nil
}

// isUsable checks if an item can be used.
// consumable:true alone is NOT enough — needs a script or effect attribute.
func isUsable(item *items.Item) bool {
	if item.OnUseScriptID != "" {
		return true
	}
	if item.Attributes != nil {
		if _, ok := item.Attributes["healthRestore"]; ok {
			return true
		}
		if _, ok := item.Attributes["manaRestore"]; ok {
			return true
		}
		if msg, ok := item.Attributes["useMessage"]; ok {
			if msgStr, isStr := msg.(string); isStr && msgStr != "" {
				return true
			}
		}
	}
	return false
}

func canLightTarget(tool, target *items.Item) bool {
	if tool == nil || target == nil {
		return false
	}
	if !isLightSource(target) {
		return false
	}
	return isFireStarter(tool)
}

func isLightSource(item *items.Item) bool {
	if item == nil {
		return false
	}
	if string(item.SubType) == "light_source" {
		return true
	}
	for _, tag := range item.Tags {
		if strings.EqualFold(tag, "light") {
			return true
		}
	}
	return false
}

func isFireStarter(item *items.Item) bool {
	if item == nil {
		return false
	}
	name := strings.ToLower(item.Name)
	if strings.Contains(name, "flint") || strings.Contains(name, "tinder") {
		return true
	}
	if string(item.SubType) == "tool" {
		for _, tag := range item.Tags {
			if strings.EqualFold(tag, "tool") || strings.EqualFold(tag, "utility") {
				return true
			}
		}
	}
	if item.Attributes != nil {
		if v, ok := item.Attributes["canLight"]; ok {
			switch t := v.(type) {
			case bool:
				return t
			case string:
				return strings.EqualFold(t, "true") || t == "1"
			}
		}
	}
	return false
}

// applyLightSourceUse lights a torch/light_source target and sets torch_lit for scripts.
func applyLightSourceUse(game def.GameCtrl, message *messages.Message, tool, target *items.Item) bool {
	if !canLightTarget(tool, target) {
		return false
	}
	if target.Attributes != nil {
		if lit, ok := target.Attributes["lit"].(bool); ok && lit {
			game.SendMessage() <- message.Reply(target.Name + " is already lit.")
			return true
		}
	}
	if target.Attributes == nil {
		target.Attributes = map[string]interface{}{}
	}
	target.Attributes["lit"] = true

	if message.Character.Flags == nil {
		message.Character.Flags = map[string]interface{}{}
	}
	message.Character.Flags["torch_lit"] = true

	// Persist target attribute change when it is an owned inventory/equipped item.
	if err := game.GetFacade().ItemsService().Update(target.ID, target); err != nil {
		log.WithField("itemID", target.ID).WithError(err).Warn("Failed to persist lit attribute")
	}

	game.SendMessage() <- message.Reply("You strike " + tool.Name + " and light " + target.Name + ".")
	return true
}

// applyBuiltInEffects applies data-driven effects from item Attributes
func applyBuiltInEffects(game def.GameCtrl, message *messages.Message, item *items.Item) bool {
	if item.Attributes == nil {
		return false
	}

	applied := false
	char := message.Character

	if val, ok := item.Attributes["healthRestore"]; ok {
		amount := toInt32(val)
		if amount > 0 {
			oldHP := char.CurrentHitPoints
			char.CurrentHitPoints += amount
			if char.CurrentHitPoints > char.MaxHitPoints {
				char.CurrentHitPoints = char.MaxHitPoints
			}
			healed := char.CurrentHitPoints - oldHP
			if healed > 0 {
				game.SendMessage() <- message.Reply("You use " + item.Name + " and restore " + itoa(int(healed)) + " health.")
			} else {
				game.SendMessage() <- message.Reply("You use " + item.Name + " but you're already at full health.")
			}
			applied = true
		}
	}

	if val, ok := item.Attributes["manaRestore"]; ok {
		amount := toInt32(val)
		if amount > 0 {
			if char.MaxMana <= 0 {
				game.SendMessage() <- message.Reply("You use " + item.Name + " but you have no use for magical energy.")
				applied = true
			} else {
				oldMana := char.CurrentMana
				char.CurrentMana += amount
				if char.CurrentMana > char.MaxMana {
					char.CurrentMana = char.MaxMana
				}
				restored := char.CurrentMana - oldMana
				if restored > 0 {
					game.SendMessage() <- message.Reply("You use " + item.Name + " and restore " + itoa(int(restored)) + " mana.")
				} else {
					game.SendMessage() <- message.Reply("You use " + item.Name + " but your mana is already full.")
				}
				applied = true
			}
		}
	}

	if !applied {
		if msg, ok := item.Attributes["useMessage"]; ok {
			if msgStr, isStr := msg.(string); isStr && msgStr != "" {
				game.SendMessage() <- message.Reply(msgStr)
				applied = true
			}
		}
	}

	return applied
}

// executeItemScript runs the OnUse Lua script
func executeItemScript(game def.GameCtrl, message *messages.Message, item *items.Item, room *rooms.Room, targetItem *items.Item, targetName string) bool {
	script, err := game.GetFacade().ScriptsService().FindByID(item.OnUseScriptID)
	if err != nil || script == nil {
		log.WithField("scriptID", item.OnUseScriptID).WithError(err).Warn("Item OnUse script not found")
		return false
	}

	ctx := scripts.NewScriptContext()
	ctx.Set("eventType", "item.use")
	ctx.Set("item", item)
	ctx.Set("character", message.Character)
	if room != nil {
		ctx.Set("room", room)
	}
	if targetName != "" {
		ctx.Set("useTarget", targetName)
	}
	if targetItem != nil {
		ctx.Set("targetItem", targetItem)
	}

	result := game.GetFacade().Runner().RunWithResult(*script, ctx)
	if result != nil && !result.Success {
		log.WithField("script", script.Name).WithField("error", result.Error).Warn("Item OnUse script failed")
		return false
	}

	return true
}

// consumeItem decrements quantity or removes the item from inventory
func consumeItem(game def.GameCtrl, message *messages.Message, item *items.Item) {
	if item.Stackable && item.Quantity > 1 {
		item.Quantity--
		if err := game.GetFacade().ItemsService().Update(item.ID, item); err != nil {
			log.WithField("itemID", item.ID).WithError(err).Warn("Failed to update consumed stack quantity")
		}
	} else {
		message.Character.Inventory.RemoveItem(item.ID)
		err := game.GetFacade().ItemsService().Delete(item.ID)
		if err != nil {
			log.WithField("itemID", item.ID).WithError(err).Warn("Failed to delete consumed item")
		}
	}
}

// toInt32 converts interface{} to int32
func toInt32(val interface{}) int32 {
	switch v := val.(type) {
	case float64:
		return int32(v)
	case int:
		return int32(v)
	case int32:
		return v
	case int64:
		return int32(v)
	}
	return 0
}
