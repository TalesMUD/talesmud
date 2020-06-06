package dto

//CreateCharacterDTO ...
type CreateCharacterDTO struct {
	TemplateID  int32  `json:"templateId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UserID      string
}
