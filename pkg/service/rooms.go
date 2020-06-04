package service

import (
	r "github.com/atla/owndnd/pkg/repository"
)

//RoomsService delives logical functions on top of the rooms Repo
type RoomsService interface {
	r.RoomsRepository
}

type roomsService struct {
	r.RoomsRepository
}

//NewRoomsService creates a new rooms service
func NewRoomsService(roomsRepository r.RoomsRepository) RoomsService {
	return &roomsService{
		roomsRepository,
	}
}
