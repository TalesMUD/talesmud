package contenthealth

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/importer"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
	"gopkg.in/yaml.v3"
)

func TestEnemyExportKeepsAuthoredStatsAndShowsEffective(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	n := &npc.NPC{
		Entity:           &entities.Entity{ID: "ENM0007"},
		Name:             "Sample Thug",
		Level:            2,
		MaxHitPoints:     48,
		CurrentHitPoints: 48,
		EnemyTrait: &npc.EnemyTrait{
			Difficulty:  "normal",
			AttackPower: 5,
			Defense:     1,
			BaseStats:   &npc.BaseStats{MaxHitPoints: 18, AttackPower: 4, Defense: 1},
		},
	}
	out, err := MarshalYAML("npc", n)
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	if strings.Contains(text, "maxHitPoints: 48") || strings.Contains(text, "attackPower: 5") {
		t.Fatalf("export wrote scaled stats into importer fields:\n%s", text)
	}
	var body map[string]any
	if err := yaml.Unmarshal(out, &body); err != nil {
		t.Fatal(err)
	}
	combat, _ := body["combat"].(map[string]any)
	if combat == nil {
		t.Fatalf("missing combat block:\n%s", text)
	}
	authored, _ := combat["authored"].(map[string]any)
	effective, _ := combat["effective"].(map[string]any)
	if num(authored["hp"]) != 18 || num(authored["attack"]) != 4 || num(authored["defense"]) != 1 {
		t.Fatalf("authored %+v", authored)
	}
	if num(effective["hp"]) != 20 || num(effective["attack"]) != 7 || num(effective["defense"]) != 1 {
		t.Fatalf("effective %+v", effective)
	}

	var back importer.YAMLNPC
	if err := yaml.Unmarshal(out, &back); err != nil {
		t.Fatal(err)
	}
	again := back.ToEntity()
	if again.MaxHitPoints != 20 || again.EnemyTrait.AttackPower != 7 {
		t.Fatalf("reimport %d/%d", again.MaxHitPoints, again.EnemyTrait.AttackPower)
	}
	out2, err := MarshalYAML("npc", again)
	if err != nil {
		t.Fatal(err)
	}
	var back2 importer.YAMLNPC
	if err := yaml.Unmarshal(out2, &back2); err != nil {
		t.Fatal(err)
	}
	if back2.MaxHitPoints != 18 || back2.EnemyTrait.AttackPower != 4 {
		t.Fatalf("second export scaled again: hp %d atk %d", back2.MaxHitPoints, back2.EnemyTrait.AttackPower)
	}
	twice := back2.ToEntity()
	if twice.MaxHitPoints != 20 {
		t.Fatalf("double apply HP %d", twice.MaxHitPoints)
	}
}

func num(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case uint64:
		return int(n)
	case float64:
		return int(n)
	default:
		return -1
	}
}
