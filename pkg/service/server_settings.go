package service

import (
	"github.com/talesmud/talesmud/pkg/entities/settings"
	r "github.com/talesmud/talesmud/pkg/repository"
)

// ServerSettingsService provides business logic for server settings.
type ServerSettingsService interface {
	Get() (*settings.ServerSettings, error)
	Update(s *settings.ServerSettings) error
}

type serverSettingsService struct {
	repo r.ServerSettingsRepository
}

// NewServerSettingsService creates a new server settings service.
func NewServerSettingsService(repo r.ServerSettingsRepository) ServerSettingsService {
	return &serverSettingsService{repo: repo}
}

func (srv *serverSettingsService) Get() (*settings.ServerSettings, error) {
	return srv.repo.Get()
}

func (srv *serverSettingsService) Update(s *settings.ServerSettings) error {
	return srv.repo.Upsert(s)
}
