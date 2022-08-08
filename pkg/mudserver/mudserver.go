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

	Clients   map[string]*Connection
	Broadcast chan interface{}
	Upgrader  websocket.Upgrader
}

func (server *server) GameCtrl() def.GameCtrl {
	return server.Game
}

//New creates a new mud server
func New(facade service.Facade) MUDServer {

	game := game.New(facade)

	srv := &server{
		Facade: facade,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		Clients:   make(map[string]*Connection),
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

	for _, con := range server.Clients {
		server.sendMessage(con.User.ID, messages.MessageResponse{
			Type: messages.MessageTypePing,
		})
	}

}

//HandleConnections asd
func (server *server) HandleConnections(c *gin.Context) {

	var user *entities.User

	if usr, exists := c.Get("user"); exists {
		log.WithField("User", usr.(*entities.User).Nickname).Info("User logged in")
		user = usr.(*entities.User)
	}

	// Upgrade initial GET request to a websocket
	ws, err := server.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	log.Info("Upgraded client connection")

	// Register our new client
	server.Clients[user.ID] = &Connection{
		User:   user,
		ws:     ws,
		active: true,
	}

	// Send Welcome message
	server.sendMessage(user.ID, messages.NewRoomBasedMessage("", "Connected to [Tales of the Red Dragon's Lair] ..."))

	server.Game.OnUserJoined <- &messages.UserJoined{
		User: user,
	}

	for {
		// Read in a new message as JSON and map it to a Message object
		var msg messages.IncomingMessage
		err := ws.ReadJSON(&msg)
		if err != nil {

			user.IsOnline = false
			server.Facade.UsersService().Update(user.RefID, user)

			log.Printf("error: %v", err)
			delete(server.Clients, user.ID)
			break
		}

		// update user online status
		user.LastSeen = time.Now()
		user.IsOnline = true
		server.Facade.UsersService().Update(user.RefID, user)

		if msg.Message != "" {
			server.Game.OnMessageReceived() <- messages.NewMessage(user, msg.Message)
		}
	}
}
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
func (server *server) sendMessage(id string, msg interface{}) {

	if client, ok := server.Clients[id]; ok {
		//dont directly write to websocket, use this mutex protected method
		err := client.send(msg)
		if err != nil {

			// tell the game that the user quit as the websocket closes/closed...
			server.Game.OnUserQuit <- &messages.UserQuit{
				User: client.User,
			}

			// update user online status
			user := client.User
			user.IsOnline = false
			server.Facade.UsersService().Update(user.RefID, user)

			log.Printf("error: %v", err)
			client.ws.Close()
			delete(server.Clients, id)
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

	//TODO build service that reads all users from
	allUsers, _ := server.Facade.UsersService().FindAll()
	updateRoom := false
	for _, usr := range allUsers {
		if usr.LastCharacter != id && contains(*room.Characters, usr.LastCharacter) {

			// check if character is really in this room or remove
			if chr, err := server.Facade.CharactersService().FindByID(usr.LastCharacter); err == nil {
				if chr.CurrentRoomID == room.ID {
					usersInRoom = append(usersInRoom, usr.ID)
				} else {
					// remove character from current room
					room.RemoveCharacter(chr.ID)
					updateRoom = true
				}
			}
		}
	}

	if updateRoom {
		server.Facade.RoomsService().Update(room.ID, room)
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
		for _, client := range server.Clients {
			err := client.send(msg)
			if err != nil {
				log.Printf("error: %v", err)

				server.Game.OnUserQuit <- &messages.UserQuit{
					User: client.User,
				}

				client.ws.Close()
				delete(server.Clients, client.User.ID)

			}
		}
	}
}

// OnMessage .. broadcast receiver
//func (server *server) OnMessage(message interface{}) {

func (server *server) receiveMessages() {

	for {
		message := <-server.Game.SendMessage()

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

// OnSystemMessage .. broadcast receiver
func (server *server) OnSystemMessage(message *messages.Message) {

	server.Broadcast <- messages.MessageResponse{
		Username: "#SYSTEM",
		Message:  message.Data,
	}
}
