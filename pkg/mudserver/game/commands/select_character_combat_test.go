package commands_test

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/service"
)

// Regression: Marcus switched characters over SSH while his first character
// was fighting. The old fight kept running on auto-attack and its turn
// prompts, HP lines and defeat screen landed in the session now playing the
// second character, which could not act in that fight.
func setupSwitchFight(t *testing.T) (*game.Game, service.Facade, *entities.User, *characters.Character, *characters.Character) {
	t.Helper()
	g, facade := newSocialTestGame(t)
	user, fighter := storeSocialPlayer(t, facade, "user-sw", "ref-sw", "char-warrior", "WarriorKing", "room-sw")
	other := &characters.Character{
		Entity:      &entities.Entity{ID: "char-ranger"},
		Name:        "Longstride",
		BelongsUser: *traits.BelongsToUser(user.ID),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "room-sw"},
	}
	if _, err := facade.CharactersService().Import(other); err != nil {
		t.Fatal(err)
	}
	exits := rooms.Exits{}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity:     &entities.Entity{ID: "room-sw"},
		Name:       "Awakening Chamber",
		Exits:      &exits,
		Characters: &rooms.Characters{fighter.ID},
	}); err != nil {
		t.Fatal(err)
	}
	g.ConnectUserSession(user)
	g.SetUserSessionCharacter(user, fighter)
	return g, facade, user, fighter, other
}

func startRatFight(t *testing.T, g *game.Game, facade service.Facade, fighter *characters.Character) {
	t.Helper()
	enemy := &npc.NPC{
		Entity:           &entities.Entity{ID: "rat-sw"},
		Name:             "Catacomb Rat",
		CurrentHitPoints: 40,
		MaxHitPoints:     40,
		Level:            1,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 1, Difficulty: "trivial"},
	}
	enemy.CurrentRoomID = "room-sw"
	g.NPCManager.RegisterExistingNPC(enemy, "room-sw")
	inst := g.GetCombatEngine().InitiateCombat("room-sw", []*characters.Character{fighter}, []*npc.NPC{enemy})
	if inst == nil {
		t.Fatal("failed to initiate combat")
	}
	fighter.InCombat = true
	fighter.CombatInstanceID = inst.ID
	_ = facade.CharactersService().Update(fighter.ID, fighter)
}

func TestSelectCharacterRefusedDuringFight(t *testing.T) {
	g, facade, user, fighter, _ := setupSwitchFight(t)
	startRatFight(t, g, facade, fighter)
	_ = drainSocialMessages(g.SendMessage())

	msg := &messages.Message{FromUser: user, Character: fighter, Data: "sc Longstride"}
	if !(&commands.SelectCharacterCommand{}).Execute(g, msg) {
		t.Fatal("select not handled")
	}
	var refused bool
	for _, out := range drainSocialMessages(g.SendMessage()) {
		switch v := out.(type) {
		case *messages.CharacterSelected:
			t.Fatalf("switched characters mid-fight: %s", v.Message)
		case messages.MessageResponse:
			if strings.Contains(v.Message, "can't switch characters during a fight") {
				refused = true
			}
		}
	}
	if !refused {
		t.Fatal("expected a refusal message")
	}
	if user.LastCharacter != fighter.ID {
		t.Fatalf("session moved to %q", user.LastCharacter)
	}
	if !g.GetCombatEngine().IsPlayerInCombat(fighter.ID) {
		t.Fatal("fight should still belong to the active character")
	}
}

func TestSelectCharacterAllowedOutOfFight(t *testing.T) {
	g, _, user, fighter, other := setupSwitchFight(t)
	_ = drainSocialMessages(g.SendMessage())
	msg := &messages.Message{FromUser: user, Character: fighter, Data: "sc Longstride"}
	(&commands.SelectCharacterCommand{}).Execute(g, msg)
	var selected bool
	for _, out := range drainSocialMessages(g.SendMessage()) {
		if v, ok := out.(*messages.CharacterSelected); ok && v.Character != nil && v.Character.ID == other.ID {
			selected = true
		}
	}
	if !selected {
		t.Fatal("switch outside a fight should work")
	}
}

func TestSelectCharacterRejectsOtherUsersCharacter(t *testing.T) {
	g, facade, _, _, _ := setupSwitchFight(t)
	guest := &entities.User{Entity: &entities.Entity{ID: "guest-1"}, RefID: "guest|1", IsGuest: true}
	if _, err := facade.UsersService().Import(guest); err != nil {
		t.Fatal(err)
	}
	g.ConnectUserSession(guest)
	_ = drainSocialMessages(g.SendMessage())
	(&commands.SelectCharacterCommand{}).Execute(g, &messages.Message{FromUser: guest, Data: "sc WarriorKing"})
	for _, out := range drainSocialMessages(g.SendMessage()) {
		if _, ok := out.(*messages.CharacterSelected); ok {
			t.Fatal("guest selected another account's character")
		}
	}
}

func TestBareSkillNameOutsideFightIsAHint(t *testing.T) {
	g, _, user, fighter, _ := setupSwitchFight(t)
	skills.RefreshCache([]*skills.Skill{{Entity: &entities.Entity{ID: "ranger_aimed_shot"}, Name: "Aimed Shot", ClassIDs: []string{"ranger"}, LevelRequired: 1}})
	fighter.EquippedSkills = []string{"ranger_aimed_shot"}
	_ = drainSocialMessages(g.SendMessage())

	if !commands.BareSkill(g, &messages.Message{FromUser: user, Character: fighter, Data: "Aimed Shot"}) {
		t.Fatal("bare skill name was not handled")
	}
	var hinted bool
	for _, out := range drainSocialMessages(g.SendMessage()) {
		if v, ok := out.(messages.MessageResponse); ok && strings.Contains(v.Message, "not in combat") {
			hinted = true
		}
	}
	if !hinted {
		t.Fatal("expected a not-in-combat hint")
	}
	if commands.BareSkill(g, &messages.Message{FromUser: user, Character: fighter, Data: "hello there"}) {
		t.Fatal("ordinary chat must not be taken as a skill")
	}
	if commands.BareSkill(g, &messages.Message{FromUser: user, Character: fighter, Data: "aimed shots are cool"}) {
		t.Fatal("a sentence starting with a skill-like word is chat")
	}
}
