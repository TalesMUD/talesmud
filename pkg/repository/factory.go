package repository

// Factory provides repository instances for a given storage backend.
type Factory interface {
	Characters() CharactersRepository
	Parties() PartiesRepository
	Users() UsersRepository
	Rooms() RoomsRepository
	Scripts() ScriptsRepository
	Items() ItemsRepository
	ItemTemplates() ItemTemplatesRepository
	CharacterTemplates() CharacterTemplatesRepository
	NPCs() NPCsRepository
	Dialogs() DialogsRepository
	Conversations() ConversationsRepository
	Close() error
}
