package game

import (
	"github.com/atla/owndnd/pkg/entities"
	log "github.com/sirupsen/logrus"

	c "github.com/atla/owndnd/pkg/mudserver/game/commands"
	m "github.com/atla/owndnd/pkg/mudserver/game/messages"
	"github.com/atla/owndnd/pkg/service"
)

//Game ... contains live game state
type Game struct {
	id    string
	title string

	// access to repository data
	Facade service.Facade

	SystemUser *entities.User

	// messages
	onMessageReceived chan *m.Message

	OnUserJoined chan *m.UserJoined
	OnUserQuit   chan *m.UserQuit

	//OnAvatarJoinedRoom chan *AvatarJoinedRoom
	//OnAvatarLeftRoom   chan *AvatarLeftRoom

	Receivers []Receiver

	CommandProcessor *c.CommandProcessor

	Avatars map[string]*Avatar

	//world *World
}

// New creates a new game instance
func New(facade service.Facade) *Game {
	return &Game{

		title: "Lair of the Dragon",

		CommandProcessor: c.NewCommandProcessor(),

		// event channels
		onMessageReceived: make(chan *m.Message, 20),
		OnUserJoined:      make(chan *m.UserJoined, 20),
		OnUserQuit:        make(chan *m.UserQuit, 20),

		// game update listeners
		Receivers: make([]Receiver, 0, 10),

		Avatars: make(map[string]*Avatar),

		Facade: facade,
	}
}

// Subscribe ... sub
func (g *Game) Subscribe(receiver Receiver) {
	g.Receivers = append(g.Receivers, receiver)
}

// Receiver ... rec
type Receiver interface {
	OnMessage(message interface{})
}

//Unsubscribe ...
func (g *Game) Unsubscribe(receiver *Receiver) {
	//TODO:
	//game.Receivers = delete(game.Receivers, receiver)
}

// SendMessage ...
func (g *Game) SendMessage(msg interface{}) {
	// broeadcast message to all game listeners (currently only websocket server)
	for _, receiver := range g.Receivers {
		receiver.OnMessage(msg)
	}
}

//OnMessageReceived returns onMessageReceived channel
func (g *Game) OnMessageReceived() chan *m.Message {
	return g.onMessageReceived
}

//Run main game loop
func (g *Game) Run() {

	go func() {
		for {
			select {
			case userJoined := <-g.OnUserJoined:
				log.Info("Received UserJoinged message")
				g.handleUserJoined(userJoined.User)
				break

			case userQuit := <-g.OnUserQuit:
				log.WithField("user", userQuit.User).Info("Received UserQuit message")
				g.handleUserQuit(userQuit.User)

				break
			case message := <-g.onMessageReceived:
				// attach current character if a user is set
				g.attachCharacterToMessage(message)

				// only broadcast if commandprocessor didnt process it
				if !g.CommandProcessor.Process(g, message) {

					// generic messages will be converted to plain OutgoingMessages (type message)
					// and send to the room audience including the origin nickname or charactername
					g.handleDefaultMessage(message)
				}
			}
		}
	}()
}
