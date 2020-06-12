package service

import (
	r "github.com/talesmud/talesmud/pkg/repository"
)

//--- Interface Definitions

//ItemsService delives logical functions on top of the charactersheets Repo
type ItemsService interface {
	r.ItemsRepository
}

//--- Implementations

type itemsService struct {
	r.ItemsRepository
}

//NewItemsService creates a nwe item service
func NewItemsService(itemsRepo r.ItemsRepository) ItemsService {
	return &itemsService{
		itemsRepo,
	}
}
