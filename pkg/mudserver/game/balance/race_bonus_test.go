package balance

import "testing"

func TestRacialWeaponBonusOnce(t *testing.T) {
	cases := []struct {
		race, weapon string
		in, want     int32
	}{
		{"dwarf", "mace", 20, 22},
		{"dwarf", "club", 20, 22},
		{"dwarf", "hammer", 20, 22},
		{"dwarf", "sword", 20, 20},
		{"dwarf", "bow", 20, 20},
		{"human", "mace", 20, 20},
		{"elf", "bow", 20, 22},
		{"elve", "bow", 20, 22},
		{"elves", "bow", 20, 22},
		{"elf", "mace", 20, 20},
		{"elf", "staff", 20, 20},
		{"construct", "mace", 20, 20},
		{"construct", "bow", 20, 20},
		{"orc", "mace", 20, 20},
		{"", "bow", 20, 20},
	}
	for _, tc := range cases {
		got := ApplyRacialWeaponBonus(tc.race, tc.weapon, tc.in)
		if got != tc.want {
			t.Fatalf("%s %s %d -> %d, want %d", tc.race, tc.weapon, tc.in, got, tc.want)
		}
		if tc.want == 22 && got == 24 {
			t.Fatalf("%s %s stacked the 10%%", tc.race, tc.weapon)
		}
	}
	if got := RacialWeaponMultiplier("dwarf", "mace"); got != 1.10 {
		t.Fatalf("dwarf blunt mult %v", got)
	}
	if got := RacialWeaponMultiplier("human", "mace"); got != 1 {
		t.Fatalf("human mult %v", got)
	}
}
