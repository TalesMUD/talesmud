package entities

import "time"

// Role constants for access levels
const (
	RolePlayer  = "player"
	RoleCreator = "creator"
	RoleAdmin   = "admin"
)

// User connects to login credentials to a user object
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
	LastCharacter string `bson:"lastCharacter" json:"lastCharacter"`

	// is set to false after the first PUT request
	IsNewUser bool `bson:"isNewUser" json:"isNewUser"`

	IsOnline bool `bson:"isOnline" json:"isOnline"`

	// Role is the user's access level: "player" (default), "creator", or "admin"
	Role string `json:"role"`

	// IsBanned indicates whether the user is banned from the game
	IsBanned bool `json:"isBanned"`

	// BannedEmail stores the email at the time of banning (for email-based ban enforcement)
	BannedEmail string `json:"bannedEmail,omitempty"`
}

// NewUser creates a new user
func NewUser() *User {
	return &User{}
}

// GetRole returns the effective role, defaulting to "player" if not set
func (u *User) GetRole() string {
	if u.Role == "" {
		return RolePlayer
	}
	return u.Role
}

// IsAdmin returns true if the user has admin role
func (u *User) IsAdmin() bool {
	return u.GetRole() == RoleAdmin
}

// IsCreator returns true if the user has creator or admin role
func (u *User) IsCreator() bool {
	role := u.GetRole()
	return role == RoleCreator || role == RoleAdmin
}
