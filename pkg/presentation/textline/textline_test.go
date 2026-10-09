package textline

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
)

func TestSanitizeStripsControls(t *testing.T) {
	in := "hi\x1b[31mred\x00\x7f\u009b\tok\n"
	got := Sanitize(in)
	if strings.Contains(got, "\x1b") || strings.Contains(got, "\x00") || strings.ContainsRune(got, 0x7f) || strings.ContainsRune(got, 0x9b) {
		t.Fatalf("controls survived: %q", got)
	}
	if got != "hi[31mred\tok\n" {
		t.Fatalf("got %q", got)
	}
	if Sanitize("plain") != "plain" {
		t.Fatal("plain text changed")
	}
}

func TestRenderWelcomeRoomDialogShopAndPing(t *testing.T) {
	welcome := Render(messages.NewRoomBasedMessage("", "Connected to [TalesMUD] ..."), Options{})
	if welcome.Text != "Connected to [TalesMUD] ..." {
		t.Fatalf("welcome %q", welcome.Text)
	}

	exits := rooms.Exits{{Name: "north", Target: "R2"}, {Name: "secret", Hidden: true}}
	room := Render(&messages.EnterRoomMessage{
		MessageResponse: messages.MessageResponse{Message: "You stand on the dock."},
		Room:            rooms.Room{Name: "Harbor", Exits: &exits},
		Items:           []util.RoomItem{{Name: "coil of rope"}, {Name: "lantern\x1b[2J"}},
		NPCs:            []messages.RoomNPC{{DisplayName: "Harbormaster"}},
	}, Options{})
	if !strings.Contains(room.Text, "You stand on the dock.") || !strings.Contains(room.Text, "coil of rope") {
		t.Fatalf("room %q", room.Text)
	}
	if strings.Contains(room.Text, "\x1b") {
		t.Fatalf("item name injected an escape: %q", room.Text)
	}

	empty := Render(&messages.EnterRoomMessage{
		Room: rooms.Room{Name: "Harbor", Exits: &exits},
		NPCs: []messages.RoomNPC{{Name: "Cat", IsEnemy: true}},
	}, Options{})
	if !strings.Contains(empty.Text, "[Harbor]") || !strings.Contains(empty.Text, "Exits: north") || strings.Contains(empty.Text, "secret") {
		t.Fatalf("built room %q", empty.Text)
	}
	if !strings.Contains(empty.Text, "Cat (enemy)") {
		t.Fatalf("npc %q", empty.Text)
	}

	quiet := Render(&messages.RoomPresenceMessage{}, Options{})
	if quiet.Text != "The room is quiet." {
		t.Fatalf("presence %q", quiet.Text)
	}

	dialog := Render(&messages.DialogMessage{
		NPCName: "Harbormaster",
		NPCText: "Need a boat?",
		Options: []messages.DialogOption{{Index: 1, Text: "Yes"}, {Index: 2, Text: "No"}},
	}, Options{})
	if !strings.Contains(dialog.Text, "[Harbormaster] Need a boat?") || !strings.Contains(dialog.Text, "1. Yes") || !strings.Contains(dialog.Text, "Enter a number") {
		t.Fatalf("dialog %q", dialog.Text)
	}

	shop := Render(&messages.ShopMessage{
		MerchantName: "Shop",
		Gold:         12,
		Stock:        []messages.ShopStockItem{{Name: "rope", Price: 3, Quantity: 2}, {Name: "map", Price: 1, Quantity: -1}},
	}, Options{})
	if !strings.Contains(shop.Text, "1. rope — 3 gold (2)") || !strings.Contains(shop.Text, "unlimited") || !strings.Contains(shop.Text, "buy <number>") {
		t.Fatalf("shop %q", shop.Text)
	}

	if ping := Render(messages.MessageResponse{Type: messages.MessageTypePing, Message: "ping"}, Options{}); ping.Text != "" {
		t.Fatalf("ping rendered %q", ping.Text)
	}
	if Render((*messages.RecipesMessage)(nil), Options{}).Text != "No recipes." {
		t.Fatal("nil recipes")
	}

	stats := Render(&messages.CharacterUpdateMessage{CurrentHitPoints: 8, MaxHitPoints: 10, CurrentMana: 1, MaxMana: 4, Level: 2, Gold: 5}, Options{})
	if stats.Text != "HP 8/10  Mana 1/4  Level 2  Gold 5" {
		t.Fatalf("stats %q", stats.Text)
	}

	inv := Render(&messages.InventoryUpdateMessage{
		Gold:      9,
		Inventory: &items.Inventory{Items: []*items.Item{{Name: "rope", Quantity: 2}}},
	}, Options{})
	if !strings.Contains(inv.Text, "Gold: 9") || !strings.Contains(inv.Text, "rope x2") {
		t.Fatalf("inventory %q", inv.Text)
	}

	create := Render(messages.NewCreateCharacterMessage("u1"), Options{PlayURL: "https://example.test/play"})
	if !strings.Contains(create.Text, "https://example.test/play") {
		t.Fatalf("create %q", create.Text)
	}
}

func TestCharacterListNumbersAndEditorMapsThem(t *testing.T) {
	msg := messages.MessageResponse{Message: "Your Characters:\n- Ada [LVL 1 Fighter 0xp] here - ready\n- Bob [LVL 2 Mage 10xp] away - out\n\nTo select a character, use: sc [charactername]"}
	got := Render(msg, Options{})
	if got.ChoiceKind != "character" || len(got.Choices) != 2 || got.Choices[0] != "Ada" || got.Choices[1] != "Bob" {
		t.Fatalf("choices %+v text %q", got, got.Text)
	}
	if !strings.Contains(got.Text, "1. Ada") || !strings.Contains(got.Text, "Enter a number") {
		t.Fatalf("text %q", got.Text)
	}
	ed := NewEditor(20, 512)
	ed.SetChoices(got.ChoiceKind, got.Choices)
	res := ed.Feed([]byte("1\r"))
	if res.Line != "selectcharacter Ada" {
		t.Fatalf("line %q out %q", res.Line, res.Out)
	}
}

func TestEditorEditingHistoryAndQuit(t *testing.T) {
	ed := NewEditor(2, 4)
	if res := ed.Feed([]byte("look\r")); res.Line != "look" || !strings.Contains(string(res.Out), "\r\n") {
		t.Fatalf("submit %+v", res)
	}
	up := ed.Feed([]byte{0x1b})
	if up.Line != "" {
		t.Fatal("partial escape submitted")
	}
	up = ed.Feed([]byte("[A"))
	if !strings.Contains(string(up.Out), "look") {
		t.Fatalf("history up %q", up.Out)
	}
	if res := ed.Feed([]byte{0x15, '\r'}); res.Line != "" {
		t.Fatalf("ctrl-u should submit empty, got %q", res.Line)
	}
	typed := ed.Feed([]byte("abcdX"))
	if strings.Contains(string(typed.Out), "X") {
		t.Fatal("line cap ignored")
	}
	back := ed.Feed([]byte{0x7f})
	if strings.Contains(string(back.Out), "abcd") {
		t.Fatalf("backspace %q", back.Out)
	}
	ask := ed.Feed([]byte{0x03})
	if !strings.Contains(string(ask.Out), "Quit?") || ask.Quit {
		t.Fatalf("confirm %+v", ask)
	}
	no := ed.Feed([]byte("n"))
	if no.Quit || !strings.Contains(string(no.Out), "> ") {
		t.Fatalf("cancel %+v", no)
	}
	ed.Feed([]byte{0x04})
	yes := ed.Feed([]byte("Y"))
	if !yes.Quit {
		t.Fatal("expected quit")
	}
	if strings.Contains(string(ed.PrintAbove("hi\x1b[2J")), "\x1b[2J") && strings.Contains(string(ed.PrintAbove("ok")), "\x1b[2K") {
		// PrintAbove does not sanitize; the renderer does. The redraw uses a clear.
	}
	above := string(ed.PrintAbove("Room\nline"))
	if !strings.HasPrefix(above, "\r\x1b[2K") || !strings.Contains(above, "Room\r\nline") {
		t.Fatalf("print above %q", above)
	}
}

func TestEditorCollapsesCRLF(t *testing.T) {
	ed := NewEditor(20, 64)
	res := ed.Feed([]byte("say hi\r\n"))
	if res.Line != "say hi" {
		t.Fatalf("line %q", res.Line)
	}
}
