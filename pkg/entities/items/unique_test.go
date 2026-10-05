package items

import "testing"

func TestTrimMatchingTemplateKeepsOne(t *testing.T) {
	inv := &Inventory{Items: []*Item{
		{TemplateID: "ITM0020", Name: "Sample Relic Shard"},
		{TemplateID: "ITM0020", Name: "Sample Relic Shard"},
		{TemplateID: "ITM0012", Name: "Weak Health Potion", Quantity: 2, Stackable: true},
	}}
	if n := inv.TrimMatchingTemplate("ITM0020"); n != 1 {
		t.Fatalf("removed %d, want 1", n)
	}
	if inv.CountMatchingTemplate("ITM0020") != 1 {
		t.Fatalf("count %d", inv.CountMatchingTemplate("ITM0020"))
	}
	if len(inv.Items) != 2 {
		t.Fatalf("len %d", len(inv.Items))
	}
	if n := inv.TrimMatchingTemplate("ITM0020"); n != 0 {
		t.Fatalf("second trim removed %d", n)
	}
	if inv.CountMatchingTemplate("ITM0012") != 2 {
		t.Fatal("potion stack changed")
	}
}

func TestTemplateKeyPrefersTemplateID(t *testing.T) {
	item := &Item{TemplateID: "ITM0020", Name: "Sample Relic Shard"}
	if TemplateKey(item) != "ITM0020" {
		t.Fatal(TemplateKey(item))
	}
	if TemplateKey(nil) != "" {
		t.Fatal("nil key")
	}
}
