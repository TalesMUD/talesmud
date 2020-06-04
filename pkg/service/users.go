package service

import (
	"time"

	"github.com/sirupsen/logrus"
	e "github.com/talesmud/talesmud/pkg/entities"
	r "github.com/talesmud/talesmud/pkg/repository"
)

//UsersService delives logical functions on top of the users Repo
type UsersService interface {
	r.UsersRepository

	FindOrCreateNewUser(id string) (*e.User, error)
}

type usersService struct {
	r.UsersRepository
}

//NewUsersService creates a new users service
func NewUsersService(usersRepository r.UsersRepository) UsersService {
	return &usersService{
		usersRepository,
	}
}

// FindOrCreateNewUser ...
func (us *usersService) FindOrCreateNewUser(refID string) (*e.User, error) {

	if user, err := us.FindByRefID(refID); err != nil {
		// Creating new user with id
		user := &e.User{
			Entity:    e.NewEntity(),
			RefID:     refID,
			Created:   time.Now(),
			LastSeen:  time.Now(),
			IsNewUser: true,
		}
		logrus.WithField("UserID", refID).Info("Creating new user")
		return us.Create(user)
	} else {
		return user, nil
	}
}