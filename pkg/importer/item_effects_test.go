package importer

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAMLItemEffectsReachEntity(t *testing.T) {
	src := `
id: ITM9
name: Test Blade
onHitScriptId: SCR9
effects:
  - trigger: onHit
    name: Vigil Burn
    text: 2 damage each turn for 3 turns.
  - trigger: onHit
`
	var y YAMLItem
	if err := yaml.Unmarshal([]byte(src), &y); err != nil {
		t.Fatal(err)
	}
	item := y.ToEntity()
	if item.OnHitScriptID != "SCR9" {
		t.Fatalf("script id = %q", item.OnHitScriptID)
	}
	if len(item.Effects) != 1 || item.Effects[0].Name != "Vigil Burn" || item.Effects[0].Trigger != "onHit" {
		t.Fatalf("effects = %+v", item.Effects)
	}
}
