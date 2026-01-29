package commands

import (
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// ListCommand shows merchant inventory
type ListCommand struct {
}

// Key returns the command key matcher
func (command *ListCommand) Key() CommandKey { return &ExactCommandKey{} }

// Execute handles the list/shop command
func (command *ListCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	if message.Character.CurrentRoomID == "" {
		game.SendMessage() <- message.Reply("You are not in a room.")
		return true
	}

	// Find merchant in room
	merchant := findMerchantInRoom(game, message.Character.CurrentRoomID)
	if merchant == nil {
		game.SendMessage() <- message.Reply("There is no merchant here.")
		return true
	}

	if merchant.MerchantTrait == nil || len(merchant.MerchantTrait.Inventory) == 0 {
		game.SendMessage() <- message.Reply(merchant.Name + " has nothing for sale.")
		return true
	}

	// Build shop display
	var sb strings.Builder
	sb.WriteString("=== ")
	sb.WriteString(merchant.Name)
	sb.WriteString("'s Shop ===\n")
	sb.WriteString("Your gold: ")
	sb.WriteString(itoa64(message.Character.Gold))
	sb.WriteString("\n\n")

	for i, invItem := range merchant.MerchantTrait.Inventory {
		// Fetch item template for name and base price
		itemTemplate, err := game.GetFacade().ItemsService().FindByID(invItem.ItemTemplateID)
		if err != nil || itemTemplate == nil {
			continue
		}

		// Calculate price
		price := merchant.MerchantTrait.GetBuyPrice(&invItem, itemTemplate.BasePrice)

		sb.WriteString(itoa(i + 1))
		sb.WriteString(". ")
		sb.WriteString(itemTemplate.Name)

		// Stock indicator
		if invItem.Quantity >= 0 {
			sb.WriteString(" [")
			sb.WriteString(itoa(int(invItem.Quantity)))
			sb.WriteString(" in stock]")
		} else {
			sb.WriteString(" [unlimited]")
		}

		// Price
		sb.WriteString(" - ")
		sb.WriteString(itoa64(price))
		sb.WriteString(" gold")

		// Level requirement
		if invItem.RequiredLevel > 0 {
			sb.WriteString(" (Lv.")
			sb.WriteString(itoa(int(invItem.RequiredLevel)))
			sb.WriteString("+)")
		}

		sb.WriteString("\n")
	}

	sb.WriteString("\nUse 'buy <item>' to purchase.")

	game.SendMessage() <- message.Reply(sb.String())
	return true
}

// BuyCommand handles buying items from a merchant
type BuyCommand struct {
}

// Key returns the command key matcher
func (command *BuyCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute handles the buy command
func (command *BuyCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	if message.Character.CurrentRoomID == "" {
		game.SendMessage() <- message.Reply("You are not in a room.")
		return true
	}

	// Parse: "buy <item> [quantity]"
	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- message.Reply("Buy what? Usage: buy <item> [quantity]")
		return true
	}

	// Check for quantity at the end
	quantity := int32(1)
	var itemName string
	lastPart := parts[len(parts)-1]
	if q, err := strconv.Atoi(lastPart); err == nil && len(parts) > 2 {
		quantity = int32(q)
		itemName = strings.Join(parts[1:len(parts)-1], " ")
	} else {
		itemName = strings.Join(parts[1:], " ")
	}

	// Find merchant
	merchant := findMerchantInRoom(game, message.Character.CurrentRoomID)
	if merchant == nil {
		game.SendMessage() <- message.Reply("There is no merchant here.")
		return true
	}

	if merchant.MerchantTrait == nil {
		game.SendMessage() <- message.Reply(merchant.Name + " is not selling anything.")
		return true
	}

	// Find item in merchant inventory
	var foundItem *npc.MerchantItem
	var foundIndex int
	itemNameLower := strings.ToLower(itemName)

	for i := range merchant.MerchantTrait.Inventory {
		invItem := &merchant.MerchantTrait.Inventory[i]
		itemTemplate, err := game.GetFacade().ItemsService().FindByID(invItem.ItemTemplateID)
		if err != nil || itemTemplate == nil {
			continue
		}

		if strings.EqualFold(itemTemplate.Name, itemName) ||
			strings.HasPrefix(strings.ToLower(itemTemplate.Name), itemNameLower) {
			foundItem = invItem
			foundIndex = i
			break
		}
	}

	if foundItem == nil {
		game.SendMessage() <- message.Reply(merchant.Name + " doesn't sell '" + itemName + "'.")
		return true
	}

	// Check stock
	if foundItem.Quantity >= 0 && foundItem.Quantity < quantity {
		if foundItem.Quantity == 0 {
			game.SendMessage() <- message.Reply("That item is out of stock.")
		} else {
			game.SendMessage() <- message.Reply("Only " + itoa(int(foundItem.Quantity)) + " in stock.")
		}
		return true
	}

	// Get item template for pricing
	itemTemplate, _ := game.GetFacade().ItemsService().FindByID(foundItem.ItemTemplateID)
	price := merchant.MerchantTrait.GetBuyPrice(foundItem, itemTemplate.BasePrice)
	totalPrice := price * int64(quantity)

	// Check level requirement
	if foundItem.RequiredLevel > 0 && message.Character.Level < foundItem.RequiredLevel {
		game.SendMessage() <- message.Reply("You need to be level " + itoa(int(foundItem.RequiredLevel)) + " to buy that.")
		return true
	}

	// Check gold
	if message.Character.Gold < totalPrice {
		game.SendMessage() <- message.Reply("You don't have enough gold. Need " + itoa64(totalPrice) + " gold.")
		return true
	}

	// Check inventory space
	if message.Character.Inventory.IsFull() {
		game.SendMessage() <- message.Reply("Your inventory is full.")
		return true
	}

	// Create item instance(s)
	for i := int32(0); i < quantity; i++ {
		newItem, err := game.GetFacade().ItemsService().CreateInstanceFromTemplate(foundItem.ItemTemplateID)
		if err != nil {
			log.WithError(err).Error("Failed to create item instance")
			game.SendMessage() <- message.Reply("Error creating item.")
			return true
		}

		// Store the item
		storedItem, err := game.GetFacade().ItemsService().Store(newItem)
		if err != nil {
			log.WithError(err).Error("Failed to store item")
			continue
		}

		// Add to inventory
		err = message.Character.Inventory.AddItem(storedItem)
		if err != nil {
			log.WithError(err).Error("Failed to add item to inventory")
			continue
		}
	}

	// Deduct gold
	message.Character.Gold -= totalPrice

	// Update stock
	if foundItem.Quantity >= 0 {
		merchant.MerchantTrait.Inventory[foundIndex].Quantity -= quantity
	}

	// Persist character
	err := game.GetFacade().CharactersService().Update(message.Character.ID, message.Character)
	if err != nil {
		log.WithError(err).Error("Failed to update character")
	}

	// Send confirmation
	var msg string
	if quantity > 1 {
		msg = "You buy " + itoa(int(quantity)) + "x " + itemTemplate.Name + " for " + itoa64(totalPrice) + " gold."
	} else {
		msg = "You buy " + itemTemplate.Name + " for " + itoa64(totalPrice) + " gold."
	}
	game.SendMessage() <- message.Reply(msg)
	if inv := messages.NewInventoryUpdateMessage(message); inv != nil {
		game.SendMessage() <- inv
	}

	return true
}

// SellCommand handles selling items to a merchant
type SellCommand struct {
}

// Key returns the command key matcher
func (command *SellCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute handles the sell command
func (command *SellCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	if message.Character.CurrentRoomID == "" {
		game.SendMessage() <- message.Reply("You are not in a room.")
		return true
	}

	// Parse: "sell <item> [quantity]"
	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- message.Reply("Sell what? Usage: sell <item> [quantity]")
		return true
	}

	// Check for quantity at the end
	quantity := int32(1)
	var itemName string
	lastPart := parts[len(parts)-1]
	if q, err := strconv.Atoi(lastPart); err == nil && len(parts) > 2 {
		quantity = int32(q)
		itemName = strings.Join(parts[1:len(parts)-1], " ")
	} else {
		itemName = strings.Join(parts[1:], " ")
	}

	// Find merchant
	merchant := findMerchantInRoom(game, message.Character.CurrentRoomID)
	if merchant == nil {
		game.SendMessage() <- message.Reply("There is no merchant here.")
		return true
	}

	if merchant.MerchantTrait == nil {
		game.SendMessage() <- message.Reply(merchant.Name + " is not buying anything.")
		return true
	}

	// Find item in inventory
	item := message.Character.Inventory.FindItemByName(itemName)
	if item == nil {
		item = message.Character.Inventory.FindItemByTargetName(itemName)
	}
	if item == nil {
		game.SendMessage() <- message.Reply("You don't have '" + itemName + "' in your inventory.")
		return true
	}

	// Check if merchant accepts this item
	if !merchant.MerchantTrait.CanBuyItem(string(item.Type), item.Tags) {
		game.SendMessage() <- message.Reply(merchant.Name + " doesn't want to buy that.")
		return true
	}

	// Check quantity for stackable items
	if item.Stackable && quantity > item.Quantity {
		game.SendMessage() <- message.Reply("You only have " + itoa(int(item.Quantity)) + " of those.")
		return true
	}

	// Calculate sell price
	price := merchant.MerchantTrait.GetSellPrice(item.BasePrice)
	if price == 0 {
		price = 1 // Minimum 1 gold
	}
	totalPrice := price * int64(quantity)

	// Handle stackable items
	if item.Stackable && quantity < item.Quantity {
		item.Quantity -= quantity
	} else {
		// Remove item from inventory
		_, err := message.Character.Inventory.RemoveItem(item.ID)
		if err != nil {
			game.SendMessage() <- message.Reply("Error removing item.")
			return true
		}

		// Delete the item from database
		game.GetFacade().ItemsService().Delete(item.ID)
	}

	// Add gold
	message.Character.Gold += totalPrice

	// Persist character
	err := game.GetFacade().CharactersService().Update(message.Character.ID, message.Character)
	if err != nil {
		log.WithError(err).Error("Failed to update character")
	}

	// Send confirmation
	var msg string
	if quantity > 1 {
		msg = "You sell " + itoa(int(quantity)) + "x " + item.Name + " for " + itoa64(totalPrice) + " gold."
	} else {
		msg = "You sell " + item.Name + " for " + itoa64(totalPrice) + " gold."
	}
	game.SendMessage() <- message.Reply(msg)
	if inv := messages.NewInventoryUpdateMessage(message); inv != nil {
		game.SendMessage() <- inv
	}

	return true
}

// ValueCommand shows what a merchant would pay for an item
type ValueCommand struct {
}

// Key returns the command key matcher
func (command *ValueCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute handles the value/price command
func (command *ValueCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	if message.Character.CurrentRoomID == "" {
		game.SendMessage() <- message.Reply("You are not in a room.")
		return true
	}

	// Parse: "value <item>"
	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- message.Reply("Check the value of what? Usage: value <item>")
		return true
	}

	itemName := strings.Join(parts[1:], " ")

	// Find merchant
	merchant := findMerchantInRoom(game, message.Character.CurrentRoomID)
	if merchant == nil {
		game.SendMessage() <- message.Reply("There is no merchant here to appraise items.")
		return true
	}

	if merchant.MerchantTrait == nil {
		game.SendMessage() <- message.Reply(merchant.Name + " can't appraise items.")
		return true
	}

	// Find item in inventory
	item := message.Character.Inventory.FindItemByName(itemName)
	if item == nil {
		item = message.Character.Inventory.FindItemByTargetName(itemName)
	}
	if item == nil {
		game.SendMessage() <- message.Reply("You don't have '" + itemName + "' in your inventory.")
		return true
	}

	// Check if merchant accepts this item
	if !merchant.MerchantTrait.CanBuyItem(string(item.Type), item.Tags) {
		game.SendMessage() <- message.Reply(merchant.Name + " doesn't want to buy that.")
		return true
	}

	// Calculate sell price
	price := merchant.MerchantTrait.GetSellPrice(item.BasePrice)
	if price == 0 {
		price = 1
	}

	game.SendMessage() <- message.Reply(merchant.Name + " will pay " + itoa64(price) + " gold for " + item.Name + ".")
	return true
}

// findMerchantInRoom finds a merchant NPC in the given room
func findMerchantInRoom(game def.GameCtrl, roomID string) *npc.NPC {
	npcManager := game.GetNPCInstanceManager()
	if npcManager == nil {
		return nil
	}

	instances := npcManager.GetInstancesInRoom(roomID)
	for _, inst := range instances {
		if inst.IsMerchant() {
			return inst
		}
	}
	return nil
}
