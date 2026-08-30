package settings

import "github.com/talesmud/talesmud/pkg/entities"

// ServerSettings holds global server configuration.
type ServerSettings struct {
	*entities.Entity `json:",inline"`
	ServerName       string `json:"serverName"`
	About            string `json:"about"`

	// GuestsAllowed controls whether guest mode is enabled (anonymous demo sessions).
	GuestsAllowed bool `json:"guestsAllowed"`

	// MaxGuestAccounts is the maximum number of concurrent guest accounts (0 = unlimited).
	MaxGuestAccounts int `json:"maxGuestAccounts"`

	// StartRoomID is where new and guest characters spawn. Empty means
	// "R0001 if that room exists". Never use an arbitrary rooms[0] while
	// a configured or Veilspan start room is available.
	StartRoomID string `json:"startRoomID"`
}

// NewDefaultServerSettings returns settings with default values.
func NewDefaultServerSettings() *ServerSettings {
	e := entities.NewEntity()
	e.ID = "server-settings"
	return &ServerSettings{
		Entity:           e,
		ServerName:       "TalesMUD",
		About:            "",
		GuestsAllowed:    true,
		MaxGuestAccounts: 20,
		StartRoomID:      "R0001",
	}
}
