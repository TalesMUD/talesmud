package messages

import (
	"encoding/json"
	"testing"
)

func TestCharacterSelectedGuestJSON(t *testing.T) {
	for _, guest := range []bool{false, true} {
		data, err := json.Marshal(CharacterSelected{Guest: guest})
		if err != nil {
			t.Fatal(err)
		}
		var wire map[string]interface{}
		if err := json.Unmarshal(data, &wire); err != nil {
			t.Fatal(err)
		}
		value, present := wire["guest"]
		if present != guest || (guest && value != true) {
			t.Fatalf("guest=%v JSON=%s", guest, data)
		}
	}
}
