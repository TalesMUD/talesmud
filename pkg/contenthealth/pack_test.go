package contenthealth

import (
	"path/filepath"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

func TestPackEvaluatorAggroFixture(t *testing.T) {
	rules, err := LoadPackRules(filepath.Join("testdata"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 || rules[0].ID != "aggro-requires-level-10" {
		t.Fatalf("rules = %+v", rules)
	}
	rule := rules[0]

	low := &npc.NPC{
		Entity: &entities.Entity{ID: "A"}, Name: "A", Level: 9,
		EnemyTrait: &npc.EnemyTrait{AggroOnSight: true, Difficulty: "normal"},
	}
	if _, ok := packHit(rule, low, "npc", "A", "A"); !ok {
		t.Fatal("level 9 aggro should fail")
	}
	boss := &npc.NPC{
		Entity: &entities.Entity{ID: "B"}, Name: "B", Level: 12,
		EnemyTrait: &npc.EnemyTrait{AggroOnSight: true, Difficulty: "boss"},
	}
	if _, ok := packHit(rule, boss, "npc", "B", "B"); !ok {
		t.Fatal("aggro boss should fail")
	}
	okNPC := &npc.NPC{
		Entity: &entities.Entity{ID: "C"}, Name: "C", Level: 10,
		EnemyTrait: &npc.EnemyTrait{AggroOnSight: true, Difficulty: "normal"},
	}
	if _, ok := packHit(rule, okNPC, "npc", "C", "C"); ok {
		t.Fatal("level 10 non-boss should pass")
	}
	quiet := &npc.NPC{
		Entity: &entities.Entity{ID: "D"}, Name: "D", Level: 1,
		EnemyTrait: &npc.EnemyTrait{AggroOnSight: false, Difficulty: "normal"},
	}
	if _, ok := packHit(rule, quiet, "npc", "D", "D"); ok {
		t.Fatal("non-aggro NPC should be skipped")
	}
	missingLevel := map[string]interface{}{
		"enemyTrait": map[string]interface{}{"aggroOnSight": true, "difficulty": "normal"},
	}
	if _, ok := packHit(rule, missingLevel, "npc", "E", "E"); !ok {
		t.Fatal("missing level should fail gte")
	}
	missingDifficulty := map[string]interface{}{
		"level":      12,
		"enemyTrait": map[string]interface{}{"aggroOnSight": true},
	}
	if _, ok := packHit(rule, missingDifficulty, "npc", "F", "F"); ok {
		t.Fatal("missing difficulty should pass ne")
	}
}

func TestPackOperators(t *testing.T) {
	rule := PackRule{
		ID: "ops", Severity: "warning", Entity: "item", Message: "nope",
		Require: []Condition{
			{Field: "name", Op: "regex", Value: "^Sword"},
			{Field: "level", Op: "lte", Value: 5},
			{Field: "slot", Op: "in", Value: []interface{}{"main_hand", "off_hand"}},
			{Field: "detail", Op: "exists", Value: true},
		},
	}
	bad := map[string]interface{}{"name": "Axe", "level": 9, "slot": "head"}
	hit, ok := packHit(rule, bad, "item", "I", "I")
	if !ok || hit.Field != "name" {
		t.Fatalf("hit = %+v ok=%v", hit, ok)
	}
	good := map[string]interface{}{"name": "Sword", "level": 5, "slot": "main_hand", "detail": "sharp"}
	if _, ok := packHit(rule, good, "item", "J", "J"); ok {
		t.Fatal("matching item should pass")
	}
	missing := map[string]interface{}{"level": 1}
	ne := PackRule{ID: "ne", Entity: "item", Require: []Condition{{Field: "slot", Op: "ne", Value: "head"}}}
	if _, ok := packHit(ne, missing, "item", "K", "K"); ok {
		t.Fatal("missing field should pass ne")
	}
}
