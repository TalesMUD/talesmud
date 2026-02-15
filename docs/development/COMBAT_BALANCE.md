# Combat Balance System

## Overview

TalesMUD uses a **difficulty-based multiplier system** to balance enemy combat stats. Enemy stats in YAML files are treated as **base values**, and multipliers are applied at runtime based on the enemy's difficulty tier.

## How It Works

### Source Files (Unchanged)
```yaml
# import/mvp-rpg-1/data/npcs/ENM0001.yaml
name: Catacomb Rat
maxHitPoints: 8        # Base HP
level: 1
enemyTrait:
  attackPower: 1       # Base Attack
  defense: 0           # Base Defense
  difficulty: trivial  # Difficulty tier
```

### Runtime Multipliers
When the enemy is loaded into the game, the system applies multipliers from `config/combat_balance.yaml`:

```yaml
difficulty_multipliers:
  trivial:
    hp: 3.75      # Catacomb Rat: 8 * 3.75 = 30 HP
    attack: 4.0   # Catacomb Rat: 1 * 4.0 = 4 ATK
    defense: 1.0  # No change
```

**Final in-game stats:** HP=30, ATK=4, DEF=0

## Benefits

✅ **Single point of control** - Adjust all enemies of a difficulty tier at once
✅ **Clean import data** - YAML files stay unmodified
✅ **Easy tuning** - Edit one config file instead of dozens of YAML files
✅ **Systematic balance** - Consistent scaling across all enemies
✅ **Fast iteration** - Test balance changes without re-importing world data

## Tuning Combat Difficulty

### Global Rebalancing
To make all combat easier or harder:

```yaml
# config/combat_balance.yaml
difficulty_multipliers:
  normal:
    hp: 2.0       # Reduce from 2.67 to make enemies die faster
    attack: 1.2   # Reduce from 1.5 to make enemies hit softer
    defense: 1.5  # Reduce from 2.0 to make enemies take more damage
```

### Per-Tier Adjustments
To make early game easier but keep endgame challenging:

```yaml
difficulty_multipliers:
  trivial:
    hp: 2.5      # Reduce from 3.75 - easier starter enemies
    attack: 3.0  # Reduce from 4.0
  boss:
    hp: 2.5      # Increase from 2.0 - tougher bosses
    attack: 2.0  # Increase from 1.5
```

## Current Balance (Level 1 Testing)

### Catacomb Rat (Trivial, L1)
**Target:** 5-6 rounds to defeat

| Class   | Rounds | Win Rate | Assessment |
|---------|--------|----------|------------|
| Warrior | 7.0    | 100%     | ✓ Close to target |
| Rogue   | 6.2    | 100%     | ✓ Perfect |
| Mage    | 5.9    | 95%      | ✓ Perfect |
| Cleric  | 5.3    | 100%     | ✓ Close to target |
| Ranger  | 5.1    | 100%     | ✓ Close to target |
| Druid   | 10.7   | 77%      | ⚠ Longer but reasonable |

### Enemy Stats (With Multipliers)

| Enemy           | Level | HP  | ATK | DEF | Difficulty |
|-----------------|-------|-----|-----|-----|------------|
| Catacomb Rat    | 1     | 30  | 4   | 0   | trivial    |
| Sewer Rat       | 2     | 34  | 5   | 0   | easy       |
| Tunnel Mole     | 2     | 43  | 5   | 2   | easy       |
| Meadow Wolf     | 2     | 48  | 6   | 2   | normal     |
| Wild Boar       | 2     | 66  | 7   | 4   | normal     |
| Burrow Brute    | 4     | 110 | 12  | 5   | boss       |
| Hollow Knight   | 6     | 280 | 18  | 6   | boss       |

## Implementation Details

### Code Flow

1. **Import Time** (`pkg/importer/converters.go`)
   - YAML files are loaded
   - `balance.ApplyMultipliers()` is called
   - Multiplied stats are saved to database

2. **Simulator** (`pkg/mudserver/game/combat/simutil/enemy_factory.go`)
   - Uses same `balance.ApplyMultipliers()` function
   - Ensures simulator matches actual game balance

3. **Config Loading** (`pkg/mudserver/game/balance/difficulty.go`)
   - Loads `config/combat_balance.yaml` on first use
   - Falls back to hardcoded defaults if file missing
   - Cached for performance

### Formula
```go
finalHP = baseHP * difficulty_multipliers[difficulty].hp
finalAttack = baseAttack * difficulty_multipliers[difficulty].attack
finalDefense = baseDefense * difficulty_multipliers[difficulty].defense
```

## Testing Balance Changes

Use the combat simulator to test changes:

```bash
# Test all L1 classes vs all enemies
go run cmd/combat_simulator/main.go -level 1 -n 100

# Test specific matchup
go run cmd/combat_simulator/main.go -class Warrior -enemy "Catacomb Rat" -level 1 -n 100

# Test all classes and levels (comprehensive)
go run cmd/combat_simulator/main.go -n 500
```

## Design Philosophy

**Why use multipliers instead of updating YAML directly?**

1. **Preserve import data** - Keep imported content clean and original
2. **Rapid iteration** - Change one config file, not dozens of YAML files
3. **Version control** - See balance changes clearly in git diffs
4. **A/B testing** - Easily compare different balance configurations
5. **Consistency** - Ensure all enemies of same tier scale uniformly

**When to update YAML instead?**

- Individual enemy is too weak/strong compared to its tier
- Enemy needs unique stats that break the formula
- New content with different base stat ranges

In these cases, adjust the base values in YAML, then re-import the world data.
