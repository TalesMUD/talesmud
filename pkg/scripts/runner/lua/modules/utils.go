package modules

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
	lua "github.com/yuin/gopher-lua"

	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RegisterUtilsModule registers the tales.utils module
func RegisterUtilsModule(L *lua.LState, runner *luarunner.LuaRunner) int {
	mod := L.NewTable()

	// tales.utils.random(min, max) - Generate random number between min and max (inclusive)
	mod.RawSetString("random", L.NewFunction(func(L *lua.LState) int {
		min := L.CheckInt(1)
		max := L.CheckInt(2)
		if max < min {
			min, max = max, min
		}
		result := min + rand.Intn(max-min+1)
		L.Push(lua.LNumber(result))
		return 1
	}))

	// tales.utils.randomFloat() - Generate random float between 0 and 1
	mod.RawSetString("randomFloat", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(rand.Float64()))
		return 1
	}))

	// tales.utils.uuid() - Generate a UUID
	mod.RawSetString("uuid", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(uuid.New().String()))
		return 1
	}))

	// tales.utils.now() - Get current Unix timestamp
	mod.RawSetString("now", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(time.Now().Unix()))
		return 1
	}))

	// tales.utils.nowMs() - Get current Unix timestamp in milliseconds
	mod.RawSetString("nowMs", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(time.Now().UnixMilli()))
		return 1
	}))

	// tales.utils.formatTime(timestamp) - Format Unix timestamp to readable string
	mod.RawSetString("formatTime", L.NewFunction(func(L *lua.LState) int {
		timestamp := L.CheckInt64(1)
		t := time.Unix(timestamp, 0)
		L.Push(lua.LString(t.Format("2006-01-02 15:04:05")))
		return 1
	}))

	// tales.utils.roll(dice) - Roll dice in standard notation (e.g., "2d6", "1d20+5")
	mod.RawSetString("roll", L.NewFunction(func(L *lua.LState) int {
		dice := L.CheckString(1)
		result := rollDice(dice)
		L.Push(lua.LNumber(result))
		return 1
	}))

	// tales.utils.chance(percentage) - Return true with given percentage chance (0-100)
	mod.RawSetString("chance", L.NewFunction(func(L *lua.LState) int {
		percentage := L.CheckInt(1)
		roll := rand.Intn(100) + 1
		L.Push(lua.LBool(roll <= percentage))
		return 1
	}))

	// tales.utils.pick(array) - Pick a random element from array
	mod.RawSetString("pick", L.NewFunction(func(L *lua.LState) int {
		tbl := L.CheckTable(1)
		length := tbl.Len()
		if length == 0 {
			L.Push(lua.LNil)
			return 1
		}
		index := rand.Intn(length) + 1
		L.Push(tbl.RawGetInt(index))
		return 1
	}))

	// tales.utils.shuffle(array) - Shuffle array in place and return it
	mod.RawSetString("shuffle", L.NewFunction(func(L *lua.LState) int {
		tbl := L.CheckTable(1)
		length := tbl.Len()

		// Fisher-Yates shuffle
		for i := length; i > 1; i-- {
			j := rand.Intn(i) + 1
			// Swap elements at i and j
			vi := tbl.RawGetInt(i)
			vj := tbl.RawGetInt(j)
			tbl.RawSetInt(i, vj)
			tbl.RawSetInt(j, vi)
		}

		L.Push(tbl)
		return 1
	}))

	// tales.utils.clamp(value, min, max) - Clamp value between min and max
	mod.RawSetString("clamp", L.NewFunction(func(L *lua.LState) int {
		value := L.CheckNumber(1)
		min := L.CheckNumber(2)
		max := L.CheckNumber(3)

		if value < min {
			value = min
		} else if value > max {
			value = max
		}

		L.Push(value)
		return 1
	}))

	// tales.utils.lerp(a, b, t) - Linear interpolation between a and b
	mod.RawSetString("lerp", L.NewFunction(func(L *lua.LState) int {
		a := float64(L.CheckNumber(1))
		b := float64(L.CheckNumber(2))
		t := float64(L.CheckNumber(3))

		result := a + (b-a)*t
		L.Push(lua.LNumber(result))
		return 1
	}))

	L.Push(mod)
	return 1
}

// rollDice parses and rolls dice notation like "2d6" or "1d20+5"
func rollDice(notation string) int {
	var count, sides, modifier int
	count = 1
	sides = 6
	modifier = 0

	// Parse notation
	var current int
	var parsingModifier bool
	var modifierSign int = 1

	for _, c := range notation {
		switch {
		case c >= '0' && c <= '9':
			current = current*10 + int(c-'0')
		case c == 'd' || c == 'D':
			if current > 0 {
				count = current
			}
			current = 0
		case c == '+':
			if current > 0 {
				sides = current
			}
			current = 0
			parsingModifier = true
			modifierSign = 1
		case c == '-':
			if current > 0 {
				sides = current
			}
			current = 0
			parsingModifier = true
			modifierSign = -1
		}
	}

	// Handle final number
	if current > 0 {
		if parsingModifier {
			modifier = current * modifierSign
		} else {
			sides = current
		}
	}

	// Roll the dice
	total := 0
	for i := 0; i < count; i++ {
		total += rand.Intn(sides) + 1
	}

	return total + modifier
}
