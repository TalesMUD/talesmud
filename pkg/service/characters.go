package service

import (
	r "github.com/talesmud/talesmud/pkg/repository"
)

//--- Interface Definitions

//CharactersService delives logical functions on top of the charactersheets Repo
type CharactersService interface {
	r.CharactersRepository
}

//--- Implementations

type charactersService struct {
	r.CharactersRepository
}

//NewCharactersService creates a nwe item service
func NewCharactersService(charactersRepo r.CharactersRepository) CharactersService {
	return &charactersService{
		charactersRepo,
	}
}
