package dialogs

import (
	"io/ioutil"
	"math/rand"

	"github.com/hoisie/mustache"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type DialogOptions []*Dialog

type DialogState struct {
	CurrentDialogID string
	// DialogVisited stores how many times a dialog has been visited
	DialogVisited  map[string]int
	Context        map[string]string
	DynamicContext map[string]func() string
}

// create DialogOptionType enum with options SINGLE and ALWAYS
type DialogOptionType int

const DOT_SINGLE = 0
const DOT_ALWAYS = 1

func NewDialogState() *DialogState {
	return &DialogState{
		CurrentDialogID: "main",
		DialogVisited:   make(map[string]int),
		Context:         make(map[string]string),
		// dont use dynamic context in player options for now
		DynamicContext: make(map[string]func() string),
	}
}

// Dialog ...
type Dialog struct {
	ID   string `bson:"id,omitempty" json:"id,omitempty" yaml:"id,omitempty"`
	Text string `bson:"text,omitempty" json:"text,omitempty" yaml:"text"`
	// if AlternateTexts is not empty then the text should be randomly selected
	AlternateTexts []string `bson:"alternateTexts,omitempty" json:"alternateTexts,omitempty" yaml:"alternateTexts,omitempty"`
	// if ordered texts is set the shown text will be based on how many times you visited the dialog
	OrderedTexts *bool `bson:"orderedTexts,omitempty" json:"orderedTexts,omitempty" yaml:"orderedTexts,omitempty"`

	// a dialog either has options the player can choose from or it has a response that is shown to the player.
	Options                []*Dialog `bson:"options,omitempty" json:"options,omitempty" yaml:"options,omitempty"`
	Answer                 *Dialog   `bson:"answer,omitempty" json:"answer,omitempty" yaml:"answer,omitempty"`
	RequiresVisitedDialogs []string  `bson:"requires_visited_dialogs,omitempty" json:"requires_visited_dialogs,omitempty" yaml:"requires_visited_dialogs,omitempty"`
	ShowOnlyOnce           *bool     `bson:"show_only_once,omitempty" json:"show_only_once,omitempty" yaml:"show_only_once,omitempty"`
	//	HasAnswer              *bool         `bson:"has_answer,omitempty" json:"has_answer,omitempty" yaml:"has_answer,omitempty"`
	IsDialogExit *bool `bson:"is_dialog_exit,omitempty" json:"is_dialog_exit,omitempty" yaml:"is_dialog_exit,omitempty"`
}

// create a new Dialog
func NewDialog(id string, text string, options []*Dialog, showOnlyOnce *bool) *Dialog {
	return &Dialog{
		ID:           id,
		Text:         text,
		Options:      options,
		ShowOnlyOnce: showOnlyOnce}
}
func NewDialogWithRequirements(id string, text string, options []*Dialog, showOnlyOnce *bool, visited ...string) *Dialog {
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
	return dialogs
}

func (d *Dialog) GetText() string {

	//TODO: add logic for ordered texts
	randTextCount := 1
	if d.AlternateTexts != nil {
		randTextCount += len(d.AlternateTexts)
	}

	randTextID := rand.Intn(randTextCount)

	if randTextID == 0 {
		return d.Text
	}

	assert.NotNil(nil, d.AlternateTexts)
	return d.AlternateTexts[randTextID-1]
}

func (d *Dialog) Render(state *DialogState) string {
	// iterate over all state.DynamicContext and add them to the state.Context
	for k, v := range state.DynamicContext {
		state.Context[k] = v()
	}

	return mustache.Render(d.GetText(), state.Context)
}

func (d *Dialog) RenderPlain(state *DialogState) string {
	// iterate over all state.DynamicContext and add them to the state.Context
	for k, v := range state.DynamicContext {
		state.Context[k] = v()
	}

	return mustache.Render(d.Text, state.Context)
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
	if d.Answer != nil {
		return d.Answer.FindDialog(id)
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

func WriteToFile(dialog *Dialog, fileName string) {
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
