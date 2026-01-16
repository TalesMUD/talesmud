package events

// EventType defines the type of game event
type EventType string

// Player events
const (
	EventPlayerEnterRoom EventType = "player.enter_room"
	EventPlayerLeaveRoom EventType = "player.leave_room"
	EventPlayerJoin      EventType = "player.join"
	EventPlayerQuit      EventType = "player.quit"
	EventPlayerDeath     EventType = "player.death"
	EventPlayerLevelUp   EventType = "player.level_up"
)

// Item events
const (
	EventItemPickup  EventType = "item.pickup"
	EventItemDrop    EventType = "item.drop"
	EventItemUse     EventType = "item.use"
	EventItemEquip   EventType = "item.equip"
	EventItemUnequip EventType = "item.unequip"
	EventItemCreate  EventType = "item.create"
)

// NPC events
const (
	EventNPCDeath  EventType = "npc.death"
	EventNPCSpawn  EventType = "npc.spawn"
	EventNPCUpdate EventType = "npc.update"
	EventNPCIdle   EventType = "npc.idle"
)

// Dialog events
const (
	EventDialogStart  EventType = "dialog.start"
	EventDialogEnd    EventType = "dialog.end"
	EventDialogOption EventType = "dialog.option"
)

// Room events
const (
	EventRoomUpdate EventType = "room.update"
	EventRoomAction EventType = "room.action"
)

// Quest events
const (
	EventQuestStart    EventType = "quest.start"
	EventQuestComplete EventType = "quest.complete"
	EventQuestFail     EventType = "quest.fail"
	EventQuestProgress EventType = "quest.progress"
)

// Combat events
const (
	EventCombatStart  EventType = "combat.start"
	EventCombatEnd    EventType = "combat.end"
	EventCombatDamage EventType = "combat.damage"
)

// Timer events
const (
	EventTimerTick EventType = "timer.tick"
)

// AllEventTypes returns all available event types
func AllEventTypes() []EventType {
	return []EventType{
		EventPlayerEnterRoom,
		EventPlayerLeaveRoom,
		EventPlayerJoin,
		EventPlayerQuit,
		EventPlayerDeath,
		EventPlayerLevelUp,
		EventItemPickup,
		EventItemDrop,
		EventItemUse,
		EventItemEquip,
		EventItemUnequip,
		EventItemCreate,
		EventNPCDeath,
		EventNPCSpawn,
		EventNPCUpdate,
		EventNPCIdle,
		EventDialogStart,
		EventDialogEnd,
		EventDialogOption,
		EventRoomUpdate,
		EventRoomAction,
		EventQuestStart,
		EventQuestComplete,
		EventQuestFail,
		EventQuestProgress,
		EventCombatStart,
		EventCombatEnd,
		EventCombatDamage,
		EventTimerTick,
	}
}
