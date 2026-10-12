package messages

import "runtime/debug"

const ProtocolVersion = 1

// HelloMessage is optional negotiation; it is never a game command.
type HelloMessage struct {
	MessageResponse
	Proto  int    `json:"proto,omitempty"`
	Server string `json:"server,omitempty"`
	Status string `json:"status,omitempty"`
}

func NewHelloMessage(userID string, clientProto int) *HelloMessage {
	status := "ok"
	if clientProto < ProtocolVersion {
		status = "too-old"
	}
	if clientProto > ProtocolVersion {
		status = "upgrade-recommended"
	}
	version := "talesmud/dev"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" && setting.Value != "" {
				version = "talesmud/" + setting.Value
				break
			}
		}
	}
	return &HelloMessage{
		MessageResponse: MessageResponse{Audience: MessageAudienceOrigin, AudienceID: userID, Type: MessageTypeHello},
		Proto:           ProtocolVersion, Server: version, Status: status,
	}
}
