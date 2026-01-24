package commands

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/entities/items"
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

	// Parse command: "use potion" or "use health potion"
	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- message.Reply("Use what? Usage: use <item>")
		return true
	}
	itemName := strings.Join(parts[1:], " ")

	// Find item in inventory
	item := message.Character.Inventory.FindItemByName(itemName)
	if item == nil {
		// Try by target name (name-suffix format)
		item = message.Character.Inventory.FindItemByTargetName(itemName)
	}

	if item == nil {
		game.SendMessage() <- message.Reply("You don't have a '" + itemName + "' in your inventory.")
		return true
	}

	// Validate item is usable (has OnUseScriptID, Type==consumable, or effect attributes)
	if !isUsable(item) {
		game.SendMessage() <- message.Reply("You can't use " + item.Name + ".")
		return true
	}

	// Get room for script context (may be nil if not in a room)
	var room interface{}
	if message.Character.CurrentRoomID != "" {
		room, _ = game.GetFacade().RoomsService().FindByID(message.Character.CurrentRoomID)
	}

	// Apply data-driven effects (healthRestore from Attributes)
	effectApplied := applyBuiltInEffects(game, message, item)

	// Execute OnUse script if defined
	scriptExecuted := false
	if item.OnUseScriptID != "" {
		scriptExecuted = executeItemScript(game, message, item, room)
	}

	// If neither effect nor script did anything meaningful, show generic message
	if !effectApplied && !scriptExecuted {
		game.SendMessage() <- message.Reply("You use " + item.Name + ".")
	}

	// Handle consumption (decrement quantity or remove item)
	if item.Consumable {
		consumeItem(game, message, item)
	}

	// Persist character changes
	err := game.GetFacade().CharactersService().Update(message.Character.ID, message.Character)
	if err != nil {
		log.WithError(err).Error("Failed to update character after item use")
	}

	return true
}

// isUsable checks if an item can be used
func isUsable(item *items.Item) bool {
	// Has explicit OnUse script
	if item.OnUseScriptID != "" {
		return true
	}
	// Has consumable type
	if item.Type == items.ItemTypeConsumable {
		return true
	}
	// Has Consumable flag
	if item.Consumable {
		return true
	}
	// Has effect attributes
	if item.Attributes != nil {
		if _, ok := item.Attributes["healthRestore"]; ok {
			return true
		}
		if _, ok := item.Attributes["manaRestore"]; ok {
			return true
		}
	}
	return false
}

// applyBuiltInEffects applies data-driven effects from item Attributes
func applyBuiltInEffects(game def.GameCtrl, message *messages.Message, item *items.Item) bool {
	if item.Attributes == nil {
		return false
	}

	applied := false
	char := message.Character

	// Health restoration
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

	// Mana restoration (placeholder for future mana system)
	if val, ok := item.Attributes["manaRestore"]; ok {
		amount := toInt32(val)
		if amount > 0 {
			game.SendMessage() <- message.Reply("You use " + item.Name + " and feel your magical energy restored.")
			applied = true
		}
	}

	// Custom use message (if defined and no other effect applied)
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
func executeItemScript(game def.GameCtrl, message *messages.Message, item *items.Item, room interface{}) bool {
	script, err := game.GetFacade().ScriptsService().FindByID(item.OnUseScriptID)
	if err != nil || script == nil {
		log.WithField("scriptID", item.OnUseScriptID).WithError(err).Warn("Item OnUse script not found")
		return false
	}

	// Build script context (following room enter script pattern from mudserver.go)
	ctx := scripts.NewScriptContext()
	ctx.Set("eventType", "item.use")
	ctx.Set("item", item)
	ctx.Set("character", message.Character)
	if room != nil {
		ctx.Set("room", room)
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
		// Decrement quantity
		item.Quantity--
	} else {
		// Remove item from inventory
		message.Character.Inventory.RemoveItem(item.ID)
		// Delete the item entity from database
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
