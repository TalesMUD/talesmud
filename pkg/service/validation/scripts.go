package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/talesmud/talesmud/pkg/scripts"
)

var knownGameFunctions = map[string]int{
	"msgToRoom":          2,
	"msgToCharacter":     2,
	"msgToUser":          2,
	"broadcast":          1,
	"msgToRoomExcept":    3,
	"log":                2,
	"hasItem":            2,
	"getFlag":            2,
	"setFlag":            3,
	"revealExit":         3,
	"hasRevealedExit":    3,
	"giveItem":           2,
	"hasEquipped":        2,
	"hasCollectedItem":   2,
	"resetCollectedItem": 2,
}

var (
	reGameCall  = regexp.MustCompile(`tales\.game\.(\w+)\s*\(`)
	reCtxRoomID = regexp.MustCompile(`ctx\.roomID\b`)
)

func ValidateLuaScript(script *scripts.Script, isOnEnterScript bool) Result {
	result := NewResult()
	if script == nil || script.Code == "" {
		return result
	}
	scriptID := scriptEntityID(script)
	if script.GetLanguage() != scripts.ScriptLanguageLua {
		result.Add(Error("non_lua_script", "script", scriptID, "language", "", string(script.GetLanguage()), "Only Lua scripts are supported in the Creator."))
		return result
	}

	matches := reGameCall.FindAllStringSubmatchIndex(script.Code, -1)
	for _, loc := range matches {
		funcName := script.Code[loc[2]:loc[3]]
		callStart := loc[0]
		parenStart := loc[1] - 1

		expectedArgs, known := knownGameFunctions[funcName]
		if !known {
			result.Add(Warning("unknown_lua_game_function", "script", scriptID, "code", "lua_function", funcName, fmt.Sprintf("Script %s (%s): calls unknown function tales.game.%s()", scriptID, script.Name, funcName)))
			continue
		}

		argCount := countArguments(script.Code, parenStart)
		if argCount >= 0 && argCount != expectedArgs {
			snippet := extractSnippet(script.Code, callStart, 80)
			result.Add(Warning("wrong_lua_game_function_arg_count", "script", scriptID, "code", "lua_function", funcName, fmt.Sprintf("Script %s (%s): tales.game.%s() called with %d args, expected %d - %s", scriptID, script.Name, funcName, argCount, expectedArgs, snippet)))
		}
	}

	if isOnEnterScript || script.Type == scripts.ScriptTypeRoom {
		if reCtxRoomID.MatchString(script.Code) {
			result.Add(Warning("room_script_ctx_room_id", "script", scriptID, "code", "", "ctx.roomID", fmt.Sprintf("Script %s (%s): uses ctx.roomID in a room-type script (onEnter); use ctx.room.ID instead", scriptID, script.Name)))
		}
	}

	return result
}

func scriptEntityID(script *scripts.Script) string {
	if script == nil || script.Entity == nil {
		return ""
	}
	return script.ID
}

func countArguments(code string, parenPos int) int {
	if parenPos >= len(code) || code[parenPos] != '(' {
		return -1
	}

	depth := 0
	commas := 0
	inString := false
	stringChar := byte(0)
	hasContent := false

	for i := parenPos; i < len(code); i++ {
		ch := code[i]

		if inString {
			if ch == stringChar && (i == 0 || code[i-1] != '\\') {
				inString = false
			}
			continue
		}

		if ch == '[' && i+1 < len(code) && code[i+1] == '[' {
			end := strings.Index(code[i+2:], "]]")
			if end >= 0 {
				i += 2 + end + 1
			}
			hasContent = true
			continue
		}

		if ch == '"' || ch == '\'' {
			inString = true
			stringChar = ch
			hasContent = true
			continue
		}

		if ch == '-' && i+1 < len(code) && code[i+1] == '-' {
			nl := strings.IndexByte(code[i:], '\n')
			if nl >= 0 {
				i += nl
			} else {
				i = len(code) - 1
			}
			continue
		}

		switch ch {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				if !hasContent {
					return 0
				}
				return commas + 1
			}
		case ',':
			if depth == 1 {
				commas++
			}
		default:
			if depth == 1 && ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
				hasContent = true
			}
		}
	}

	return -1
}

func extractSnippet(code string, pos, maxLen int) string {
	end := pos + maxLen
	if end > len(code) {
		end = len(code)
	}
	snippet := code[pos:end]
	if nl := strings.IndexByte(snippet, '\n'); nl >= 0 {
		snippet = snippet[:nl]
	}
	snippet = strings.TrimSpace(snippet)
	if len(snippet) > maxLen {
		snippet = snippet[:maxLen-3] + "..."
	}
	return snippet
}
