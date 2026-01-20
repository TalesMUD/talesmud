package repository

import (
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/conversations"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/scripts"
)

// ItemsQuery holds query parameters for filtering items.
type ItemsQuery struct {
	Name string `form:"name"`
	Type string `form:"type"`
	Slot string `form:"slot"`
}

// matches returns true if the item matches the query filters.
func (q ItemsQuery) matches(item *items.Item) bool {
	if q.Name != "" && item.Name != q.Name {
		return false
	}
	if q.Type != "" && string(item.Type) != q.Type {
		return false
	}
	if q.Slot != "" && string(item.Slot) != q.Slot {
		return false
	}
	return true
}

// RoomsQuery holds query parameters for filtering rooms.
type RoomsQuery struct {
	Name string `form:"name"`
	Area string `form:"area"`
}

// CharactersRepository provides access to character data.
type CharactersRepository interface {
	Drop() error
	FindByID(id string) (*characters.Character, error)
	FindAllForUser(userID string) ([]*characters.Character, error)
	FindByName(name string) ([]*characters.Character, error)
	FindAll() ([]*characters.Character, error)
	Update(id string, character *characters.Character) error
	Delete(id string) error
	Store(character *characters.Character) (*characters.Character, error)
	Import(character *characters.Character) (*characters.Character, error)
}

// PartiesRepository provides access to party data.
type PartiesRepository interface {
	FindAll() ([]*entities.Party, error)
	Store(party *entities.Party) (*entities.Party, error)
	FindByID(id string) (*entities.Party, error)
	Update(id string, party *entities.Party) error
	Delete(id string) error
}

// UsersRepository provides access to user data.
type UsersRepository interface {
	Drop() error
	Import(user *entities.User) (*entities.User, error)
	Create(user *entities.User) (*entities.User, error)
	FindAll() ([]*entities.User, error)
	FindAllOnline() ([]*entities.User, error)
	Update(refID string, user *entities.User) error
	FindByID(id string) (*entities.User, error)
	FindByRefID(refID string) (*entities.User, error)
	Delete(id string) error
}

// RoomsRepository provides access to room data.
type RoomsRepository interface {
	Drop() error
	FindByID(id string) (*rooms.Room, error)
	FindByName(name string) ([]*rooms.Room, error)
	FindAll() ([]*rooms.Room, error)
	FindAllWithQuery(query RoomsQuery) ([]*rooms.Room, error)
	Update(id string, room *rooms.Room) error
	Delete(id string) error
	Store(room *rooms.Room) (*rooms.Room, error)
	Import(room *rooms.Room) (*rooms.Room, error)
}

// ScriptsRepository provides access to script data.
type ScriptsRepository interface {
	Drop() error
	FindByID(id string) (*scripts.Script, error)
	FindByName(name string) ([]*scripts.Script, error)
	FindAll() ([]*scripts.Script, error)
	Update(id string, script *scripts.Script) error
	Delete(id string) error
	Store(script *scripts.Script) (*scripts.Script, error)
	Import(script *scripts.Script) (*scripts.Script, error)
}

// ItemsRepository provides access to item data.
type ItemsRepository interface {
	Drop() error
	FindByID(id string) (*items.Item, error)
	FindByName(name string) ([]*items.Item, error)
	FindAll(query ItemsQuery) ([]*items.Item, error)
	Update(id string, item *items.Item) error
	Delete(id string) error
	Store(item *items.Item) (*items.Item, error)
	Import(item *items.Item) (*items.Item, error)
}

// ItemTemplatesRepository provides access to item template data.
type ItemTemplatesRepository interface {
	Drop() error
	FindByID(id string) (*items.ItemTemplate, error)
	FindByName(name string) ([]*items.ItemTemplate, error)
	FindAll(query ItemsQuery) ([]*items.ItemTemplate, error)
	Update(id string, item *items.ItemTemplate) error
	Delete(id string) error
	Store(item *items.ItemTemplate) (*items.ItemTemplate, error)
	Import(item *items.ItemTemplate) (*items.ItemTemplate, error)
}

// CharacterTemplatesRepository provides access to character template data.
type CharacterTemplatesRepository interface {
	Drop() error
	FindByID(id string) (*characters.CharacterTemplate, error)
	FindByName(name string) ([]*characters.CharacterTemplate, error)
	FindAll() ([]*characters.CharacterTemplate, error)
	Count() (int, error)
	Update(id string, template *characters.CharacterTemplate) error
	Delete(id string) error
	Store(template *characters.CharacterTemplate) (*characters.CharacterTemplate, error)
	Import(template *characters.CharacterTemplate) (*characters.CharacterTemplate, error)
}

// NPCsRepository provides access to NPC data.
type NPCsRepository interface {
	FindAll() ([]*npc.NPC, error)
	FindByID(id string) (*npc.NPC, error)
	FindByName(name string) ([]*npc.NPC, error)
	FindByRoom(roomID string) ([]*npc.NPC, error)
	Store(npc *npc.NPC) (*npc.NPC, error)
	Import(npc *npc.NPC) (*npc.NPC, error)
	Update(id string, npc *npc.NPC) error
	Delete(id string) error
	Drop() error
}

// DialogsRepository provides access to dialog data.
type DialogsRepository interface {
	FindAll() ([]*dialogs.Dialog, error)
	FindByID(id string) (*dialogs.Dialog, error)
	FindByName(name string) (*dialogs.Dialog, error)
	Store(dialog *dialogs.Dialog) (*dialogs.Dialog, error)
	Import(dialog *dialogs.Dialog) (*dialogs.Dialog, error)
	Update(id string, dialog *dialogs.Dialog) error
	Delete(id string) error
	Drop() error
}

// ConversationsRepository provides access to conversation data.
type ConversationsRepository interface {
	FindByID(id string) (*conversations.Conversation, error)
	FindByCharacterAndTarget(characterID, targetID string) (*conversations.Conversation, error)
	FindAllForCharacter(characterID string) ([]*conversations.Conversation, error)
	Store(conv *conversations.Conversation) (*conversations.Conversation, error)
	Update(id string, conv *conversations.Conversation) error
	Delete(id string) error
}
