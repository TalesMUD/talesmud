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

	// Combat messages
	MessageTypeCombatStart  = "combatStart"  // Combat initiated
	MessageTypeCombatTurn   = "combatTurn"   // It's your turn
	MessageTypeCombatAction = "combatAction" // Action result (attack, defend, etc.)
	MessageTypeCombatEnd    = "combatEnd"    // Combat ended (victory, defeat, fled)
	MessageTypeCombatStatus = "combatStatus" // Combat status update
)
