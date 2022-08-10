package dialogs

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

type DialogOptions map[string]*Dialog

type DialogState struct {
	CurrentDialog string
	DialogVisited map[string]bool
}

// Dialog ...
type Dialog struct {
	ID                     string        `bson:"id,omitempty" json:"id,omitempty"`
	Text                   string        `bson:"text,omitempty" json:"text,omitempty"`
	Options                DialogOptions `bson:"options,omitempty" json:"options,omitempty"`
	RequiresVisitedDialogs []string      `bson:"requiresVisitedDialogs,omitempty" json:"requiresVisitedDialogs,omitempty"`
}

// create a new Dialog
func NewDialog(id string, text string, options map[string]*Dialog) *Dialog {
	return &Dialog{
		ID:      id,
		Text:    text,
		Options: options,
	}
}
func NewDialogWithRequirements(id string, text string, options map[string]*Dialog, visited ...string) *Dialog {
	return &Dialog{
		ID:                     id,
		Text:                   text,
		Options:                options,
		RequiresVisitedDialogs: visited,
	}
}
func NewResponse(id string, text string) DialogOptions {
	return Options(NewDialog(id, text, nil))
}

func SingleOption(dialog *Dialog) DialogOptions {
	return DialogOptions{
		dialog.ID: dialog,
	}
}

// create dialog map structure and add all parameters to it
func Options(dialogs ...*Dialog) DialogOptions {
	dialogMap := make(DialogOptions)
	for _, v := range dialogs {
		dialogMap[v.ID] = v
	}
	return dialogMap
}

// FindDialogue ...
func (d *Dialog) FindDialog(id string) *Dialog {

	if d.ID == id {
		return d
	}

	for _, v := range d.Options {
		if r := v.FindDialog(id); r != nil {
			return r
		}
	}
	return nil
}

func ReadFromFile(fileName string) *Dialog {
	data, err := ioutil.ReadFile(fileName)
	if err == nil {
		var d Dialog
		err = json.Unmarshal(data, &d)
		if err == nil {
			return &d
		}
	}
	return nil
}

func WrtieToFile(dialog *Dialog, fileName string) {
	result, err := json.MarshalIndent(dialog, "", "    ")
	if err == nil {
		err = ioutil.WriteFile(fileName, result, 0644)
		if err != nil {
			log.Error("Error writing out.json")
		}
	}
}
