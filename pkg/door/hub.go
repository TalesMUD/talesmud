package door

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/talesmud/talesmud/pkg/daily"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
	"github.com/talesmud/talesmud/pkg/service"
)

const (
	screenName   = "name"
	screenSex    = "sex"
	screenTrack  = "track"
	screenTown   = "town"
	screenFight  = "fight"
	screenResult = "result"
	screenHealer = "healer"
	screenBank   = "bank"
	screenAmount = "amount"
	screenStats  = "stats"
	screenBoard  = "board"
	screenNews   = "news"
	screenInn    = "inn"
	screenArmory = "armory"

	innKey       = "inn_draught"
	starterGold  = 50
	doorLevelCap = 12
)

var warriorName = regexp.MustCompile(`^[A-Za-z][A-Za-z' -]{0,22}[A-Za-z]$`)

// Roller supplies combat dice. Tests install a fixed sequence.
type Roller interface {
	Intn(n int) int
}

type mathRoll struct{}

func (mathRoll) Intn(n int) int {
	if n <= 1 {
		return 0
	}
	return rand.Intn(n)
}

// Hub is the door menu state machine for one process.
type Hub struct {
	facade   service.Facade
	daily    *daily.Store
	pack     *Pack
	fights   int
	loc      *time.Location
	now      func() time.Time
	roll     Roller
	mu       sync.Mutex
	sessions map[string]*session
}

type session struct {
	screen    string
	inputMode string
	notice    string
	purpose   string
	draftName string
	draftSex  string
	returnTo  string
	fight     *fight
}

// NewHub builds the Ashmarket session router.
func NewHub(facade service.Facade, resources *daily.Store, pack *Pack, fights int, loc *time.Location) *Hub {
	if pack == nil {
		pack = DefaultPack()
	}
	if fights <= 0 {
		fights = pack.DailyFights
	}
	if loc == nil {
		loc = time.UTC
	}
	return &Hub{
		facade:   facade,
		daily:    resources,
		pack:     pack,
		fights:   fights,
		loc:      loc,
		now:      time.Now,
		roll:     mathRoll{},
		sessions: map[string]*session{},
	}
}

// SetRoller replaces the dice source.
func (h *Hub) SetRoller(r Roller) {
	if r != nil {
		h.roll = r
	}
}

// SetNow replaces the clock used on the status line.
func (h *Hub) SetNow(now func() time.Time) {
	if now != nil {
		h.now = now
	}
}

// Active reports that this hook owns the websocket session.
func (h *Hub) Active() bool { return true }

// OnConnect opens the town, or warrior creation when the account has no character.
func (h *Hub) OnConnect(user *entities.User, send func(any)) {
	sess := &session{screen: h.homeScreen(), inputMode: "hotkey"}
	h.mu.Lock()
	h.sessions[user.ID] = sess
	h.mu.Unlock()

	chars, err := h.facade.CharactersService().FindAllForUser(user.ID)
	if err != nil || len(chars) == 0 {
		sess.screen = screenName
		sess.inputMode = "line"
	} else {
		if user.LastCharacter == "" {
			user.LastCharacter = chars[0].ID
			_ = h.facade.UsersService().Update(user.RefID, user)
		}
		h.ensureNewDay(user, sess)
	}
	h.render(user, sess, send)
}

// ensureNewDay applies the LORD-style calendar rollover: full HP once per
// Europe/Berlin day, and touches daily budgets so forest/inn refill on login.
// Same-day reconnects do not re-heal.
func (h *Hub) ensureNewDay(user *entities.User, sess *session) {
	if user == nil || h.daily == nil {
		return
	}
	ch, err := h.load(user)
	if err != nil || ch == nil {
		return
	}
	today := h.now().In(h.loc).Format("2006-01-02")
	if ch.Door != nil && ch.Door.LastDay == today {
		return
	}
	healed := false
	_, err = h.mutate(user, func(loaded *characters.Character) error {
		if loaded.Door == nil {
			loaded.Door = &characters.DoorProfile{}
		}
		if loaded.Door.LastDay == today {
			return nil
		}
		if h.healOnNewDay() {
			loaded.CurrentHitPoints = loaded.MaxHitPoints
			healed = true
		}
		loaded.Door.LastDay = today
		return nil
	})
	if err != nil {
		return
	}
	if healed && sess != nil {
		sess.notice = "A new day dawns in Ashmarket. Your wounds close with the morning light."
	}
	_, _ = h.daily.Get(ch.ID, daily.ForestFightsKey, h.allowance())
	_, _ = h.daily.Get(ch.ID, innKey, 1)
}

func (h *Hub) healOnNewDay() bool {
	if h.pack == nil {
		return true
	}
	return h.pack.HealOnNewDay
}

// OnInput handles one hotkey or one finished line. It always consumes input in door mode.
func (h *Hub) OnInput(user *entities.User, text string, send func(any)) bool {
	sess := h.session(user.ID)
	if sess == nil {
		h.OnConnect(user, send)
		sess = h.session(user.ID)
		if sess == nil {
			return true
		}
	}
	text = strings.TrimRight(text, "\r\n")
	if sess.inputMode != "line" && strings.TrimSpace(text) == "" {
		return true
	}
	sess.notice = ""
	if sess.inputMode == "line" {
		h.handleLine(user, sess, strings.TrimSpace(text))
	} else {
		h.handleKey(user, sess, strings.ToLower(strings.TrimSpace(text)))
	}
	h.render(user, sess, send)
	return true
}

// OnDisconnect drops the in-memory menu. Character rows stay in SQLite.
func (h *Hub) OnDisconnect(user *entities.User) {
	if user == nil {
		return
	}
	h.mu.Lock()
	delete(h.sessions, user.ID)
	h.mu.Unlock()
}

func (h *Hub) session(id string) *session {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.sessions[id]
}

func (h *Hub) allowance() int {
	if h.fights > 0 {
		return h.fights
	}
	return h.pack.DailyFights
}

func (h *Hub) handleLine(user *entities.User, sess *session, text string) {
	switch sess.screen {
	case screenName:
		name, ok := cleanName(text)
		if !ok {
			sess.notice = "Use 2 to 24 letters. Apostrophes, spaces, and hyphens are fine."
			return
		}
		sess.draftName = name
		sess.screen = screenSex
		sess.inputMode = "hotkey"
	case screenAmount:
		if sess.purpose == "buy" {
			h.applyBuy(user, sess, text)
		} else {
			h.applyAmount(user, sess, text)
		}
	default:
		sess.inputMode = "hotkey"
		sess.screen = h.homeScreen()
	}
}

func (h *Hub) homeScreen() string {
	if h.pack != nil && h.pack.Screens != nil {
		if _, ok := h.pack.Screens["town_square"]; ok {
			return "town_square"
		}
	}
	return screenTown
}

func (h *Hub) handleKey(user *entities.User, sess *session, key string) {
	if sess.screen != screenSex && sess.screen != screenTrack && h.dispatchPack(user, sess, key) {
		return
	}
	if key == "?" {
		sess.notice = "Press the letter in parentheses. One key is enough."
		return
	}
	switch sess.screen {
	case screenSex:
		switch key {
		case "m", "f", "x":
			sess.draftSex = key
			sess.screen = screenTrack
		default:
			sess.notice = "Press (M), (F), or (X)."
		}
	case screenTrack:
		n, err := strconv.Atoi(key)
		if err != nil {
			sess.notice = "Press a number for a path."
			return
		}
		tr, ok := h.pack.TrackByIndex(n)
		if !ok {
			sess.notice = "That path is not on the board."
			return
		}
		if err := h.createWarrior(user, sess.draftName, sess.draftSex, tr); err != nil {
			sess.notice = err.Error()
			return
		}
		sess.screen = h.homeScreen()
	case screenTown:
		h.townKey(user, sess, key)
	case screenFight:
		h.fightKey(user, sess, key)
	case screenResult:
		kind := ""
		if sess.fight != nil {
			kind = sess.fight.Kind
		}
		sess.fight = nil
		if kind == fightMaster {
			if _, ok := h.pack.Screens["trainer"]; ok {
				sess.screen = "trainer"
			} else {
				sess.screen = h.homeScreen()
			}
		} else {
			sess.screen = h.homeScreen()
		}
		sess.inputMode = "hotkey"
	case screenHealer:
		h.healerKey(user, sess, key)
	case screenBank:
		h.bankKey(user, sess, key)
	case screenArmory:
		h.armoryKey(user, sess, key)
	case screenInn:
		h.innKey(user, sess, key)
	case screenStats, screenBoard, screenNews:
		if key == "r" {
			h.goBack(sess)
		}
	default:
		sess.screen = screenTown
	}
}

func (h *Hub) goBack(sess *session) {
	if sess.returnTo != "" {
		sess.screen = sess.returnTo
		sess.returnTo = ""
		sess.inputMode = "hotkey"
		return
	}
	sess.screen = h.homeScreen()
	sess.inputMode = "hotkey"
}

func (h *Hub) townKey(user *entities.User, sess *session, key string) {
	sess.returnTo = ""
	switch key {
	case "f":
		h.enterForest(user, sess)
	case "h":
		sess.screen = screenHealer
	case "b":
		sess.screen = screenBank
	case "a":
		sess.screen = screenArmory
	case "y":
		sess.screen = screenStats
	case "l":
		sess.screen = screenBoard
	case "n":
		sess.screen = screenNews
	case "i":
		sess.screen = screenInn
	case "q":
		sess.notice = "Close the window to leave. Coin in the vault stays put."
	default:
		sess.notice = "That key does nothing on the square. Press ? for the list."
	}
}

func (h *Hub) enterForest(user *entities.User, sess *session) {
	ch, err := h.load(user)
	if err != nil || ch == nil {
		sess.notice = "You have no warrior yet."
		return
	}
	res, err := h.daily.Consume(ch.ID, daily.ForestFightsKey, h.allowance(), 1)
	if err == daily.ErrExhausted {
		sess.notice = "The gate is shut. No walks left until the day turns."
		return
	}
	if err != nil {
		sess.notice = "The gate will not open. Try again."
		return
	}
	spawned := h.spawn(ch.Level)
	spawned.Log = []string{fmt.Sprintf("A %s steps onto the path.", spawned.Name)}
	sess.fight = &spawned
	sess.screen = screenFight
	sess.notice = fmt.Sprintf("Walks left today: %d/%d.", res.Remaining, res.Allowance)
}

func (h *Hub) fightKey(user *entities.User, sess *session, key string) {
	if sess.fight == nil {
		sess.screen = screenTown
		return
	}
	if key == "s" {
		sess.returnTo = screenFight
		sess.screen = screenStats
		return
	}
	if key != "a" && key != "r" {
		sess.notice = "Press (A)ttack, (R)un, or (S)tats."
		return
	}
	_, err := h.mutate(user, func(ch *characters.Character) error {
		f := sess.fight
		var lines []string
		switch key {
		case "a":
			lines = h.playerAttack(ch, f).Lines
		case "r":
			fled, fleeLines := h.tryFlee(ch, f)
			lines = fleeLines
			if fled {
				h.finishFight(ch, f, "fled")
			}
		}
		f.Log = append(f.Log, lines...)
		if !f.Done && f.HP <= 0 {
			h.finishFight(ch, f, "win")
		}
		if !f.Done && ch.CurrentHitPoints <= 0 {
			h.finishFight(ch, f, "lose")
		}
		if len(f.Log) > 12 {
			f.Log = f.Log[len(f.Log)-12:]
		}
		if f.Done {
			sess.screen = screenResult
		}
		return nil
	})
	if err != nil {
		sess.notice = "The fight book slipped. Try the key again."
	}
}

func (h *Hub) finishFight(ch *characters.Character, f *fight, outcome string) {
	if f.Done {
		return
	}
	f.Done = true
	f.Outcome = outcome
	if ch.Door == nil {
		ch.Door = &characters.DoorProfile{}
	}
	switch outcome {
	case "win":
		if f.Kind == fightMaster {
			h.finishMasterWin(ch, f)
			return
		}
		ch.Gold += f.Gold
		ch.XP += f.XP
		ch.Door.Wins++
		if h.roll.Intn(100) < 8 {
			ch.Door.Gems++
			f.Log = append(f.Log, "A dull gem catches in the moss.")
		}
		// Forest awards XP/gold only. Level-ups are gated behind the trainer duel.
		f.Log = append(f.Log, fmt.Sprintf("You take %d coin and %d experience.", f.Gold, f.XP))
		if levels, _ := leveling.CheckLevelUp(ch); levels > 0 {
			f.Log = append(f.Log, "You feel ready. Seek the Ashmarket Master to train.")
		}
	case "lose":
		lost := ch.Gold * 15 / 100
		ch.Gold -= lost
		if ch.CurrentHitPoints < 1 {
			ch.CurrentHitPoints = 1
		}
		ch.Door.Losses++
		if f.Kind == fightMaster {
			f.Log = append(f.Log, fmt.Sprintf("The master lowers the blade. %d coin is gone from your belt. Mend, then try again.", lost))
		} else {
			f.Log = append(f.Log, fmt.Sprintf("You wake at the gate. %d coin is gone from your belt. The vault is untouched.", lost))
		}
	case "fled":
		if f.Kind == fightMaster {
			f.Log = append(f.Log, "You step out of the circle. No level is granted.")
		} else {
			f.Log = append(f.Log, "The walk is spent either way.")
		}
	}
}

// finishMasterWin applies the classic door train loop: win the duel, then level
// by one only when CheckLevelUp says the XP threshold is met.
func (h *Hub) finishMasterWin(ch *characters.Character, f *fight) {
	ch.Door.Wins++
	levels, _ := leveling.CheckLevelUp(ch)
	if levels <= 0 {
		f.Log = append(f.Log, "The master nods. You are not ready — walk the Ashwood more, then return.")
		return
	}
	if ch.Level >= doorLevelCap {
		f.Log = append(f.Log, "The master has nothing left to teach.")
		return
	}
	leveling.ApplyLevelUp(ch, 1)
	f.Log = append(f.Log, fmt.Sprintf("The chalk circle holds. You reach level %d. Your wounds close as the lesson settles.", ch.Level))
}

// trainMaster starts (or resumes) the Ashmarket Master duel for a level-up.
func (h *Hub) trainMaster(user *entities.User, sess *session) {
	ch, err := h.load(user)
	if err != nil || ch == nil {
		sess.notice = "You have no warrior yet."
		return
	}
	if ch.Level >= doorLevelCap {
		sess.notice = "The Ashmarket Master has nothing left to teach. The road beyond is yours alone."
		return
	}
	if ch.CurrentHitPoints < 1 {
		sess.notice = "You can barely stand. Mend first, then step into the circle."
		return
	}
	if sess.fight != nil && !sess.fight.Done && sess.fight.Kind == fightMaster {
		sess.screen = screenFight
		sess.inputMode = "hotkey"
		return
	}
	spawned := h.spawnMaster(ch.Level)
	spawned.Log = []string{fmt.Sprintf("%s steps into the chalk circle.", spawned.Name)}
	sess.fight = &spawned
	sess.screen = screenFight
	sess.inputMode = "hotkey"
	sess.returnTo = "trainer"
	sess.notice = "Steel rings. Press (A)ttack or (R)un."
}

func (h *Hub) healerKey(user *entities.User, sess *session, key string) {
	switch key {
	case "r":
		h.goBack(sess)
	case "h":
		_, err := h.mutate(user, func(ch *characters.Character) error {
			if ch.CurrentHitPoints >= ch.MaxHitPoints {
				sess.notice = h.pack.Healer + " finds nothing to mend."
				return nil
			}
			cost := int64(h.pack.HealCost)
			if ch.Gold < cost {
				sess.notice = fmt.Sprintf("The mend costs %d coin. You are short.", cost)
				return nil
			}
			ch.Gold -= cost
			ch.CurrentHitPoints = ch.MaxHitPoints
			sess.notice = h.pack.Healer + " knits the ache out of you."
			return nil
		})
		if err != nil {
			sess.notice = "The healer is busy. Try again."
		}
	default:
		sess.notice = "Press (H) to pay for mending, or (R) to return."
	}
}

func (h *Hub) bankKey(user *entities.User, sess *session, key string) {
	switch key {
	case "r":
		h.goBack(sess)
	case "d":
		sess.purpose = "deposit"
		sess.screen = screenAmount
		sess.inputMode = "line"
	case "w":
		sess.purpose = "withdraw"
		sess.screen = screenAmount
		sess.inputMode = "line"
	default:
		sess.notice = "Press (D)eposit, (W)ithdraw, or (R)eturn."
	}
}

func (h *Hub) applyAmount(user *entities.User, sess *session, text string) {
	if strings.EqualFold(text, "x") || text == "" {
		sess.screen = screenBank
		sess.inputMode = "hotkey"
		sess.notice = "Left the amount unwritten."
		return
	}
	_, err := h.mutate(user, func(ch *characters.Character) error {
		if ch.Door == nil {
			ch.Door = &characters.DoorProfile{}
		}
		var available int64
		if sess.purpose == "withdraw" {
			available = ch.Door.BankGold
		} else {
			available = ch.Gold
		}
		amount, ok := parseAmount(text, available)
		if !ok {
			sess.notice = "That is not a sum you can move."
			return nil
		}
		if sess.purpose == "withdraw" {
			ch.Door.BankGold -= amount
			ch.Gold += amount
			sess.notice = fmt.Sprintf("Ilen Vos counts out %d coin.", amount)
		} else {
			ch.Gold -= amount
			ch.Door.BankGold += amount
			sess.notice = fmt.Sprintf("The vault takes %d coin off your belt.", amount)
		}
		return nil
	})
	sess.screen = screenBank
	sess.inputMode = "hotkey"
	if err != nil {
		sess.notice = "The vault ledger sticks. Try again."
	}
}

func (h *Hub) applyBuy(user *entities.User, sess *session, text string) {
	back := sess.returnTo
	if back == "" {
		back = "shop"
		if h.pack.Screens == nil || h.pack.Screens["shop"] == nil {
			back = screenArmory
		}
	}
	sess.screen = back
	sess.returnTo = ""
	sess.inputMode = "hotkey"
	if text == "" || strings.EqualFold(text, "x") {
		sess.notice = "Left the counter."
		return
	}
	h.armoryKey(user, sess, strings.ToLower(strings.TrimSpace(text)))
}

func (h *Hub) armoryKey(user *entities.User, sess *session, key string) {
	if key == "r" {
		h.goBack(sess)
		return
	}
	n, err := strconv.Atoi(key)
	if err != nil {
		sess.notice = "Press a number to buy, or (R) to return."
		return
	}
	goods := h.shopGoodsFor(sess.screen)
	if n < 1 || n > len(goods) {
		sess.notice = "That peg is empty."
		return
	}
	item := goods[n-1]
	_, err = h.mutate(user, func(ch *characters.Character) error {
		if ch.Door == nil {
			ch.Door = &characters.DoorProfile{}
		}
		if (item.kind == "weapon" && ch.Door.WeaponID == item.id) || (item.kind == "armor" && ch.Door.ArmorID == item.id) {
			sess.notice = "You already carry that."
			return nil
		}
		price := int64(item.price)
		if ch.Gold < price {
			sess.notice = h.pack.Smith + " shakes his head. Not enough coin."
			return nil
		}
		ch.Gold -= price
		if item.kind == "weapon" {
			ch.Door.WeaponID = item.id
		} else {
			ch.Door.ArmorID = item.id
		}
		sess.notice = "Bought " + item.name + "."
		return nil
	})
	if err != nil {
		sess.notice = "The smith is at the forge. Try again."
	}
}

func (h *Hub) innKey(user *entities.User, sess *session, key string) {
	switch key {
	case "r":
		h.goBack(sess)
	case "d":
		ch, err := h.load(user)
		if err != nil || ch == nil {
			sess.notice = "Maela cannot find your name."
			return
		}
		if _, err := h.daily.Consume(ch.ID, innKey, 1, 1); err == daily.ErrExhausted {
			sess.notice = h.pack.Innkeeper + " already poured for you today."
			return
		} else if err != nil {
			sess.notice = "The kettle is off the hook. Try again."
			return
		}
		_, err = h.mutate(user, func(loaded *characters.Character) error {
			missing := loaded.MaxHitPoints - loaded.CurrentHitPoints
			heal := missing / 3
			if heal < 4 {
				heal = 4
			}
			loaded.CurrentHitPoints += heal
			if loaded.CurrentHitPoints > loaded.MaxHitPoints {
				loaded.CurrentHitPoints = loaded.MaxHitPoints
			}
			return nil
		})
		if err != nil {
			sess.notice = "The draught spills. Try again."
			return
		}
		sess.notice = h.pack.Innkeeper + " slides you a cup of ember draught."
	default:
		sess.notice = "Press (D) to drink, or (R) to return."
	}
}

func (h *Hub) createWarrior(user *entities.User, name, sex string, tr Track) error {
	if h.facade.CharactersService().IsCharacterNameTaken(name) {
		return fmt.Errorf("someone on this door already carries the name %s", name)
	}
	hp := int32(16 + tr.CON)
	if hp < 12 {
		hp = 12
	}
	ch := &characters.Character{
		Name:             name,
		Description:      tr.Blurb,
		Race:             characters.RaceHuman,
		Class:            classFor(tr.Class),
		CurrentHitPoints: hp,
		MaxHitPoints:     hp,
		XP:               0,
		Level:            1,
		Gold:             starterGold,
		MaxLevelCap:      doorLevelCap,
		Attributes: characters.Attributes{
			{Name: "Strength", Short: "STR", Value: int32(tr.STR)},
			{Name: "Dexterity", Short: "DEX", Value: int32(tr.DEX)},
			{Name: "Constitution", Short: "CON", Value: int32(tr.CON)},
			{Name: "Intelligence", Short: "INT", Value: int32(tr.INT)},
			{Name: "Wisdom", Short: "WIS", Value: int32(tr.WIS)},
		},
		Door: &characters.DoorProfile{
			Sex:   sex,
			Track: tr.ID,
		},
	}
	ch.BelongsUserID = user.ID
	created, err := h.facade.CharactersService().Store(ch)
	if err != nil {
		return err
	}
	user.LastCharacter = created.ID
	user.Nickname = name
	return h.facade.UsersService().Update(user.RefID, user)
}

func (h *Hub) load(user *entities.User) (*characters.Character, error) {
	if user.LastCharacter == "" {
		chars, err := h.facade.CharactersService().FindAllForUser(user.ID)
		if err != nil {
			return nil, err
		}
		if len(chars) == 0 {
			return nil, fmt.Errorf("no character")
		}
		user.LastCharacter = chars[0].ID
		_ = h.facade.UsersService().Update(user.RefID, user)
	}
	return h.facade.CharactersService().FindByID(user.LastCharacter)
}

func (h *Hub) mutate(user *entities.User, fn func(*characters.Character) error) (*characters.Character, error) {
	ch, err := h.load(user)
	if err != nil || ch == nil {
		return nil, err
	}
	if err := h.facade.CharactersService().Modify(ch.ID, func(loaded *characters.Character) error {
		if loaded.Door == nil {
			loaded.Door = &characters.DoorProfile{WeaponID: weaponID(loaded), ArmorID: armorID(loaded)}
		}
		return fn(loaded)
	}); err != nil {
		return nil, err
	}
	return h.facade.CharactersService().FindByID(ch.ID)
}

func (h *Hub) render(user *entities.User, sess *session, send func(any)) {
	var ch *characters.Character
	if user != nil && user.LastCharacter != "" {
		ch, _ = h.facade.CharactersService().FindByID(user.LastCharacter)
	}
	send(h.compose(sess, ch).frame())
}

func (h *Hub) compose(sess *session, ch *characters.Character) view {
	name := ""
	if ch != nil {
		name = ch.Name
	} else if sess.draftName != "" {
		name = sess.draftName
	}
	if scr, ok := h.pack.Screens[sess.screen]; ok && sess.screen != screenAmount {
		return h.composeArt(sess, ch, scr)
	}
	v := view{
		Location:  h.pack.Town,
		Prompt:    "Your command",
		Footer:    h.footer(ch),
		InputMode: sess.inputMode,
		ScreenID:  sess.screen,
	}
	if name != "" && sess.inputMode != "line" {
		v.Prompt = "Your command, " + name + "? "
	} else if sess.inputMode != "line" {
		v.Prompt = "Your command? "
	}
	var body []string
	if sess.notice != "" {
		body = append(body, ansiKey+sess.notice+ansiReset)
		body = append(body, "")
	}
	switch sess.screen {
	case screenName:
		v.Location = "Name your warrior"
		v.Prompt = "Warrior name: "
		v.InputMode = "line"
		body = append(body, h.prose([]string{"The square asks for a name before it will keep your coin."})...)
	case screenSex:
		v.Location = "How should the square address you?"
		body = append(body,
			hotkey("M", "Man", ""),
			hotkey("F", "Woman", ""),
			hotkey("X", "Neither word fits", ""),
		)
	case screenTrack:
		v.Location = "Choose a path"
		for i, tr := range h.pack.Tracks {
			body = append(body, hotkey(strconv.Itoa(i+1), tr.Name, ""))
			body = append(body, h.prose([]string{"    " + tr.Blurb})...)
		}
	case screenTown:
		v.Location = h.pack.Town + " — " + h.pack.Setting
		body = append(body, h.prose(h.pack.Intro)...)
		body = append(body, "")
		body = append(body,
			hotkey("F", h.pack.Forest, "daily walks"),
			hotkey("H", h.pack.Healer, fmt.Sprintf("%d coin to full health", h.pack.HealCost)),
			hotkey("B", h.pack.Bank, "deposit or withdraw"),
			hotkey("A", h.pack.Armory, h.pack.Smith),
			hotkey("Y", "Your stats", ""),
			hotkey("L", "Ledger of names", "leaderboard"),
			hotkey("N", "Day's whisper", "news"),
			hotkey("I", h.pack.Inn, "one draught a day"),
			hotkey("?", "Show this menu", ""),
		)
		if ch != nil {
			bank := int64(0)
			if ch.Door != nil {
				bank = ch.Door.BankGold
			}
			body = append(body, "", fmt.Sprintf("Coin on hand: %d    Vault: %d", ch.Gold, bank))
		}
	case screenFight:
		body = append(body, h.fightLines(ch, sess.fight)...)
		body = append(body, "",
			hotkey("A", "Attack", ""),
			hotkey("R", "Run", ""),
			hotkey("S", "Stats", ""),
		)
	case screenResult:
		body = append(body, h.fightLines(ch, sess.fight)...)
		body = append(body, "", hotkey("C", "Continue", "back to the square"))
	case screenHealer:
		v.Location = h.pack.Healer
		body = append(body, h.prose([]string{
			h.pack.Healer + " works under a canvas awning. Full mending costs " + strconv.Itoa(h.pack.HealCost) + " coin.",
		})...)
		body = append(body, "", hotkey("H", "Pay and mend", ""), hotkey("R", "Return", ""))
	case screenBank:
		v.Location = h.pack.Bank
		bank := int64(0)
		hand := int64(0)
		if ch != nil {
			hand = ch.Gold
			if ch.Door != nil {
				bank = ch.Door.BankGold
			}
		}
		body = append(body, fmt.Sprintf("%s keeps the book.", h.pack.Banker))
		body = append(body, fmt.Sprintf("On hand: %d    In the vault: %d", hand, bank))
		body = append(body, "Coin on your belt can be lost in the trees. The vault cannot.")
		body = append(body, "", hotkey("D", "Deposit", ""), hotkey("W", "Withdraw", ""), hotkey("R", "Return", ""))
	case screenAmount:
		v.InputMode = "line"
		if sess.purpose == "withdraw" {
			v.Prompt = "Withdraw how much (or all, or x)? "
			v.Location = "Withdraw"
		} else {
			v.Prompt = "Deposit how much (or all, or x)? "
			v.Location = "Deposit"
		}
	case screenStats:
		v.Location = "Your stats"
		body = append(body, h.statLines(ch)...)
		body = append(body, "", hotkey("R", "Return", ""))
	case screenBoard:
		v.Location = "Ledger of names"
		body = append(body, "Ranked by level, then experience. Challenges between warriors come later.")
		body = append(body, "")
		body = append(body, h.leaderLines()...)
		body = append(body, "", hotkey("R", "Return", ""))
	case screenNews:
		v.Location = "Day's whisper"
		if len(h.pack.News) == 0 {
			body = append(body, "The board is blank.")
		}
		for _, line := range h.pack.News {
			body = append(body, h.prose([]string{"- " + line})...)
		}
		body = append(body, "", hotkey("R", "Return", ""))
	case screenInn:
		v.Location = h.pack.Inn
		body = append(body, h.prose([]string{
			h.pack.Innkeeper + " keeps a kettle on. One ember draught a day, no coin.",
		})...)
		body = append(body, "", hotkey("D", "Drink", ""), hotkey("R", "Return", ""))
	case screenArmory:
		v.Location = h.pack.Armory
		body = append(body, h.pack.Smith+" sells one weapon and one coat at a time.")
		body = append(body, "")
		for _, g := range h.shopGoodsFor(sess.screen) {
			body = append(body, hotkey(strconv.Itoa(g.n), g.name, fmt.Sprintf("%d coin, %s", g.price, g.stat)))
		}
		body = append(body, "", hotkey("R", "Return", ""))
	}
	v.Body = body
	if v.InputMode == "" {
		v.InputMode = "hotkey"
	}
	return v
}

func (h *Hub) fightLines(ch *characters.Character, f *fight) []string {
	if f == nil {
		return []string{"The path is empty."}
	}
	youHP, youMax := 0, 1
	if ch != nil {
		youHP = int(ch.CurrentHitPoints)
		youMax = int(ch.MaxHitPoints)
	}
	lines := []string{
		fmt.Sprintf("%s  %s  %d/%d", f.Name, hpBar(f.HP, f.MaxHP), f.HP, f.MaxHP),
		fmt.Sprintf("You  %s  %d/%d", hpBar(youHP, youMax), youHP, youMax),
		"",
	}
	logLines := f.Log
	if len(logLines) > 8 {
		logLines = logLines[len(logLines)-8:]
	}
	lines = append(lines, logLines...)
	return lines
}

func (h *Hub) statLines(ch *characters.Character) []string {
	if ch == nil {
		return []string{"No warrior is written in the book yet."}
	}
	track := "unset"
	sex := ""
	bank := int64(0)
	wins, losses, gems := 0, 0, 0
	weapon := h.pack.WeaponByID(weaponID(ch)).Name
	armor := h.pack.ArmorByID(armorID(ch)).Name
	if ch.Door != nil {
		track = ch.Door.Track
		sex = ch.Door.Sex
		bank = ch.Door.BankGold
		wins = ch.Door.Wins
		losses = ch.Door.Losses
		gems = ch.Door.Gems
	}
	walks := ""
	if res, err := h.daily.Get(ch.ID, daily.ForestFightsKey, h.allowance()); err == nil {
		walks = fmt.Sprintf("%d/%d", res.Remaining, res.Allowance)
	}
	return []string{
		fmt.Sprintf("%s   path %s   address %s", ch.Name, track, sexLabel(sex)),
		fmt.Sprintf("Level %d   XP %d   HP %d/%d", ch.Level, ch.XP, ch.CurrentHitPoints, ch.MaxHitPoints),
		fmt.Sprintf("STR %d  DEX %d  CON %d  INT %d  WIS %d", ch.GetAttribute("STR"), ch.GetAttribute("DEX"), ch.GetAttribute("CON"), ch.GetAttribute("INT"), ch.GetAttribute("WIS")),
		fmt.Sprintf("Weapon: %s    Armor: %s", weapon, armor),
		fmt.Sprintf("Coin %d    Vault %d    Gems %d", ch.Gold, bank, gems),
		fmt.Sprintf("Wins %d    Losses %d    Walks %s", wins, losses, walks),
	}
}

func (h *Hub) leaderLines() []string {
	chars, err := h.facade.CharactersService().FindAll()
	if err != nil || len(chars) == 0 {
		return []string{"The ledger is empty."}
	}
	ranked := make([]*characters.Character, 0, len(chars))
	for _, ch := range chars {
		if ch != nil && ch.Door != nil {
			ranked = append(ranked, ch)
		}
	}
	if len(ranked) == 0 {
		return []string{"The ledger is empty."}
	}
	for i := 0; i < len(ranked); i++ {
		for j := i + 1; j < len(ranked); j++ {
			if worse(ranked[i], ranked[j]) {
				ranked[i], ranked[j] = ranked[j], ranked[i]
			}
		}
	}
	lines := make([]string, 0, 12)
	limit := len(ranked)
	if limit > 12 {
		limit = 12
	}
	for i := 0; i < limit; i++ {
		ch := ranked[i]
		name := ch.Name
		if len(name) > 18 {
			name = name[:18]
		}
		lines = append(lines, fmt.Sprintf("%2d  %-18s Lv %-2d  XP %d", i+1, name, ch.Level, ch.XP))
	}
	return lines
}

func worse(a, b *characters.Character) bool {
	if a.Level != b.Level {
		return a.Level < b.Level
	}
	if a.XP != b.XP {
		return a.XP < b.XP
	}
	aw, bw := 0, 0
	if a.Door != nil {
		aw = a.Door.Wins
	}
	if b.Door != nil {
		bw = b.Door.Wins
	}
	return aw < bw
}

func (h *Hub) footer(ch *characters.Character) string {
	clock := h.now().In(h.loc).Format("15:04")
	if ch == nil {
		return h.pack.Title + "  " + clock
	}
	bank := int64(0)
	if ch.Door != nil {
		bank = ch.Door.BankGold
	}
	walks := ""
	if res, err := h.daily.Get(ch.ID, daily.ForestFightsKey, h.allowance()); err == nil {
		walks = fmt.Sprintf("  walks %d/%d", res.Remaining, res.Allowance)
	}
	name := ch.Name
	if len(name) > 16 {
		name = name[:16]
	}
	return fmt.Sprintf("%s  Lv %d  HP %d/%d  coin %d  vault %d%s  %s",
		name, ch.Level, ch.CurrentHitPoints, ch.MaxHitPoints, ch.Gold, bank, walks, clock)
}

func (h *Hub) prose(lines []string) []string {
	var out []string
	for _, ln := range lines {
		out = append(out, wrapText(ln, 76)...)
	}
	return out
}

type goods struct {
	n     int
	kind  string
	id    string
	name  string
	price int
	stat  string
}

// shopKind returns "weapon", "armor", or "" (both) for a shop screen id.
func shopKind(screen string) string {
	switch screen {
	case "weapons", "shop":
		return "weapon"
	case "armor", "armory":
		return "armor"
	default:
		return ""
	}
}

func (h *Hub) shopGoods() []goods {
	return h.shopGoodsFor("")
}

func (h *Hub) shopGoodsFor(screen string) []goods {
	kind := shopKind(screen)
	var list []goods
	n := 1
	if kind == "" || kind == "weapon" {
		for _, w := range h.pack.Weapons {
			if w.Price <= 0 {
				continue
			}
			list = append(list, goods{n: n, kind: "weapon", id: w.ID, name: w.Name, price: w.Price, stat: fmt.Sprintf("atk %d", w.Attack)})
			n++
		}
	}
	if kind == "" || kind == "armor" {
		for _, a := range h.pack.Armor {
			if a.Price <= 0 {
				continue
			}
			list = append(list, goods{n: n, kind: "armor", id: a.ID, name: a.Name, price: a.Price, stat: fmt.Sprintf("def %d", a.Defense)})
			n++
		}
	}
	return list
}

func cleanName(text string) (string, bool) {
	name := strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if !warriorName.MatchString(name) {
		return "", false
	}
	return name, true
}

func parseAmount(text string, available int64) (int64, bool) {
	text = strings.TrimSpace(text)
	if strings.EqualFold(text, "all") {
		if available <= 0 {
			return 0, false
		}
		return available, true
	}
	n, err := strconv.ParseInt(text, 10, 64)
	if err != nil || n <= 0 || n > available {
		return 0, false
	}
	return n, true
}

func classFor(id string) characters.Class {
	switch id {
	case "wizard":
		return characters.ClassWizard
	case "rogue":
		return characters.ClassRogue
	default:
		return characters.ClassWarrior
	}
}

func sexLabel(sex string) string {
	switch sex {
	case "m":
		return "man"
	case "f":
		return "woman"
	case "x":
		return "unspecified"
	default:
		return "unspecified"
	}
}
