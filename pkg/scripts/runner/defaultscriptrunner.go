package runner

import (
	"encoding/json"
	"time"

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

// DefaultScriptRunner implements ScriptRunner for JavaScript using Otto
// Deprecated: Use LuaRunner for new scripts
type DefaultScriptRunner struct {
	RoomsService      service.RoomsService
	CharactersService service.CharactersService
	ItemsService      service.ItemsService
	Game              def.GameCtrl
}

// NewDefaultScriptRunner creates a new JavaScript script runner
func NewDefaultScriptRunner() *DefaultScriptRunner {
	return &DefaultScriptRunner{}
}

// SetServices injects the required services into the runner
func (runner *DefaultScriptRunner) SetServices(facade service.Facade, game def.GameCtrl) {
	runner.ItemsService = facade.ItemsService()
	runner.RoomsService = facade.RoomsService()
	runner.CharactersService = facade.CharactersService()
	runner.Game = game
}

// SupportsLanguage returns true if the runner supports the given language
func (runner *DefaultScriptRunner) SupportsLanguage(lang scripts.ScriptLanguage) bool {
	return lang == scripts.ScriptLanguageJavaScript
}

// Shutdown gracefully shuts down the script runner
func (runner *DefaultScriptRunner) Shutdown() {
	// Otto doesn't require cleanup
}

// Run executes a JavaScript script with the given context
func (runner *DefaultScriptRunner) Run(script scripts.Script, ctx interface{}) interface{} {
	result := runner.RunWithResult(script, &scripts.ScriptContext{Data: map[string]interface{}{"ctx": ctx}})
	if !result.Success {
		return result.Error
	}
	return result.Result
}

// RunWithResult executes a script and returns detailed result information
func (runner *DefaultScriptRunner) RunWithResult(script scripts.Script, ctx *scripts.ScriptContext) *scripts.ScriptResult {
	start := time.Now()

	logrus.WithField("Script", script.Name).WithField("Language", "javascript").Info("Executing script...")

	vm := runner.newScriptRuntime()

	// Set context variables
	if ctx != nil {
		for key, value := range ctx.Data {
			vm.Set(key, value)
		}
	}

	_, err := vm.Run(script.Code)
	if err != nil {
		return &scripts.ScriptResult{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start),
		}
	}

	// Try to get the modified context
	if value, err := vm.Get("ctx"); err == nil {
		if str, err := value.ToString(); err == nil {
			// Try to unmarshal as item (for backward compatibility)
			bytes := []byte(str)
			var item items.Item
			if err := json.Unmarshal(bytes, &item); err == nil {
				return &scripts.ScriptResult{
					Success:  true,
					Result:   item,
					Duration: time.Since(start),
				}
			}
			// Return raw string if not an item
			return &scripts.ScriptResult{
				Success:  true,
				Result:   str,
				Duration: time.Since(start),
			}
		}
	}

	return &scripts.ScriptResult{
		Success:  true,
		Result:   nil,
		Duration: time.Since(start),
	}
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
		templates, _ := runner.ItemsService.FindTemplateByName(itemTemplate)
		result, _ := vm.ToValue(items.ItemsToJSONString(templates))
		return result
	})
	vm.Set("T_getItemTemplate", func(call otto.FunctionCall) otto.Value {
		itemTemplateID, _ := call.Argument(0).ToString()
		template, _ := runner.ItemsService.FindByID(itemTemplateID)
		if template != nil && template.IsTemplate {
			result, _ := vm.ToValue(items.ItemToJSONString(*template))
			return result
		}
		return otto.NullValue()
	})
	vm.Set("T_createItemFromTemplate", func(call otto.FunctionCall) otto.Value {
		templateID, _ := call.Argument(0).ToString()
		item, _ := runner.ItemsService.CreateInstanceFromTemplate(templateID)
		if item != nil {
			result, _ := vm.ToValue(items.ItemToJSONString(*item))
			return result
		}
		return otto.NullValue()
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
		r, _ := runner.RoomsService.FindAllWithQuery(repository.RoomsQuery{Name: room})

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
