package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/skills"
)

type sqliteSkillsRepository struct {
	*sqliteGenericRepo
}

// NewSQLiteSkillsRepository creates a new SQLite skills repository.
func NewSQLiteSkillsRepository(client *dbsqlite.Client) SkillsRepository {
	return &sqliteSkillsRepository{
		sqliteGenericRepo: newSQLiteGenericRepo(client.DB(), "skills", func() interface{} {
			return &skills.Skill{}
		}),
	}
}

func (repo *sqliteSkillsRepository) Drop() error {
	return repo.sqliteGenericRepo.DropCollection()
}

func (repo *sqliteSkillsRepository) FindByID(id string) (*skills.Skill, error) {
	if id == "" {
		log.Error("Skills::FindByID - id is empty")
		return nil, errors.New("empty id")
	}
	result, err := repo.sqliteGenericRepo.FindByID(id)
	if err == nil {
		return result.(*skills.Skill), nil
	}
	return nil, err
}

func (repo *sqliteSkillsRepository) FindAll() ([]*skills.Skill, error) {
	results := make([]*skills.Skill, 0)
	if err := repo.sqliteGenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*skills.Skill))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteSkillsRepository) Update(id string, skill *skills.Skill) error {
	return repo.sqliteGenericRepo.Update(skill, id)
}

func (repo *sqliteSkillsRepository) Delete(id string) error {
	return repo.sqliteGenericRepo.Delete(id)
}

func (repo *sqliteSkillsRepository) Store(skill *skills.Skill) (*skills.Skill, error) {
	skill.Entity = entities.NewEntity()
	return repo.Import(skill)
}

func (repo *sqliteSkillsRepository) Import(skill *skills.Skill) (*skills.Skill, error) {
	result, err := repo.sqliteGenericRepo.Store(skill)
	if result == nil {
		return nil, err
	}
	return result.(*skills.Skill), nil
}
