package main

import (
	"fmt"
	"time"

	//	"reflect"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	d "github.com/talesmud/talesmud/pkg/entities/dialogs"

	"github.com/joho/godotenv"
	"github.com/manifoldco/promptui"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Error("Error loading .env file")
	}

	d.WriteToFile(createSampleDialog(), "dialog1.yaml")
	dialog := d.ReadFromFile("dialog1.yaml")

	// setup dialog state
	state := d.NewDialogState()
	state.CurrentDialogID = "intro"
	state.Context["PLAYER"] = "Hercules"
	state.Context["NPC"] = "Guard"
	state.DynamicContext["TIME"] = func() string {
		return time.Now().Format(time.Kitchen)
	}

	dialogRunning := true

	for dialogRunning {

		prompt := promptui.Select{
			Label: "Talk to Oldtown guard?",
			Items: []string{"Yes", "No"},
		}

		_, result, err := prompt.Run()
		if err != nil {
			log.Error(err)
		}

		if result == "Yes" {
			run(dialog, state)
		} else {
			dialogRunning = false
		}
	}

	d.WriteToFile(createIdleDialog(), "oldtown_townguard_idle_dialog.yaml")
	idleDialog := d.ReadFromFile("oldtown_townguard_idle_dialog.yaml")

	state2 := d.NewDialogState()
	state2.Context["NPC"] = "Guard"
	run(idleDialog, state2)

	log.Info("Sandbox ended.")

}

func run(dialog *d.Dialog, state *d.DialogState) {

	// restart the dialog where we last left off
	currentDialog := dialog.FindDialog(state.CurrentDialogID)

	// if we are at the end of the dialog, restart from the main
	if currentDialog.IsDialogExit != nil && *currentDialog.IsDialogExit {
		state.CurrentDialogID = "main"
		currentDialog = dialog.FindDialog(state.CurrentDialogID)
	}

	running := true
	playerInteracted := false
	for running {
		running, playerInteracted = runDialog(currentDialog, state, playerInteracted, func(key string, interacted bool) (bool, bool) {

			// initialize visited dialogs with 0 if not initialized
			if _, ok := state.DialogVisited[key]; !ok {
				state.DialogVisited[key] = 0
			}

			state.DialogVisited[key]++
			state.CurrentDialogID = key
			currentDialog = dialog.FindDialog(state.CurrentDialogID)
			if currentDialog.IsDialogExit != nil && *currentDialog.IsDialogExit {
				return false, interacted
			}
			return true, interacted
		})
	}

	log.Info("Dialog ended.")
}

func getAvailableOptions(dialog *d.Dialog, state *d.DialogState) []string {

	//TODO: add alternate options as well

	options := make([]string, 0)
	for _, v := range dialog.Options {
		// check if v.RequiresVisitedDialogs is not empty and add to options if all required visited dialogs are matching the ones in state
		if len(v.RequiresVisitedDialogs) > 0 {

			requirementsMatched := true
			for _, r := range v.RequiresVisitedDialogs {
				// if requirement is not in visited dialogs, don't add to options
				if _, ok := state.DialogVisited[r]; !ok {
					requirementsMatched = false
					break
				}
			}
			if requirementsMatched {
				options = append(options, v.RenderPlain(state))
				// should this go here?
			}

		} else {
			options = append(options, v.RenderPlain(state))
		}

	}
	return options
}

func runDialog(dialog *d.Dialog, state *d.DialogState, playerInteracted bool, choiceFunc func(string, bool) (bool, bool)) (bool, bool) {

	assert.NotNil(nil, dialog)

	dialogSource := "NPC"
	if playerInteracted {
		dialogSource = "PLAYER"
	}
	dialogOutPrefix := ""
	if val, ok := state.Context[dialogSource]; ok {
		dialogOutPrefix = val + ": "
	}
	// dialog has no options
	if dialog.Options == nil && dialog.Answer == nil {
		fmt.Println(dialogOutPrefix + dialog.Render(state))
		time.Sleep(3 * time.Second)

		if dialog.IsDialogExit != nil && *dialog.IsDialogExit {
			return choiceFunc("end", false)
		} else {
			return choiceFunc("main", false)
		}
	}
	// dialog has a single option as a response
	if dialog.Answer != nil {
		fmt.Println(dialogOutPrefix + dialog.Render(state))
		if !playerInteracted {
			time.Sleep(3 * time.Second)
		} else {
			time.Sleep(1 * time.Second)
		}

		return runDialog(dialog.Answer, state, false, choiceFunc)
	}

	// dialog has multiple options, get input from user
	options := getAvailableOptions(dialog, state)

	prompt := promptui.Select{
		Label: dialogOutPrefix + dialog.Render(state),
		Items: options,
	}

	_, result, err := prompt.Run()
	if err != nil {
		log.Error(err)
	}

	// get ID of option of user selected item and passinto choiceFunc
	selectedID := ""
	for _, k := range dialog.Options {
		if k.RenderPlain(state) == result {
			selectedID = k.ID
		}
	}

	return choiceFunc(selectedID, true)

}

func createIdleDialog() *d.Dialog {
	return &d.Dialog{
		ID:   "main",
		Text: "Have you checked out the well in the towns center yet?",
		Answer: &d.Dialog{
			ID:   "rambling2",
			Text: "I heard a sip from it could work wonders on your health.",
			Answer: &d.Dialog{
				ID:   "rambling3",
				Text: "I'm sure it will at least be refreshing.",
				Answer: &d.Dialog{
					ID:           "end",
					Text:         "Just make sure to not fall into the well.",
					IsDialogExit: d.RefTrue(),
				},
			},
		},
	}
}

func createSampleDialog() *d.Dialog {

	return &d.Dialog{
		ID:   "intro",
		Text: "Hello stranger, are you visiting Oldtown for the first time?",
		AlternateTexts: []string{
			"Hi there, new to Oldtown?",
			"Welcome to Oldtown!",
		},
		Options: d.Options(
			&d.Dialog{
				ID:   "intro_yes",
				Text: "Yes, i go by the name {{PLAYER}}.",
				Answer: &d.Dialog{
					ID:   "main",
					Text: "Welcome to Oldtown {{PLAYER}}, let me show you around. How can i help you today?",
					AlternateTexts: []string{
						"Greetings {{PLAYER}}, let me show you around. How can serve you?",
						"Let me welcome you to Oldtown {{PLAYER}}, how can i help you today?",
					},
					Options: d.Options(
						&d.Dialog{
							ID:   "smith",
							Text: "Where can i find the next smith?",
							Answer: &d.Dialog{
								ID:   "smith_question",
								Text: "You can find the next smith in the town center on the left side of the road.",
								AlternateTexts: []string{
									"Head to the town center and then go to the left side of the road you will reach the smith from there.",
								},
							},
						},
						&d.Dialog{
							ID:                     "hidden",
							Text:                   "You can only see this if you asked me where to find the mayor",
							RequiresVisitedDialogs: []string{"mayor"},
							Answer: &d.Dialog{
								ID:   "hidden_question",
								Text: "Secretly, i am the Mayor. Dont tell anyone!",
							},
						},
						&d.Dialog{
							ID:   "inn",
							Text: "Where can i grab a drink?",
						},
						&d.Dialog{
							ID:   "time",
							Text: "How late is it?",
							Answer: &d.Dialog{
								ID:   "time_answer",
								Text: "It is currently {{TIME}}.",
							},
						},
						&d.Dialog{
							ID:   "mayor",
							Text: "Where do i find the mayor of Oldtown?",
							Answer: &d.Dialog{
								ID:   "mayor_question",
								Text: "Not sure, why do you ask?",
							},
						},
						&d.Dialog{
							ID:   "quests",
							Text: "Where could i offer my services for some coins?",
						},
						&d.Dialog{
							ID:   "leave",
							Text: "I'm done, goodbye.",
							Answer: &d.Dialog{
								ID:           "end",
								Text:         "'Til next time {{PLAYER}}.",
								IsDialogExit: d.RefTrue(),
							},
						},
					),
				},
			},
		),
	}
}
