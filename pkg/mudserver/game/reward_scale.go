package game

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// rawEnemyReward is one dead enemy's XP and gold before the level-gap multiplier.
type rawEnemyReward struct {
	name       string
	level      int32
	boss       bool
	bossKey    string
	baseXP     int64
	baseGold   int64
	scaledXP   int64
	scaledGold int64
	tier       string
}

// bossIdentity is the persisted first-kill key. Template id wins over name so
// every copy of a boss shares one flag. World-specific names are data, not code.
func bossIdentity(n *npc.NPC) string {
	if n == nil {
		return ""
	}
	if n.TemplateID != "" {
		return "tpl:" + n.TemplateID
	}
	name := strings.ToLower(strings.TrimSpace(n.Name))
	if name != "" {
		return "name:" + name
	}
	if n.Entity != nil && n.Entity.ID != "" {
		return "id:" + n.Entity.ID
	}
	return ""
}

func isBossDifficulty(difficulty string) bool {
	return strings.EqualFold(strings.TrimSpace(difficulty), "boss")
}

func hasBossKill(kills []string, key string) bool {
	if key == "" {
		return true
	}
	for _, k := range kills {
		if k == key {
			return true
		}
	}
	return false
}

// victoryReferenceLevel is the highest level among characters who receive
// this victory's split (living fighters plus same-room online party).
// Levels below 1 count as 1. That single number scales every enemy in the fight.
func (c *CombatController) victoryReferenceLevel(instance *combat.CombatInstance, living []*combat.CombatantRef) int32 {
	max := int32(1)
	if c == nil {
		return max
	}
	shares, _ := c.planVictoryShares(instance, living, 0, 0)
	if c.game == nil || c.game.Facade == nil {
		return max
	}
	for _, share := range shares {
		char, err := c.game.Facade.CharactersService().FindByID(share.ID)
		if err != nil || char == nil {
			continue
		}
		lvl := char.Level
		if lvl < 1 {
			lvl = 1
		}
		if lvl > max {
			max = lvl
		}
	}
	return max
}

func formatRewardLines(b messages.RewardBreakdown) string {
	return fmt.Sprintf(
		"  Base: %d XP, %d Gold\n  Level modifier (highest in the split, L%d): %d XP, %d Gold\n  First-kill bonus: %d XP, %d Gold\n  Party split: %d recipients, your share %d XP, %d Gold\n",
		b.BaseXP, b.BaseGold, b.ReferenceLevel, b.LevelModXP, b.LevelModGold,
		b.FirstKillXP, b.FirstKillGold, b.PartySize, b.ShareXP, b.ShareGold,
	)
}

// applyRewardScale multiplies each enemy's base XP and gold by the threat tier
// against referenceLevel (highest level among split recipients).
func applyRewardScale(raw []rawEnemyReward, referenceLevel int32) (baseXP, baseGold, scaledXP, scaledGold int64) {
	for i := range raw {
		tier := balance.ThreatTier(referenceLevel, raw[i].level)
		mult := balance.RewardMultiplier(tier)
		raw[i].tier = tier
		raw[i].scaledXP = balance.ScaleReward(raw[i].baseXP, mult)
		raw[i].scaledGold = balance.ScaleReward(raw[i].baseGold, mult)
		baseXP += raw[i].baseXP
		baseGold += raw[i].baseGold
		scaledXP += raw[i].scaledXP
		scaledGold += raw[i].scaledGold
	}
	return baseXP, baseGold, scaledXP, scaledGold
}

func firstKillBonus(kills []string, enemies []rawEnemyReward, ids []string, charID, killerID string, keepRemainder bool) (xp, gold int64, marked []string) {
	rate := balance.FirstKillBonusRate()
	seen := map[string]bool{}
	for _, k := range kills {
		seen[k] = true
	}
	for _, en := range enemies {
		if !en.boss || en.bossKey == "" || seen[en.bossKey] {
			continue
		}
		seen[en.bossKey] = true
		marked = append(marked, en.bossKey)
		partXP := splitVictoryAmount(en.scaledXP, ids, killerID, keepRemainder)[charID]
		partGold := splitVictoryAmount(en.scaledGold, ids, killerID, keepRemainder)[charID]
		xp += balance.BonusReward(partXP, rate)
		gold += balance.BonusReward(partGold, rate)
	}
	return xp, gold, marked
}
