/** Test-only pack payload. Production source must not contain these names. */

function skill(id, name, classId, level, kit, keeps, once, cd, mult, description, target) {
  return {
    id, name, description,
    classIds: [classId],
    levelRequired: level,
    kit, keepsSwing: keeps, oncePerFight: once,
    cooldownRounds: cd, swingMult: mult,
    target: target || "enemy",
    resourceType: "cooldown", manaCost: 0, effect: "damage",
  };
}

export function veilspanPayload() {
  return {
    source: "pack",
    races: [
      { id: "human", name: "Human", blurb: "No weapon bonus. A few extra coins when you start." },
      { id: "dwarf", name: "Dwarf", blurb: "Blunt weapons deal 10% more damage." },
      { id: "elf", name: "Elf", blurb: "Bows deal 10% more damage." },
      { id: "construct", name: "Construct", blurb: "Rigger only. Poison never sticks. No weapon bonus." },
    ],
    classes: [
      {
        id: "warrior", name: "Fenwatch", skillClass: "warrior",
        description: "The door. You stand in it until they don't. Brace once when a blow comes in.",
        aliases: ["warrior", "fenwatch", "tpl-fenwatch"],
        races: ["human", "dwarf"], templateId: "tpl-fenwatch", portraitClass: "warrior",
        hotbarCap: 4, caster: false,
        skills: [
          skill("warrior_brace", "Brace", "warrior", 1, "brace", true, true, 0, 0, "Once a fight. The next hit on you is halved. You still swing."),
          skill("warrior_slam", "Slam", "warrior", 4, "slam", false, false, 4, 1.4, "A heavy swing, 1.40×. Replaces your swing this round."),
          skill("warrior_stand", "Stand", "warrior", 8, "stand", false, true, 0, 0, "Once a fight, for two rounds hits aimed at others hit you."),
        ],
      },
      {
        id: "rogue", name: "Alley", skillClass: "rogue",
        description: "Back street. Two cuts, then you Slip the first one that comes back.",
        aliases: ["rogue", "alley", "tpl-alley"],
        races: ["human", "dwarf", "elf"], templateId: "tpl-alley", portraitClass: "rogue",
        hotbarCap: 4, caster: false,
        skills: [
          skill("rogue_slip", "Slip", "rogue", 1, "slip", false, true, 0, 0, "Once a fight. Drop the fight and take one exit."),
          skill("rogue_nick", "Nick", "rogue", 4, "nick", false, false, 3, 0.55, "Extra 0.55× swing on top of your cadence."),
          skill("rogue_smoke", "Smoke", "rogue", 8, "smoke", false, true, 0, 0, "Once a fight. The target misses their next swing."),
        ],
      },
      {
        id: "wizard", name: "Rune Hand", skillClass: "mage",
        description: "Vault runes. One heavy swing, then you Inscribe. The mark burns for three rounds.",
        aliases: ["wizard", "mage", "runehand", "rune_hand", "rune hand", "tpl-runehand"],
        races: ["human", "elf"], templateId: "tpl-runehand", portraitClass: "mage",
        hotbarCap: 4, caster: true,
        skills: [
          skill("mage_inscribe", "Inscribe", "mage", 1, "inscribe", false, false, 4, 0, "Mark them. 4 a round for 3 rounds."),
          skill("mage_sear", "Sear", "mage", 4, "sear", false, false, 4, 1.8, "A cast at 1.80×. No burn. Replaces your swing this round."),
          skill("mage_glyph", "Glyph", "mage", 8, "glyph", false, true, 0, 0, "Once a fight. The next hit on you is reduced by 4."),
        ],
      },
      {
        id: "ward", name: "Ward", skillClass: "ward",
        description: "Heavy plate. You start slow. Hits you take stack Grit, and Slam and the hit you throw back get heavier.",
        aliases: ["ward", "hitch", "tpl-ward", "tpl-hitch"],
        races: ["human", "dwarf"], templateId: "tpl-ward", portraitClass: "ward",
        hotbarCap: 4, caster: false,
        skills: [
          skill("ward_guard", "Guard", "ward", 1, "guard", true, true, 0, 0, "Once a fight. The next hit aimed at an ally hits you.", "ally"),
          skill("ward_slam", "Slam", "ward", 1, "slam", false, false, 4, 1, "Replaces your swing. 1.00×, plus 0.20× for each Grit, up to 2.00×."),
        ],
      },
      {
        id: "rigger", name: "Rigger", skillClass: "rigger",
        description: "Constructs. Bolt scrap onto someone in the room; the next hit still lands, and the attacker takes the same amount back. Rig drops a turret that does not chase.",
        aliases: ["rigger", "tpl-rigger"],
        races: ["construct"], templateId: "tpl-rigger", portraitClass: "",
        hotbarCap: 4, caster: false,
        skills: [
          skill("rigger_bolt", "Bolt", "rigger", 1, "bolt", false, true, 0, 0, "Once a fight. Spend your swing."),
          skill("rigger_rig", "Rig", "rigger", 1, "rig", false, true, 0, 0.5, "Once a fight. Drop a turret. It hits twice at 0.50×, then falls apart."),
          skill("rigger_overload", "Overload", "rigger", 6, "overload", false, true, 0, 0.8, "Once a fight, while the turret is up."),
        ],
      },
    ],
  };
}
