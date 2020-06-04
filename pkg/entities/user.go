package entities

import "time"

//User connects to login credentials to a user object
type User struct {
	*Entity `bson:",inline"`

	// main *unique* reference id, its additional to ID because it might be a list of reference ids in the future so a user
	// can merge multiple accounts (refids) into the same user object
	RefID string `json:"refid,omitempty"`

	// name and email should never be displayed in a game scenario
	Name  string `json:"name"`
	Email string `json:"email"`

	// nickname can be used to display a player/user in the case where character name is not applicable
	Nickname string `json:"nickname"`

	Created  time.Time `bson:"created" json:"created,omitempty"`
	LastSeen time.Time `bson:"lastSeen" json:"lastSeen,omitempty"`
	Picture  string    `json:"picture"`

	// every time the user logs in the last character is automatically loaded. When switched we track the last character in the user object while the
	// game server will switch completely to the new character
	LastCharacter string `json:"lastCharacter"`

	// is set to false after the first PUT request
	IsNewUser bool `json:"isNewUser"`
}

// NewUser creates a new user
func NewUser() *User {
	return &User{}
}
