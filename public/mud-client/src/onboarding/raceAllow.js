/** Create allow-list and portraits. Names and blurbs come from the class catalog. */

import { get } from "svelte/store";
import { catalogRaces, catalogTemplates, classCatalog, lookupClass } from "./classCatalog.js";

export const FALLBACK_TEMPLATES = [
  { id: "tpl-warrior", name: "Warrior", description: "Sample warrior. Plate, one swing, and a brace that halves the next hit.", class: { id: "warrior", name: "Warrior" } },
  { id: "tpl-rogue", name: "Rogue", description: "Sample rogue. Leather, two lighter swings, and one slip out of a fight.", class: { id: "rogue", name: "Rogue" } },
  { id: "tpl-mage", name: "Mage", description: "Sample mage. Cloth, one heavy swing, then a mark that burns.", class: { id: "wizard", name: "Mage" } },
];

const STATIC_PORTRAIT = {
  ranger: "ranger",
  hunter: "ranger",
  cleric: "cleric",
  druid: "druid",
};

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

function raceById(id) {
  const key = canonicalRace(id);
  return catalogRaces().find((race) => race.id === key) || null;
}

export function classBlurb(id) {
  const hit = lookupClass(id);
  return hit ? hit.description || "" : "";
}

/** Races the selected class card may offer. Empty when the class is unknown. */
export function racesForTemplate(template) {
  let ids = null;
  for (const key of templateKeys(template)) {
    const hit = lookupClass(key);
    if (hit && Array.isArray(hit.races) && hit.races.length) {
      ids = hit.races;
      break;
    }
  }
  if (!ids) return [];
  return ids.map((id) => raceById(id)).filter(Boolean);
}

export function raceAllowed(template, raceId) {
  const id = canonicalRace(raceId);
  return racesForTemplate(template).some((race) => race.id === id);
}

/** Painted player portrait for a create card. Empty when that race/class has no file. */
export function originPortraitSrc(template, raceId) {
  let cls = "";
  for (const key of templateKeys(template)) {
    const hit = lookupClass(key);
    if (hit && hit.portraitClass) {
      cls = hit.portraitClass;
      break;
    }
  }
  if (!cls) {
    for (const key of templateKeys(template)) {
      if (STATIC_PORTRAIT[key]) {
        cls = STATIC_PORTRAIT[key];
        break;
      }
    }
  }
  if (!cls) return "";
  let race = canonicalRace(raceId);
  if (!["human", "dwarf", "elf"].includes(race)) {
    const races = racesForTemplate(template);
    race = (races[0] && races[0].id) || "";
  }
  if (!["human", "dwarf", "elf"].includes(race)) return "";
  if (cls === "ward" && race === "elf") return "";
  return `/api/portraits/player-${race}-${cls}.png`;
}

/**
 * Hosts that show the guest class and race picker. Other hosts stay a random button.
 * A deployment sets window.TALESMUD_GUEST_PICKER_HOSTS (array or comma-separated
 * string) before the bundle loads. Empty by default.
 */
export function guestPickerHosts() {
  const raw = typeof globalThis !== "undefined" ? globalThis.TALESMUD_GUEST_PICKER_HOSTS : undefined;
  const list = Array.isArray(raw) ? raw : String(raw || "").split(",");
  return list.map((h) => String(h || "").trim().toLowerCase()).filter(Boolean);
}

export function guestPickerEnabled(hostname) {
  const host = String(hostname || "").trim().toLowerCase();
  return host !== "" && guestPickerHosts().includes(host);
}

/**
 * Body for POST /guest.
 * Other hosts always send {} so class and race are never sent.
 * Both empty on a picker host is random ({}).
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

/** Create cards from the catalog currently loaded. */
export function catalogFallbackTemplates() {
  const cards = catalogTemplates();
  if (get(classCatalog).source === "pack" && cards.length) return cards;
  return FALLBACK_TEMPLATES;
}
