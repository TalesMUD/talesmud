/**
 * Central registry of known field values for suggestions and filter options.
 * Used by editor forms (datalist suggestions) and table column filters.
 * Users can always enter custom values beyond these lists.
 */

/** Item attribute keys recognized by game logic (use.go, equipment.go) */
export const knownItemAttributes = [
  "healthRestore",
  "manaRestore",
  "useMessage",
  "damage",
  "defense",
  "armor",
  "strength",
  "agility",
  "intelligence",
];

/** Known character stat attribute names (used in Character Templates) */
export const knownCharacterAttributes = [
  "Strength",
  "Dexterity",
  "Intelligence",
  "Wisdom",
  "Stamina",
  "Charisma",
];

/** Known NPC races — display names matching accessor output */
export const knownRaces = [
  "Human", "Elve", "Dwarf", "Halfling", "Orc", "Goblin", "Undead", "Demon", "Beast",
];

/** Known NPC classes — display names matching accessor output */
export const knownClasses = [
  "Warrior", "Ranger", "Wizard", "Rogue", "Cleric", "Merchant", "Guard", "Commoner",
];

/**
 * Extract unique non-empty string values for a given key from a list of elements.
 * Useful for deriving dynamic filter options (e.g., room areas from loaded rooms).
 * @param {Array} elements - array of objects
 * @param {string} key - dot-path key to extract (e.g., "area", "areaType")
 * @returns {string[]} sorted unique values
 */
export function uniqueValues(elements, key) {
  const vals = new Set();
  for (const el of elements) {
    const val = key.split(".").reduce((acc, part) => acc?.[part], el);
    if (val && typeof val === "string") vals.add(val);
  }
  return [...vals].sort();
}
