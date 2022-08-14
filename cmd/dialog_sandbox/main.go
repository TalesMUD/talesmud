package main

import (
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	d "github.com/talesmud/talesmud/pkg/entities/dialogs"

	"github.com/joho/godotenv"
	"github.com/manifoldco/promptui"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Error("Error loading .env file")
	}

	d.WrtieToFile(createSampleDialog(), "dialog1.json")

	dialog := d.ReadFromFile("dialog1.json")
	state := d.DialogState{
		CurrentDialog: "main",
		DialogVisited: make(map[string]bool),
	}

	run(dialog, &state)

}

func run(dialog *d.Dialog, state *d.DialogState) {

	currentDialog := dialog

	for {
		runDialog(currentDialog, state, func(key string) {
			state.DialogVisited[key] = true
			state.CurrentDialog = dialog.Options[key].ID
			currentDialog = dialog.FindDialog(state.CurrentDialog)
		})
	}
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
				options = append(options, v.ID)
			}

		} else {
			options = append(options, v.Text)
		}

	}
	return options
}

func runDialog(dialog *d.Dialog, state *d.DialogState, choiceFunc func(string)) {

	if dialog.Options == nil {
		fmt.Println(dialog.Text)
		choiceFunc("main")
		return
	}

	if len(dialog.Options) == 1 {
		keys := reflect.ValueOf(dialog.Options).MapKeys()
		runDialog(dialog.Options[keys[0].String()], state, choiceFunc)
	} else {
		options := getAvailableOptions(dialog, state)

		prompt := promptui.Select{
			Label: dialog.Text,
			Items: options,
		}

		_, result, err := prompt.Run()
		if err != nil {
			log.Error(err)
		}

		// get ID of option of user selected item and passinto choiceFunc
		selectedID := ""
		for _, k := range dialog.Options {
			if k.Text == result {
				selectedID = k.ID
			}
		}
		choiceFunc(selectedID)
	}

}

func createSampleDialog() *d.Dialog {
	dialog := d.NewDialog("main", "Hi there how are you doing?",
		d.Options(
			d.NewDialog("1", "I'm fine, thanks",
				d.NewResponse("1_1", "Glad to hear!"),
				d.DOT_SINGLE,
			),
			d.NewDialog("2", "I'm doing fine, thanks", nil, d.DOT_SINGLE),
			d.NewDialog("3", "Almost great!", nil, d.DOT_SINGLE),
			d.NewDialogWithRequirements("4", "Until next time",
				d.NewResponse("4_1", "Bye!"),
				d.DOT_SINGLE,
				"1"), // requires visited dialog 1
		),
		d.DOT_ALWAYS,
	)
	return dialog
}
