package dialogs

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type DialogOptions map[string]*Dialog

type DialogState struct {
	CurrentDialog string
	DialogVisited map[string]bool
	Context       map[string]string
}

// create DialogOptionType enum with options SINGLE and ALWAYS
type DialogOptionType int

const DOT_SINGLE = 0
const DOT_ALWAYS = 1

func NewDialogState() *DialogState {
	return &DialogState{
		CurrentDialog: "main",
		DialogVisited: make(map[string]bool),
		Context:       make(map[string]string),
	}
}

// Dialog ...
type Dialog struct {
	ID                     string        `bson:"id,omitempty" json:"id,omitempty" yaml:"id,omitempty"`
	Text                   string        `bson:"text,omitempty" json:"text,omitempty" yaml:"text"`
	Options                DialogOptions `bson:"options,omitempty" json:"options,omitempty" yaml:"options,omitempty"`
	RequiresVisitedDialogs []string      `bson:"requires_visited_dialogs,omitempty" json:"requires_visited_dialogs,omitempty" yaml:"requires_visited_dialogs,omitempty"`
	ShowOnlyOnce           *bool         `bson:"show_only_once,omitempty" json:"show_only_once,omitempty" yaml:"show_only_once,omitempty"`
	HasAnswer              *bool         `bson:"has_answer,omitempty" json:"has_answer,omitempty" yaml:"has_answer,omitempty"`
	IsDialogExit           *bool         `bson:"is_dialog_exit,omitempty" json:"is_dialog_exit,omitempty" yaml:"is_dialog_exit,omitempty"`
}

// create a new Dialog
func NewDialog(id string, text string, options map[string]*Dialog, showOnlyOnce *bool) *Dialog {
	return &Dialog{
		ID:           id,
		Text:         text,
		Options:      options,
		ShowOnlyOnce: showOnlyOnce}
}
func NewDialogWithRequirements(id string, text string, options map[string]*Dialog, showOnlyOnce *bool, visited ...string) *Dialog {
	return &Dialog{
		ID:                     id,
		Text:                   text,
		Options:                options,
		RequiresVisitedDialogs: visited,
		ShowOnlyOnce:           showOnlyOnce}
}

// creates a dialog that is used as a response (defaults to DOT_ALWAYS)
func NewResponse(id string, text string) DialogOptions {
	return Options(NewDialog(id, text, nil, nil))
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
		err = yaml.Unmarshal(data, &d)
		if err == nil {
			return &d
		}
	}
	return nil
}

func WrtieToFile(dialog *Dialog, fileName string) {
	result, err := yaml.Marshal(dialog) //.MarshalIndent(dialog, "", "    ")
	if err == nil {
		err = ioutil.WriteFile(fileName, result, 0644)
		if err != nil {
			log.Error("Error writing %s", fileName)
		}
	}
}

func RefTrue() *bool {
	b := true
	return &b
}
