package dto

//CreateCharacterDTO ...
type CreateCharacterDTO struct {
	TemplateID  string `json:"templateId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UserID      string
}
