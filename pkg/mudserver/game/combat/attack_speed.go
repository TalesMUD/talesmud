package combat

import "math"

const (
	attackSpeedFloor = 0.25
	attackSpeedCeil  = 3
)

func clampAttackSpeed(speed float64) float64 {
	if speed < attackSpeedFloor {
		return attackSpeedFloor
	}
	if speed > attackSpeedCeil {
		return attackSpeedCeil
	}
	return speed
}

// EnemySwingCount is how many swings one NPC attack action spends.
// speed <= 0 (unset) is one swing. Positive speed is attacks per round,
// clamped to [0.25, 3]. speed >= 1 rounds to 1..3 swings. Below 1 is one
// swing; EnemyHoldsAttack inserts the gap. This does not change the turn beat.
func EnemySwingCount(speed float64) int {
	if speed <= 0 {
		return 1
	}
	speed = clampAttackSpeed(speed)
	if speed < 1 {
		return 1
	}
	n := int(math.Round(speed))
	if n < 1 {
		return 1
	}
	if n > 3 {
		return 3
	}
	return n
}

// EnemyHoldsAttack reports whether this attack action is a hold instead of a swing.
// actionIndex is how many attack actions this enemy has already taken (0 is the first).
// speed <= 0 or speed >= 1 never holds. Below 1, period N is round(1/speed),
// clamped to 2..4, and the enemy holds when actionIndex % N != 0.
func EnemyHoldsAttack(speed float64, actionIndex int) bool {
	if speed <= 0 || speed >= 1 {
		return false
	}
	speed = clampAttackSpeed(speed)
	if speed >= 1 {
		return false
	}
	n := int(math.Round(1 / speed))
	if n < 2 {
		n = 2
	}
	if n > 4 {
		n = 4
	}
	if actionIndex < 0 {
		actionIndex = 0
	}
	return actionIndex%n != 0
}
