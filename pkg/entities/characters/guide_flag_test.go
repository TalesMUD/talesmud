package characters

import "testing"

func TestGuideDismissedFlag(t *testing.T) {
	c := &Character{}
	if c.GuideDismissed() {
		t.Fatal("empty character looks dismissed")
	}
	c.SetGuideDismissed()
	if !c.GuideDismissed() {
		t.Fatal("set flag did not stick")
	}
	for _, v := range []interface{}{"true", "1", "yes", float64(1)} {
		c.Flags = map[string]interface{}{FlagGuideDismissed: v}
		if !c.GuideDismissed() {
			t.Fatalf("value %v should count as dismissed", v)
		}
	}
	c.Flags = map[string]interface{}{FlagGuideDismissed: "false"}
	if c.GuideDismissed() {
		t.Fatal("false string counted as dismissed")
	}
}
