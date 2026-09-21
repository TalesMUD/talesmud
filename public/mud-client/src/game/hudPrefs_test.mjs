import assert from 'assert';
import {
  ACTION_BAR_CHROME,
  ACTION_BAR_LAYOUT_REVISION,
  DEFAULT_ACTION_BAR_PINS,
  DEFAULT_HOTBAR_BINDS,
  DEFAULT_INVENTORY_OPEN_MODE,
  DEFAULT_REST_SLOT,
  HOTBAR_SLOT_COUNT,
  INVENTORY_OPEN_OVERLAY,
  INVENTORY_OPEN_WIDGET,
  PINNABLE_COMMANDS,
  commandForPin,
  makeItemBind,
  makeSkillBind,
  migrateActionBarPins,
  normalizeActionBarPins,
  normalizeHotbarBinds,
  normalizeInventoryOpenMode,
  resolveActionBarChrome,
  resolveHotbarActivation,
  resolvePinnedCommands,
  makeActionBind,
  HOTBAR_ACTIONS,
  scrubLegacySearchBinds,
  seedRestOnEmptyHotbar,
  skillDisplayName,
  skillGenericArtUrl,
  togglePin,
  SKILL_CATALOG,
  bindSkillToFirstEmptyHotbar,
  classifySkills,
  firstEmptyHotbarIndex,
  formatSkillCost,
  formatSkillEffects,
  isSkillOnHotbar,
  maxSkillSlots,
  normalizeClassId,
  skillById,
  skillsForClass,
} from './hudPrefs.js';

assert.deepStrictEqual(DEFAULT_ACTION_BAR_PINS, ['recipes'], 'Recipes seeded for crafting discoverability');
assert.ok(ACTION_BAR_LAYOUT_REVISION >= 3, 'layout revision bumped for Recipes seed');

assert.deepStrictEqual(
  ACTION_BAR_CHROME.map((c) => c.id),
  ['inv', 'map', 'say'],
  'INV/MAP/SAY are fixed chrome'
);
assert.ok(
  !PINNABLE_COMMANDS.some((c) => ['inv', 'map', 'say'].includes(c.id)),
  'chrome ids are not pinnable'
);
assert.ok(
  !DEFAULT_ACTION_BAR_PINS.includes('look'),
  'Look is default OFF the action bar'
);
assert.ok(
  DEFAULT_ACTION_BAR_PINS.includes('recipes'),
  'Recipes is default ON the action bar'
);
assert.ok(
  PINNABLE_COMMANDS.some((c) => c.id === 'look'),
  'Look remains pinnable via ⋯'
);
assert.ok(
  PINNABLE_COMMANDS.some((c) => c.id === 'recipes'),
  'Recipes remains pinnable via ⋯'
);

assert.deepStrictEqual(
  normalizeActionBarPins(null),
  ['recipes'],
  'null pins → Recipes default'
);
assert.deepStrictEqual(
  normalizeActionBarPins([]),
  [],
  'empty pins stay empty (explicit clear)'
);
assert.deepStrictEqual(
  normalizeActionBarPins(['look', 'inv', 'map', 'look', 'nope', 'say']),
  ['look'],
  'chrome ids stripped; look kept if explicitly stored'
);
assert.deepStrictEqual(
  normalizeActionBarPins(['WHO', ' Help ']),
  ['who', 'help'],
  'normalize case/whitespace'
);

assert.deepStrictEqual(
  migrateActionBarPins(['look', 'inv', 'map'], 1),
  ['recipes'],
  'legacy look+inv+map defaults migrate to Recipes seed'
);
assert.deepStrictEqual(
  migrateActionBarPins(['look', 'inv', 'map', 'rest', 'help', 'say'], 1),
  ['recipes'],
  'cluttered legacy pins migrate to Recipes seed on revision bump'
);
assert.deepStrictEqual(
  migrateActionBarPins(['look', 'who'], 2),
  ['recipes', 'look', 'who'],
  'revision 2 → 3 seeds Recipes without wiping optional pins'
);
assert.deepStrictEqual(
  migrateActionBarPins(['recipes', 'look'], 2),
  ['recipes', 'look'],
  'revision 2 → 3 does not duplicate Recipes'
);
assert.deepStrictEqual(
  migrateActionBarPins(['look', 'who'], 3),
  ['look', 'who'],
  'revision 3+ preserves optional pins as-is'
);

const toggledOn = togglePin([], 'who');
assert.deepStrictEqual(toggledOn, ['who'], 'pin who onto empty bar');
const toggledOff = togglePin(['who', 'help'], 'help');
assert.deepStrictEqual(toggledOff, ['who'], 'unpin help (empty allowed)');
assert.deepStrictEqual(togglePin(['who'], 'inv'), ['who'], 'cannot pin chrome inv');
assert.deepStrictEqual(togglePin([], 'look'), ['look'], 'Look remains pinnable via ⋯');

const chrome = resolveActionBarChrome();
assert.strictEqual(chrome.length, 3);
assert.strictEqual(chrome[0].kind, 'inventory');
assert.strictEqual(chrome[1].kind, 'map');
assert.strictEqual(chrome[2].kind, 'say');
assert.strictEqual(commandForPin(chrome[2]), null, 'Say chrome must not emit bare say');

const resolved = resolvePinnedCommands(['look', 'help']);
assert.strictEqual(resolved.length, 2);
assert.strictEqual(resolved[0].id, 'look');

assert.strictEqual(normalizeInventoryOpenMode(undefined), DEFAULT_INVENTORY_OPEN_MODE);
assert.strictEqual(normalizeInventoryOpenMode('overlay'), INVENTORY_OPEN_OVERLAY);
assert.strictEqual(normalizeInventoryOpenMode('widget'), INVENTORY_OPEN_WIDGET);
assert.strictEqual(normalizeInventoryOpenMode('bogus'), INVENTORY_OPEN_OVERLAY);

// --- Hotbar binds ---
assert.strictEqual(DEFAULT_HOTBAR_BINDS.length, HOTBAR_SLOT_COUNT);
assert.strictEqual(DEFAULT_REST_SLOT, 6, 'Rest seeds into slot 7');
assert.deepStrictEqual(DEFAULT_HOTBAR_BINDS[DEFAULT_REST_SLOT], {
  kind: 'action',
  id: 'rest',
  name: 'Rest',
  command: 'rest',
});
assert.ok(
  DEFAULT_HOTBAR_BINDS.every((b, i) => i === DEFAULT_REST_SLOT || b === null),
  'default hotbar only seeds Rest'
);

const normalized = normalizeHotbarBinds([
  { kind: 'skill', id: 'mage_fireball' },
  { kind: 'item', name: 'Health Potion', id: 'ITM0099' },
  { kind: 'nope' },
  'junk',
  null,
  { kind: 'item' },
]);
assert.strictEqual(normalized.length, HOTBAR_SLOT_COUNT);
assert.deepStrictEqual(normalized[0], {
  kind: 'skill',
  id: 'mage_fireball',
  name: 'Fireball',
});
assert.deepStrictEqual(normalized[1], {
  kind: 'item',
  name: 'Health Potion',
  id: 'ITM0099',
});
assert.strictEqual(normalized[2], null, 'junk kind → null');
assert.strictEqual(normalized[3], null, 'string junk → null');
assert.strictEqual(normalized[5], null, 'item without name/id → null');
assert.ok(normalized.slice(6).every((b) => b === null), 'pad to 8 slots');

assert.strictEqual(skillDisplayName('mage_fireball'), 'Fireball');
assert.strictEqual(skillDisplayName('warrior_power_strike'), 'Power Strike');

const skillBind = makeSkillBind('mage_fireball');
const outOfCombat = resolveHotbarActivation(skillBind, { inCombat: false, inventory: [] });
assert.strictEqual(outOfCombat.ok, false);
assert.strictEqual(outOfCombat.gated, true);
assert.match(outOfCombat.reason, /only use skills in combat/i);
assert.strictEqual(outOfCombat.command, null);

const inCombat = resolveHotbarActivation(skillBind, { inCombat: true, inventory: [] });
assert.strictEqual(inCombat.ok, true);
assert.strictEqual(inCombat.command, 'cast Fireball');

const potion = { name: 'Health Potion', type: 'consumable', templateId: 'ITM0099' };
const itemBind = makeItemBind(potion);
const itemOoC = resolveHotbarActivation(itemBind, {
  inCombat: false,
  inventory: [potion],
});
assert.strictEqual(itemOoC.ok, true, 'consumables usable out of combat');
assert.strictEqual(itemOoC.command, 'use Health Potion');

const missing = resolveHotbarActivation(itemBind, { inCombat: true, inventory: [] });
assert.strictEqual(missing.ok, false);
assert.strictEqual(missing.missing, true);
assert.strictEqual(missing.command, null);

console.log('hudPrefs: pins + chrome + hotbar binds/combat gate OK');

assert.strictEqual(skillGenericArtUrl('mage_fireball'), '/api/item-art/generic-spell-fire.png');
assert.strictEqual(skillGenericArtUrl('Fireball'), '/api/item-art/generic-spell-fire.png');
assert.strictEqual(skillGenericArtUrl('cleric_heal'), '/api/item-art/generic-spell-heal.png');
assert.strictEqual(skillGenericArtUrl('ranger_aimed_shot'), '/api/item-art/generic-action-ranged.png');

const lookBind = makeActionBind('look');
assert.deepStrictEqual(lookBind, { kind: 'action', id: 'look', name: 'Look', command: 'look' });
const lookAct = resolveHotbarActivation(lookBind, { inCombat: false, inventory: [] });
assert.strictEqual(lookAct.ok, true);
assert.strictEqual(lookAct.command, 'look');

assert.ok(
  !HOTBAR_ACTIONS.some((a) => a.id === 'search'),
  'Search hotbar action that only sent look is removed'
);
assert.strictEqual(makeActionBind('search'), null, 'Search cannot be bound');
const scrubbed = scrubLegacySearchBinds([
  { kind: 'action', id: 'search', command: 'look' },
  { kind: 'action', id: 'look', command: 'look' },
]);
assert.strictEqual(scrubbed[0], null, 'legacy Search=look bind scrubbed');
assert.strictEqual(scrubbed[1]?.id, 'look', 'Look bind preserved');

assert.ok(HOTBAR_ACTIONS.some((a) => a.id === 'look'), 'Look remains bindable');
assert.ok(HOTBAR_ACTIONS.some((a) => a.id === 'rest'), 'Rest remains bindable');
assert.ok(HOTBAR_ACTIONS.some((a) => a.id === 'talk'), 'Talk remains bindable');
assert.ok(HOTBAR_ACTIONS.some((a) => a.id === 'flee'), 'Flee remains bindable');

const seeded = seedRestOnEmptyHotbar([null, null, null, null, null, null, null, null]);
assert.strictEqual(seeded[DEFAULT_REST_SLOT]?.id, 'rest', 'empty bar seeds Rest');
const custom = seedRestOnEmptyHotbar([
  { kind: 'skill', id: 'mage_fireball' },
  null, null, null, null, null, null, null,
]);
assert.strictEqual(custom[0]?.kind, 'skill', 'custom bar kept');
assert.strictEqual(custom[DEFAULT_REST_SLOT], null, 'custom bar is not injected with Rest');
const already = seedRestOnEmptyHotbar(DEFAULT_HOTBAR_BINDS);
assert.strictEqual(already[DEFAULT_REST_SLOT]?.id, 'rest');
console.log('hudPrefs: Option C (room + chrome INV/MAP/SAY, Rest seeded on empty bar) OK');

// --- Skill catalog / slots (Character → Skills) ---
assert.ok(SKILL_CATALOG.length >= 24, 'catalog covers seeded class skills');
assert.strictEqual(normalizeClassId('wizard'), 'mage');
assert.strictEqual(normalizeClassId('Warrior'), 'warrior');

const warriorSkills = skillsForClass('warrior');
assert.strictEqual(warriorSkills.length, 5, 'warrior has 5 skills');
assert.ok(warriorSkills.some((s) => s.id === 'warrior_power_strike'));
assert.ok(warriorSkills.some((s) => s.id === 'warrior_cleave'));
assert.ok(warriorSkills.some((s) => s.id === 'warrior_berserker_rage'));
assert.strictEqual(skillsForClass('wizard').length, 5, 'wizard aliases to mage skills');

assert.strictEqual(maxSkillSlots('warrior', 1), 1);
assert.strictEqual(maxSkillSlots('warrior', 13), 2, 'warrior L13 → 2 slots');
assert.strictEqual(maxSkillSlots('warrior', 20), 3);
assert.strictEqual(maxSkillSlots('mage', 1), 2);
assert.strictEqual(maxSkillSlots('wizard', 15), 3);

const power = skillById('warrior_power_strike');
assert.strictEqual(power.name, 'Power Strike');
assert.strictEqual(formatSkillCost(power), '3 round CD');
assert.ok(formatSkillEffects(power).some((c) => /150% STR dmg/.test(c)));

const bash = skillById('Shield Bash');
assert.strictEqual(formatSkillCost(bash), '4 round CD');
assert.ok(formatSkillEffects(bash).some((c) => /stun/.test(c)));

const cry = skillById('warrior_battle_cry');
assert.ok(formatSkillEffects(cry).some((c) => /\+30% attack/.test(c)));

const fireball = skillById('mage_fireball');
assert.strictEqual(formatSkillCost(fireball), '8 mana');

const classified = classifySkills('warrior', 13, ['warrior_power_strike']);
assert.strictEqual(classified.equipped.length, 1);
assert.ok(classified.available.some((s) => s.name === 'Shield Bash'));
assert.ok(classified.available.some((s) => s.name === 'Battle Cry'));
assert.ok(!classified.available.some((s) => s.id === 'warrior_power_strike'));
assert.ok(classified.locked.some((s) => s.name === 'Cleave'));
assert.ok(classified.locked.some((s) => s.name === 'Berserker Rage'));
assert.ok(classified.locked.every((s) => s.levelRequired > 13));

const emptyIdx = firstEmptyHotbarIndex(DEFAULT_HOTBAR_BINDS);
assert.strictEqual(emptyIdx, 0, 'default bar first empty is slot 1');
const boundOnce = bindSkillToFirstEmptyHotbar(DEFAULT_HOTBAR_BINDS, 'warrior_power_strike');
assert.strictEqual(boundOnce.status, 'bound');
assert.strictEqual(boundOnce.index, 0);
assert.strictEqual(boundOnce.binds[0].id, 'warrior_power_strike');
assert.ok(isSkillOnHotbar(boundOnce.binds, 'warrior_power_strike'));
const boundAgain = bindSkillToFirstEmptyHotbar(boundOnce.binds, 'warrior_power_strike');
assert.strictEqual(boundAgain.status, 'already');
const boundSecond = bindSkillToFirstEmptyHotbar(boundOnce.binds, 'warrior_shield_bash');
assert.strictEqual(boundSecond.status, 'bound');
assert.strictEqual(boundSecond.index, 1);
console.log('hudPrefs: skill catalog + slot helpers OK');
