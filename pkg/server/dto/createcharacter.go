package dto

// CreateCharacterDTO ...
type CreateCharacterDTO struct {
	TemplateID  string `json:"templateId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Race        string `json:"race"`
	UserID      string
}
