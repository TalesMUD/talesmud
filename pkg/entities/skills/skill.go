package skills

import "github.com/talesmud/talesmud/pkg/entities"

// TargetType defines what a skill can target
type TargetType string

const (
	TargetEnemy      TargetType = "enemy"
	TargetSelf       TargetType = "self"
	TargetAllEnemies TargetType = "all_enemies"
)

// ResourceType defines what resource a skill consumes
type ResourceType string

const (
	ResourceMana     ResourceType = "mana"
	ResourceCooldown ResourceType = "cooldown"
)

// EffectType defines what a skill does
type EffectType string

const (
	EffectDamage EffectType = "damage"
	EffectHeal   EffectType = "heal"
	EffectBuff   EffectType = "buff"
	EffectDebuff EffectType = "debuff"
	EffectDot    EffectType = "dot" // Damage over time
	EffectHot    EffectType = "hot" // Heal over time
)

// Skill defines a combat skill or spell
type Skill struct {
	*entities.Entity `bson:",inline"`

	Name          string   `bson:"name" json:"name"`
	Description   string   `bson:"description" json:"description"`
	ClassIDs      []string `bson:"classIds" json:"classIds"`
	LevelRequired int32    `bson:"levelRequired" json:"levelRequired"`

	// Resource cost
	ResourceType   ResourceType `bson:"resourceType" json:"resourceType"`
	ManaCost       int32        `bson:"manaCost" json:"manaCost"`
	CooldownRounds int          `bson:"cooldownRounds" json:"cooldownRounds"`

	// Targeting
	Target TargetType `bson:"target" json:"target"`

	// Effect
	Effect        EffectType `bson:"effect" json:"effect"`
	ScalingAttr   string     `bson:"scalingAttr" json:"scalingAttr"`
	BasePower     int32      `bson:"basePower" json:"basePower"`
	ScalingFactor float64    `bson:"scalingFactor" json:"scalingFactor"`

	// Duration for buffs/debuffs/dots/hots (0 = instant)
	Duration int `bson:"duration" json:"duration"`

	// Buff/Debuff specifics
	BuffStat    string  `bson:"buffStat" json:"buffStat"`
	BuffPercent float64 `bson:"buffPercent" json:"buffPercent"`

	// Special flags
	IgnoresDefense bool `bson:"ignoresDefense" json:"ignoresDefense"`
	HitCount       int  `bson:"hitCount" json:"hitCount"`

	// Secondary effect (e.g. Holy Strike: damage + small heal)
	SecondaryEffect    EffectType `bson:"secondaryEffect,omitempty" json:"secondaryEffect,omitempty"`
	SecondaryBasePower int32      `bson:"secondaryBasePower,omitempty" json:"secondaryBasePower,omitempty"`
	SecondaryScaling   float64    `bson:"secondaryScaling,omitempty" json:"secondaryScaling,omitempty"`
	SecondaryTarget    TargetType `bson:"secondaryTarget,omitempty" json:"secondaryTarget,omitempty"`
}

// HasClass returns true if this skill is available to the given class ID
func (s *Skill) HasClass(classID string) bool {
	for _, c := range s.ClassIDs {
		if c == classID {
			return true
		}
	}
	return false
}
