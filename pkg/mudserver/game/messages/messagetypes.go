package messages

// MessageType type
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

	// Inventory messages
	MessageTypeInventoryUpdate = "inventoryUpdate" // Inventory/equipment changed

	// Character update messages
	MessageTypeCharacterUpdate = "characterUpdate" // Character stats changed (HP, XP, etc.)
	MessageTypeLevelUp         = "levelUp"         // Player leveled up

	// Room update (silent refresh of exits/items/NPCs without re-rendering description)
	MessageTypeRoomUpdate = "roomUpdate"
	// Room presence update (silent refresh of online player list only)
	MessageTypeRoomPresence = "roomPresence"

	// Quest messages
	MessageTypeQuestAccepted  = "questAccepted"  // Quest accepted
	MessageTypeQuestProgress  = "questProgress"  // Quest objective progress updated
	MessageTypeQuestReady     = "questReady"     // Quest objectives complete and ready for turn-in
	MessageTypeQuestCompleted = "questCompleted" // Quest completed
	MessageTypeQuestAbandoned = "questAbandoned" // Quest abandoned
	MessageTypeQuestLog       = "questLog"       // Full quest log

	// Discovered-world atlas for web/mobile map widgets
	MessageTypeAtlas = "atlas"

	// Merchant shop overlay (structured stock for web/mobile)
	MessageTypeShop = "shop"

	// Crafting recipes overlay (structured list for web/mobile)
	MessageTypeRecipes = "recipes"

	// Friends list overlay (structured roster for web/mobile)
	MessageTypeFriends = "friends"

	// Party roster overlay (structured members for web/mobile)
	MessageTypeParty = "party"

	// Pending party invite toast/banner (Accept / Decline without typing)
	MessageTypePartyInvite = "party_invite"
)
