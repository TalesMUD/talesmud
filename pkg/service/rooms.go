package service

import (
	r "github.com/talesmud/talesmud/pkg/repository"
)

// RoomValueHelpEntry ...
type RoomValueHelpEntry struct {
	ID   string
	Name string
}

//RoomsService delives logical functions on top of the rooms Repo
type RoomsService interface {
	r.RoomsRepository

	ValueHelp() ([]RoomValueHelpEntry, error)
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

func (service *roomsService) ValueHelp() ([]RoomValueHelpEntry, error) {

	results := make([]RoomValueHelpEntry, 0)

	if rooms, err := service.FindAll(); err == nil {

		for _, room := range rooms {
			results = append(results, RoomValueHelpEntry{
				ID:   room.ID,
				Name: room.Name,
			})
		}
	}
	return results, nil
}
