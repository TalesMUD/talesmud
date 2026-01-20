package repository

import "github.com/talesmud/talesmud/pkg/db"

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

// MongoFactory builds repositories backed by MongoDB.
type MongoFactory struct {
	db *db.Client
}

// NewMongoFactory returns a Mongo-backed repository factory.
func NewMongoFactory(client *db.Client) *MongoFactory {
	return &MongoFactory{db: client}
}

func (f *MongoFactory) Characters() CharactersRepository {
	return NewMongoDBcharactersRepository(f.db)
}

func (f *MongoFactory) Parties() PartiesRepository {
	return NewMongoDBPartiesRepository(f.db)
}

func (f *MongoFactory) Users() UsersRepository {
	return NewMongoDBUsersRepository(f.db)
}

func (f *MongoFactory) Rooms() RoomsRepository {
	return NewMongoDBRoomsRepository(f.db)
}

func (f *MongoFactory) Scripts() ScriptsRepository {
	return NewMongoDBScriptRepository(f.db)
}

func (f *MongoFactory) Items() ItemsRepository {
	return NewMongoDBItemsRepository(f.db)
}

func (f *MongoFactory) ItemTemplates() ItemTemplatesRepository {
	return NewMongoDBItemTemplatesRepository(f.db)
}

func (f *MongoFactory) CharacterTemplates() CharacterTemplatesRepository {
	return NewMongoDBCharacterTemplatesRepository(f.db)
}

func (f *MongoFactory) NPCs() NPCsRepository {
	return NewMongoDBNPCsRepository(f.db)
}

func (f *MongoFactory) Dialogs() DialogsRepository {
	return NewMongoDBDialogsRepository(f.db)
}

func (f *MongoFactory) Conversations() ConversationsRepository {
	return NewMongoDBConversationsRepository(f.db)
}

func (f *MongoFactory) Close() error {
	return f.db.Close()
}
