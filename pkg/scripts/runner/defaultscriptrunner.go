package runner

import (
	"encoding/json"
	"fmt"

	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"

	"github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts"
	"github.com/talesmud/talesmud/pkg/service"
)

// DefaultScriptRunner ...
type DefaultScriptRunner struct {
	RoomsService      service.RoomsService
	CharactersService service.CharactersService
	ItemsService      service.ItemsService
	Game              def.GameCtrl

	//	ScriptService     service.ScriptsService
}

// NewDefaultScriptRunner ...
func NewDefaultScriptRunner() *DefaultScriptRunner {
	return &DefaultScriptRunner{}
}

// SetServices ...
func (runner *DefaultScriptRunner) SetServices(facade service.Facade, game def.GameCtrl) {
	runner.ItemsService = facade.ItemsService()
	runner.RoomsService = facade.RoomsService()
	runner.CharactersService = facade.CharactersService()
	runner.Game = game
}

//Run ...
func (runner *DefaultScriptRunner) Run(script scripts.Script, ctx interface{}) interface{} {

	logrus.WithField("Script", script.Name).Info("Executing script ...")

	//TODO: do some checks with Code
	//TODO: run code in goroutine that will be killed after 5 seconds
	vm := runner.newScriptRuntime()
	vm.Set("ctx", ctx)

	_, err := vm.Run(script.Code)
	if err != nil {
		return err.Error()
	}

	if value, err := vm.Get("ctx"); err == nil {
		if str, err := value.ToString(); err == nil {

			//st, _ := strconv.Unquote(str)

			// depending on type unmarshal to different entities
			//TYPE ITEM:
			fmt.Println(str)

			bytes := []byte(str)
			var item items.Item
			if err := json.Unmarshal(bytes, &item); err != nil {
				logrus.WithField("string", str).WithError(err).Info("Could not unmarshal")
			}

			fmt.Println(item)

			return item
		}
	}
	return ctx
}

func (runner DefaultScriptRunner) newScriptRuntime() *otto.Otto {
	vm := otto.New()

	runner.addItemFunctions(vm)
	runner.addRoomFunctions(vm)
	runner.addGameFunctions(vm)

	return vm
}

func (runner DefaultScriptRunner) addItemFunctions(vm *otto.Otto) {

	vm.Set("T_findItemTemplate", func(call otto.FunctionCall) otto.Value {
		itemTemplate, _ := call.Argument(0).ToString()
		templates, _ := runner.ItemsService.ItemTemplates().FindByName(itemTemplate)
		result, _ := vm.ToValue(items.ItemTemplatesToJSONString(templates))
		return result
	})
	vm.Set("T_getItemTemplate", func(call otto.FunctionCall) otto.Value {
		itemTemplateID, _ := call.Argument(0).ToString()
		template, _ := runner.ItemsService.ItemTemplates().FindByID(itemTemplateID)
		result, _ := vm.ToValue(items.ItemTemplateToJSONString(*template))
		return result
	})
	vm.Set("T_createItemFromTemplate", func(call otto.FunctionCall) otto.Value {
		templateID, _ := call.Argument(0).ToString()
		item, _ := runner.ItemsService.CreateItemFromTemplate(templateID)
		result, _ := vm.ToValue(items.ItemToJSONString(*item))
		return result
	})
}

func (runner DefaultScriptRunner) addGameFunctions(vm *otto.Otto) {

	vm.Set("T_msgToRoom", func(call otto.FunctionCall) otto.Value {
		roomID, _ := call.Argument(0).ToString()
		message, _ := call.Argument(1).ToString()

		if runner.Game != nil {
			msg := messages.NewRoomBasedMessage("SYSTEM", message)
			msg.Audience = messages.MessageAudienceRoom
			msg.AudienceID = roomID
			runner.Game.SendMessage() <- msg
		}
		return otto.TrueValue()
	})

}

func (runner DefaultScriptRunner) addRoomFunctions(vm *otto.Otto) {

	vm.Set("T_findRoom", func(call otto.FunctionCall) otto.Value {
		room, _ := call.Argument(0).ToString()
		r, _ := runner.RoomsService.FindAllWithQuery(repository.RoomsQuery{Name: &room})

		result, _ := vm.ToValue(rooms.RoomsToJSONString(r))
		return result
	})
	vm.Set("T_getRoom", func(call otto.FunctionCall) otto.Value {
		roomID, _ := call.Argument(0).ToString()
		room, _ := runner.RoomsService.FindByID(roomID)
		result, _ := vm.ToValue(rooms.RoomToJSONString(*room))
		return result
	})
	vm.Set("T_updateRoom", func(call otto.FunctionCall) otto.Value {
		roomString, _ := call.Argument(0).ToString()
		if room, err := rooms.RoomFromJSONString(roomString); err == nil {
			runner.RoomsService.Update(room.ID, room)
		}
		return otto.Value{}
	})
}
