package contenthealth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

// LoadWorld reads stored content through repositories.
// Runtime copies stay in the lists; Run drops them from the rule snapshot.
func LoadWorld(repos repository.Factory, startRoomID, commit string) (World, error) {
	var world World
	var err error
	world.StartRoomID = startRoomID
	world.ContentCommit = commit
	world.Classes = CatalogClasses()
	if world.Rooms, err = repos.Rooms().FindAll(); err != nil {
		return world, fmt.Errorf("rooms: %w", err)
	}
	if world.NPCs, err = repos.NPCs().FindAll(); err != nil {
		return world, fmt.Errorf("npcs: %w", err)
	}
	if world.Items, err = repos.Items().FindAll(repository.ItemsQuery{}); err != nil {
		return world, fmt.Errorf("items: %w", err)
	}
	if world.LootTables, err = repos.LootTables().FindAll(); err != nil {
		return world, fmt.Errorf("loot tables: %w", err)
	}
	if world.Spawners, err = repos.NPCSpawners().FindAll(); err != nil {
		return world, fmt.Errorf("spawners: %w", err)
	}
	if world.Dialogs, err = repos.Dialogs().FindAll(); err != nil {
		return world, fmt.Errorf("dialogs: %w", err)
	}
	if world.Quests, err = repos.Quests().FindAll(); err != nil {
		return world, fmt.Errorf("quests: %w", err)
	}
	if world.Scripts, err = repos.Scripts().FindAll(); err != nil {
		return world, fmt.Errorf("scripts: %w", err)
	}
	if world.Skills, err = repos.Skills().FindAll(); err != nil {
		return world, fmt.Errorf("skills: %w", err)
	}
	if world.CharacterTemplates, err = repos.CharacterTemplates().FindAll(); err != nil {
		return world, fmt.Errorf("character templates: %w", err)
	}
	return world, nil
}

// RecordBaseline stores the entities just imported, plus the pack rules from that folder.
func RecordBaseline(repos repository.Factory, importPath string) error {
	world, err := LoadWorld(repos, "", ContentCommit(importPath))
	if err != nil {
		return err
	}
	rules, err := LoadPackRules(filepath.Join(importPath, "data", "rules"))
	if err != nil {
		return err
	}
	doc := Baseline{
		ContentCommit: world.ContentCommit,
		ImportedAt:    time.Now().UTC(),
		Rules:         rules,
		Entries:       baselineEntries(world),
	}
	payload, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	return repos.ContentHealth().Save(payload)
}

// ReadBaseline returns the stored baseline. A missing row is a nil baseline.
func ReadBaseline(repo repository.ContentHealthRepository) (*Baseline, error) {
	if repo == nil {
		return nil, nil
	}
	raw, err := repo.Get()
	if err != nil || len(raw) == 0 {
		return nil, err
	}
	var doc Baseline
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

// RunFromFacade checks the live database.
func RunFromFacade(facade service.Facade, repo repository.ContentHealthRepository) (Report, error) {
	settings, err := facade.ServerSettingsService().Get()
	if err != nil {
		return Report{}, err
	}
	start := ""
	var muted []string
	if settings != nil {
		start = settings.StartRoomID
		muted = append([]string(nil), settings.MutedHealthRuleIDs...)
	}
	world, err := worldFromFacade(facade, start)
	if err != nil {
		return Report{}, err
	}
	base, err := ReadBaseline(repo)
	if err != nil {
		return Report{}, err
	}
	if base != nil {
		world.ContentCommit = base.ContentCommit
	}
	live, err := liveView(facade, world.Rooms)
	if err != nil {
		return Report{}, err
	}
	opt := Options{Muted: muted, Live: live, Baseline: base}
	if base != nil {
		opt.Rules = base.Rules
	}
	if wd, err := os.Getwd(); err == nil {
		opt.DeployDir = wd
	}
	return Run(world, opt), nil
}

// WorldFromFacade loads the live content snapshot.
// The start room is settings.StartRoomID when that value is set.
func WorldFromFacade(facade service.Facade) (World, error) {
	start := ""
	if facade != nil && facade.ServerSettingsService() != nil {
		settings, err := facade.ServerSettingsService().Get()
		if err != nil {
			return World{}, err
		}
		if settings != nil {
			start = settings.StartRoomID
		}
	}
	return worldFromFacade(facade, start)
}

func worldFromFacade(facade service.Facade, start string) (World, error) {
	var world World
	var err error
	world.StartRoomID = start
	world.Classes = CatalogClasses()
	if world.Rooms, err = facade.RoomsService().FindAll(); err != nil {
		return world, fmt.Errorf("rooms: %w", err)
	}
	if world.NPCs, err = facade.NPCsService().FindAll(); err != nil {
		return world, fmt.Errorf("npcs: %w", err)
	}
	if world.Items, err = facade.ItemsService().FindAll(repository.ItemsQuery{}); err != nil {
		return world, fmt.Errorf("items: %w", err)
	}
	if world.LootTables, err = facade.LootTablesService().FindAll(); err != nil {
		return world, fmt.Errorf("loot tables: %w", err)
	}
	if world.Spawners, err = facade.NPCSpawnersService().FindAll(); err != nil {
		return world, fmt.Errorf("spawners: %w", err)
	}
	if world.Dialogs, err = facade.DialogsService().FindAll(); err != nil {
		return world, fmt.Errorf("dialogs: %w", err)
	}
	if world.Quests, err = facade.QuestsService().FindAll(); err != nil {
		return world, fmt.Errorf("quests: %w", err)
	}
	if world.Scripts, err = facade.ScriptsService().FindAll(); err != nil {
		return world, fmt.Errorf("scripts: %w", err)
	}
	if world.Skills, err = facade.SkillsService().FindAll(); err != nil {
		return world, fmt.Errorf("skills: %w", err)
	}
	if world.CharacterTemplates, err = facade.CharacterTemplatesRepo().FindAll(); err != nil {
		return world, fmt.Errorf("character templates: %w", err)
	}
	return world, nil
}

func liveView(facade service.Facade, roomList []*rooms.Room) (*LiveView, error) {
	characters, err := facade.CharactersService().FindAll()
	if err != nil {
		return nil, err
	}
	live := &LiveView{RoomIDs: map[string]bool{}}
	for _, room := range roomList {
		if room == nil || room.Entity == nil || room.ID == "" {
			continue
		}
		live.RoomIDs[room.ID] = true
		if strings.Contains(room.ID, "~") {
			live.InstanceRoomIDs = append(live.InstanceRoomIDs, room.ID)
		}
	}
	for _, character := range characters {
		if character == nil || character.Entity == nil {
			continue
		}
		live.Characters = append(live.Characters, LiveCharacter{
			ID:            character.ID,
			Name:          character.Name,
			CurrentRoomID: character.CurrentRoomID,
		})
	}
	return live, nil
}
