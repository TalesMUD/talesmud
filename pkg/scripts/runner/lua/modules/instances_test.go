package modules

import (
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func TestProcSpecFromLua(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	if err := L.DoString(`spec = {
		count = 2,
		returnRoom = "town",
		templates = {"glade", "thicket"},
		timeout = 15,
		seed = 4,
		encounters = { {id = "wolf", minLevel = 1, maxLevel = 3, weight = 2} }
	}`); err != nil {
		t.Fatal(err)
	}
	tbl, ok := L.GetGlobal("spec").(*lua.LTable)
	if !ok {
		t.Fatal("spec table missing")
	}
	spec, err := procSpecFromLua(tbl)
	if err != nil {
		t.Fatal(err)
	}
	if spec.Count != 2 || spec.ReturnRoomID != "town" || spec.Seed != 4 || spec.Timeout != 15*time.Second {
		t.Fatalf("%+v", spec)
	}
	if len(spec.TemplateIDs) != 2 || len(spec.Encounters) != 1 || spec.Encounters[0].TemplateID != "wolf" || spec.Encounters[0].Weight != 2 {
		t.Fatalf("%+v", spec)
	}
}
