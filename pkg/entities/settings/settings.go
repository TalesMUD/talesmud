package settings

import "github.com/talesmud/talesmud/pkg/entities"

// ServerSettings holds global server configuration.
type ServerSettings struct {
	*entities.Entity `json:",inline"`
	ServerName       string `json:"serverName"`
	About            string `json:"about"`
}

// NewDefaultServerSettings returns settings with default values.
func NewDefaultServerSettings() *ServerSettings {
	e := entities.NewEntity()
	e.ID = "server-settings"
	return &ServerSettings{
		Entity:     e,
		ServerName: "TalesMUD",
		About:      "",
	}
}
