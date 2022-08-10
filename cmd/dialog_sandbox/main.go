package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	d "github.com/talesmud/talesmud/pkg/entities/dialogs"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Error("Error loading .env file")
	}

	d.WrtieToFile(createSampleDialog(), "dialog1.json")

	dialog := d.ReadFromFile("dialog1.json")

	fmt.Println(dialog.Text)

}

func createSampleDialog() *d.Dialog {
	dialog := d.NewDialog("main", "Hi there how are you doing?",
		d.Options(
			d.NewDialog("1", "I'm fine, thanks",
				d.NewResponse("1_1", "Glad to hear!"),
			),
			d.NewDialog("2", "I'm doing fine, thanks", nil),
			d.NewDialog("3", "I'm doing fine, thanks", nil),
			d.NewDialogWithRequirements("4", "Until next time",
				d.NewResponse("4_1", "Bye!"),
				"1"), // requires visited dialog 1
		),
	)
	return dialog
}
