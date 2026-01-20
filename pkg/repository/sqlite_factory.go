package repository

import "github.com/talesmud/talesmud/pkg/db/sqlite"

// SQLiteFactory builds repositories backed by SQLite.
type SQLiteFactory struct {
	client *sqlite.Client
}

// NewSQLiteFactory returns a SQLite-backed repository factory.
func NewSQLiteFactory(client *sqlite.Client) *SQLiteFactory {
	return &SQLiteFactory{client: client}
}

func (f *SQLiteFactory) Characters() CharactersRepository {
	return NewSQLiteCharactersRepository(f.client)
}

func (f *SQLiteFactory) Parties() PartiesRepository {
	return NewSQLitePartiesRepository(f.client)
}

func (f *SQLiteFactory) Users() UsersRepository {
	return NewSQLiteUsersRepository(f.client)
}

func (f *SQLiteFactory) Rooms() RoomsRepository {
	return NewSQLiteRoomsRepository(f.client)
}

func (f *SQLiteFactory) Scripts() ScriptsRepository {
	return NewSQLiteScriptsRepository(f.client)
}

func (f *SQLiteFactory) Items() ItemsRepository {
	return NewSQLiteItemsRepository(f.client)
}

func (f *SQLiteFactory) ItemTemplates() ItemTemplatesRepository {
	return NewSQLiteItemTemplatesRepository(f.client)
}

func (f *SQLiteFactory) CharacterTemplates() CharacterTemplatesRepository {
	return NewSQLiteCharacterTemplatesRepository(f.client)
}

func (f *SQLiteFactory) NPCs() NPCsRepository {
	return NewSQLiteNPCsRepository(f.client)
}

func (f *SQLiteFactory) Dialogs() DialogsRepository {
	return NewSQLiteDialogsRepository(f.client)
}

func (f *SQLiteFactory) Conversations() ConversationsRepository {
	return NewSQLiteConversationsRepository(f.client)
}

func (f *SQLiteFactory) Close() error {
	return f.client.Close()
}
