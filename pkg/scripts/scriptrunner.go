package scripts

// ScriptRunner ...
type ScriptRunner interface {
	Run(script Script, ctx interface{}) interface{}
}
