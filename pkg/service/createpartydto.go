package service

//CreatePartyDTO data
type CreatePartyDTO struct {
	Name       string   `json:"name,omitempty"`
	Characters []string `json:"characters,omitempty"`
}
