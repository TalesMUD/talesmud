package mudserver

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/scripts"
	"github.com/talesmud/talesmud/pkg/service"
)

// MUDServer ... server application connecting the websocket clients with the game instance, providing utility functions etc.
type MUDServer interface {
	Run()
	GameCtrl() def.GameCtrl
	HandleConnections(*gin.Context)
}

// Connection ...
type Connection struct {
	User *entities.User
	ws   *websocket.Conn
	mu   sync.Mutex

	active bool
}

func (p *Connection) send(v interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.ws.WriteJSON(v)
}

/*CheckOrigin:
 */
type server struct {
	Facade service.Facade
	port   string

	Game *game.Game

	Clients   *clientRegistry
	Broadcast chan interface{}
	Upgrader  websocket.Upgrader
}

func (server *server) GameCtrl() def.GameCtrl {
	return server.Game
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

	return srv
}

func (server *server) Run() {

	log.WithTime(time.Now()).Info("MUD Server starting ...")

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
		log.WithField("User", usr.(*entities.User).Nickname).Info("User logged in")
		user = usr.(*entities.User)
	}
	if user == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Upgrade initial GET request to a websocket
	ws, err := server.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.WithError(err).Warn("Failed to upgrade websocket connection")
		return
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	log.Info("Upgraded client connection")

	// Register our new client
	server.Clients.Set(user.ID, &Connection{
		User:   user,
		ws:     ws,
		active: true,
	})

	// Send Welcome message with dynamic server name
	serverName := "TalesMUD"
	if ss, err := server.Facade.ServerSettingsService().Get(); err == nil && ss.ServerName != "" {
		serverName = ss.ServerName
	}
	server.sendMessage(user.ID, messages.NewRoomBasedMessage("", "Connected to ["+serverName+"] ..."))

	server.Game.OnUserJoined <- &messages.UserJoined{
		User: user,
	}

	// Guest session timeout: warn 5 minutes before expiry, then disconnect
	if user.IsGuest && !user.GuestExpiresAt.IsZero() {
		go func() {
			warnAt := time.Until(user.GuestExpiresAt) - 5*time.Minute
			if warnAt > 0 {
				time.Sleep(warnAt)
				server.sendMessage(user.ID, messages.MessageResponse{
					Type:    messages.MessageTypeDefault,
					Message: "\n[SYSTEM] Your guest session expires in 5 minutes. Create an account to save your progress!\n",
				})
			}
		}()
		go func() {
			timeout := time.Until(user.GuestExpiresAt)
			if timeout <= 0 {
				timeout = 1 * time.Second
			}
			time.Sleep(timeout)
			server.sendMessage(user.ID, messages.MessageResponse{
				Type:    messages.MessageTypeDefault,
				Message: "\n[SYSTEM] Your guest session has expired. Thank you for playing! Create an account to continue your adventure.\n",
			})
			// Brief delay so the message can be delivered before close
			time.Sleep(500 * time.Millisecond)
			if client, ok := server.Clients.Get(user.ID); ok {
				client.ws.Close()
			}
		}()
	}

	for {
		// Read in a new message as JSON and map it to a Message object
		var msg messages.IncomingMessage
		err := ws.ReadJSON(&msg)
		if err != nil {

			server.Game.OnUserQuit <- &messages.UserQuit{
				User: user,
			}

			log.Printf("error: %v", err)
			server.Clients.Delete(user.ID)

			// Guest disconnect cleanup with 5-minute grace period for reconnection
			if user.IsGuest {
				go func(userID string) {
					time.Sleep(5 * time.Minute)
					// Check if user reconnected during grace period
					if _, ok := server.Clients.Get(userID); ok {
						return // Reconnected, don't clean up
					}
					// Delete all characters for this guest
					if chars, err := server.Facade.CharactersService().FindAllForUser(userID); err == nil {
						for _, ch := range chars {
							server.Facade.CharactersService().Delete(ch.ID)
						}
					}
					// Delete the guest user
					server.Facade.UsersService().Delete(userID)
					log.WithField("userID", userID).Info("Guest user cleaned up after disconnect grace period")
				}(user.ID)
			}

			break
		}

		// update user online status
		server.Game.ConnectUserSession(user)

		if msg.Message != "" {
			server.Game.OnMessageReceived() <- messages.NewMessage(user, msg.Message)
		}
	}
}
func (server *server) sendMessage(id string, msg interface{}) {

	if client, ok := server.Clients.Get(id); ok {
		//dont directly write to websocket, use this mutex protected method
		err := client.send(msg)
		if err != nil {

			// tell the game that the user quit as the websocket closes/closed...
			server.Game.OnUserQuit <- &messages.UserQuit{
				User: client.User,
			}

			log.Printf("error: %v", err)
			client.ws.Close()
			server.Clients.Delete(id)
		}
	}
}

func (server *server) sendToRoom(room *rooms.Room, msg interface{}) {
	server.sendToRoomWithout("", room, msg)
}

// sendToRoomWithout sends a message to all clients except the one with the given id
func (server *server) sendToRoomWithout(id string, room *rooms.Room, msg interface{}) {

	if id != "" {
		log.WithField("origin", id).Info("Sending to room without origin")
	}

	if room == nil {
		log.WithField("origin", id).Info("MUDServer::sendToRoomWithout - room is nil (user has no character?)")
		return
	}

	usersInRoom := []string{}

	for _, player := range server.Game.GetRoomPlayers(room.ID, "") {
		if player.CharacterID == id {
			continue
		}
		usersInRoom = append(usersInRoom, player.UserID)
	}

	for _, usr := range usersInRoom {
		server.sendMessage(usr, msg)
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
				log.Printf("error: %v", err)

				server.Game.OnUserQuit <- &messages.UserQuit{
					User: client.User,
				}

				client.ws.Close()
				server.Clients.Delete(client.User.ID)

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
		if enter, ok := message.(*messages.EnterRoomMessage); ok {
			server.runRoomEnterScript(enter)
		}

		if msg, ok := message.(messages.MessageResponder); ok {
			switch msg.GetAudience() {
			case messages.MessageAudienceOrigin:
				server.sendMessage(msg.GetAudienceID(), msg)
				break
			case messages.MessageAudienceUser:
				server.sendMessage(msg.GetAudienceID(), msg)
				break
			case messages.MessageAudienceRoom:
				room, _ := server.Facade.RoomsService().FindByID(msg.GetAudienceID())
				server.sendToRoom(room, msg)
				break

			case messages.MessageAudienceRoomWithoutOrigin:
				room, _ := server.Facade.RoomsService().FindByID(msg.GetAudienceID())
				server.sendToRoomWithout(msg.GetOriginID(), room, msg)
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
