package messages

//MessageType type
type MessageType string

const (
	MessageTypeDefault           = "message"
	MessageTypeEnterRoom         = "enterRoom"
	MessageTypeCreateCharacter   = "createCharacter"
	MessageTypeSelectCharacter   = "selectCharacter"
	MessageTypeCharacterSelected = "characterSelected"

	MessageTypePing = "ping"
)
