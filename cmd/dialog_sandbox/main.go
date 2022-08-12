package main

import (
	"fmt"

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

	run(dialog, state)

}

func run(dialog *d.Dialog, state d.DialogState) {
	for {
		runDialog(dialog, func(key string) {
			fmt.Printf("You selected: %s\n", dialog.Options[key])

			state.CurrentDialog = dialog.Options[key].ID
		})

	}
}

func runDialog(dialog *d.Dialog, choiceFunc func(string)) {

	if dialog.Options == nil {
		fmt.Println(dialog.Text)
		return
	}

	// get string array slice of options texts from dialog
	options := make([]string, 0)
	for _, v := range dialog.Options {
		options = append(options, v.Text)
	}

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

func createSampleDialog() *d.Dialog {
	dialog := d.NewDialog("main", "Hi there how are you doing?",
		d.Options(
			d.NewDialog("1", "I'm fine, thanks",
				d.NewResponse("1_1", "Glad to hear!"),
				d.DOT_SINGLE,
			),
			d.NewDialog("2", "I'm doing fine, thanks", nil, d.DOT_SINGLE),
			d.NewDialog("3", "I'm doing fine, thanks", nil, d.DOT_SINGLE),
			d.NewDialogWithRequirements("4", "Until next time",
				d.NewResponse("4_1", "Bye!"),
				"1"), // requires visited dialog 1
		),
	)
	return dialog
}
