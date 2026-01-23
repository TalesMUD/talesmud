package modules

import (
	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
)

// RegisterAllModules registers all Lua API modules with the runner
func RegisterAllModules(runner *luarunner.LuaRunner) {
	runner.RegisterModule("items", RegisterItemsModule)
	runner.RegisterModule("rooms", RegisterRoomsModule)
	runner.RegisterModule("characters", RegisterCharactersModule)
	runner.RegisterModule("npcs", RegisterNPCsModule)
	runner.RegisterModule("dialogs", RegisterDialogsModule)
	runner.RegisterModule("game", RegisterGameModule)
	runner.RegisterModule("utils", RegisterUtilsModule)
}
