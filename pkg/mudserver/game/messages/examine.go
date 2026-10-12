package messages

import "github.com/talesmud/talesmud/pkg/entities/items"

// ExamineResult is an additive companion to the existing item inspection prose.
type ExamineResult struct {
	MessageResponse
	Name        string                 `json:"name,omitempty"`
	Slot        items.ItemSlot         `json:"slot,omitempty"`
	Quality     items.ItemQuality      `json:"quality,omitempty"`
	Stats       map[string]interface{} `json:"stats,omitempty"`
	Value       int64                  `json:"value,omitempty"`
	Effects     []string               `json:"effects,omitempty"`
	Description string                 `json:"description,omitempty"`
	Detail      string                 `json:"detail,omitempty"`
}

func NewExamineResult(userID string, item *items.Item, scriptName func(string) string) *ExamineResult {
	if item == nil {
		return nil
	}
	result := &ExamineResult{
		MessageResponse: MessageResponse{Audience: MessageAudienceOrigin, AudienceID: userID, Type: MessageTypeExamineResult},
		Name:            item.Name, Slot: item.Slot, Quality: item.Quality,
		Stats: item.Attributes, Value: item.BasePrice, Description: item.Description, Detail: item.Detail,
	}
	for _, effect := range item.DisplayEffects(scriptName) {
		result.Effects = append(result.Effects, items.FormatEffect(effect))
	}
	return result
}
