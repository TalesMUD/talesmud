package service

import (
	r "github.com/talesmud/talesmud/pkg/repository"
)

// RoomValueHelpEntry ...
type RoomValueHelpEntry struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
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

// Implementation of the Value Help interface for rooms using the RoomValueHelpEntry object
func (service *roomsService) ValueHelp() ([]RoomValueHelpEntry, error) {

	results := make([]RoomValueHelpEntry, 0)

	if rooms, err := service.FindAll(); err == nil {
		for i := range rooms {
			results = append(results, RoomValueHelpEntry{
				ID:   rooms[i].ID,
				Name: rooms[i].Name,
			})
		}
	}
	return results, nil
}
