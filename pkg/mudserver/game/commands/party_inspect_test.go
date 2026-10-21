package commands_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/portraits"
	"github.com/talesmud/talesmud/pkg/service"
)

func TestPartyInspect(t *testing.T) {
	t.Run("party member", func(t *testing.T) {
		g, facade := newSocialTestGame(t)
		callerUser, caller := storeSocialPlayer(t, facade, "user-caller", "ref-caller", "char-caller", "Aryn", "room-1")
		targetUser, target := storeSocialPlayer(t, facade, "user-target", "ref-target", "char-target", "Bran", "room-1")

		callerChest := &items.Item{
			Entity:     &entities.Entity{ID: "caller-chest"},
			Name:       "Plain Shirt",
			Type:       items.ItemTypeArmor,
			Slot:       items.ItemSlotChest,
			Attributes: map[string]interface{}{"defense": float64(6)},
		}
		dressInspectCharacter(caller, callerChest)
		caller.EquippedItems[items.ItemSlotPurse] = &items.Item{
			Entity: &entities.Entity{ID: "caller-purse"},
			Name:   "Hidden Purse",
			Slot:   items.ItemSlotPurse,
		}

		jerkin := &items.Item{
			Entity:        &entities.Entity{ID: "jerkin-1"},
			Name:          "Studded Jerkin",
			Type:          items.ItemTypeArmor,
			Slot:          items.ItemSlotChest,
			Attributes:    map[string]interface{}{"defense": float64(8)},
			Durability:    4,
			MaxDurability: 4,
		}
		dressInspectCharacter(target, jerkin)
		target.EquippedItems[items.ItemSlotPurse] = &items.Item{
			Entity: &entities.Entity{ID: "purse-1"},
			Name:   "Hidden Purse",
			Slot:   items.ItemSlotPurse,
		}
		jerkin.LockedBy = "secret-locker"
		jerkin.CreatedBy = "secret-creator"
		jerkin.Items = items.Items{{Entity: &entities.Entity{ID: "stash-1"}, Name: "Secret Stash"}}
		if err := facade.CharactersService().Update(target.ID, target); err != nil {
			t.Fatalf("update target: %v", err)
		}
		inspectParty(t, facade, caller.ID, target.ID)

		g.ConnectUserSession(callerUser)
		g.SetUserSessionCharacter(callerUser, caller)
		g.ConnectUserSession(targetUser)
		g.SetUserSessionCharacter(targetUser, target)

		if !(&commands.PartyInspectCommand{}).Execute(g, &messages.Message{
			FromUser: callerUser, Character: caller, Data: "partyinspect char-target",
		}) {
			t.Fatal("partyinspect did not handle the message")
		}
		frame := onePartyInspect(t, drainSocialMessages(g.SendMessage()))
		assertInspectDoll(t, frame, callerUser.ID, target, 1, 8)
		chest := frame.Equipped[items.ItemSlotChest]
		if chest == nil || chest.Name != "Studded Jerkin" || chest.ID != "jerkin-1" {
			t.Fatalf("equipped chest = %#v", chest)
		}
		if chest.LockedBy != "" || chest.CreatedBy != "" || len(chest.Items) != 0 {
			t.Fatalf("chest leaked private fields: %#v", chest)
		}
		if _, ok := frame.Equipped[items.ItemSlotPurse]; ok {
			t.Fatal("purse is not a doll slot")
		}
		if target.CurrentHitPoints != 17 || target.EquippedItems[items.ItemSlotChest].Durability != 4 {
			t.Fatal("inspect mutated the target")
		}

		if !(&commands.PartyInspectCommand{}).Execute(g, &messages.Message{
			FromUser: callerUser, Character: caller, Data: "partyinspect bRaN",
		}) {
			t.Fatal("name inspect did not handle the message")
		}
		byName := onePartyInspect(t, drainSocialMessages(g.SendMessage()))
		if byName.ID != target.ID || byName.Equipped[items.ItemSlotChest] == nil {
			t.Fatalf("name inspect = %#v", byName)
		}

		if !(&commands.PartyInspectCommand{}).Execute(g, &messages.Message{
			FromUser: callerUser, Character: caller, Data: "partyinspect char-caller",
		}) {
			t.Fatal("self inspect did not handle the message")
		}
		self := onePartyInspect(t, drainSocialMessages(g.SendMessage()))
		assertInspectDoll(t, self, callerUser.ID, caller, 1, 6)
		worn := self.Equipped[items.ItemSlotChest]
		if worn == nil || worn.ID != callerChest.ID || callerChest.Durability != 0 || callerChest.MaxDurability != 0 {
			t.Fatalf("self inspect mutated gear: %+v", worn)
		}
	})

	t.Run("party member in another room", func(t *testing.T) {
		g, facade := newSocialTestGame(t)
		callerUser, caller := storeSocialPlayer(t, facade, "user-caller", "ref-caller", "char-caller", "Aryn", "room-1")
		otherUser, other := storeSocialPlayer(t, facade, "user-other", "ref-other", "char-other", "Cyn", "room-2")
		inspectParty(t, facade, caller.ID, other.ID)
		g.ConnectUserSession(callerUser)
		g.SetUserSessionCharacter(callerUser, caller)
		g.ConnectUserSession(otherUser)
		g.SetUserSessionCharacter(otherUser, other)

		(&commands.PartyInspectCommand{}).Execute(g, &messages.Message{
			FromUser: callerUser, Character: caller, Data: "partyinspect char-other",
		})
		frame := onePartyInspect(t, drainSocialMessages(g.SendMessage()))
		if frame.ID != other.ID {
			t.Fatalf("frame id %q", frame.ID)
		}
	})

	t.Run("not in party", func(t *testing.T) {
		g, facade := newSocialTestGame(t)
		callerUser, caller := storeSocialPlayer(t, facade, "user-caller", "ref-caller", "char-caller", "Aryn", "room-1")
		otherUser, other := storeSocialPlayer(t, facade, "user-other", "ref-other", "char-other", "Cyn", "room-1")
		g.ConnectUserSession(callerUser)
		g.SetUserSessionCharacter(callerUser, caller)
		g.ConnectUserSession(otherUser)
		g.SetUserSessionCharacter(otherUser, other)

		if !(&commands.PartyInspectCommand{}).Execute(g, &messages.Message{
			FromUser: callerUser, Character: caller, Data: "partyinspect char-other",
		}) {
			t.Fatal("partyinspect did not handle the message")
		}
		out := drainSocialMessages(g.SendMessage())
		if len(out) != 1 {
			t.Fatalf("expected one text reply, got %d: %#v", len(out), out)
		}
		if _, ok := out[0].(*messages.PartyInspectMessage); ok {
			t.Fatal("other room sent a partyInspect frame")
		}
		text := socialReplyText(out[0])
		if text == "" || strings.Contains(text, "partyInspect") {
			t.Fatalf("text reply = %q", text)
		}
		raw, err := json.Marshal(out[0])
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), "partyInspect") {
			t.Fatalf("frame leaked: %s", raw)
		}
	})

	t.Run("examine and inspect stay prose", func(t *testing.T) {
		g, facade := newSocialTestGame(t)
		user, ch := storeSocialPlayer(t, facade, "user-caller", "ref-caller", "char-caller", "Aryn", "room-1")
		_, target := storeSocialPlayer(t, facade, "user-target", "ref-target", "char-target", "Bran", "room-1")
		dressInspectCharacter(target, &items.Item{
			Entity:        &entities.Entity{ID: "jerkin-1"},
			Name:          "Studded Jerkin",
			Type:          items.ItemTypeArmor,
			Slot:          items.ItemSlotChest,
			Attributes:    map[string]interface{}{"defense": float64(8)},
			Durability:    4,
			MaxDurability: 4,
		})
		if err := facade.CharactersService().Update(target.ID, target); err != nil {
			t.Fatalf("update target: %v", err)
		}
		g.ConnectUserSession(user)
		g.SetUserSessionCharacter(user, ch)

		for _, verb := range []string{"examine", "inspect"} {
			(&commands.ExamineCommand{}).Execute(g, &messages.Message{
				FromUser: user, Character: ch, Data: verb + " Bran",
			})
			out := drainSocialMessages(g.SendMessage())
			if len(out) == 0 {
				t.Fatalf("%s sent nothing", verb)
			}
			for _, msg := range out {
				if _, ok := msg.(*messages.PartyInspectMessage); ok {
					t.Fatalf("%s returned a partyInspect frame", verb)
				}
				raw, err := json.Marshal(msg)
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(string(raw), "partyInspect") || strings.Contains(string(raw), "Studded Jerkin") {
					t.Fatalf("%s leaked the doll: %s", verb, raw)
				}
			}
		}
	})
}

func inspectParty(t *testing.T, facade service.Facade, ids ...string) {
	t.Helper()
	if _, err := facade.PartiesService().CreateParty(&service.CreatePartyDTO{Name: "Inspectors", Characters: ids}); err != nil {
		t.Fatalf("create party: %v", err)
	}
}

func dressInspectCharacter(ch *characters.Character, chest *items.Item) {
	ch.Race = characters.RaceHuman
	ch.Class = characters.ClassWarrior
	ch.Level = 6
	ch.CurrentHitPoints = 17
	ch.MaxHitPoints = 40
	ch.CurrentMana = 2
	ch.MaxMana = 11
	ch.Gold = 777
	ch.XP = 55
	ch.Attributes = characters.Attributes{
		{Name: "Strength", Short: "STR", Value: 10},
		{Name: "Dexterity", Short: "DEX", Value: 10},
		{Name: "Constitution", Short: "CON", Value: 10},
		{Name: "Intelligence", Short: "INT", Value: 10},
		{Name: "Wisdom", Short: "WIS", Value: 10},
	}
	ch.FriendIDs = []string{"friend-secret"}
	ch.Flags = map[string]interface{}{"secretFlag": true}
	ch.Inventory.Items = []*items.Item{{
		Entity: &entities.Entity{ID: "bag-" + ch.ID},
		Name:   "Secret Satchel",
		Slot:   items.ItemSlotInventory,
	}}
	if ch.EquippedItems == nil {
		ch.EquippedItems = map[items.ItemSlot]*items.Item{}
	}
	ch.EquippedItems[items.ItemSlotChest] = chest
}

func onePartyInspect(t *testing.T, out []interface{}) *messages.PartyInspectMessage {
	t.Helper()
	if len(out) != 1 {
		t.Fatalf("expected one partyInspect message, got %d: %#v", len(out), out)
	}
	frame, ok := out[0].(*messages.PartyInspectMessage)
	if !ok || frame.Type != messages.MessageTypePartyInspect {
		t.Fatalf("expected partyInspect, got %T %#v", out[0], out[0])
	}
	if frame.GetAudience() != messages.MessageAudienceOrigin {
		t.Fatalf("audience %v", frame.GetAudience())
	}
	return frame
}

func assertInspectDoll(t *testing.T, frame *messages.PartyInspectMessage, userID string, ch *characters.Character, attack, defense int32) {
	t.Helper()
	if frame.GetAudienceID() != userID {
		t.Fatalf("audience id %q", frame.GetAudienceID())
	}
	raw, err := json.Marshal(frame)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{
		"type": true, "username": true, "message": true,
		"id": true, "name": true, "class": true, "level": true, "portrait": true,
		"currentHitPoints": true, "maxHitPoints": true, "currentMana": true, "maxMana": true,
		"attackPower": true, "defense": true, "equipped": true,
	}
	for key := range payload {
		if !allowed[key] {
			t.Fatalf("unexpected field %s in %s", key, raw)
		}
	}
	for _, key := range []string{"id", "name", "class", "level", "portrait", "currentHitPoints", "maxHitPoints", "currentMana", "maxMana", "attackPower", "defense", "equipped"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("missing %s in %s", key, raw)
		}
	}
	equipped := strings.TrimSpace(string(payload["equipped"]))
	if equipped == "null" || !strings.HasPrefix(equipped, "{") {
		t.Fatalf("equipped = %s", equipped)
	}
	for _, leaked := range []string{"Secret Satchel", "Hidden Purse", "friend-secret", "secretFlag", `"inventory"`, `"gold"`, `"xp"`} {
		if strings.Contains(string(raw), leaked) {
			t.Fatalf("payload leaked %s: %s", leaked, raw)
		}
	}
	className := strings.TrimSpace(ch.Class.Name)
	if className == "" {
		className = ch.Class.ID
	}
	if frame.ID != ch.ID || frame.Name != ch.Name || frame.Class != className || frame.Level != ch.Level {
		t.Fatalf("identity = %+v", frame)
	}
	if frame.CurrentHitPoints != ch.CurrentHitPoints || frame.MaxHitPoints != ch.MaxHitPoints || frame.CurrentMana != ch.CurrentMana || frame.MaxMana != ch.MaxMana {
		t.Fatalf("vitals = %+v", frame)
	}
	if frame.Portrait != portraits.ForPlayer(ch) {
		t.Fatalf("portrait %q", frame.Portrait)
	}
	if frame.AttackPower != attack || frame.Defense != defense {
		t.Fatalf("attack %d defense %d", frame.AttackPower, frame.Defense)
	}
}

// Leaving a two-member party deletes it; the leave refresh used to panic on the missing row.
func TestPartyLeaveLastPairDoesNotPanic(t *testing.T) {
	g, facade := newSocialTestGame(t)
	aUser, a := storeSocialPlayer(t, facade, "user-a", "ref-a", "char-a", "Aryn", "room-1")
	bUser, b := storeSocialPlayer(t, facade, "user-b", "ref-b", "char-b", "Bran", "room-1")
	inspectParty(t, facade, a.ID, b.ID)
	g.ConnectUserSession(aUser)
	g.SetUserSessionCharacter(aUser, a)
	g.ConnectUserSession(bUser)
	g.SetUserSessionCharacter(bUser, b)

	(&commands.PartyCommand{}).Execute(g, &messages.Message{FromUser: bUser, Character: b, Data: "party leave"})
	drainSocialMessages(g.SendMessage())
	(&commands.PartyCommand{}).Execute(g, &messages.Message{FromUser: aUser, Character: a, Data: "party leave"})
	drainSocialMessages(g.SendMessage())
	if party, err := facade.PartiesService().FindByCharacterID(a.ID); err == nil && party != nil {
		t.Fatalf("party still exists: %#v", party)
	}
}
