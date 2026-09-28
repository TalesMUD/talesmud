// Package ruleset is the process profile beside config/combat_balance.yaml.
// It does not own level gap, threat, reward scale, or class tuning.
// Missing file and the shipped defaults behave like today's play.
package ruleset

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/talesmud/talesmud/pkg/resources"
)

const (
	ModeAuto    = "auto"
	ModeTrainer = "trainer"

	RespawnBindpoint = "bindpoint"
	RespawnNextReset = "next_reset"

	PacingAuto      = "auto"
	PacingTurnBased = "turn_based"

	// SafeStay leaves a disconnecting character in a real room.
	// SafeBind moves a mid-combat disconnect to BoundRoomID.
	// SafeStart moves that disconnect to the world's start room.
	// A room that is about to be destroyed still relocates, whatever this says.
	SafeStay  = "stay"
	SafeBind  = "bind"
	SafeStart = "start"

	// DisconnectContinue leaves a dropped connection in the fight.
	// DisconnectRelease ends that fight without a defeat and may move the character.
	DisconnectContinue = "continue"
	DisconnectRelease  = "release"

	// BareAttackAsk replies "Attack whom?" when no target is named.
	// BareAttackFirst starts a fight against the first hostile in the room.
	BareAttackAsk   = "ask"
	BareAttackFirst = "first_hostile"

	defaultPath = "config/ruleset.yaml"
)

var forbiddenKeys = map[string]bool{
	"difficulty_multipliers": true,
	"named_overrides":        true,
	"level_gap":              true,
	"threat":                 true,
	"reward_scale":           true,
	"class_balance":          true,
	"first_kill_bonus":       true,
}

type fileShape struct {
	Progression struct {
		LevelCap           int32         `yaml:"level_cap"`
		LevelUpMode        string        `yaml:"level_up_mode"`
		XPRequired         map[int]int32 `yaml:"xp_required"`
		BaseXPByEnemyLevel map[int]int64 `yaml:"base_xp_by_enemy_level"`
	} `yaml:"progression"`
	Death struct {
		XPLossPercent    float64 `yaml:"xp_loss_percent"`
		GoldLossFlat     int64   `yaml:"gold_loss_flat"`
		GoldLossPercent  float64 `yaml:"gold_loss_percent"`
		Respawn          string  `yaml:"respawn"`
		RespawnHPPercent float64 `yaml:"respawn_hp_percent"`
		DamageArmor      *bool   `yaml:"damage_armor"`
	} `yaml:"death"`
	NewDay struct {
		FullHeal *bool  `yaml:"full_heal"`
		Timezone string `yaml:"timezone"`
	} `yaml:"new_day"`
	Resources map[string]struct {
		Allowance int    `yaml:"allowance"`
		Reset     string `yaml:"reset"`
		Timezone  string `yaml:"timezone"`
		Interval  string `yaml:"interval"`
	} `yaml:"resources"`
	Combat struct {
		Pacing     string `yaml:"pacing"`
		SafeRoom   string `yaml:"safe_room"`
		Disconnect string `yaml:"disconnect"`
		BareAttack string `yaml:"bare_attack"`
	} `yaml:"combat"`
}

type resourceSpec struct {
	allowance int
	reset     string
	timezone  string
	interval  time.Duration
}

type state struct {
	levelCap    int32
	levelUpMode string
	xpRequired  map[int32]int32
	baseXP      map[int32]int64
	death       DeathPolicy
	fullHeal    bool
	timezone    string
	resources   map[string]resourceSpec
	pacing      string
	safeRoom    string
	disconnect  string
	bareAttack  string
}

var (
	mu      sync.RWMutex
	current = builtin()
)

func init() {
	if err := LoadDefault(); err != nil {
		log.WithError(err).Warn("Ruleset file failed to load; using built-in defaults")
		mu.Lock()
		current = builtin()
		mu.Unlock()
	}
}

func builtin() state {
	return state{
		levelCap:    50,
		levelUpMode: ModeAuto,
		death: DeathPolicy{
			XPLossPercent:    10,
			GoldLossFlat:     1,
			GoldLossPercent:  0,
			Respawn:          RespawnBindpoint,
			RespawnHPPercent: 50,
			DamageArmor:      true,
		},
		fullHeal:   false,
		timezone:   "UTC",
		pacing:     PacingAuto,
		safeRoom:   SafeStay,
		disconnect: DisconnectContinue,
		bareAttack: BareAttackAsk,
	}
}

// LoadDefault reads RULESET or config/ruleset.yaml. A missing default path
// leaves the built-in profile in place.
func LoadDefault() error {
	path := strings.TrimSpace(os.Getenv("RULESET"))
	if path == "" {
		path = findConfig()
		if path == "" {
			mu.Lock()
			current = builtin()
			mu.Unlock()
			return nil
		}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read ruleset %s: %w", path, err)
	}
	return LoadBytes(raw)
}

// LoadBytes parses a ruleset document and installs it.
func LoadBytes(raw []byte) error {
	var node yaml.Node
	if err := yaml.Unmarshal(raw, &node); err != nil {
		return fmt.Errorf("parse ruleset: %w", err)
	}
	if err := rejectForbidden(&node); err != nil {
		return err
	}
	next, err := decode(raw)
	if err != nil {
		return err
	}
	mu.Lock()
	current = next
	mu.Unlock()
	return nil
}

// Reset restores the built-in profile, then loads the default file again.
func Reset() {
	mu.Lock()
	current = builtin()
	mu.Unlock()
	_ = LoadDefault()
}

func decode(raw []byte) (state, error) {
	base := builtin()
	var file fileShape
	file.Progression.LevelCap = base.levelCap
	file.Progression.LevelUpMode = base.levelUpMode
	file.Death.XPLossPercent = base.death.XPLossPercent
	file.Death.GoldLossFlat = base.death.GoldLossFlat
	file.Death.GoldLossPercent = base.death.GoldLossPercent
	file.Death.Respawn = base.death.Respawn
	file.Death.RespawnHPPercent = base.death.RespawnHPPercent
	armor := base.death.DamageArmor
	file.Death.DamageArmor = &armor
	heal := base.fullHeal
	file.NewDay.FullHeal = &heal
	file.NewDay.Timezone = base.timezone
	file.Combat.Pacing = base.pacing
	file.Combat.SafeRoom = base.safeRoom
	file.Combat.Disconnect = base.disconnect
	file.Combat.BareAttack = base.bareAttack
	if err := yaml.Unmarshal(raw, &file); err != nil {
		return state{}, fmt.Errorf("parse ruleset: %w", err)
	}

	next := builtin()
	next.levelCap = file.Progression.LevelCap
	if next.levelCap <= 0 {
		next.levelCap = 50
	}
	next.levelUpMode = strings.TrimSpace(file.Progression.LevelUpMode)
	if next.levelUpMode == "" {
		next.levelUpMode = ModeAuto
	}
	if next.levelUpMode != ModeAuto && next.levelUpMode != ModeTrainer {
		return state{}, fmt.Errorf("level_up_mode %q", next.levelUpMode)
	}
	if len(file.Progression.XPRequired) > 0 {
		next.xpRequired = map[int32]int32{}
		for level, xp := range file.Progression.XPRequired {
			if level < 0 {
				continue
			}
			next.xpRequired[int32(level)] = xp
		}
	}
	if len(file.Progression.BaseXPByEnemyLevel) > 0 {
		next.baseXP = map[int32]int64{}
		for level, xp := range file.Progression.BaseXPByEnemyLevel {
			if level < 0 {
				continue
			}
			next.baseXP[int32(level)] = xp
		}
	}
	next.death = DeathPolicy{
		XPLossPercent:    file.Death.XPLossPercent,
		GoldLossFlat:     file.Death.GoldLossFlat,
		GoldLossPercent:  file.Death.GoldLossPercent,
		Respawn:          strings.TrimSpace(file.Death.Respawn),
		RespawnHPPercent: file.Death.RespawnHPPercent,
		DamageArmor:      file.Death.DamageArmor == nil || *file.Death.DamageArmor,
	}
	if next.death.Respawn == "" {
		next.death.Respawn = RespawnBindpoint
	}
	if next.death.Respawn != RespawnBindpoint && next.death.Respawn != RespawnNextReset {
		return state{}, fmt.Errorf("respawn %q", next.death.Respawn)
	}
	if next.death.XPLossPercent < 0 || next.death.GoldLossPercent < 0 || next.death.RespawnHPPercent < 0 {
		return state{}, fmt.Errorf("death percents must be >= 0")
	}
	next.fullHeal = file.NewDay.FullHeal != nil && *file.NewDay.FullHeal
	next.timezone = strings.TrimSpace(file.NewDay.Timezone)
	if next.timezone == "" {
		next.timezone = "UTC"
	}
	next.pacing = strings.TrimSpace(file.Combat.Pacing)
	if next.pacing == "" {
		next.pacing = PacingAuto
	}
	if next.pacing != PacingAuto && next.pacing != PacingTurnBased {
		return state{}, fmt.Errorf("combat pacing %q", next.pacing)
	}
	next.safeRoom = strings.TrimSpace(file.Combat.SafeRoom)
	if next.safeRoom == "" {
		next.safeRoom = SafeStay
	}
	if next.safeRoom != SafeStay && next.safeRoom != SafeBind && next.safeRoom != SafeStart {
		return state{}, fmt.Errorf("combat safe_room %q", next.safeRoom)
	}
	next.disconnect = strings.TrimSpace(file.Combat.Disconnect)
	if next.disconnect == "" {
		next.disconnect = DisconnectContinue
	}
	if next.disconnect != DisconnectContinue && next.disconnect != DisconnectRelease {
		return state{}, fmt.Errorf("combat disconnect %q", next.disconnect)
	}
	next.bareAttack = strings.TrimSpace(file.Combat.BareAttack)
	if next.bareAttack == "" {
		next.bareAttack = BareAttackAsk
	}
	if next.bareAttack != BareAttackAsk && next.bareAttack != BareAttackFirst {
		return state{}, fmt.Errorf("combat bare_attack %q", next.bareAttack)
	}
	if len(file.Resources) > 0 {
		next.resources = map[string]resourceSpec{}
		for key, spec := range file.Resources {
			key = strings.TrimSpace(key)
			if key == "" {
				continue
			}
			item := resourceSpec{
				allowance: spec.Allowance,
				reset:     strings.TrimSpace(spec.Reset),
				timezone:  strings.TrimSpace(spec.Timezone),
			}
			if item.reset == "" {
				item.reset = resources.ResetCalendar
			}
			if spec.Interval != "" {
				d, err := time.ParseDuration(spec.Interval)
				if err != nil {
					return state{}, fmt.Errorf("resource %s interval: %w", key, err)
				}
				item.interval = d
			}
			if item.allowance < 0 {
				item.allowance = 0
			}
			next.resources[key] = item
		}
	}
	return next, nil
}

func rejectForbidden(node *yaml.Node) error {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case yaml.DocumentNode, yaml.SequenceNode:
		for _, child := range node.Content {
			if err := rejectForbidden(child); err != nil {
				return err
			}
		}
	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			key := node.Content[i].Value
			if forbiddenKeys[key] {
				return fmt.Errorf("ruleset must not set %q; it belongs to config/combat_balance.yaml", key)
			}
			if err := rejectForbidden(node.Content[i+1]); err != nil {
				return err
			}
		}
	}
	return nil
}

func findConfig() string {
	candidates := []string{defaultPath, filepath.Join(".", defaultPath)}
	if cwd, err := os.Getwd(); err == nil {
		dir := cwd
		for i := 0; i < 8; i++ {
			candidates = append(candidates, filepath.Join(dir, defaultPath))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	seen := map[string]bool{}
	for _, path := range candidates {
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		if st, err := os.Stat(path); err == nil && !st.IsDir() {
			return path
		}
	}
	return ""
}

// LevelCap is the global level cap. A character MaxLevelCap still wins when lower.
// A cap above 50 is honored only when the XP table defines that level.
func LevelCap() int32 {
	mu.RLock()
	defer mu.RUnlock()
	cap := current.levelCap
	if cap <= 0 {
		cap = 50
	}
	if cap > 50 {
		if _, ok := current.xpRequired[cap]; !ok {
			return 50
		}
	}
	return cap
}

// LevelUpMode is auto or trainer.
func LevelUpMode() string {
	mu.RLock()
	defer mu.RUnlock()
	if current.levelUpMode == "" {
		return ModeAuto
	}
	return current.levelUpMode
}

// Pacing is auto or turn_based. Combat reads this when it resolves turns.
func Pacing() string {
	mu.RLock()
	defer mu.RUnlock()
	if current.pacing == "" {
		return PacingAuto
	}
	return current.pacing
}

// Disconnect is continue or release. continue is the unconfigured default:
// closing a session does not end the fight and does not move the character.
func Disconnect() string {
	mu.RLock()
	defer mu.RUnlock()
	if current.disconnect == DisconnectRelease {
		return DisconnectRelease
	}
	return DisconnectContinue
}

// BareAttack is ask or first_hostile. ask is the unconfigured default.
// It only applies outside combat. A bare attack during a fight still swings.
func BareAttack() string {
	mu.RLock()
	defer mu.RUnlock()
	if current.bareAttack == BareAttackFirst {
		return BareAttackFirst
	}
	return BareAttackAsk
}

// SafeRoom is stay, bind, or start. stay is the unconfigured default.
// It is used only when Disconnect is release. A destroyed instance room
// still relocates on instance teardown.
func SafeRoom() string {
	mu.RLock()
	defer mu.RUnlock()
	switch current.safeRoom {
	case SafeBind, SafeStart:
		return current.safeRoom
	default:
		return SafeStay
	}
}

// XPRequired returns a cumulative XP threshold when the profile defines one.
func XPRequired(level int32) (int32, bool) {
	mu.RLock()
	defer mu.RUnlock()
	if len(current.xpRequired) == 0 {
		return 0, false
	}
	xp, ok := current.xpRequired[level]
	return xp, ok
}

// BaseXPForEnemyLevel is the base XP for an enemy whose authored reward is 0.
// The boolean is false when this level is not in the table.
func BaseXPForEnemyLevel(level int32) (int64, bool) {
	mu.RLock()
	defer mu.RUnlock()
	if len(current.baseXP) == 0 {
		return 0, false
	}
	xp, ok := current.baseXP[level]
	return xp, ok
}

// ResourceAllowances is the catalog for the refilling store. Empty by default.
func ResourceAllowances() []resources.Allowance {
	mu.RLock()
	defer mu.RUnlock()
	keys := make([]string, 0, len(current.resources))
	for key := range current.resources {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]resources.Allowance, 0, len(keys))
	for _, key := range keys {
		spec := current.resources[key]
		out = append(out, resources.Allowance{
			Key:      key,
			Amount:   spec.allowance,
			Reset:    spec.reset,
			Timezone: spec.timezone,
			Interval: spec.interval,
		})
	}
	return out
}

// SetLevelUpMode overrides the mode until Reset. Unknown values are ignored.
func SetLevelUpMode(mode string) {
	mode = strings.TrimSpace(mode)
	if mode != ModeAuto && mode != ModeTrainer {
		return
	}
	mu.Lock()
	current.levelUpMode = mode
	mu.Unlock()
}

// SetLevelCap overrides the global cap until Reset.
func SetLevelCap(cap int32) {
	if cap <= 0 {
		cap = 50
	}
	mu.Lock()
	current.levelCap = cap
	mu.Unlock()
}

// SetXPRequired replaces the cumulative table. Nil clears it.
func SetXPRequired(table map[int32]int32) {
	mu.Lock()
	if len(table) == 0 {
		current.xpRequired = nil
	} else {
		current.xpRequired = map[int32]int32{}
		for k, v := range table {
			current.xpRequired[k] = v
		}
	}
	mu.Unlock()
}

// SetBaseEnemyXP replaces the fallback enemy base table. Nil clears it.
func SetBaseEnemyXP(table map[int32]int64) {
	mu.Lock()
	if len(table) == 0 {
		current.baseXP = nil
	} else {
		current.baseXP = map[int32]int64{}
		for k, v := range table {
			current.baseXP[k] = v
		}
	}
	mu.Unlock()
}

// SetDeath replaces the death policy until Reset.
func SetDeath(policy DeathPolicy) {
	mu.Lock()
	current.death = policy
	mu.Unlock()
}

// SetDisconnect overrides combat.disconnect until Reset. Unknown values are ignored.
func SetDisconnect(mode string) {
	mode = strings.TrimSpace(mode)
	if mode != DisconnectContinue && mode != DisconnectRelease {
		return
	}
	mu.Lock()
	current.disconnect = mode
	mu.Unlock()
}

// SetBareAttack overrides combat.bare_attack until Reset. Unknown values are ignored.
func SetBareAttack(mode string) {
	mode = strings.TrimSpace(mode)
	if mode != BareAttackAsk && mode != BareAttackFirst {
		return
	}
	mu.Lock()
	current.bareAttack = mode
	mu.Unlock()
}

// SetSafeRoom overrides the disconnect room until Reset. Unknown values are ignored.
func SetSafeRoom(mode string) {
	mode = strings.TrimSpace(mode)
	if mode != SafeStay && mode != SafeBind && mode != SafeStart {
		return
	}
	mu.Lock()
	current.safeRoom = mode
	mu.Unlock()
}

// SetPacing overrides combat pacing until Reset. Unknown values are ignored.
func SetPacing(mode string) {
	mode = strings.TrimSpace(mode)
	if mode != PacingAuto && mode != PacingTurnBased {
		return
	}
	mu.Lock()
	current.pacing = mode
	mu.Unlock()
}

// SetNewDay overrides the dawn heal switch until Reset.
func SetNewDay(fullHeal bool, timezone string) {
	if strings.TrimSpace(timezone) == "" {
		timezone = "UTC"
	}
	mu.Lock()
	current.fullHeal = fullHeal
	current.timezone = timezone
	mu.Unlock()
}
