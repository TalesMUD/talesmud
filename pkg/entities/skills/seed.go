package skills

import "github.com/talesmud/talesmud/pkg/entities"

// SeedSkills returns the default skill definitions used to populate the database on first run.
func SeedSkills() []*Skill {
	return []*Skill{
		// ============================================================
		// WARRIOR — Cooldown-based, STR scaling
		// ============================================================
		{
			Entity: &entities.Entity{ID: "warrior_power_strike"}, Name: "Power Strike", ClassIDs: []string{"warrior"}, LevelRequired: 1,
			Description:    "A powerful strike dealing 150% weapon damage.",
			ResourceType:   ResourceCooldown, CooldownRounds: 3,
			Target:         TargetEnemy,
			Effect:         EffectDamage,
			ScalingAttr:    "STR", BasePower: 3, ScalingFactor: 1.5,
			IgnoresDefense: false,
		},
		{
			Entity: &entities.Entity{ID: "warrior_shield_bash"}, Name: "Shield Bash", ClassIDs: []string{"warrior"}, LevelRequired: 5,
			Description:    "Bash the target with your shield, dealing damage and stunning for 1 round.",
			ResourceType:   ResourceCooldown, CooldownRounds: 4,
			Target:         TargetEnemy,
			Effect:         EffectDamage,
			ScalingAttr:    "STR", BasePower: 2, ScalingFactor: 1.0,
			Duration:       1,
			BuffStat:       "stun",
			IgnoresDefense: false,
		},
		{
			Entity: &entities.Entity{ID: "warrior_battle_cry"}, Name: "Battle Cry", ClassIDs: []string{"warrior"}, LevelRequired: 10,
			Description:  "Let out a battle cry, increasing attack power by 30% for 3 rounds.",
			ResourceType: ResourceCooldown, CooldownRounds: 5,
			Target:       TargetSelf,
			Effect:       EffectBuff,
			BuffStat:     "attack", BuffPercent: 0.30,
			Duration:     3,
		},
		{
			Entity: &entities.Entity{ID: "warrior_cleave"}, Name: "Cleave", ClassIDs: []string{"warrior"}, LevelRequired: 15,
			Description:  "Swing your weapon in a wide arc, hitting all enemies for 80% damage.",
			ResourceType: ResourceCooldown, CooldownRounds: 4,
			Target:       TargetAllEnemies,
			Effect:       EffectDamage,
			ScalingAttr:  "STR", BasePower: 2, ScalingFactor: 0.8,
		},
		{
			Entity: &entities.Entity{ID: "warrior_berserker_rage"}, Name: "Berserker Rage", ClassIDs: []string{"warrior"}, LevelRequired: 20,
			Description:  "Enter a berserker rage: +50% attack but -25% defense for 3 rounds.",
			ResourceType: ResourceCooldown, CooldownRounds: 6,
			Target:       TargetSelf,
			Effect:       EffectBuff,
			BuffStat:     "attack", BuffPercent: 0.50,
			Duration:     3,
		},

		// ============================================================
		// ROGUE — Cooldown-based, DEX scaling
		// ============================================================
		{
			Entity: &entities.Entity{ID: "rogue_backstab"}, Name: "Backstab", ClassIDs: []string{"rogue"}, LevelRequired: 1,
			Description:  "Strike from the shadows for 200% DEX-scaled damage.",
			ResourceType: ResourceCooldown, CooldownRounds: 3,
			Target:       TargetEnemy,
			Effect:       EffectDamage,
			ScalingAttr:  "DEX", BasePower: 4, ScalingFactor: 2.0,
		},
		{
			Entity: &entities.Entity{ID: "rogue_poison_strike"}, Name: "Poison Strike", ClassIDs: []string{"rogue"}, LevelRequired: 5,
			Description:  "Coat your blade with poison. Target takes damage each round for 3 rounds.",
			ResourceType: ResourceCooldown, CooldownRounds: 4,
			Target:       TargetEnemy,
			Effect:       EffectDot,
			ScalingAttr:  "DEX", BasePower: 3, ScalingFactor: 0.5,
			Duration:     3,
		},
		{
			Entity: &entities.Entity{ID: "rogue_evasion"}, Name: "Evasion", ClassIDs: []string{"rogue"}, LevelRequired: 10,
			Description:  "Heighten your reflexes, gaining +75% dodge chance for 2 rounds.",
			ResourceType: ResourceCooldown, CooldownRounds: 5,
			Target:       TargetSelf,
			Effect:       EffectBuff,
			BuffStat:     "dodge", BuffPercent: 0.75,
			Duration:     2,
		},
		{
			Entity: &entities.Entity{ID: "rogue_shadow_strike"}, Name: "Shadow Strike", ClassIDs: []string{"rogue"}, LevelRequired: 15,
			Description:    "Strike from the shadows, ignoring the target's armor.",
			ResourceType:   ResourceCooldown, CooldownRounds: 5,
			Target:         TargetEnemy,
			Effect:         EffectDamage,
			ScalingAttr:    "DEX", BasePower: 5, ScalingFactor: 1.5,
			IgnoresDefense: true,
		},
		{
			Entity: &entities.Entity{ID: "rogue_flurry"}, Name: "Flurry", ClassIDs: []string{"rogue"}, LevelRequired: 20,
			Description:  "Unleash a flurry of 3 rapid strikes, each at 60% damage.",
			ResourceType: ResourceCooldown, CooldownRounds: 6,
			Target:       TargetEnemy,
			Effect:       EffectDamage,
			ScalingAttr:  "DEX", BasePower: 2, ScalingFactor: 0.6,
			HitCount:     3,
		},

		// ============================================================
		// MAGE — Mana-based, INT scaling
		// ============================================================
		{
			Entity: &entities.Entity{ID: "mage_fireball"}, Name: "Fireball", ClassIDs: []string{"mage"}, LevelRequired: 1,
			Description:  "Hurl a ball of fire at the target.",
			ResourceType: ResourceMana, ManaCost: 8,
			Target:       TargetEnemy,
			Effect:       EffectDamage,
			ScalingAttr:  "INT", BasePower: 6, ScalingFactor: 1.5,
		},
		{
			Entity: &entities.Entity{ID: "mage_frost_shield"}, Name: "Frost Shield", ClassIDs: []string{"mage"}, LevelRequired: 1,
			Description:  "Surround yourself with a shield of ice, increasing defense by 50% for 2 rounds.",
			ResourceType: ResourceMana, ManaCost: 6,
			Target:       TargetSelf,
			Effect:       EffectBuff,
			BuffStat:     "defense", BuffPercent: 0.50,
			Duration:     2,
		},
		{
			Entity: &entities.Entity{ID: "mage_lightning_bolt"}, Name: "Lightning Bolt", ClassIDs: []string{"mage"}, LevelRequired: 5,
			Description:    "Call down a bolt of lightning that ignores armor.",
			ResourceType:   ResourceMana, ManaCost: 15,
			Target:         TargetEnemy,
			Effect:         EffectDamage,
			ScalingAttr:    "INT", BasePower: 10, ScalingFactor: 2.0,
			IgnoresDefense: true,
		},
		{
			Entity: &entities.Entity{ID: "mage_arcane_burst"}, Name: "Arcane Burst", ClassIDs: []string{"mage"}, LevelRequired: 10,
			Description:  "Release a burst of arcane energy, damaging all enemies.",
			ResourceType: ResourceMana, ManaCost: 20,
			Target:       TargetAllEnemies,
			Effect:       EffectDamage,
			ScalingAttr:  "INT", BasePower: 5, ScalingFactor: 1.2,
		},
		{
			Entity: &entities.Entity{ID: "mage_mana_shield"}, Name: "Mana Shield", ClassIDs: []string{"mage"}, LevelRequired: 15,
			Description:  "Create a shield that absorbs damage by consuming mana.",
			ResourceType: ResourceMana, ManaCost: 12,
			Target:       TargetSelf,
			Effect:       EffectBuff,
			BuffStat:     "mana_shield",
			ScalingAttr:  "INT", BasePower: 20, ScalingFactor: 2.0,
			Duration:     3,
		},

		// ============================================================
		// CLERIC — Mana-based, WIS scaling
		// ============================================================
		{
			Entity: &entities.Entity{ID: "cleric_heal"}, Name: "Heal", ClassIDs: []string{"cleric"}, LevelRequired: 1,
			Description:  "Channel divine energy to heal yourself.",
			ResourceType: ResourceMana, ManaCost: 8,
			Target:       TargetSelf,
			Effect:       EffectHeal,
			ScalingAttr:  "WIS", BasePower: 8, ScalingFactor: 1.5,
		},
		{
			Entity: &entities.Entity{ID: "cleric_holy_strike"}, Name: "Holy Strike", ClassIDs: []string{"cleric"}, LevelRequired: 1,
			Description:     "Strike with holy power, dealing damage and healing yourself slightly.",
			ResourceType:    ResourceMana, ManaCost: 6,
			Target:          TargetEnemy,
			Effect:          EffectDamage,
			ScalingAttr:     "WIS", BasePower: 4, ScalingFactor: 1.0,
			SecondaryEffect: EffectHeal, SecondaryBasePower: 3, SecondaryScaling: 0.5, SecondaryTarget: TargetSelf,
		},
		{
			Entity: &entities.Entity{ID: "cleric_shield_of_faith"}, Name: "Shield of Faith", ClassIDs: []string{"cleric"}, LevelRequired: 5,
			Description:  "Invoke divine protection, increasing defense by 40% for 3 rounds.",
			ResourceType: ResourceMana, ManaCost: 10,
			Target:       TargetSelf,
			Effect:       EffectBuff,
			BuffStat:     "defense", BuffPercent: 0.40,
			Duration:     3,
		},
		{
			Entity: &entities.Entity{ID: "cleric_smite"}, Name: "Smite", ClassIDs: []string{"cleric"}, LevelRequired: 10,
			Description:  "Smite your foe with holy wrath.",
			ResourceType: ResourceMana, ManaCost: 15,
			Target:       TargetEnemy,
			Effect:       EffectDamage,
			ScalingAttr:  "WIS", BasePower: 10, ScalingFactor: 2.0,
		},
		{
			Entity: &entities.Entity{ID: "cleric_divine_light"}, Name: "Divine Light", ClassIDs: []string{"cleric"}, LevelRequired: 15,
			Description:  "Bathe yourself in divine light, restoring a large amount of health.",
			ResourceType: ResourceMana, ManaCost: 25,
			Target:       TargetSelf,
			Effect:       EffectHeal,
			ScalingAttr:  "WIS", BasePower: 20, ScalingFactor: 2.5,
		},

		// ============================================================
		// RANGER — Cooldown-based, DEX scaling
		// ============================================================
		{
			Entity: &entities.Entity{ID: "ranger_aimed_shot"}, Name: "Aimed Shot", ClassIDs: []string{"ranger"}, LevelRequired: 1,
			Description:  "Take careful aim for 150% DEX-scaled damage.",
			ResourceType: ResourceCooldown, CooldownRounds: 3,
			Target:       TargetEnemy,
			Effect:       EffectDamage,
			ScalingAttr:  "DEX", BasePower: 3, ScalingFactor: 1.5,
		},
		{
			Entity: &entities.Entity{ID: "ranger_volley"}, Name: "Volley", ClassIDs: []string{"ranger"}, LevelRequired: 5,
			Description:  "Fire a volley of arrows, hitting all enemies for 70% damage.",
			ResourceType: ResourceCooldown, CooldownRounds: 4,
			Target:       TargetAllEnemies,
			Effect:       EffectDamage,
			ScalingAttr:  "DEX", BasePower: 2, ScalingFactor: 0.7,
		},
		{
			Entity: &entities.Entity{ID: "ranger_natures_gift"}, Name: "Nature's Gift", ClassIDs: []string{"ranger"}, LevelRequired: 10,
			Description:  "Call upon nature to heal 30% of your maximum HP.",
			ResourceType: ResourceCooldown, CooldownRounds: 5,
			Target:       TargetSelf,
			Effect:       EffectHeal,
			BasePower:    0, ScalingFactor: 0.30,
			ScalingAttr:  "maxhp",
		},
		{
			Entity: &entities.Entity{ID: "ranger_pin_down"}, Name: "Pin Down", ClassIDs: []string{"ranger"}, LevelRequired: 15,
			Description:  "Pin the target down, dealing DEX damage and reducing defense by 50% for 2 rounds.",
			ResourceType: ResourceCooldown, CooldownRounds: 5,
			Target:       TargetEnemy,
			Effect:       EffectDamage,
			ScalingAttr:  "DEX", BasePower: 4, ScalingFactor: 1.0,
			SecondaryEffect: EffectDebuff, SecondaryTarget: TargetEnemy,
			Duration:        2,
			BuffStat:        "defense", BuffPercent: -0.50,
		},

		// ============================================================
		// DRUID — Mana-based, INT scaling
		// ============================================================
		{
			Entity: &entities.Entity{ID: "druid_wrath"}, Name: "Wrath", ClassIDs: []string{"druid"}, LevelRequired: 1,
			Description:  "Call down nature's wrath on your target.",
			ResourceType: ResourceMana, ManaCost: 6,
			Target:       TargetEnemy,
			Effect:       EffectDamage,
			ScalingAttr:  "INT", BasePower: 5, ScalingFactor: 1.2,
		},
		{
			Entity: &entities.Entity{ID: "druid_rejuvenation"}, Name: "Rejuvenation", ClassIDs: []string{"druid"}, LevelRequired: 1,
			Description:  "Regenerate health over time for 3 rounds.",
			ResourceType: ResourceMana, ManaCost: 8,
			Target:       TargetSelf,
			Effect:       EffectHot,
			ScalingAttr:  "WIS", BasePower: 4, ScalingFactor: 0.8,
			Duration:     3,
		},
		{
			Entity: &entities.Entity{ID: "druid_entangle"}, Name: "Entangle", ClassIDs: []string{"druid"}, LevelRequired: 5,
			Description:  "Entangle the target with roots, reducing attack by 40% for 2 rounds.",
			ResourceType: ResourceMana, ManaCost: 10,
			Target:       TargetEnemy,
			Effect:       EffectDebuff,
			BuffStat:     "attack", BuffPercent: -0.40,
			Duration:     2,
		},
		{
			Entity: &entities.Entity{ID: "druid_starfire"}, Name: "Starfire", ClassIDs: []string{"druid"}, LevelRequired: 10,
			Description:  "Call down a beam of starlight for heavy damage.",
			ResourceType: ResourceMana, ManaCost: 18,
			Target:       TargetEnemy,
			Effect:       EffectDamage,
			ScalingAttr:  "INT", BasePower: 10, ScalingFactor: 2.0,
		},
		{
			Entity: &entities.Entity{ID: "druid_barkskin"}, Name: "Barkskin", ClassIDs: []string{"druid"}, LevelRequired: 15,
			Description:  "Coat yourself in bark, increasing defense by 60% for 3 rounds.",
			ResourceType: ResourceMana, ManaCost: 12,
			Target:       TargetSelf,
			Effect:       EffectBuff,
			BuffStat:     "defense", BuffPercent: 0.60,
			Duration:     3,
		},
	}
}
