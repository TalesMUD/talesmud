package scripts

import (
	"github.com/sirupsen/logrus"
)

// ScriptRunner ...
type ScriptRunner struct {
}

//Run ...
func (runner ScriptRunner) Run(script Script, ctx interface{}) {

	logrus.WithField("Script", script.Name).Info("Executing script ...")
}
