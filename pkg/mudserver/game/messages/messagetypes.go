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

	// Dialog messages
	MessageTypeDialog    = "dialog"    // NPC dialog with options
	MessageTypeDialogEnd = "dialogEnd" // Conversation ended
)
