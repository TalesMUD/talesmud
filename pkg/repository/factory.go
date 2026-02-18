package repository

// Factory provides repository instances for a given storage backend.
type Factory interface {
	Characters() CharactersRepository
	Parties() PartiesRepository
	Users() UsersRepository
	Rooms() RoomsRepository
	Scripts() ScriptsRepository
	Items() ItemsRepository
	CharacterTemplates() CharacterTemplatesRepository
	NPCs() NPCsRepository
	NPCSpawners() NPCSpawnersRepository
	Dialogs() DialogsRepository
	Conversations() ConversationsRepository
	LootTables() LootTablesRepository
	ServerSettings() ServerSettingsRepository
	Quests() QuestsRepository
	QuestProgress() QuestProgressRepository
	Skills() SkillsRepository
	GuestStats() GuestStatsRepository
	Close() error
}
