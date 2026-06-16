package items

import "testing"

func TestItemsFromJSONStringParsesArray(t *testing.T) {
	got, err := ItemsFromJSONString(`[{"id":"item-1","name":"Torch"},{"id":"item-2","name":"Rope"}]`)
	if err != nil {
		t.Fatalf("ItemsFromJSONString returned error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	if got[0].ID != "item-1" || got[0].Name != "Torch" {
		t.Fatalf("first item not populated: %#v", got[0])
	}
	if got[1].ID != "item-2" || got[1].Name != "Rope" {
		t.Fatalf("second item not populated: %#v", got[1])
	}
}
