// Package textline turns classic game messages into terminal lines.
// Door frames are not rendered here. Unknown messages fall back to .Message.
package textline

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// Options tunes a few lines that need a web address.
type Options struct {
	// PlayURL is where an account with no character is sent, usually https://host/play.
	PlayURL string
}

// Result is the text to print and an optional numbered choice list.
type Result struct {
	Text       string
	Choices    []string
	ChoiceKind string
}

// Render formats one outbound game value. An empty Text means "say nothing"
// (pings, and updates the web client only paints in widgets).
func Render(v any, opt Options) Result {
	if v == nil {
		return Result{}
	}
	switch msg := v.(type) {
	case *messages.EnterRoomMessage:
		return renderEnter(msg)
	case messages.EnterRoomMessage:
		return renderEnter(&msg)
	case *messages.RoomPresenceMessage:
		return renderPresence(msg)
	case messages.RoomPresenceMessage:
		return renderPresence(&msg)
	case *messages.DialogMessage:
		return renderDialog(msg)
	case messages.DialogMessage:
		return renderDialog(&msg)
	case *messages.ShopMessage:
		return renderShop(msg)
	case messages.ShopMessage:
		return renderShop(&msg)
	case *messages.QuestUpdateMessage:
		return renderQuest(msg)
	case messages.QuestUpdateMessage:
		return renderQuest(&msg)
	case *messages.QuestLogMessage:
		return renderQuestLog(msg)
	case messages.QuestLogMessage:
		return renderQuestLog(&msg)
	case *messages.CombatStartMessage:
		return renderCombatStart(msg)
	case messages.CombatStartMessage:
		return renderCombatStart(&msg)
	case *messages.CombatTurnMessage:
		return textOnly(msg.Message, msg.ActorName, msg.Round)
	case messages.CombatTurnMessage:
		return textOnly(msg.Message, msg.ActorName, msg.Round)
	case *messages.CombatActionMessage:
		return textOnly(msg.Message, "", 0)
	case messages.CombatActionMessage:
		return textOnly(msg.Message, "", 0)
	case *messages.CombatStatusMessage:
		return textOnly(msg.Message, "", 0)
	case messages.CombatStatusMessage:
		return textOnly(msg.Message, "", 0)
	case *messages.CombatEndMessage:
		return renderCombatEnd(msg)
	case messages.CombatEndMessage:
		return renderCombatEnd(&msg)
	case *messages.LevelUpMessage:
		return renderLevel(msg)
	case messages.LevelUpMessage:
		return renderLevel(&msg)
	case *messages.CharacterUpdateMessage:
		return renderStats(msg)
	case messages.CharacterUpdateMessage:
		return renderStats(&msg)
	case *messages.CharacterSelected:
		return renderSelected(msg)
	case messages.CharacterSelected:
		return renderSelected(&msg)
	case *messages.InventoryUpdateMessage:
		return renderInventory(msg)
	case messages.InventoryUpdateMessage:
		return renderInventory(&msg)
	case *messages.PartyInviteMessage:
		return renderInvite(msg)
	case messages.PartyInviteMessage:
		return renderInvite(&msg)
	case *messages.FriendsMessage:
		return renderFriends(msg)
	case messages.FriendsMessage:
		return renderFriends(&msg)
	case *messages.PartyMessage:
		return renderParty(msg)
	case messages.PartyMessage:
		return renderParty(&msg)
	case *messages.RecipesMessage:
		return renderRecipes(msg)
	case messages.RecipesMessage:
		return renderRecipes(&msg)
	case messages.MessageResponse:
		return renderPlain(msg, opt)
	case *messages.MessageResponse:
		if msg == nil {
			return Result{}
		}
		return renderPlain(*msg, opt)
	default:
		if typer, ok := v.(interface{ GetType() messages.MessageType }); ok && typer.GetType() == messages.MessageTypePing {
			return Result{}
		}
		if m, ok := v.(interface{ GetMessage() string }); ok {
			text := strings.TrimSpace(Sanitize(m.GetMessage()))
			if text == "" {
				return Result{}
			}
			return Result{Text: text}
		}
		return Result{}
	}
}

func renderPlain(msg messages.MessageResponse, opt Options) Result {
	if msg.Type == messages.MessageTypePing {
		return Result{}
	}
	text := Sanitize(msg.Message)
	if msg.Username != "" {
		name := Sanitize(msg.Username)
		body := strings.TrimSpace(text)
		if body == "" {
			return Result{}
		}
		text = name + ": " + body
	}
	if msg.Type == messages.MessageTypeCreateCharacter {
		url := strings.TrimSpace(opt.PlayURL)
		if url == "" {
			url = "/play"
		}
		text = strings.TrimRight(text, "\n") + "\nCreate a character at " + url + "\n"
	}
	text = strings.TrimRight(text, "\n")
	if strings.HasPrefix(strings.TrimSpace(text), "Your Characters:") {
		return numberCharacters(text)
	}
	if strings.TrimSpace(text) == "" {
		return Result{}
	}
	return Result{Text: text}
}

func numberCharacters(text string) Result {
	lines := strings.Split(text, "\n")
	var choices []string
	var b strings.Builder
	n := 0
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "- ") {
			n++
			body := strings.TrimPrefix(trim, "- ")
			name := body
			if i := strings.Index(body, " ["); i > 0 {
				name = body[:i]
			}
			choices = append(choices, Sanitize(name))
			fmt.Fprintf(&b, "%d. %s\n", n, Sanitize(body))
			continue
		}
		if strings.Contains(trim, "sc [charactername]") {
			b.WriteString("Enter a number, or: sc [charactername]\n")
			continue
		}
		if trim == "" {
			b.WriteByte('\n')
			continue
		}
		b.WriteString(Sanitize(line))
		b.WriteByte('\n')
	}
	return Result{Text: strings.TrimRight(b.String(), "\n"), Choices: choices, ChoiceKind: "character"}
}

func renderEnter(msg *messages.EnterRoomMessage) Result {
	if msg == nil {
		return Result{}
	}
	text := strings.TrimRight(Sanitize(msg.Message), "\n")
	if text == "" {
		var b strings.Builder
		name := strings.TrimSpace(msg.Room.Name)
		if name == "" {
			name = "Room"
		}
		fmt.Fprintf(&b, "[%s]", Sanitize(name))
		if exits := exitNames(msg); exits != "" {
			fmt.Fprintf(&b, "\nExits: %s", exits)
		}
		if npcs := npcNames(msg); npcs != "" {
			fmt.Fprintf(&b, "\nNPCs: %s", npcs)
		}
		if players := playerNames(msg.Players); players != "" {
			fmt.Fprintf(&b, "\nPlayers: %s", players)
		}
		if items := itemLine(msg); items != "" {
			fmt.Fprintf(&b, "\nItems: %s", items)
		}
		return Result{Text: b.String()}
	}
	if extra := itemLine(msg); extra != "" && !strings.Contains(text, extra) {
		text += "\nItems: " + extra
	}
	return Result{Text: text}
}

func exitNames(msg *messages.EnterRoomMessage) string {
	if msg.Room.Exits == nil {
		return ""
	}
	var names []string
	for _, exit := range *msg.Room.Exits {
		if exit.Hidden || strings.TrimSpace(exit.Name) == "" {
			continue
		}
		names = append(names, Sanitize(exit.Name))
	}
	return strings.Join(names, ", ")
}

func npcNames(msg *messages.EnterRoomMessage) string {
	var names []string
	for _, n := range msg.NPCs {
		name := n.DisplayName
		if name == "" {
			name = n.Name
		}
		if strings.TrimSpace(name) == "" {
			continue
		}
		if n.IsEnemy {
			name += " (enemy)"
		}
		names = append(names, Sanitize(name))
	}
	return strings.Join(names, ", ")
}

func playerNames(players []messages.RoomPlayer) string {
	var names []string
	for _, p := range players {
		name := strings.TrimSpace(p.Name)
		if name == "" {
			continue
		}
		if p.IsYou {
			name += " (you)"
		}
		names = append(names, Sanitize(name))
	}
	return strings.Join(names, ", ")
}

func itemLine(msg *messages.EnterRoomMessage) string {
	var names []string
	for _, item := range msg.Items {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}
		names = append(names, Sanitize(item.Name))
	}
	return strings.Join(names, ", ")
}

func renderPresence(msg *messages.RoomPresenceMessage) Result {
	if msg == nil {
		return Result{}
	}
	names := playerNames(msg.Players)
	if names == "" {
		return Result{Text: "The room is quiet."}
	}
	return Result{Text: "Here: " + names}
}

func renderDialog(msg *messages.DialogMessage) Result {
	if msg == nil {
		return Result{}
	}
	var b strings.Builder
	name := Sanitize(msg.NPCName)
	body := Sanitize(msg.NPCText)
	if body == "" {
		body = Sanitize(msg.Message)
	}
	if name != "" {
		fmt.Fprintf(&b, "[%s] %s", name, body)
	} else {
		b.WriteString(body)
	}
	if len(msg.Options) > 0 {
		b.WriteString("\n")
		for _, opt := range msg.Options {
			fmt.Fprintf(&b, "\n%d. %s", opt.Index, Sanitize(opt.Text))
		}
		b.WriteString("\n\nEnter a number to respond:")
	}
	return Result{Text: b.String()}
}

func renderShop(msg *messages.ShopMessage) Result {
	if msg == nil {
		return Result{}
	}
	var b strings.Builder
	if strings.TrimSpace(msg.Message) != "" {
		b.WriteString(Sanitize(msg.Message))
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, "%s — %d gold", Sanitize(msg.MerchantName), msg.Gold)
	for i, row := range msg.Stock {
		qty := "unlimited"
		if row.Quantity >= 0 {
			qty = fmt.Sprintf("%d", row.Quantity)
		}
		fmt.Fprintf(&b, "\n%d. %s — %d gold (%s)", i+1, Sanitize(row.Name), row.Price, qty)
	}
	if len(msg.Stock) > 0 {
		b.WriteString("\nBuy with: buy <number>")
	}
	return Result{Text: strings.TrimRight(b.String(), "\n")}
}

func renderQuest(msg *messages.QuestUpdateMessage) Result {
	if msg == nil {
		return Result{}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Quest %s (%s)", Sanitize(msg.QuestName), Sanitize(msg.Status))
	if strings.TrimSpace(msg.Message) != "" && Sanitize(msg.Message) != Sanitize(msg.QuestName) {
		fmt.Fprintf(&b, "\n%s", Sanitize(msg.Message))
	}
	for _, obj := range msg.Objectives {
		mark := " "
		if obj.Completed {
			mark = "x"
		}
		fmt.Fprintf(&b, "\n[%s] %s %d/%d", mark, Sanitize(obj.Description), obj.Current, obj.Required)
	}
	return Result{Text: b.String()}
}

func renderQuestLog(msg *messages.QuestLogMessage) Result {
	if msg == nil || len(msg.Quests) == 0 {
		return Result{Text: "Quest log is empty."}
	}
	var b strings.Builder
	b.WriteString("Quest log")
	for _, q := range msg.Quests {
		fmt.Fprintf(&b, "\n- %s (%s)", Sanitize(q.QuestName), Sanitize(q.Status))
		if q.ReadyToTurnIn {
			b.WriteString(" ready")
		}
	}
	return Result{Text: b.String()}
}

func renderCombatStart(msg *messages.CombatStartMessage) Result {
	if msg == nil {
		return Result{}
	}
	text := strings.TrimRight(Sanitize(msg.Message), "\n")
	roster := rosterLine(append(append([]messages.CombatantView{}, msg.Players...), msg.Enemies...))
	if text == "" {
		text = "Combat begins."
	}
	if roster != "" && !strings.Contains(text, roster) {
		text += "\n" + roster
	}
	return Result{Text: text}
}

func renderCombatEnd(msg *messages.CombatEndMessage) Result {
	if msg == nil {
		return Result{}
	}
	text := strings.TrimRight(Sanitize(msg.Message), "\n")
	if text == "" {
		text = "Combat ends (" + Sanitize(msg.Outcome) + ")."
	}
	if msg.LevelUp != nil && msg.LevelUp.NewLevel > 0 && !strings.Contains(text, "level") {
		text += fmt.Sprintf("\nLevel %d.", msg.LevelUp.NewLevel)
	}
	return Result{Text: text}
}

func rosterLine(views []messages.CombatantView) string {
	var parts []string
	for _, c := range views {
		if strings.TrimSpace(c.Name) == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s %d/%d", Sanitize(c.Name), c.HP, c.MaxHP))
	}
	return strings.Join(parts, ", ")
}

func textOnly(message, actor string, round int) Result {
	text := strings.TrimSpace(Sanitize(message))
	if text == "" && actor != "" {
		text = fmt.Sprintf("Round %d — %s", round, Sanitize(actor))
	}
	if text == "" {
		return Result{}
	}
	return Result{Text: text}
}

func renderLevel(msg *messages.LevelUpMessage) Result {
	if msg == nil {
		return Result{}
	}
	var b strings.Builder
	if strings.TrimSpace(msg.Message) != "" {
		b.WriteString(strings.TrimRight(Sanitize(msg.Message), "\n"))
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, "Level %d → %d. HP +%d.", msg.OldLevel, msg.NewLevel, msg.HPGained)
	if msg.ManaGained > 0 {
		fmt.Fprintf(&b, " Mana +%d.", msg.ManaGained)
	}
	if msg.UnspentAttributePoints > 0 {
		fmt.Fprintf(&b, " Unspent points: %d.", msg.UnspentAttributePoints)
	}
	return Result{Text: strings.TrimRight(b.String(), "\n")}
}

func renderStats(msg *messages.CharacterUpdateMessage) Result {
	if msg == nil {
		return Result{}
	}
	line := fmt.Sprintf("HP %d/%d  Mana %d/%d  Level %d  Gold %d", msg.CurrentHitPoints, msg.MaxHitPoints, msg.CurrentMana, msg.MaxMana, msg.Level, msg.Gold)
	if msg.InCombat {
		line += "  (in combat)"
	}
	if strings.TrimSpace(msg.Message) != "" {
		return Result{Text: Sanitize(msg.Message) + "\n" + line}
	}
	return Result{Text: line}
}

func renderSelected(msg *messages.CharacterSelected) Result {
	if msg == nil {
		return Result{}
	}
	var b strings.Builder
	if msg.Character != nil && msg.Character.Name != "" {
		fmt.Fprintf(&b, "%s, level %d.", Sanitize(msg.Character.Name), msg.Character.Level)
	}
	if strings.TrimSpace(msg.Message) != "" {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(Sanitize(msg.Message))
	}
	return Result{Text: b.String()}
}

func renderInventory(msg *messages.InventoryUpdateMessage) Result {
	if msg == nil {
		return Result{}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Gold: %d", msg.Gold)
	if names := inventoryNames(msg.Inventory); names != "" {
		fmt.Fprintf(&b, "\nCarrying: %s", names)
	}
	if worn := equippedNames(msg.EquippedItems); worn != "" {
		fmt.Fprintf(&b, "\nWearing: %s", worn)
	}
	return Result{Text: b.String()}
}

func inventoryNames(inv any) string {
	var list []*items.Item
	switch v := inv.(type) {
	case *items.Inventory:
		if v != nil {
			list = v.Items
		}
	case items.Inventory:
		list = v.Items
	default:
		return ""
	}
	var names []string
	for _, item := range list {
		if item == nil || strings.TrimSpace(item.Name) == "" {
			continue
		}
		name := Sanitize(item.Name)
		if item.Quantity > 1 {
			name = fmt.Sprintf("%s x%d", name, item.Quantity)
		}
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}

func equippedNames(equipped any) string {
	switch v := equipped.(type) {
	case map[items.ItemSlot]*items.Item:
		var names []string
		for slot, item := range v {
			if item == nil || strings.TrimSpace(item.Name) == "" {
				continue
			}
			names = append(names, fmt.Sprintf("%s %s", slot, Sanitize(item.Name)))
		}
		return strings.Join(names, ", ")
	default:
		return ""
	}
}

func renderInvite(msg *messages.PartyInviteMessage) Result {
	if msg == nil || !msg.Pending {
		return Result{}
	}
	name := Sanitize(msg.InviterName)
	if name == "" {
		name = "Someone"
	}
	return Result{Text: name + " invites you to a party. Type accept or decline."}
}

func renderFriends(msg *messages.FriendsMessage) Result {
	if msg == nil || len(msg.Friends) == 0 {
		return Result{Text: "Friends: none."}
	}
	var names []string
	for _, f := range msg.Friends {
		name := Sanitize(f.Name)
		if f.Online {
			name += " (online)"
		}
		names = append(names, name)
	}
	return Result{Text: "Friends: " + strings.Join(names, ", ")}
}

func renderParty(msg *messages.PartyMessage) Result {
	if msg == nil || !msg.InParty || len(msg.Members) == 0 {
		return Result{Text: "You are not in a party."}
	}
	var names []string
	for _, m := range msg.Members {
		name := Sanitize(m.Name)
		if m.IsLeader {
			name += " (leader)"
		}
		names = append(names, name)
	}
	title := "Party"
	if msg.PartyName != "" {
		title = Sanitize(msg.PartyName)
	}
	return Result{Text: title + ": " + strings.Join(names, ", ")}
}

func renderRecipes(msg *messages.RecipesMessage) Result {
	if msg == nil {
		return Result{Text: "No recipes."}
	}
	if len(msg.Recipes) == 0 {
		text := strings.TrimSpace(Sanitize(msg.Message))
		if text == "" {
			text = "No recipes."
		}
		return Result{Text: text}
	}
	var b strings.Builder
	b.WriteString("Recipes")
	for i, row := range msg.Recipes {
		fmt.Fprintf(&b, "\n%d. %s", i+1, Sanitize(row.Name))
	}
	return Result{Text: b.String()}
}
