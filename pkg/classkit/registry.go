package classkit

import "sync"

var (
	mu        sync.RWMutex
	defs      []*Def
	byID      map[string]*Def
	byResolve map[string]*Def
	fromPack  bool
	onChange  func()
)

// OnChange runs after the catalog is replaced. One callback; the characters
// package uses it to refresh the legacy class vars.
func OnChange(fn func()) {
	mu.Lock()
	onChange = fn
	mu.Unlock()
}

func fire() {
	mu.RLock()
	fn := onChange
	mu.RUnlock()
	if fn != nil {
		fn()
	}
}

// UseDefaults installs the generic Warrior / Rogue / Mage sample.
func UseDefaults() {
	install(sampleDefs(), false)
}

func install(list []*Def, pack bool) {
	idMap := map[string]*Def{}
	resMap := map[string]*Def{}
	cleaned := make([]*Def, 0, len(list))
	for _, d := range list {
		if d == nil || norm(d.ID) == "" {
			continue
		}
		d.ID = norm(d.ID)
		if d.SkillClass == "" {
			d.SkillClass = d.ID
		} else {
			d.SkillClass = norm(d.SkillClass)
		}
		for i := range d.Skills {
			d.Skills[i].ClassID = d.skillClass()
			if d.Skills[i].Target == "" {
				d.Skills[i].Target = "enemy"
			}
		}
		cleaned = append(cleaned, d)
		addKey(idMap, d.ID, d)
		addKey(resMap, d.ID, d)
		addKey(idMap, d.skillClass(), d)
		addKey(resMap, d.skillClass(), d)
		if d.Template != nil && d.Template.ID != "" {
			addKey(idMap, d.Template.ID, d)
			addKey(resMap, d.Template.ID, d)
		}
		for _, a := range d.Aliases {
			addKey(idMap, a, d)
			addKey(resMap, a, d)
		}
		if d.Name != "" {
			addKey(idMap, d.Name, d)
			addKey(resMap, d.Name, d)
		}
		for _, a := range d.Also {
			addKey(resMap, a, d)
		}
	}
	mu.Lock()
	defs = cleaned
	byID = idMap
	byResolve = resMap
	fromPack = pack
	mu.Unlock()
	fire()
}

func addKey(m map[string]*Def, key string, d *Def) {
	key = norm(key)
	if key == "" {
		return
	}
	if _, ok := m[key]; ok {
		return
	}
	m[key] = d
}

// FromPack reports whether the current rows came from a world pack.
func FromPack() bool {
	mu.RLock()
	defer mu.RUnlock()
	return fromPack
}

// Lookup resolves an id, alias, template id, or display name to its class.
// Shared rows (ranger, hunter) do not match here.
func Lookup(id string) *Def {
	mu.RLock()
	d := byID[norm(id)]
	mu.RUnlock()
	return d.copy()
}

func resolve(id string) *Def {
	mu.RLock()
	d := byResolve[norm(id)]
	mu.RUnlock()
	return d
}

// Playable returns classes that have a create template, in order.
func Playable() []*Def {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]*Def, 0, len(defs))
	for _, d := range defs {
		if d.Template != nil && d.Template.ID != "" {
			out = append(out, d.copy())
		}
	}
	return out
}

// All returns every loaded class.
func All() []*Def {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]*Def, 0, len(defs))
	for _, d := range defs {
		out = append(out, d.copy())
	}
	return out
}

// SkillClass is the id used on skill rows. Empty when id is not a catalog class.
func SkillClass(id string) string {
	d := Lookup(id)
	if d == nil {
		return ""
	}
	return d.skillClass()
}

// HotbarCap is >0 for a kit class. Shared rows such as ranger stay 0.
func HotbarCap(id string) int {
	d := Lookup(id)
	if d == nil {
		return 0
	}
	return d.Hotbar
}

// Balance returns the damage row for id, including shared rows.
func Balance(id string) (Row, bool) {
	d := resolve(id)
	if d == nil {
		return Row{}, false
	}
	return d.Balance, true
}

// SetBalance replaces one class row. id may be an alias. Used by tests.
func SetBalance(id string, row Row) bool {
	mu.Lock()
	defer mu.Unlock()
	d := byResolve[norm(id)]
	if d == nil {
		return false
	}
	d.Balance = row
	return true
}

// BalanceMap keys each class by its id and, when different, its skill class.
func BalanceMap() map[string]Row {
	mu.RLock()
	defer mu.RUnlock()
	out := map[string]Row{}
	for _, d := range defs {
		out[d.ID] = d.Balance
		if d.skillClass() != d.ID {
			out[d.skillClass()] = d.Balance
		}
	}
	return out
}

// HPMultiplier is the create-time max HP scale. Unknown classes are 1.
func HPMultiplier(id string) float64 {
	d := resolve(id)
	if d == nil || d.HPMultiplier <= 0 {
		return 1
	}
	return d.HPMultiplier
}

// Races is the create allow-list. Nil when the class is unknown.
func Races(id string) []string {
	d := resolve(id)
	if d == nil {
		return nil
	}
	return append([]string(nil), d.Races...)
}

// Primary is the basic-attack attribute, or empty when id is not in the catalog.
func Primary(id string) string {
	d := resolve(id)
	if d == nil {
		return ""
	}
	return d.Primary
}

// Caster reports a mana class.
func Caster(id string) bool {
	d := resolve(id)
	return d != nil && d.Caster
}

// Inscribes reports a class whose basic hit applies the inscribe primitive.
func Inscribes(id string) bool {
	d := resolve(id)
	return d != nil && d.InscribeBasic
}

// HasGrit reports the soak primitive.
func HasGrit(id string) bool {
	return GritOf(id) != nil
}

// GritOf returns the soak block, or nil.
func GritOf(id string) *Grit {
	d := resolve(id)
	if d == nil || d.Grit == nil {
		return nil
	}
	g := *d.Grit
	return &g
}

// Charges is brace and slip uses granted at combat start.
func Charges(id string) (brace, slip int) {
	d := resolve(id)
	if d == nil {
		return 0, 0
	}
	return d.BraceCharges, d.SlipCharges
}

// BoltRig is bolt and rig uses granted at combat start.
func BoltRig(id string) (bolt, rig int) {
	d := resolve(id)
	if d == nil {
		return 0, 0
	}
	return d.BoltCharges, d.RigCharges
}

// Skills flattens every kit skill in class order.
func Skills() []SkillSpec {
	mu.RLock()
	defer mu.RUnlock()
	var out []SkillSpec
	for _, d := range defs {
		for _, s := range d.Skills {
			s.ClassID = d.skillClass()
			out = append(out, s)
		}
	}
	return out
}

// RaceBlurbOverrides collects per-race blurbs supplied by the pack.
func RaceBlurbOverrides() map[string]string {
	mu.RLock()
	defer mu.RUnlock()
	out := map[string]string{}
	for _, d := range defs {
		for k, v := range d.RaceBlurbs {
			if norm(k) == "" || v == "" {
				continue
			}
			out[norm(k)] = v
		}
	}
	return out
}
