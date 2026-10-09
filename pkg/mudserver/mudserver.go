package mudserver

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/presentation/doorview"
	"github.com/talesmud/talesmud/pkg/resources"
	"github.com/talesmud/talesmud/pkg/scripts"
	"github.com/talesmud/talesmud/pkg/service"
)

// MUDServer ... server application connecting the websocket clients with the game instance, providing utility functions etc.
type MUDServer interface {
	Run()
	GameCtrl() def.GameCtrl
	HandleConnections(*gin.Context)
	SetResourceStore(*resources.Store)
	SetSessionHook(SessionHook)
	// AttachExternal takes a non-WebSocket session. input receives a line
	// (isKey is true for a door hotkey, including an empty Enter). done
	// detaches this session only.
	AttachExternal(user *entities.User, t Transport) (input func(text string, isKey bool), done func())
}

// WS close codes (application-specific, RFC6455 4000-4999).
const (
	// closeSessionReplaced is sent when a newer socket for the same user takes over.
	// Clients must NOT auto-reconnect on this code (stops desktop↔mobile flap).
	closeSessionReplaced = 4001
)

// Connection ...
type Connection struct {
	User *entities.User
	t    Transport
	mu   sync.Mutex

	active   bool
	remoteIP string
}

func (p *Connection) send(v interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.t == nil {
		return errors.New("no transport")
	}
	return p.t.Send(v)
}

func (p *Connection) closeTransport(code int, reason string) {
	if p == nil || p.t == nil {
		return
	}
	p.t.Close(code, reason)
}

func (p *Connection) Kind() string {
	if p == nil || p.t == nil {
		return ""
	}
	return p.t.Kind()
}

/*CheckOrigin:
 */
type server struct {
	Facade service.Facade
	port   string

	Game *game.Game

	hook SessionHook

	Clients   *clientRegistry
	Broadcast chan interface{}
	Upgrader  websocket.Upgrader
}

func (server *server) GameCtrl() def.GameCtrl {
	return server.Game
}

// SetSessionHook installs a presentation-mode input owner. Nil keeps classic play.
func (server *server) SetSessionHook(hook SessionHook) {
	if server == nil {
		return
	}
	server.hook = hook
}

func (server *server) ansiSession() bool {
	return server.hook != nil && server.hook.Active()
}

// SetResourceStore keeps the game and any later caller on the same catalog.
func (server *server) SetResourceStore(store *resources.Store) {
	if server == nil || server.Game == nil {
		return
	}
	server.Game.Resources = store
}

// New creates a new mud server
func New(facade service.Facade) MUDServer {

	game := game.New(facade)

	srv := &server{
		Facade: facade,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		Clients:   newClientRegistry(),
		Broadcast: make(chan interface{}),
		Game:      game,
	}

	if gamemode.ANSI() {
		srv.SetSessionHook(&doorview.View{Game: game, Title: gamemode.Current().Title})
	}
	return srv
}

func (server *server) Run() {

	log.WithTime(time.Now()).Info("MUD Server starting ...")

	// Drop leftover instance copies before any player can connect.
	// ListenAndServe runs only after setupRoutes returns, and this call
	// returns before the game loop starts, so the sweep is done first.
	if server.Game != nil {
		server.Game.SweepInstanceRooms()
	}

	go server.receiveMessages()
	go server.Game.Run()
	go server.handleBroadcastMessages()
	go server.handleClientTimeouts()

	log.WithTime(time.Now()).Info("MUD Server running")
}

func (server *server) handleClientTimeouts() {

	pingTicker := time.NewTicker(60 * time.Second)

	for {
		select {
		case <-pingTicker.C:
			server.sendUserPings()
		}
	}
}

func (server *server) sendUserPings() {

	server.Clients.ForEach(func(_ string, con *Connection) {
		server.sendMessage(con.User.ID, messages.MessageResponse{
			Type: messages.MessageTypePing,
		})
	})

}

// HandleConnections asd
func (server *server) HandleConnections(c *gin.Context) {

	var user *entities.User

	if usr, exists := c.Get("user"); exists {
		user = usr.(*entities.User)
	}
	if user == nil {
		log.WithField("ip", c.ClientIP()).Warn("WS auth missing user")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	remoteIP := c.ClientIP()
	log.WithFields(log.Fields{
		"userId":   user.ID,
		"nickname": user.Nickname,
		"ip":       remoteIP,
	}).Info("WS connect")

	// Upgrade initial GET request to a websocket
	ws, err := server.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"userId":   user.ID,
			"nickname": user.Nickname,
			"ip":       remoteIP,
		}).Warn("WS upgrade failed")
		return
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	connection := server.attachSession(user, &wsTransport{conn: ws, ip: remoteIP})

	for {
		// Read in a new message as JSON and map it to a Message object
		var msg messages.IncomingMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"userId":      user.ID,
				"nickname":    user.Nickname,
				"ip":          remoteIP,
				"characterId": user.LastCharacter,
			}).Info("WS read error")
			server.detachSession(user, connection)
			break
		}

		text := msg.Message
		isKey := false
		if strings.EqualFold(msg.Type, "door_key") && msg.Key != "" {
			text = msg.Key
			isKey = true
		}
		if server.ansiSession() {
			// Empty door keys stay ignored on the websocket, matching the
			// previous read loop. SSH passes isKey for an empty Enter.
			if text != "" {
				server.handleInput(user, text, isKey)
			}
			continue
		}
		server.handleInput(user, msg.Message, false)
	}
}

// attachSession installs t as the user's live session and kicks the previous one.
// Ordering matters: Replace before Close, so the old read-loop's DeleteIf misses
// the new session and skips UserQuit.
func (server *server) attachSession(user *entities.User, t Transport) *Connection {
	remoteIP := ""
	if t != nil {
		remoteIP = t.RemoteIP()
	}
	connection := &Connection{
		User:     user,
		t:        t,
		active:   true,
		remoteIP: remoteIP,
	}
	old := server.Clients.Replace(user.ID, connection)
	if old != nil && old.t != nil {
		replaceLog := "WS replace-existing"
		if old.Kind() != "ws" || connection.Kind() != "ws" {
			replaceLog = "session replace"
		}
		log.WithFields(log.Fields{
			"userId":      user.ID,
			"nickname":    user.Nickname,
			"ip":          remoteIP,
			"oldIP":       old.remoteIP,
			"characterId": user.LastCharacter,
			"reason":      "session replaced",
			"oldKind":     old.Kind(),
			"kind":        connection.Kind(),
		}).Info(replaceLog)
		old.closeTransport(closeSessionReplaced, "session replaced")
	}

	log.WithFields(log.Fields{
		"userId":      user.ID,
		"nickname":    user.Nickname,
		"ip":          remoteIP,
		"characterId": user.LastCharacter,
		"replaced":    old != nil,
		"kind":        connection.Kind(),
	}).Info("WS upgrade")

	user.LastSeen = time.Now()
	user.IsOnline = true
	server.Facade.UsersService().Update(user.RefID, user)

	serverName := "TalesMUD"
	if ss, err := server.Facade.ServerSettingsService().Get(); err == nil && ss.ServerName != "" {
		serverName = ss.ServerName
	}
	if server.ansiSession() {
		server.hook.OnConnect(user, func(v any) {
			server.sendMessage(user.ID, v)
		})
	} else {
		server.sendMessage(user.ID, messages.NewRoomBasedMessage("", "Connected to ["+serverName+"] ..."))

		server.Game.OnUserJoined <- &messages.UserJoined{
			User: user,
		}
	}

	server.armGuestExpiry(user, connection)
	return connection
}

// armGuestExpiry warns five minutes before a guest lapses, then closes that
// same session. A newer replacement is left alone.
func (server *server) armGuestExpiry(user *entities.User, connection *Connection) {
	if user == nil || !user.IsGuest || user.GuestExpiresAt.IsZero() {
		return
	}
	expires := user.GuestExpiresAt
	go func() {
		warnAt := time.Until(expires) - 5*time.Minute
		if warnAt > 0 {
			time.Sleep(warnAt)
			if client, ok := server.Clients.Get(user.ID); ok && client == connection {
				server.sendMessage(user.ID, messages.MessageResponse{
					Type:    messages.MessageTypeDefault,
					Message: "\n[SYSTEM] Your guest session expires in 5 minutes. Create an account to save your progress!\n",
				})
			}
		}
	}()
	go func() {
		timeout := time.Until(expires)
		if timeout <= 0 {
			timeout = 1 * time.Second
		}
		time.Sleep(timeout)
		if client, ok := server.Clients.Get(user.ID); !ok || client != connection {
			return
		}
		server.sendMessage(user.ID, messages.MessageResponse{
			Type:    messages.MessageTypeDefault,
			Message: "\n[SYSTEM] Your guest session has expired. Thank you for playing! Create an account to continue your adventure.\n",
		})
		time.Sleep(500 * time.Millisecond)
		if client, ok := server.Clients.Get(user.ID); ok && client == connection {
			client.closeTransport(0, "")
		}
	}()
}

// handleInput delivers one line. isKey is set for a door hotkey, including Enter.
func (server *server) handleInput(user *entities.User, text string, isKey bool) {
	if user == nil {
		return
	}
	if server.ansiSession() {
		if text == "" && !isKey {
			return
		}
		server.hook.OnInput(user, text, func(v any) {
			server.sendMessage(user.ID, v)
		})
		return
	}
	server.Game.ConnectUserSession(user)
	user.LastSeen = time.Now()
	user.IsOnline = true
	server.Facade.UsersService().Update(user.RefID, user)
	if text != "" {
		server.Game.OnMessageReceived() <- messages.NewMessage(user, text)
	}
}

// detachSession drops this session when it is still the live one.
// A replaced socket returns false and does not quit the user.
func (server *server) detachSession(user *entities.User, connection *Connection) bool {
	if !server.handleConnectionClosed(user, connection) {
		return false
	}
	if user != nil && user.IsGuest {
		go server.cleanupGuestLater(user.ID)
	}
	return true
}

func (server *server) cleanupGuestLater(userID string) {
	time.Sleep(5 * time.Minute)
	if _, ok := server.Clients.Get(userID); ok {
		return
	}
	if chars, err := server.Facade.CharactersService().FindAllForUser(userID); err == nil {
		for _, ch := range chars {
			server.Facade.CharactersService().Delete(ch.ID)
		}
	}
	server.Facade.UsersService().Delete(userID)
	log.WithField("userID", userID).Info("Guest user cleaned up after disconnect grace period")
}

// AttachExternal is the SSH (and any future) entry. The returned done func is
// safe to call once; a later session for the same user is not removed.
func (server *server) AttachExternal(user *entities.User, t Transport) (func(string, bool), func()) {
	if server == nil || user == nil || t == nil {
		return func(string, bool) {}, func() {}
	}
	connection := server.attachSession(user, t)
	var once sync.Once
	input := func(text string, isKey bool) {
		server.handleInput(user, text, isKey)
	}
	done := func() {
		once.Do(func() {
			server.detachSession(user, connection)
		})
	}
	return input, done
}

func (server *server) handleConnectionClosed(user *entities.User, connection *Connection) bool {
	if user == nil || connection == nil {
		return false
	}
	// Stale/replaced socket: a newer session already owns this user id.
	if server.hook != nil {
		server.hook.OnDisconnect(user)
	}
	if !server.Clients.DeleteIf(user.ID, connection) {
		log.WithFields(log.Fields{
			"userId":   user.ID,
			"nickname": user.Nickname,
			"ip":       connection.remoteIP,
			"reason":   "stale socket after replace",
		}).Info("WS close ignored")
		return false
	}

	connection.active = false

	log.WithFields(log.Fields{
		"userId":      user.ID,
		"nickname":    user.Nickname,
		"ip":          connection.remoteIP,
		"characterId": user.LastCharacter,
		"reason":      "connection closed",
	}).Info("WS close")

	server.Game.OnUserQuit <- &messages.UserQuit{
		User: user,
	}

	user.IsOnline = false
	user.LastSeen = time.Now()
	server.Facade.UsersService().Update(user.RefID, user)

	return true
}

func (server *server) sendMessage(id string, msg interface{}) {

	if client, ok := server.Clients.Get(id); ok {
		//dont directly write to websocket, use this mutex protected method
		err := client.send(msg)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"userId":   client.User.ID,
				"nickname": client.User.Nickname,
				"ip":       client.remoteIP,
			}).Warn("WS send error")
			client.closeTransport(0, "")
			server.handleConnectionClosed(client.User, client)
		}
	}
}

func (server *server) sendToRoom(room *rooms.Room, msg interface{}) {
	if room == nil {
		log.Info("MUDServer::sendToRoom - room is nil (user has no character?)")
		return
	}
	server.sendToRoomID(room.ID, "", msg)
}

// sendToRoomWithout sends a message to all clients except the one with the given id
func (server *server) sendToRoomWithout(id string, room *rooms.Room, msg interface{}) {
	if room == nil {
		log.WithField("origin", id).Info("MUDServer::sendToRoomWithout - room is nil (user has no character?)")
		return
	}
	server.sendToRoomID(room.ID, id, msg)
}

func (server *server) sendToRoomID(roomID, exceptCharacterID string, msg interface{}) {
	if exceptCharacterID != "" {
		log.WithField("origin", exceptCharacterID).Info("Sending to room without origin")
	}
	if roomID == "" {
		return
	}
	for _, player := range server.Game.GetRoomPlayers(roomID, "") {
		if player.CharacterID == exceptCharacterID {
			continue
		}
		server.sendMessage(player.UserID, msg)
	}
}

func (server *server) handleBroadcastMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-server.Broadcast

		// Send it out to every client that is currently connected
		server.Clients.ForEach(func(_ string, client *Connection) {
			err := client.send(msg)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"userId":   client.User.ID,
					"nickname": client.User.Nickname,
					"ip":       client.remoteIP,
				}).Warn("WS broadcast send error")
				client.closeTransport(0, "")
				server.handleConnectionClosed(client.User, client)
			}
		})
	}
}

// OnMessage .. broadcast receiver
//func (server *server) OnMessage(message interface{}) {

func (server *server) receiveMessages() {

	for {
		message := <-server.Game.SendMessage()

		// Room-enter trigger: run a room-attached script whenever a player enters a room.
		// This is observed via the outgoing EnterRoomMessage (sent to the entering player).
		if enter, ok := message.(*messages.EnterRoomMessage); ok && enter.Type == messages.MessageTypeEnterRoom {
			server.runRoomEnterScript(enter)
		}

		if msg, ok := message.(messages.MessageResponder); ok {
			switch msg.GetAudience() {
			case messages.MessageAudienceOrigin:
				server.sendMessage(msg.GetAudienceID(), msg)
				server.noteANSI(msg)
				break
			case messages.MessageAudienceUser:
				server.sendMessage(msg.GetAudienceID(), msg)
				server.noteANSI(msg)
				break
			case messages.MessageAudienceRoom:
				// Do not load rooms from SQLite here: this goroutine drains
				// sendMessage. A DB wait while the game loop is also sending
				// deadlocks movement after on-enter scripts.
				server.sendToRoomID(msg.GetAudienceID(), "", msg)
				break

			case messages.MessageAudienceRoomWithoutOrigin:
				server.sendToRoomID(msg.GetAudienceID(), msg.GetOriginID(), msg)
				break

			case messages.MessageAudienceGlobal:
				server.Broadcast <- msg
				break
			case messages.MessageAudienceSystem:

				server.Broadcast <- messages.MessageResponse{
					Username: "#SYSTEM",
					Message:  msg.GetMessage(),
				}
				break
			}
		}
	}
}

func (server *server) noteANSI(msg messages.MessageResponder) {
	if server == nil || !server.ansiSession() || server.hook == nil || msg == nil {
		return
	}
	text := strings.TrimSpace(msg.GetMessage())
	userID := msg.GetAudienceID()
	if text == "" || userID == "" {
		return
	}
	kind := ""
	var gen uint64
	if typed, ok := msg.(interface{ GetType() messages.MessageType }); ok {
		kind = string(typed.GetType())
	}
	if stamped, ok := msg.(interface{ GetNoticeGen() uint64 }); ok {
		gen = stamped.GetNoticeGen()
	}
	go server.repaintANSI(userID, text, kind, gen)
}

func (server *server) repaintANSI(userID, text, kind string, gen uint64) {
	if server == nil || server.hook == nil || userID == "" {
		return
	}
	client, ok := server.Clients.Get(userID)
	if !ok || client == nil || client.User == nil {
		return
	}
	server.hook.OnNotice(client.User, text, kind, gen, func(v any) {
		server.sendMessage(userID, v)
	})
}

func (server *server) runRoomEnterScript(enter *messages.EnterRoomMessage) {
	if enter == nil {
		return
	}

	scriptID := enter.Room.OnEnterScriptID
	if scriptID == "" {
		return
	}

	// AudienceID is set by the caller before sending the EnterRoomMessage.
	userID := enter.GetAudienceID()
	if userID == "" {
		return
	}

	// Run asynchronously so we don't delay room rendering.
	go func() {
		script, err := server.Facade.ScriptsService().FindByID(scriptID)
		if err != nil || script == nil {
			log.WithField("scriptID", scriptID).WithError(err).Warn("Room on-enter script not found")
			return
		}

		// Load user + character (best effort)
		user, _ := server.Facade.UsersService().FindByID(userID)
		var character interface{}
		if user != nil && user.LastCharacter != "" {
			if chr, err := server.Facade.CharactersService().FindByID(user.LastCharacter); err == nil {
				character = chr
			}
		}

		// Load the canonical room (best effort) so scripts see the latest saved version.
		roomObj := interface{}(&enter.Room)
		if enter.Room.ID != "" {
			if room, err := server.Facade.RoomsService().FindByID(enter.Room.ID); err == nil && room != nil {
				roomObj = room
			}
		}

		ctx := scripts.NewScriptContext()
		ctx.Set("eventType", "player.enter_room")
		ctx.Set("room", roomObj)
		ctx.Set("toRoom", roomObj)
		if user != nil {
			ctx.Set("user", user)
		}
		if character != nil {
			ctx.Set("character", character)
		}

		result := server.Facade.Runner().RunWithResult(*script, ctx)
		if result != nil && !result.Success {
			log.WithField("script", script.Name).WithField("scriptID", scriptID).WithField("error", result.Error).
				Warn("Room on-enter script failed")
		}
	}()
}

// OnSystemMessage .. broadcast receiver
func (server *server) OnSystemMessage(message *messages.Message) {

	server.Broadcast <- messages.MessageResponse{
		Username: "#SYSTEM",
		Message:  message.Data,
	}
}
