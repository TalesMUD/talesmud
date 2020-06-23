package service

import (
	r "github.com/talesmud/talesmud/pkg/repository"
)

//--- Interface Definitions

//ScriptsService delives logical functions on top of the charactersheets Repo
type ScriptsService interface {
	r.ScriptsRepository
}

//--- Implementations

type scriptsService struct {
	r.ScriptsRepository
}

//NewScriptsService creates a nwe item service
func NewScriptsService(repo r.ScriptsRepository) ScriptsService {
	return &scriptsService{
		repo,
	}
}
