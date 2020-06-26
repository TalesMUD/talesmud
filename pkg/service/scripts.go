package service

import (
	r "github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts"
)

//--- Interface Definitions

//ScriptsService delives logical functions on top of the charactersheets Repo
type ScriptsService interface {
	r.ScriptsRepository

	ScriptTypes() scripts.ScriptTypes
}

//--- Implementations

type scriptsService struct {
	r.ScriptsRepository
}

func (service *scriptsService) ScriptTypes() scripts.ScriptTypes {
	return scripts.ScriptTypes{
		scripts.ScriptTypeNone,
		scripts.ScriptTypeCustom,
		scripts.ScriptTypeItem,
		scripts.ScriptTypeRoom,
		scripts.ScriptTypeQuest,
		scripts.ScriptTypeNPC,
	}
}

//NewScriptsService creates a nwe item service
func NewScriptsService(repo r.ScriptsRepository) ScriptsService {
	return &scriptsService{
		repo,
	}
}
