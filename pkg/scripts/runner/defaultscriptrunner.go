package runner

import (
	"encoding/json"
	"fmt"

	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"

	"github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/scripts"
	"github.com/talesmud/talesmud/pkg/service"
)

// DefaultScriptRunner ...
type DefaultScriptRunner struct {
	//	RoomsService      service.RoomsService
	//	CharactersService service.CharactersService/
	//	UserService       service.UsersService
	ItemsService service.ItemsService
	//	ScriptService     service.ScriptsService
}

// NewDefaultScriptRunner ...
func NewDefaultScriptRunner() *DefaultScriptRunner {
	return &DefaultScriptRunner{}
}

// SetServices ...
func (runner *DefaultScriptRunner) SetServices(facade service.Facade) {
	runner.ItemsService = facade.ItemsService()
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
	vm.Set("T_findItemTemplate", func(call otto.FunctionCall) otto.Value {
		itemTemplate, _ := call.Argument(0).ToString()
		templates, _ := runner.ItemsService.ItemTemplates().FindByName(itemTemplate)

		bytes, _ := json.MarshalIndent(templates, "", "  ")
		result, _ := vm.ToValue(string(bytes))
		return result
	})

	vm.Set("T_createItemFromTemplate", func(call otto.FunctionCall) otto.Value {
		templateID, _ := call.Argument(0).ToString()
		item, _ := runner.ItemsService.CreateItemFromTemplate(templateID)

		bytes, _ := json.MarshalIndent(item, "", "  ")
		result, _ := vm.ToValue(string(bytes))
		return result
	})

	return vm
}
