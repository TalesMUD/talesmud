package service

import (
	"github.com/talesmud/talesmud/pkg/db"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts"
)

//Facade ...
type Facade interface {
	CharactersService() CharactersService
	PartiesService() PartiesService
	UsersService() UsersService
	RoomsService() RoomsService
	ScriptsService() ScriptsService
	ItemsService() ItemsService

	Runner() scripts.ScriptRunner
}

type facade struct {
	css CharactersService
	ps  PartiesService
	us  UsersService
	rs  RoomsService
	is  ItemsService
	ss  ScriptsService
	sr  scripts.ScriptRunner
	db  *db.Client
}

//NewFacade creates a new service facade
func NewFacade(db *db.Client, runner scripts.ScriptRunner) Facade {
	charactersRepo := repository.NewMongoDBcharactersRepository(db)
	partiesRepo := repository.NewMongoDBPartiesRepository(db)
	usersRepo := repository.NewMongoDBUsersRepository(db)
	roomsRepo := repository.NewMongoDBRoomsRepository(db)
	scriptsRepo := repository.NewMongoDBScriptRepository(db)
	ss := NewScriptsService(scriptsRepo)
	itemsRepo := repository.NewMongoDBItemsRepository(db)
	itemTemplatesRepo := repository.NewMongoDBItemTemplatesRepository(db)

	is := NewItemsService(itemsRepo, itemTemplatesRepo, ss, runner)

	return &facade{
		css: NewCharactersService(charactersRepo),
		ps:  NewPartiesService(partiesRepo),
		us:  NewUsersService(usersRepo),
		rs:  NewRoomsService(roomsRepo),
		ss:  ss,
		is:  is,
		sr:  runner,
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
