package main

import (
	"fmt"
	"reflect"

	//	"reflect"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	d "github.com/talesmud/talesmud/pkg/entities/dialogs"

	"github.com/hoisie/mustache"
	"github.com/joho/godotenv"
	"github.com/manifoldco/promptui"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Error("Error loading .env file")
	}

	d.WrtieToFile(createSampleDialog(), "dialog1.yaml")

	dialog := d.ReadFromFile("dialog1.yaml")
	state := d.DialogState{
		CurrentDialog: "intro",
		DialogVisited: make(map[string]bool),
		Context: map[string]string{
			"PLAYER": "Hercules",
		},
	}

	run(dialog, &state)

}

func run(dialog *d.Dialog, state *d.DialogState) {

	currentDialog := dialog

	for runDialog(currentDialog, state, func(key string) bool {
		state.DialogVisited[key] = true
		state.CurrentDialog = key
		currentDialog = dialog.FindDialog(state.CurrentDialog)
		if currentDialog.IsDialogExit != nil && *currentDialog.IsDialogExit {
			return false
		}
		return true
	}) {
		// ...
	}

	log.Info("Dialog ended.")
}

func getAvailableOptions(dialog *d.Dialog, state *d.DialogState) []string {
	options := make([]string, 0)
	for _, v := range dialog.Options {
		// check if v.RequiresVisitedDialogs is not empty and add to options if all required visited dialogs are matching the ones in state
		if len(v.RequiresVisitedDialogs) > 0 {

			requirementsMatched := true
			for _, r := range v.RequiresVisitedDialogs {
				// if requirement is not in visited dialogs, don't add to options
				if !state.DialogVisited[r] {
					requirementsMatched = false
					break
				}
			}
			if requirementsMatched {
				options = append(options, mustache.Render(v.Text, state.Context))
				// should this go here?
			}

		} else {
			options = append(options, mustache.Render(v.Text, state.Context))
		}

	}
	return options
}

func runDialog(dialog *d.Dialog, state *d.DialogState, choiceFunc func(string) bool) bool {

	assert.NotNil(nil, dialog)

	// dialog has no options
	if dialog.Options == nil {
		fmt.Println(mustache.Render(dialog.Text, state.Context))

		if dialog.IsDialogExit != nil && *dialog.IsDialogExit {
			return choiceFunc("end")
		} else {
			return choiceFunc("main")
		}
	}
	// dialog has a single option as a response
	if dialog.HasAnswer != nil && *dialog.HasAnswer {
		keys := reflect.ValueOf(dialog.Options).MapKeys()
		nextDialog := dialog.Options[keys[0].String()]
		return runDialog(nextDialog, state, choiceFunc)
	}

	// dialog has multiple options, get input from user
	options := getAvailableOptions(dialog, state)

	prompt := promptui.Select{
		Label: mustache.Render(dialog.Text, state.Context),
		Items: options,
	}

	_, result, err := prompt.Run()
	if err != nil {
		log.Error(err)
	}

	// get ID of option of user selected item and passinto choiceFunc
	selectedID := ""
	for _, k := range dialog.Options {
		if mustache.Render(k.Text, state.Context) == result {
			selectedID = k.ID
		}
	}

	return choiceFunc(selectedID)

}

func createSampleDialog() *d.Dialog {

	return &d.Dialog{
		ID:   "intro",
		Text: "Hello stranger, are you visiting Oldtown for the first time?",
		Options: d.Options(
			&d.Dialog{
				ID:        "intro_yes",
				Text:      "Yes, i go by the name {{PLAYER}}.",
				HasAnswer: d.RefTrue(),
				Options: d.Options(
					&d.Dialog{
						ID:   "main",
						Text: "Welcome to Oldtown {{PLAYER}}, let me show you around. How can i help you today?",
						Options: d.Options(
							&d.Dialog{
								ID:        "smith",
								Text:      "Where can i find the next smith?",
								HasAnswer: d.RefTrue(),
								Options: d.Options(
									&d.Dialog{
										ID:   "smith_question",
										Text: "You can find the next smith in the town center on the left side of the road.",
									},
								),
							},
							&d.Dialog{
								ID:                     "hidden",
								Text:                   "You can only see this if you asked me where to find the mayor",
								RequiresVisitedDialogs: []string{"mayor"},
								HasAnswer:              d.RefTrue(),
								Options: d.Options(
									&d.Dialog{
										ID:   "hidden_question",
										Text: "Secretly, i am the Mayor. Dont tell anyone!",
									},
								),
							},
							&d.Dialog{
								ID:   "inn",
								Text: "Where can i grab a drink?",
							},
							&d.Dialog{
								ID:        "mayor",
								Text:      "Where do i find the mayor of Oldtown?",
								HasAnswer: d.RefTrue(),
								Options: d.Options(
									&d.Dialog{
										ID:   "mayor_question",
										Text: "Not sure, why do you ask?",
									},
								),
							},
							&d.Dialog{
								ID:   "quests",
								Text: "Where could i offer my services for some coins?",
							},
							&d.Dialog{
								ID:        "leave",
								Text:      "I'm done, goodbye.",
								HasAnswer: d.RefTrue(),
								Options: d.Options(
									&d.Dialog{
										ID:           "end",
										Text:         "'Til next time {{PLAYER}}.",
										IsDialogExit: d.RefTrue(),
									},
								),
							},
						),
					}),
			},
		),
	}
	/*
		dialog := d.NewDialog("intro", "")

		d.NewDialog("main", "Hi there how are you doing?",
			d.Options(
				d.NewDialog("1", "I'm fine, thanks",
					d.NewResponse("1_1", "Glad to hear!"),
					d.NewShowOnlyOnce(),
				),
				d.NewDialog("2", "I'm doing fine, thanks", nil, nil),
				d.NewDialog("3", "Almost great!", nil, nil),
				d.NewDialogWithRequirements("4", "Until next time",
					d.NewResponse("4_1", "Bye!"),
					d.NewShowOnlyOnce(),
					"1"), // requires visited dialog 1
			),
			nil,
		)
		return dialog*/
}
