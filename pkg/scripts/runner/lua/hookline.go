package lua

import lua "github.com/yuin/gopher-lua"

const hookLineRegistryKey = "tales.hookLine"

type hookLine struct {
	Hook   string
	Source string
}

// SetHookLine stamps this LState so room lines sent while it runs carry the hook.
// The stamp lives on the state, not the shared runner.
func SetHookLine(L *lua.LState, hook, source string) {
	if L == nil || hook == "" {
		return
	}
	ud := L.NewUserData()
	ud.Value = hookLine{Hook: hook, Source: source}
	L.SetField(L.Get(lua.RegistryIndex), hookLineRegistryKey, ud)
}

// ClearHookLine removes the stamp so the next script on this state is plain.
func ClearHookLine(L *lua.LState) {
	if L == nil {
		return
	}
	L.SetField(L.Get(lua.RegistryIndex), hookLineRegistryKey, lua.LNil)
}

// HookLineFrom reads the stamp for the state that is executing the call.
func HookLineFrom(L *lua.LState) (hook, source string, ok bool) {
	if L == nil {
		return "", "", false
	}
	v := L.GetField(L.Get(lua.RegistryIndex), hookLineRegistryKey)
	ud, isUD := v.(*lua.LUserData)
	if !isUD || ud == nil || ud.Value == nil {
		return "", "", false
	}
	line, isLine := ud.Value.(hookLine)
	if !isLine || line.Hook == "" {
		return "", "", false
	}
	return line.Hook, line.Source, true
}
