package commands

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/entities/recipes"
	"github.com/talesmud/talesmud/pkg/itemart"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// RecipesCommand lists all crafting recipes (visible to everyone — no profession gate).
type RecipesCommand struct{}

// Key ...
func (command *RecipesCommand) Key() CommandKey { return &ExactCommandKey{} }

// Execute ...
func (command *RecipesCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		return true
	}
	list := recipes.All()
	if len(list) == 0 {
		game.SendMessage() <- message.Reply("No recipes are known yet.")
		return true
	}

	var roomTags []string
	if room, err := game.GetFacade().RoomsService().FindByID(message.Character.CurrentRoomID); err == nil && room != nil {
		roomTags = room.Tags
	}

	rows := make([]messages.RecipeRow, 0, len(list))
	for _, r := range list {
		stationLabel := "anywhere"
		if r.Station != "" {
			stationLabel = r.Station
		}
		stationOK := recipes.StationOK(r, roomTags)
		ings := make([]messages.RecipeIngredientRow, 0, len(r.Ingredients))
		haveAll := true
		for _, ing := range r.Ingredients {
			have := message.Character.Inventory.CountMatchingTemplate(ing.Item)
			if have < ing.Qty {
				haveAll = false
			}
			ings = append(ings, messages.RecipeIngredientRow{
				Item:       ing.Item,
				Name:       itemDisplayName(game, ing.Item),
				Qty:        ing.Qty,
				Have:       have,
				Image:      itemart.URL(ing.Item, ing.Item),
				HaveEnough: have >= ing.Qty,
			})
		}
		outQty := r.Output.Qty
		if outQty < 1 {
			outQty = 1
		}
		rows = append(rows, messages.RecipeRow{
			ID:           r.ID,
			Key:          r.Key,
			Name:         r.Name,
			Description:  r.Description,
			Category:     r.Category,
			Station:      r.Station,
			StationLabel: stationLabel,
			StationHint:  recipes.StationHint(r.Station),
			StationOK:    stationOK,
			CanCraft:     stationOK && haveAll,
			Ingredients:  ings,
			Output: messages.RecipeOutputRow{
				Item:  r.Output.Item,
				Name:  itemDisplayName(game, r.Output.Item),
				Qty:   outQty,
				Image: itemart.URL(r.Output.Item, r.Output.Item),
			},
		})
	}

	if message.FromUser != nil {
		game.SendMessage() <- messages.NewRecipesMessage(message.FromUser.ID, rows, roomTags)
	} else {
		// Fallback text for non-user contexts
		var b strings.Builder
		b.WriteString("Crafting recipes (everyone can craft — no profession required):\n")
		b.WriteString("Use: craft <recipe>\n\n")
		for _, row := range rows {
			b.WriteString(fmt.Sprintf("• %s (%s) [%s] — %s\n", row.Name, row.Key, row.Category, row.StationLabel))
		}
		game.SendMessage() <- message.Reply(b.String())
	}
	return true
}

func itemDisplayName(game def.GameCtrl, templateID string) string {
	if game == nil || templateID == "" {
		return templateID
	}
	item, err := game.GetFacade().ItemsService().FindByID(templateID)
	if err != nil || item == nil || item.Name == "" {
		return templateID
	}
	return item.Name
}

// CraftCommand crafts an item from a recipe: craft <recipe>
type CraftCommand struct{}

// Key ...
func (command *CraftCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute ...
func (command *CraftCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		return true
	}

	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		// Bare "craft" opens the recipes overlay (same as recipes).
		return (&RecipesCommand{}).Execute(game, message)
	}
	query := strings.Join(parts[1:], " ")
	recipe := recipes.Find(query)
	if recipe == nil {
		game.SendMessage() <- message.Reply(fmt.Sprintf("Unknown recipe '%s'. Type 'recipes' for the full list.", query))
		return true
	}

	room, err := game.GetFacade().RoomsService().FindByID(message.Character.CurrentRoomID)
	if err != nil || room == nil {
		game.SendMessage() <- message.Reply("You can't craft here.")
		return true
	}
	if !recipes.StationOK(recipe, room.Tags) {
		hint := recipes.StationHint(recipe.Station)
		if hint == "" {
			hint = fmt.Sprintf("Requires a %s.", recipe.Station)
		}
		game.SendMessage() <- message.Reply(fmt.Sprintf("You can't craft %s here. %s", recipe.Name, hint))
		return true
	}

	// Check ingredients
	missing := make([]string, 0)
	for _, ing := range recipe.Ingredients {
		have := message.Character.Inventory.CountMatchingTemplate(ing.Item)
		if have < ing.Qty {
			missing = append(missing, fmt.Sprintf("%dx %s (have %d)", ing.Qty, itemDisplayName(game, ing.Item), have))
		}
	}
	if len(missing) > 0 {
		game.SendMessage() <- message.Reply(fmt.Sprintf("Missing materials for %s: %s", recipe.Name, strings.Join(missing, "; ")))
		return true
	}

	// Consume ingredients
	for _, ing := range recipe.Ingredients {
		if err := message.Character.Inventory.ConsumeMatchingTemplate(ing.Item, ing.Qty); err != nil {
			game.SendMessage() <- message.Reply("Something went wrong consuming materials.")
			log.WithError(err).WithField("item", ing.Item).Warn("craft: consume failed")
			return true
		}
	}

	// Create output
	outQty := recipe.Output.Qty
	if outQty < 1 {
		outQty = 1
	}
	newItem, createErr := game.GetFacade().ItemsService().CreateInstanceFromTemplate(recipe.Output.Item)
	if createErr != nil {
		log.WithError(createErr).WithField("template", recipe.Output.Item).Error("craft: create output failed")
		game.SendMessage() <- message.Reply("The materials are ready, but the result failed to form. Try again.")
		// Best-effort: do not refund (simple v1). Character already consumed — persist anyway.
		_ = game.GetFacade().CharactersService().Update(message.Character.ID, message.Character)
		return true
	}
	if newItem.Stackable {
		newItem.Quantity = outQty
	}

	if addErr := message.Character.Inventory.AddItem(newItem); addErr != nil {
		_ = game.GetFacade().ItemsService().Delete(newItem.ID)
		game.SendMessage() <- message.Reply("Your inventory is full — clear space and craft again. (Materials were already used.)")
		_ = game.GetFacade().CharactersService().Update(message.Character.ID, message.Character)
		return true
	}

	if err := game.GetFacade().CharactersService().Update(message.Character.ID, message.Character); err != nil {
		log.WithError(err).Warn("craft: failed to persist character")
	}
	if inv := messages.NewInventoryUpdateMessage(message); inv != nil {
		game.SendMessage() <- inv
	}

	outName := newItem.Name
	if outName == "" {
		outName = itemDisplayName(game, recipe.Output.Item)
	}
	qtyNote := ""
	if outQty > 1 {
		qtyNote = fmt.Sprintf(" x%d", outQty)
	}
	game.SendMessage() <- message.Reply(fmt.Sprintf("You craft %s%s. (%s)", outName, qtyNote, recipe.Name))
	return true
}
