package service

import (
	"errors"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	e "github.com/talesmud/talesmud/pkg/entities"
	r "github.com/talesmud/talesmud/pkg/repository"
)

// UsersService delivers logical functions on top of the users Repo
type UsersService interface {
	r.UsersRepository

	FindOrCreateNewUser(id string) (*e.User, error)
	IsOnline(id string) bool
	SetRole(userID string, role string) error
	BanUser(userID string) error
	UnbanUser(userID string) error
}

type usersService struct {
	r.UsersRepository
}

// NewUsersService creates a new users service
func NewUsersService(usersRepository r.UsersRepository) UsersService {
	return &usersService{
		usersRepository,
	}
}

// IsOnline returns the online status as bool
func (us *usersService) IsOnline(id string) bool {

	if user, err := us.UsersRepository.FindByID(id); err == nil {
		return user.IsOnline
	}

	return false
}

// FindOrCreateNewUser finds an existing user by RefID or creates a new one.
// It also syncs the admin role from the MUD_ADMIN_OAUTHID environment variable.
func (us *usersService) FindOrCreateNewUser(refID string) (*e.User, error) {

	var user *e.User
	var err error

	if user, err = us.FindByRefID(refID); err != nil {
		// Creating new user with id
		user = &e.User{
			Entity:    e.NewEntity(),
			RefID:     refID,
			Created:   time.Now(),
			LastSeen:  time.Now(),
			IsNewUser: true,
		}
		logrus.WithField("UserID", refID).Info("Creating new user")
		user, err = us.Create(user)
		if err != nil {
			return nil, err
		}
	}

	// Sync admin role from env var on every login
	adminRefID := os.Getenv("MUD_ADMIN_OAUTHID")
	if adminRefID != "" && user.RefID == adminRefID {
		if user.GetRole() != e.RoleAdmin {
			user.Role = e.RoleAdmin
			us.Update(user.RefID, user)
			logrus.WithField("UserID", user.ID).Info("Assigned admin role from MUD_ADMIN_OAUTHID")
		}
	} else if adminRefID != "" && user.GetRole() == e.RoleAdmin && user.RefID != adminRefID {
		// If this user was previously admin but env var now points elsewhere, demote to player
		user.Role = e.RolePlayer
		us.Update(user.RefID, user)
		logrus.WithField("UserID", user.ID).Info("Revoked admin role (MUD_ADMIN_OAUTHID changed)")
	}

	return user, nil
}

// SetRole sets the role for a user. Only "player" and "creator" are allowed via API.
// Admin role is managed exclusively via the MUD_ADMIN_OAUTHID env var.
func (us *usersService) SetRole(userID string, role string) error {
	if role != e.RolePlayer && role != e.RoleCreator {
		return errors.New("invalid role: must be 'player' or 'creator'")
	}

	user, err := us.UsersRepository.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if user.IsAdmin() {
		return errors.New("cannot change admin role via API")
	}

	user.Role = role
	return us.Update(user.RefID, user)
}

// BanUser bans a user by ID, storing their email for email-based enforcement.
func (us *usersService) BanUser(userID string) error {
	user, err := us.UsersRepository.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if user.IsAdmin() {
		return errors.New("cannot ban the admin")
	}

	user.IsBanned = true
	user.BannedEmail = user.Email
	return us.Update(user.RefID, user)
}

// UnbanUser removes the ban from a user by ID.
func (us *usersService) UnbanUser(userID string) error {
	user, err := us.UsersRepository.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	user.IsBanned = false
	user.BannedEmail = ""
	return us.Update(user.RefID, user)
}
