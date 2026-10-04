/** Signed class and race blurbs, and the create allow-list. */

export const CLASS_BLURBS = {
  fenwatch: "The door. You stand in it until they don't. Brace once when a blow comes in.",
  alley: "Back street. Two cuts, then you Slip the first one that comes back.",
  runehand: "Vault runes. One heavy swing, then you Inscribe. The mark burns for three rounds.",
  ward: "Heavy plate. You start slow. Hits you take stack Grit, and Slam and the hit you throw back get heavier.",
  rigger: "Constructs. Bolt scrap onto someone in the room; the next hit still lands, and the attacker takes the same amount back. Rig drops a turret that does not chase.",
};

export const RACE_BLURBS = {
  human: { id: "human", name: "Human", blurb: "No weapon bonus. A few extra coins when you start." },
  dwarf: { id: "dwarf", name: "Dwarf", blurb: "Blunt weapons deal 10% more damage." },
  elf: { id: "elf", name: "Elf", blurb: "Bows deal 10% more damage." },
  construct: { id: "construct", name: "Construct", blurb: "Rigger only. Poison never sticks. No weapon bonus." },
};

const ALLOW = {
  "tpl-fenwatch": ["human", "dwarf"],
  warrior: ["human", "dwarf"],
  fenwatch: ["human", "dwarf"],
  "tpl-alley": ["human", "dwarf", "elf"],
  rogue: ["human", "dwarf", "elf"],
  alley: ["human", "dwarf", "elf"],
  ranger: ["human", "dwarf", "elf"],
  hunter: ["human", "dwarf", "elf"],
  "tpl-runehand": ["human", "elf"],
  wizard: ["human", "elf"],
  mage: ["human", "elf"],
  runehand: ["human", "elf"],
  "rune hand": ["human", "elf"],
  "tpl-ward": ["human", "dwarf"],
  ward: ["human", "dwarf"],
  "tpl-hitch": ["human", "dwarf"],
  hitch: ["human", "dwarf"],
  "tpl-rigger": ["construct"],
  rigger: ["construct"],
};

export const FALLBACK_TEMPLATES = [
  { id: "tpl-fenwatch", name: "Fenwatch", description: CLASS_BLURBS.fenwatch, class: { id: "warrior", name: "Fenwatch" } },
  { id: "tpl-alley", name: "Alley", description: CLASS_BLURBS.alley, class: { id: "rogue", name: "Alley" } },
  { id: "tpl-runehand", name: "Rune Hand", description: CLASS_BLURBS.runehand, class: { id: "wizard", name: "Rune Hand" } },
  { id: "tpl-ward", name: "Ward", description: CLASS_BLURBS.ward, class: { id: "ward", name: "Ward" } },
  { id: "tpl-rigger", name: "Rigger", description: CLASS_BLURBS.rigger, class: { id: "rigger", name: "Rigger" } },
];

function canonicalRace(id) {
  const race = String(id || "").trim().toLowerCase();
  if (race === "elve" || race === "elves") return "elf";
  return race;
}

function templateKeys(template) {
  const keys = [];
  if (!template) return keys;
  if (template.id) keys.push(String(template.id).trim().toLowerCase());
  const cls = template.class;
  const classId = typeof cls === "object" ? (cls.id || cls.name || "") : (cls || "");
  if (classId) keys.push(String(classId).trim().toLowerCase());
  if (template.archetype) keys.push(String(template.archetype).trim().toLowerCase());
  if (template.name) keys.push(String(template.name).trim().toLowerCase());
  return keys;
}

/** Races the selected class card may offer. Empty when the class is unknown. */
export function racesForTemplate(template) {
  let ids = null;
  for (const key of templateKeys(template)) {
    if (ALLOW[key]) {
      ids = ALLOW[key];
      break;
    }
  }
  if (!ids) return [];
  return ids.map((id) => RACE_BLURBS[id]).filter(Boolean);
}

export function raceAllowed(template, raceId) {
  const id = canonicalRace(raceId);
  return racesForTemplate(template).some((race) => race.id === id);
}

/** Veilspan may show the guest picker. talesmud.io stays a random button. */
export function guestPickerEnabled(hostname) {
  const host = String(hostname || "").trim().toLowerCase();
  return host === "veilspan.com" || host === "www.veilspan.com";
}

/**
 * Body for POST /guest.
 * Non-veilspan hosts always send {} so class and race are never sent.
 * Both empty on veilspan is random ({}).
 * A legal pair is sent. Anything else is null and must not be posted.
 */
export function guestCreateBody(hostname, templateId, raceId) {
  if (!guestPickerEnabled(hostname)) return {};
  const template = String(templateId || "").trim();
  const race = canonicalRace(raceId);
  if (!template && !race) return {};
  if (!template || !race) return null;
  if (!raceAllowed({ id: template }, race)) return null;
  return { templateId: template, race };
}
