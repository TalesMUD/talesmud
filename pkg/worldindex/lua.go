package worldindex

import (
	"strings"
	"unicode"
)

// callArg describes one argument of a tales.* call that is an entity id or an exit name.
type callArg struct {
	index    int
	kind     Kind
	exitName bool
}

type callSpec struct {
	module string
	name   string
	how    string
	args   []callArg
}

// luaCalls is the static catalog of tales.* functions whose string arguments are world ids.
// Argument indexes are 0-based and follow pkg/scripts/runner/lua/modules.
var luaCalls = []callSpec{
	{module: "game", name: "revealExit", how: "revealExit", args: []callArg{
		{index: 0, kind: KindRoom},
		{index: 1, exitName: true},
	}},
	{module: "game", name: "giveItem", how: "giveItem", args: []callArg{
		{index: 1, kind: KindItem},
	}},
	{module: "game", name: "hasItem", how: "hasItem", args: []callArg{
		{index: 1, kind: KindItem},
	}},
	{module: "game", name: "hasCollectedItem", how: "hasCollectedItem", args: []callArg{
		{index: 1, kind: KindItem},
	}},
	{module: "game", name: "resetCollectedItem", how: "resetCollectedItem", args: []callArg{
		{index: 1, kind: KindItem},
	}},
	{module: "combat", name: "summon", how: "summon", args: []callArg{
		{index: 0, kind: KindNPC},
	}},
	{module: "npcs", name: "moveTo", how: "moveTo", args: []callArg{
		{index: 0, kind: KindNPC},
		{index: 1, kind: KindRoom},
	}},
	{module: "npcs", name: "spawnFromTemplate", how: "spawnFromTemplate", args: []callArg{
		{index: 0, kind: KindNPC},
		{index: 1, kind: KindRoom},
	}},
	{module: "characters", name: "teleport", how: "teleport", args: []callArg{
		{index: 1, kind: KindRoom},
	}},
	{module: "characters", name: "setBind", how: "setBind", args: []callArg{
		{index: 1, kind: KindRoom},
	}},
	{module: "quests", name: "accept", how: "startQuest", args: []callArg{
		{index: 1, kind: KindQuest},
	}},
	{module: "quests", name: "grantQuest", how: "startQuest", args: []callArg{
		{index: 1, kind: KindQuest},
	}},
	{module: "quests", name: "complete", how: "completeQuest", args: []callArg{
		{index: 1, kind: KindQuest},
	}},
}

// CurrentRoom is the room id of a revealExit whose argument is the script's
// current room (ctx.roomID, ctx.room.ID, or a local copied from those).
// Reachability applies it to each room that actually runs the script.
const CurrentRoom = "@ctx.room"

const currentRoomObj = "@ctx.roomobj"

func specByName(module, name string) (callSpec, bool) {
	for _, spec := range luaCalls {
		if spec.module == module && spec.name == name {
			return spec, true
		}
	}
	return callSpec{}, false
}

// ExtractLuaEdges reads tales.module.func(...) calls and records edges for
// string-literal ids. A local assigned a string literal earlier in the script
// counts as that literal (local roomID = "R1"; revealExit(roomID, "down")).
// ctx.roomID, ctx.room.ID, and locals copied from them count as CurrentRoom.
func ExtractLuaEdges(scriptID, source string) []Edge {
	if strings.TrimSpace(source) == "" || scriptID == "" {
		return nil
	}
	s := stripLuaComments(source)
	env := map[string]string{}
	var edges []Edge
	i := 0
	for i < len(s) {
		if s[i] == '"' || s[i] == '\'' {
			_, next := readQuoted(s, i)
			i = next
			continue
		}
		if isIdentStart(rune(s[i])) {
			start := i
			i++
			for i < len(s) && isIdentCont(rune(s[i])) {
				i++
			}
			ident := s[start:i]
			if ident == "tales" {
				if edgesAt, next, ok := parseTalesCall(s, start, scriptID, env); ok {
					edges = append(edges, edgesAt...)
					i = next
					continue
				}
			}
			if lit, next, ok := parseAssignment(s, i, ident, env); ok {
				if lit == "" {
					delete(env, ident)
				} else {
					env[ident] = lit
				}
				i = next
				continue
			}
			continue
		}
		i++
	}
	return edges
}

func parseAssignment(s string, afterIdent int, name string, env map[string]string) (lit string, next int, ok bool) {
	j := skipSpace(s, afterIdent)
	if j >= len(s) || s[j] != '=' || (j+1 < len(s) && s[j+1] == '=') {
		return "", 0, false
	}
	j = skipSpace(s, j+1)
	if j >= len(s) {
		return "", 0, false
	}
	if val, n, known := readStatic(s, j, env); known {
		return val, n, true
	}
	// Reassigned to something we cannot resolve. Drop the previous binding.
	return "", j, true
}

func parseTalesCall(s string, talesAt int, scriptID string, env map[string]string) ([]Edge, int, bool) {
	j := talesAt + len("tales")
	j = skipSpace(s, j)
	if j >= len(s) || s[j] != '.' {
		return nil, 0, false
	}
	j = skipSpace(s, j+1)
	module, j := readIdent(s, j)
	j = skipSpace(s, j)
	if module == "" || j >= len(s) || s[j] != '.' {
		return nil, 0, false
	}
	j = skipSpace(s, j+1)
	name, j := readIdent(s, j)
	j = skipSpace(s, j)
	if name == "" || j >= len(s) || s[j] != '(' {
		return nil, 0, false
	}
	args, next, ok := parseArgs(s, j, env)
	if !ok {
		return nil, 0, false
	}
	spec, known := specByName(module, name)
	if !known {
		return nil, next, true
	}
	var edges []Edge
	var exitName string
	var roomID string
	for _, arg := range spec.args {
		if arg.index < 0 || arg.index >= len(args) {
			continue
		}
		val := args[arg.index]
		if val == "" {
			continue
		}
		if arg.exitName {
			exitName = val
			continue
		}
		if arg.kind == KindRoom && spec.name == "revealExit" {
			roomID = val
		}
		edges = append(edges, Edge{
			FromType: KindScript,
			FromID:   scriptID,
			Field:    "code",
			ToType:   arg.kind,
			ToID:     val,
			How:      spec.how,
		})
	}
	if spec.name == "revealExit" && roomID != "" && exitName != "" {
		for i := range edges {
			if edges[i].ToType == KindRoom && edges[i].ToID == roomID && edges[i].How == "revealExit" {
				edges[i].ExitName = exitName
				edges[i].How = "revealExit"
			}
		}
	} else if spec.name == "revealExit" {
		// A reveal without both literals is not a static reveal edge.
		filtered := edges[:0]
		for _, e := range edges {
			if e.How == "revealExit" && e.ExitName == "" {
				continue
			}
			filtered = append(filtered, e)
		}
		edges = filtered
	}
	return edges, next, true
}

// parseArgs reads the argument list starting at '('. Empty string means the arg was not a static literal.
func parseArgs(s string, paren int, env map[string]string) ([]string, int, bool) {
	if paren >= len(s) || s[paren] != '(' {
		return nil, paren, false
	}
	var args []string
	i := paren + 1
	for i < len(s) {
		i = skipSpace(s, i)
		if i >= len(s) {
			return nil, i, false
		}
		if s[i] == ')' {
			return args, i + 1, true
		}
		val, next := readArg(s, i, env)
		args = append(args, val)
		i = skipSpace(s, next)
		if i < len(s) && s[i] == ',' {
			i++
			continue
		}
		if i < len(s) && s[i] == ')' {
			return args, i + 1, true
		}
		// Unexpected token. Bail out of this call.
		return args, i, true
	}
	return args, i, false
}

func readArg(s string, i int, env map[string]string) (string, int) {
	i = skipSpace(s, i)
	if i >= len(s) {
		return "", i
	}
	if val, next, ok := readStatic(s, i, env); ok && val != currentRoomObj {
		return val, next
	}
	return "", skipValue(s, i)
}

// readStatic resolves a string literal, a copied local, or the current room.
func readStatic(s string, i int, env map[string]string) (string, int, bool) {
	i = skipSpace(s, i)
	if i >= len(s) {
		return "", i, false
	}
	if s[i] == '"' || s[i] == '\'' {
		val, next := readQuoted(s, i)
		return val, next, true
	}
	if !isIdentStart(rune(s[i])) {
		return "", i, false
	}
	name, next := readIdent(s, i)
	parts := []string{name}
	j := next
	for {
		k := skipSpace(s, j)
		if k >= len(s) || s[k] != '.' {
			break
		}
		k = skipSpace(s, k+1)
		field, n := readIdent(s, k)
		if field == "" {
			break
		}
		parts = append(parts, field)
		j = n
	}
	if len(parts) == 1 {
		if env != nil {
			if lit, ok := env[name]; ok && lit != "" {
				return lit, j, true
			}
		}
		return "", j, false
	}
	if resolved, ok := resolveMember(parts, env); ok {
		return resolved, j, true
	}
	return "", j, false
}

func resolveMember(parts []string, env map[string]string) (string, bool) {
	if len(parts) == 2 && parts[0] == "ctx" && (parts[1] == "roomID" || parts[1] == "roomId") {
		return CurrentRoom, true
	}
	if len(parts) == 2 && parts[0] == "ctx" && parts[1] == "room" {
		return currentRoomObj, true
	}
	if len(parts) == 3 && parts[0] == "ctx" && parts[1] == "room" && (parts[2] == "ID" || parts[2] == "id") {
		return CurrentRoom, true
	}
	if len(parts) == 2 && (parts[1] == "ID" || parts[1] == "id") {
		bound := ""
		if env != nil {
			bound = env[parts[0]]
		}
		if bound == currentRoomObj || bound == CurrentRoom {
			return CurrentRoom, true
		}
	}
	return "", false
}

func skipValue(s string, i int) int {
	depth := 0
	for i < len(s) {
		switch s[i] {
		case '"', '\'':
			_, i = readQuoted(s, i)
			continue
		case '(', '{', '[':
			depth++
		case ')', '}', ']':
			if depth == 0 {
				return i
			}
			depth--
		case ',':
			if depth == 0 {
				return i
			}
		}
		i++
	}
	return i
}

func readIdent(s string, i int) (string, int) {
	if i >= len(s) || !isIdentStart(rune(s[i])) {
		return "", i
	}
	start := i
	i++
	for i < len(s) && isIdentCont(rune(s[i])) {
		i++
	}
	return s[start:i], i
}

func readQuoted(s string, i int) (string, int) {
	if i >= len(s) {
		return "", i
	}
	quote := s[i]
	i++
	var b strings.Builder
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			b.WriteByte(s[i+1])
			i += 2
			continue
		}
		if s[i] == quote {
			return b.String(), i + 1
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String(), i
}

func skipSpace(s string, i int) int {
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}
	return i
}

func isIdentStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentCont(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func stripLuaComments(src string) string {
	var b strings.Builder
	b.Grow(len(src))
	i := 0
	for i < len(src) {
		if i+3 < len(src) && src[i:i+4] == "--[[" {
			end := strings.Index(src[i+4:], "]]")
			if end < 0 {
				break
			}
			i += 4 + end + 2
			b.WriteByte(' ')
			continue
		}
		if i+1 < len(src) && src[i] == '-' && src[i+1] == '-' {
			for i < len(src) && src[i] != '\n' {
				i++
			}
			continue
		}
		if src[i] == '"' || src[i] == '\'' {
			_, next := readQuoted(src, i)
			b.WriteString(src[i:next])
			i = next
			continue
		}
		b.WriteByte(src[i])
		i++
	}
	return b.String()
}
