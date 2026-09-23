package characters

// DoorProfile is per-character state for the daily menu RPG.
// Classic characters leave Character.Door nil.
type DoorProfile struct {
	Sex      string `json:"sex,omitempty"`
	Track    string `json:"track,omitempty"`
	BankGold int64  `json:"bankGold"`
	WeaponID string `json:"weaponId,omitempty"`
	ArmorID  string `json:"armorId,omitempty"`
	Wins     int    `json:"wins"`
	Losses   int    `json:"losses"`
	Gems     int    `json:"gems"`
	// LastDay is the Europe/Berlin calendar day (YYYY-MM-DD) of the last
	// new-day full HP heal. Empty means never healed by ensureNewDay.
	LastDay string `json:"lastDay,omitempty"`
}
