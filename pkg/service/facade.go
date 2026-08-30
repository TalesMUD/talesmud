package service

import (
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts"
)

// Facade provides access to all services
type Facade interface {
	CharactersService() CharactersService
	PartiesService() PartiesService
	UsersService() UsersService
	RoomsService() RoomsService
	ScriptsService() ScriptsService
	ItemsService() ItemsService
	NPCsService() NPCsService
	NPCSpawnersService() NPCSpawnersService
	DialogsService() DialogsService
	ConversationsService() ConversationsService
	LootTablesService() LootTablesService
	ServerSettingsService() ServerSettingsService
	QuestsService() QuestsService
	SkillsService() SkillsService
	CharacterTemplatesRepo() repository.CharacterTemplatesRepository
	GuestService() GuestService
	GuestStatsService() GuestStatsService

	Runner() scripts.ScriptRunner
}

type facade struct {
	css   CharactersService
	ps    PartiesService
	us    UsersService
	rs    RoomsService
	is    ItemsService
	ss    ScriptsService
	ns    NPCsService
	nss   NPCSpawnersService
	ds    DialogsService
	convs ConversationsService
	lts   LootTablesService
	sss   ServerSettingsService
	qs    QuestsService
	skls  SkillsService
	gs    GuestService
	gss   GuestStatsService
	sr    scripts.ScriptRunner
	repos repository.Factory
}

// NewFacade creates a new service facade
func NewFacade(repos repository.Factory, runner scripts.ScriptRunner) Facade {
	// Create repositories
	charactersRepo := repos.Characters()
	partiesRepo := repos.Parties()
	usersRepo := repos.Users()
	roomsRepo := repos.Rooms()
	scriptsRepo := repos.Scripts()
	itemsRepo := repos.Items()
	npcsRepo := repos.NPCs()
	npcSpawnersRepo := repos.NPCSpawners()
	dialogsRepo := repos.Dialogs()
	conversationsRepo := repos.Conversations()
	characterTemplatesRepo := repos.CharacterTemplates()
	lootTablesRepo := repos.LootTables()
	serverSettingsRepo := repos.ServerSettings()
	questsRepo := repos.Quests()
	questProgressRepo := repos.QuestProgress()
	skillsRepo := repos.Skills()

	// Create services
	ss := NewScriptsService(scriptsRepo)
	is := NewItemsService(itemsRepo)
	lts := NewLootTablesService(lootTablesRepo, is)
	qs := NewQuestsService(questsRepo, questProgressRepo)

	skls := NewSkillsService(skillsRepo)

	rs := NewRoomsService(roomsRepo)
	sss := NewServerSettingsService(serverSettingsRepo)

	f := &facade{
		css:   NewCharactersService(charactersRepo, characterTemplatesRepo, sss, rs),
		ps:    NewPartiesService(partiesRepo),
		us:    NewUsersService(usersRepo),
		rs:    rs,
		ss:    ss,
		is:    is,
		ns:    NewNPCsService(npcsRepo),
		nss:   NewNPCSpawnersService(npcSpawnersRepo),
		ds:    NewDialogsService(dialogsRepo),
		convs: NewConversationsService(conversationsRepo),
		lts:   lts,
		sss:   sss,
		qs:    qs,
		skls:  skls,
		sr:    runner,
		repos: repos,
	}

	// Set facade reference in quests service for reward granting
	qs.SetFacade(f)

	// Initialize guest stats service (uses its own repo, no facade dependency)
	guestStatsRepo := repos.GuestStats()
	f.gss = NewGuestStatsService(guestStatsRepo)

	// Initialize guest service (needs facade + stats service)
	f.gs = NewGuestService(f, f.gss)

	return f
}
func (f *facade) RoomsService() RoomsService {
	return f.rs
}
func (f *facade) CharactersService() CharactersService {
	return f.css
}

func (f *facade) ItemsService() ItemsService {
	return f.is
}
func (f *facade) ScriptsService() ScriptsService {
	return f.ss
}
func (f *facade) PartiesService() PartiesService {
	return f.ps
}
func (f *facade) UsersService() UsersService {
	return f.us
}
func (f *facade) Runner() scripts.ScriptRunner {
	return f.sr
}

func (f *facade) NPCsService() NPCsService {
	return f.ns
}

func (f *facade) NPCSpawnersService() NPCSpawnersService {
	return f.nss
}

func (f *facade) DialogsService() DialogsService {
	return f.ds
}

func (f *facade) ConversationsService() ConversationsService {
	return f.convs
}

func (f *facade) LootTablesService() LootTablesService {
	return f.lts
}

func (f *facade) ServerSettingsService() ServerSettingsService {
	return f.sss
}

func (f *facade) QuestsService() QuestsService {
	return f.qs
}

func (f *facade) SkillsService() SkillsService {
	return f.skls
}

func (f *facade) CharacterTemplatesRepo() repository.CharacterTemplatesRepository {
	return f.repos.CharacterTemplates()
}

func (f *facade) GuestService() GuestService {
	return f.gs
}

func (f *facade) GuestStatsService() GuestStatsService {
	return f.gss
}
