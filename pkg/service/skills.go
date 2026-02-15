package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/skills"
	r "github.com/talesmud/talesmud/pkg/repository"
)

// SkillsService provides logical functions on top of the skills repository.
type SkillsService interface {
	r.SkillsRepository
}

type skillsService struct {
	r.SkillsRepository
}

// NewSkillsService creates a new skills service.
// It loads all skills into the in-memory cache and seeds the DB if empty.
func NewSkillsService(repo r.SkillsRepository) SkillsService {
	svc := &skillsService{repo}

	// Seed DB if empty
	all, err := repo.FindAll()
	if err != nil {
		log.WithError(err).Error("SkillsService: failed to load skills from DB")
	} else if len(all) == 0 {
		log.Info("SkillsService: no skills in DB, seeding with defaults...")
		seed := skills.SeedSkills()
		for _, s := range seed {
			if _, err := repo.Import(s); err != nil {
				log.WithError(err).WithField("skill", s.Name).Error("SkillsService: failed to seed skill")
			}
		}
		// Reload after seeding
		all, err = repo.FindAll()
		if err != nil {
			log.WithError(err).Error("SkillsService: failed to reload skills after seeding")
		}
		log.WithField("count", len(all)).Info("SkillsService: seeded skills")
	}

	// Populate the in-memory cache
	skills.LoadFromDB(all)
	log.WithField("count", len(all)).Info("SkillsService: loaded skills into cache")

	return svc
}

// Store creates a new skill and refreshes the cache.
func (svc *skillsService) Store(skill *skills.Skill) (*skills.Skill, error) {
	result, err := svc.SkillsRepository.Store(skill)
	if err == nil {
		svc.refreshCache()
	}
	return result, err
}

// Import stores a skill with a pre-set ID and refreshes the cache.
func (svc *skillsService) Import(skill *skills.Skill) (*skills.Skill, error) {
	result, err := svc.SkillsRepository.Import(skill)
	if err == nil {
		svc.refreshCache()
	}
	return result, err
}

// Update updates a skill and refreshes the cache.
func (svc *skillsService) Update(id string, skill *skills.Skill) error {
	err := svc.SkillsRepository.Update(id, skill)
	if err == nil {
		svc.refreshCache()
	}
	return err
}

// Delete deletes a skill and refreshes the cache.
func (svc *skillsService) Delete(id string) error {
	err := svc.SkillsRepository.Delete(id)
	if err == nil {
		svc.refreshCache()
	}
	return err
}

func (svc *skillsService) refreshCache() {
	all, err := svc.SkillsRepository.FindAll()
	if err != nil {
		log.WithError(err).Error("SkillsService: failed to refresh cache")
		return
	}
	skills.RefreshCache(all)
}
