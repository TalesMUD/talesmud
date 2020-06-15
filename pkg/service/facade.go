package service

import (
	"github.com/talesmud/talesmud/pkg/db"
	"github.com/talesmud/talesmud/pkg/repository"
)

//Facade ...
type Facade interface {
	CharactersService() CharactersService
	PartiesService() PartiesService
	UsersService() UsersService
	RoomsService() RoomsService

	ItemsService() ItemsService
}

type facade struct {
	css CharactersService
	ps  PartiesService
	us  UsersService
	rs  RoomsService
	is  ItemsService
	db  *db.Client
}

//NewFacade creates a new service facade
func NewFacade(db *db.Client) Facade {
	charactersRepo := repository.NewMongoDBcharactersRepository(db)
	partiesRepo := repository.NewMongoDBPartiesRepository(db)
	usersRepo := repository.NewMongoDBUsersRepository(db)
	roomsRepo := repository.NewMongoDBRoomsRepository(db)

	itemsRepo := repository.NewMongoDBItemsRepository(db)
	itemTemplatesRepo := repository.NewMongoDBItemTemplatesRepository(db)

	return &facade{
		css: NewCharactersService(charactersRepo),
		ps:  NewPartiesService(partiesRepo),
		us:  NewUsersService(usersRepo),
		rs:  NewRoomsService(roomsRepo),
		is:  NewItemsService(itemsRepo, itemTemplatesRepo),
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

func (f *facade) PartiesService() PartiesService {
	return f.ps
}
func (f *facade) UsersService() UsersService {
	return f.us
}
