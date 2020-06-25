package scripts

import (
	"github.com/robertkrimen/otto"
	"github.com/sirupsen/logrus"
)

// ScriptRunner ...
type ScriptRunner struct {
}

//Run ...
func (runner ScriptRunner) Run(script Script, ctx interface{}) interface{} {

	logrus.WithField("Script", script.Name).Info("Executing script ...")

	//TODO: do some checks with Code
	//TODO: run code in goroutine that will be killed after 5 seconds
	vm := otto.New()
	vm.Set("ctx", ctx)
	vm.Run(script.Code)

	if value, err := vm.Get("ctx"); err == nil {
		if str, err := value.ToString(); err == nil {
			return str
		}
	}
	return ctx
}
