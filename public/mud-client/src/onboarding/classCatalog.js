/**
 * Class catalog. The sample trio is the only built-in set.
 * A world pack replaces it from GET /api/classes.
 */
import { writable, get } from "svelte/store";
import { fetchClasses } from "../api/classes.js";

function skill(id, name, classId, level, kit, keeps, once, cd, mult, description, target) {
  return {
    id,
    name,
    description,
    classIds: [classId],
    levelRequired: level,
    kit,
    keepsSwing: keeps,
    oncePerFight: once,
    cooldownRounds: cd,
    swingMult: mult,
    target: target || "enemy",
    resourceType: "cooldown",
    manaCost: 0,
    effect: "damage",
  };
}

const SAMPLE_CLASSES = [
  {
    id: "warrior",
    name: "Warrior",
    description: "Sample warrior. Plate, one swing, and a brace that halves the next hit.",
    skillClass: "warrior",
    aliases: ["warrior", "tpl-warrior"],
    races: ["human", "dwarf"],
    templateId: "tpl-warrior",
    portraitClass: "warrior",
    armorType: "Plate",
    combatType: "Melee",
    hotbarCap: 4,
    caster: false,
    skills: [
      skill("warrior_brace", "Brace", "warrior", 1, "brace", true, true, 0, 0,
        "Once a fight. The next hit on you is halved. You still swing."),
      skill("warrior_slam", "Slam", "warrior", 4, "slam", false, false, 4, 1.4,
        "A heavy swing, 1.40×. Replaces your swing this round."),
      skill("warrior_stand", "Stand", "warrior", 8, "stand", false, true, 0, 0,
        "Once a fight, for two rounds hits aimed at others hit you."),
    ],
  },
  {
    id: "rogue",
    name: "Rogue",
    description: "Sample rogue. Leather, two lighter swings, and one slip out of a fight.",
    skillClass: "rogue",
    aliases: ["rogue", "tpl-rogue"],
    races: ["human", "dwarf", "elf"],
    templateId: "tpl-rogue",
    portraitClass: "rogue",
    armorType: "Leather",
    combatType: "Melee",
    hotbarCap: 4,
    caster: false,
    skills: [
      skill("rogue_slip", "Slip", "rogue", 1, "slip", false, true, 0, 0,
        "Once a fight. Drop the fight and take one exit."),
      skill("rogue_nick", "Nick", "rogue", 4, "nick", false, false, 3, 0.55,
        "Extra 0.55× swing on top of your cadence. Replaces the autoattack this round."),
      skill("rogue_smoke", "Smoke", "rogue", 8, "smoke", false, true, 0, 0,
        "Once a fight. The target misses their next swing."),
    ],
  },
  {
    id: "wizard",
    name: "Mage",
    description: "Sample mage. Cloth, one heavy swing, then a mark that burns.",
    skillClass: "mage",
    aliases: ["wizard", "mage", "tpl-mage"],
    races: ["human", "elf"],
    templateId: "tpl-mage",
    portraitClass: "mage",
    armorType: "Cloth",
    combatType: "Magic",
    hotbarCap: 4,
    caster: true,
    skills: [
      skill("mage_inscribe", "Inscribe", "mage", 1, "inscribe", false, false, 4, 0,
        "Mark them. 4 a round for 3 rounds. Refresh, no stack. Costs no mana."),
      skill("mage_sear", "Sear", "mage", 4, "sear", false, false, 4, 1.8,
        "A cast at 1.80×. No burn. Replaces your swing this round."),
      skill("mage_glyph", "Glyph", "mage", 8, "glyph", false, true, 0, 0,
        "Once a fight. The next hit on you is reduced by 4."),
    ],
  },
];

const SAMPLE_RACES = [
  { id: "human", name: "Human", blurb: "No weapon bonus. A few extra coins when you start." },
  { id: "dwarf", name: "Dwarf", blurb: "Blunt weapons deal 10% more damage." },
  { id: "elf", name: "Elf", blurb: "Bows deal 10% more damage." },
  { id: "construct", name: "Construct", blurb: "Poison never sticks. No weapon bonus." },
];

function samplePayload() {
  return { source: "sample", classes: SAMPLE_CLASSES, races: SAMPLE_RACES };
}

export const classCatalog = writable(samplePayload());

let inflight = null;

export function installClassCatalog(payload) {
  if (!payload || !Array.isArray(payload.classes) || payload.classes.length === 0) return;
  classCatalog.set({
    source: payload.source || "pack",
    classes: payload.classes,
    races: Array.isArray(payload.races) && payload.races.length ? payload.races : SAMPLE_RACES,
  });
}

export function resetClassCatalog() {
  classCatalog.set(samplePayload());
}

export function ensureClassCatalog() {
  if (get(classCatalog).source === "pack") return Promise.resolve(get(classCatalog));
  if (inflight) return inflight;
  inflight = fetchClasses()
    .then((data) => {
      installClassCatalog(data);
      return get(classCatalog);
    })
    .catch(() => get(classCatalog))
    .finally(() => {
      inflight = null;
    });
  return inflight;
}

function norm(value) {
  return String(value || "").trim().toLowerCase();
}

export function lookupClass(id) {
  const key = norm(id);
  if (!key) return null;
  const classes = get(classCatalog).classes || [];
  for (const cls of classes) {
    if (!cls) continue;
    const keys = [cls.id, cls.skillClass, cls.name, cls.templateId, ...(cls.aliases || [])];
    if (keys.some((k) => norm(k) === key)) return cls;
  }
  return null;
}

export function catalogRaces() {
  const races = get(classCatalog).races;
  return Array.isArray(races) && races.length ? races : SAMPLE_RACES;
}

export function catalogTemplates() {
  return (get(classCatalog).classes || [])
    .filter((cls) => cls && cls.templateId)
    .map((cls) => ({
      id: cls.templateId,
      name: cls.name,
      description: cls.description,
      class: { id: cls.id, name: cls.name },
    }));
}
