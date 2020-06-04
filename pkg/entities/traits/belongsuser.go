package traits

//BelongsUser ...
type BelongsUser struct {
	BelongsUserID string `bson:"belongsUser,omitempty" json:"belongsUser,omitempty"`
}

// BelongsToUser ...
func BelongsToUser(id string) *BelongsUser {
	return &BelongsUser{
		BelongsUserID: id,
	}
}
