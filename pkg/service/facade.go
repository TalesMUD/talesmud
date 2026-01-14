package service

import (
	"github.com/talesmud/talesmud/pkg/db"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts"
)

// Facade provides access to all services
type Facade interface {
	CharactersService() CharactersService
	PartiesService() PartiesService
	UsersService() UsersService
	RoomsService() RoomsService
	ScriptsService() ScriptsService
	ItemsService() ItemsService
	NPCsService() NPCsService
	DialogsService() DialogsService
	ConversationsService() ConversationsService

	Runner() scripts.ScriptRunner
}

type facade struct {
	css   CharactersService
	ps    PartiesService
	us    UsersService
	rs    RoomsService
	is    ItemsService
	ss    ScriptsService
	ns    NPCsService
	ds    DialogsService
	convs ConversationsService
	sr    scripts.ScriptRunner
	db    *db.Client
}

// NewFacade creates a new service facade
func NewFacade(db *db.Client, runner scripts.ScriptRunner) Facade {
	// Create repositories
	charactersRepo := repository.NewMongoDBcharactersRepository(db)
	partiesRepo := repository.NewMongoDBPartiesRepository(db)
	usersRepo := repository.NewMongoDBUsersRepository(db)
	roomsRepo := repository.NewMongoDBRoomsRepository(db)
	scriptsRepo := repository.NewMongoDBScriptRepository(db)
	itemsRepo := repository.NewMongoDBItemsRepository(db)
	itemTemplatesRepo := repository.NewMongoDBItemTemplatesRepository(db)
	npcsRepo := repository.NewMongoDBNPCsRepository(db)
	dialogsRepo := repository.NewMongoDBDialogsRepository(db)
	conversationsRepo := repository.NewMongoDBConversationsRepository(db)

	// Create services
	ss := NewScriptsService(scriptsRepo)
	is := NewItemsService(itemsRepo, itemTemplatesRepo, ss, runner)

	return &facade{
		css:   NewCharactersService(charactersRepo),
		ps:    NewPartiesService(partiesRepo),
		us:    NewUsersService(usersRepo),
		rs:    NewRoomsService(roomsRepo),
		ss:    ss,
		is:    is,
		ns:    NewNPCsService(npcsRepo),
		ds:    NewDialogsService(dialogsRepo),
		convs: NewConversationsService(conversationsRepo),
		sr:    runner,
	}
}
func (f *facade) RoomsService() RoomsService {
	return f.rs
}
func (f *facade) CharactersService() CharactersService {
	return f.css
}

func (f *facade) ItemsService() ItemsService {
	return f.is
}
func (f *facade) ScriptsService() ScriptsService {
	return f.ss
}
func (f *facade) PartiesService() PartiesService {
	return f.ps
}
func (f *facade) UsersService() UsersService {
	return f.us
}
func (f *facade) Runner() scripts.ScriptRunner {
	return f.sr
}

func (f *facade) NPCsService() NPCsService {
	return f.ns
}

func (f *facade) DialogsService() DialogsService {
	return f.ds
}

func (f *facade) ConversationsService() ConversationsService {
	return f.convs
}
