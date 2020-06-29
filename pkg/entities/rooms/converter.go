package rooms

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// RoomFromJSONString ...
func RoomFromJSONString(input string) (*Room, error) {
	bytes := []byte(input)
	var room Room
	if err := json.Unmarshal(bytes, &room); err != nil {
		logrus.WithField("input", input).WithError(err).Info("Could not unmarshal room from string")
		return nil, err
	}
	return &room, nil
}

// RoomsFromJSONString ...
func RoomsFromJSONString(input string) (Rooms, error) {
	bytes := []byte(input)
	var its Rooms
	if err := json.Unmarshal(bytes, its); err != nil {
		logrus.WithField("input", input).WithError(err).Info("Could not unmarshal rooms array from string")
		return nil, err
	}
	return its, nil
}

// RoomToJSONString ...
func RoomToJSONString(room Room) string {
	bytes, _ := json.MarshalIndent(room, "", " ")
	return string(bytes)
}

// RoomsToJSONString ...
func RoomsToJSONString(rooms Rooms) string {
	bytes, _ := json.MarshalIndent(rooms, "", " ")
	return string(bytes)
}
