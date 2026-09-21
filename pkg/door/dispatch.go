package door

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/daily"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
)

// dispatchPack handles a key when the current screen comes from the world pack.
// It returns false when this screen is not pack-driven.
func (h *Hub) dispatchPack(user *entities.User, sess *session, key string) bool {
	if h.pack == nil || len(h.pack.Screens) == 0 {
		return false
	}
	scr, ok := h.pack.Screens[sess.screen]
	if !ok {
		return false
	}
	hk, ok := hotkeyFor(scr, key)
	if !ok {
		sess.notice = "That key does nothing here."
		return true
	}
	switch hk.Action {
	case "go_screen":
		h.goScreen(user, sess, hk.Screen)
	case "fight":
		h.engage(user, sess, "a")
	case "run":
		h.engage(user, sess, "r")
	case "heal_full":
		h.healerKey(user, sess, "h")
	case "deposit":
		sess.purpose = "deposit"
		sess.screen = screenAmount
		sess.inputMode = "line"
	case "withdraw":
		sess.purpose = "withdraw"
		sess.screen = screenAmount
		sess.inputMode = "line"
	case "buy":
		sess.purpose = "buy"
		sess.screen = screenAmount
		sess.inputMode = "line"
		sess.notice = "Enter the item number, or x to step back."
	case "sell":
		h.sellGear(user, sess)
	case "quit":
		sess.notice = "Close the window to leave. Coin in the vault stays put."
	case "drink", "inn_draught":
		h.innKey(user, sess, "d")
	case "go_back", "return":
		h.goBack(sess)
	case "help":
		sess.notice = helpLine(scr)
	default:
		sess.notice = "That action is not ready yet."
	}
	return true
}

func (h *Hub) goScreen(user *entities.User, sess *session, id string) {
	if id == "" {
		id = h.homeScreen()
	}
	if id == "forest" {
		if ch, err := h.load(user); err == nil && ch != nil {
			if res, err := h.daily.Get(ch.ID, daily.ForestFightsKey, h.allowance()); err == nil && res.Remaining <= 0 {
				sess.notice = "The Ashwood gate will not open until tomorrow."
				return
			}
		}
	}
	sess.screen = id
	sess.inputMode = "hotkey"
}

func (h *Hub) engage(user *entities.User, sess *session, key string) {
	if sess.fight == nil || sess.fight.Done {
		h.enterForest(user, sess)
		if sess.fight == nil || sess.fight.Done {
			return
		}
	}
	h.fightKey(user, sess, key)
	if _, ok := h.pack.Screens["forest"]; ok {
		sess.screen = "forest"
		sess.inputMode = "hotkey"
	}
}

func (h *Hub) sellGear(user *entities.User, sess *session) {
	_, err := h.mutate(user, func(ch *characters.Character) error {
		if ch.Door == nil {
			sess.notice = "You have nothing priced to sell."
			return nil
		}
		if ch.Door.WeaponID != "" {
			item := h.pack.WeaponByID(ch.Door.WeaponID)
			refund := int64(item.Price / 2)
			ch.Gold += refund
			ch.Door.WeaponID = ""
			sess.notice = fmt.Sprintf("Sold %s for %d coin.", item.Name, refund)
			return nil
		}
		if ch.Door.ArmorID != "" {
			item := h.pack.ArmorByID(ch.Door.ArmorID)
			refund := int64(item.Price / 2)
			ch.Gold += refund
			ch.Door.ArmorID = ""
			sess.notice = fmt.Sprintf("Sold %s for %d coin.", item.Name, refund)
			return nil
		}
		sess.notice = "You have nothing priced to sell."
		return nil
	})
	if err != nil {
		sess.notice = "The smith will not take it. Try again."
	}
}

func hotkeyFor(scr *Screen, key string) (Hotkey, bool) {
	if scr == nil || len(scr.Hotkeys) == 0 {
		return Hotkey{}, false
	}
	if hk, ok := scr.Hotkeys[key]; ok {
		return hk, true
	}
	if hk, ok := scr.Hotkeys[strings.ToUpper(key)]; ok {
		return hk, true
	}
	return Hotkey{}, false
}

func helpLine(scr *Screen) string {
	if scr == nil {
		return "Press the letter in parentheses."
	}
	parts := make([]string, 0, len(scr.Hotkeys))
	for key, hk := range scr.Hotkeys {
		label := hk.Action
		if hk.Screen != "" {
			label = hk.Screen
		}
		parts = append(parts, key+" "+label)
	}
	return "Keys: " + strings.Join(parts, ", ")
}
