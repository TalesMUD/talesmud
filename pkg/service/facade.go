package service

import (
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
	CharacterTemplatesRepo() repository.CharacterTemplatesRepository

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
	repos repository.Factory
}

// NewFacade creates a new service facade
func NewFacade(repos repository.Factory, runner scripts.ScriptRunner) Facade {
	// Create repositories
	charactersRepo := repos.Characters()
	partiesRepo := repos.Parties()
	usersRepo := repos.Users()
	roomsRepo := repos.Rooms()
	scriptsRepo := repos.Scripts()
	itemsRepo := repos.Items()
	itemTemplatesRepo := repos.ItemTemplates()
	npcsRepo := repos.NPCs()
	dialogsRepo := repos.Dialogs()
	conversationsRepo := repos.Conversations()
	characterTemplatesRepo := repos.CharacterTemplates()

	// Create services
	ss := NewScriptsService(scriptsRepo)
	is := NewItemsService(itemsRepo, itemTemplatesRepo, ss, runner)

	return &facade{
		css:   NewCharactersService(charactersRepo, characterTemplatesRepo),
		ps:    NewPartiesService(partiesRepo),
		us:    NewUsersService(usersRepo),
		rs:    NewRoomsService(roomsRepo),
		ss:    ss,
		is:    is,
		ns:    NewNPCsService(npcsRepo),
		ds:    NewDialogsService(dialogsRepo),
		convs: NewConversationsService(conversationsRepo),
		sr:    runner,
		repos: repos,
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

func (f *facade) CharacterTemplatesRepo() repository.CharacterTemplatesRepository {
	return f.repos.CharacterTemplates()
}
