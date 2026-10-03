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
  filterHotbarSkillsForCharacter,
  reconcileHotbarForCharacter,
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
assert.strictEqual(skillDisplayName('mage_inscribe'), 'Inscribe');
assert.strictEqual(skillDisplayName('rigger_overload'), 'Overload');

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
assert.strictEqual(skillGenericArtUrl('mage_inscribe'), '/api/item-art/generic-spell-arcane.png');
assert.strictEqual(skillGenericArtUrl('Inscribe'), '/api/item-art/generic-spell-arcane.png');
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
assert.ok(HOTBAR_ACTIONS.some((a) => a.id === 'defend'), 'Defend remains bindable');
assert.equal(HOTBAR_ACTIONS.find((a) => a.id === 'defend')?.command, 'defend');

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
assert.strictEqual(SKILL_CATALOG.length, 15 + 5 + 4 + 5, 'kit + cleric + ranger + druid');
assert.ok(SKILL_CATALOG.every((s) => s.kit || !['warrior', 'rogue', 'mage', 'hitch', 'rigger'].includes(s.classIds[0])));
assert.strictEqual(normalizeClassId('wizard'), 'mage');
assert.strictEqual(normalizeClassId('runehand'), 'mage');
assert.strictEqual(normalizeClassId('rune_hand'), 'mage');
assert.strictEqual(normalizeClassId('Rune Hand'), 'mage');
assert.strictEqual(normalizeClassId('alley'), 'rogue');
assert.strictEqual(normalizeClassId('fenwatch'), 'warrior');
assert.strictEqual(normalizeClassId('hitch'), 'hitch');
assert.strictEqual(normalizeClassId('rigger'), 'rigger');
assert.strictEqual(normalizeClassId('Warrior'), 'warrior');

const warriorSkills = skillsForClass('warrior');
assert.strictEqual(warriorSkills.length, 3, 'fenwatch kit is Brace/Slam/Stand');
assert.deepStrictEqual(warriorSkills.map((s) => s.id), ['warrior_brace', 'warrior_slam', 'warrior_stand']);
assert.ok(!warriorSkills.some((s) => s.id === 'warrior_power_strike'));
assert.strictEqual(skillsForClass('wizard').length, 3, 'wizard aliases to rune hand kit');
assert.ok(!skillsForClass('wizard').some((s) => s.id === 'mage_fireball' || s.id === 'mage_frost_shield'));
assert.ok(!skillsForClass('rune hand').some((s) => /fireball|frost/i.test(s.id + s.name)));
assert.strictEqual(skillsForClass('alley').length, 3);
assert.ok(!skillsForClass('alley').some((s) => s.id === 'rogue_backstab'));
assert.strictEqual(skillsForClass('hitch').map((s) => s.id).join(','), 'hitch_pin,hitch_hobble,hitch_reel');
assert.strictEqual(skillsForClass('rigger').find((s) => s.id === 'rigger_overload')?.levelRequired, 6);

assert.strictEqual(maxSkillSlots('warrior', 1), 4);
assert.strictEqual(maxSkillSlots('fenwatch', 13), 4);
assert.strictEqual(maxSkillSlots('mage', 1), 4);
assert.strictEqual(maxSkillSlots('wizard', 15), 4);
assert.strictEqual(maxSkillSlots('runehand', 1), 4);
assert.strictEqual(maxSkillSlots('hitch', 1), 4);
assert.strictEqual(maxSkillSlots('rigger', 30), 4);
assert.strictEqual(maxSkillSlots('cleric', 1), 2, 'non-kit casters keep the old curve');
assert.strictEqual(maxSkillSlots('ranger', 1), 1);

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
assert.ok(!SKILL_CATALOG.some((s) => s.id === 'mage_fireball'), 'fireball is display-only');

const runeOpen = classifySkills('runehand', 1, []);
assert.deepStrictEqual(runeOpen.available.map((s) => s.id), ['mage_inscribe']);
assert.deepStrictEqual(runeOpen.locked.map((s) => s.id), ['mage_sear', 'mage_glyph']);
assert.strictEqual(runeOpen.locked[0].levelRequired, 4);
assert.strictEqual(runeOpen.locked[1].levelRequired, 8);
assert.ok(![...runeOpen.available, ...runeOpen.locked].some((s) => /fireball|frost shield/i.test(s.name)));

const runeEquipped = classifySkills('Rune Hand', 1, ['mage_inscribe']);
assert.strictEqual(runeEquipped.equipped[0]?.name, 'Inscribe');
assert.strictEqual(runeEquipped.available.length, 0, 'equipped Inscribe leaves Available empty');
assert.deepStrictEqual(runeEquipped.locked.map((s) => s.name), ['Sear', 'Glyph']);
assert.strictEqual(formatSkillCost(runeEquipped.equipped[0]), '4 round CD');
assert.strictEqual(formatSkillCost(skillById('mage_glyph')), 'once / fight');
assert.ok(formatSkillEffects(skillById('mage_sear')).some((c) => c === '1.80× swing'));

const classified = classifySkills('fenwatch', 4, ['warrior_brace']);
assert.strictEqual(classified.equipped.length, 1);
assert.strictEqual(classified.equipped[0].name, 'Brace');
assert.ok(classified.available.some((s) => s.id === 'warrior_slam'));
assert.ok(!classified.available.some((s) => s.id === 'warrior_power_strike' || s.id === 'warrior_brace'));
assert.ok(classified.locked.some((s) => s.id === 'warrior_stand' && s.levelRequired === 8));
assert.ok(classified.locked.every((s) => s.levelRequired > 4));

const emptyIdx = firstEmptyHotbarIndex(DEFAULT_HOTBAR_BINDS);
assert.strictEqual(emptyIdx, 0, 'default bar first empty is slot 1');
const boundOnce = bindSkillToFirstEmptyHotbar(DEFAULT_HOTBAR_BINDS, 'mage_inscribe');
assert.strictEqual(boundOnce.status, 'bound');
assert.strictEqual(boundOnce.index, 0);
assert.strictEqual(boundOnce.binds[0].id, 'mage_inscribe');
assert.strictEqual(boundOnce.binds[0].name, 'Inscribe');
assert.ok(isSkillOnHotbar(boundOnce.binds, 'mage_inscribe'));
const boundAgain = bindSkillToFirstEmptyHotbar(boundOnce.binds, 'mage_inscribe');
assert.strictEqual(boundAgain.status, 'already');
const boundSecond = bindSkillToFirstEmptyHotbar(boundOnce.binds, 'warrior_brace');
assert.strictEqual(boundSecond.status, 'bound');
assert.strictEqual(boundSecond.index, 1);
console.log('hudPrefs: skill catalog + slot helpers OK');

// Combat's temporary cover leaves all nine saved binds usable and unchanged.
{
  const binds = normalizeHotbarBinds(Array.from({ length: 9 }, () => makeActionBind('flee')));
  const before = JSON.stringify(binds);
  for (const inCombat of [false, true, false]) {
    for (const bind of binds) {
      assert.equal(resolveHotbarActivation(bind, { inCombat, inventory: [] }).ok, true);
    }
  }
  assert.equal(JSON.stringify(binds), before);
}

// --- Per-character hotbar: drop other class's spells ---
{
  const mageBar = normalizeHotbarBinds([
    makeSkillBind('mage_fireball'),
    makeSkillBind('mage_frost_shield'),
    makeActionBind('rest'),
    makeItemBind({ name: 'Health Potion', templateId: 'ITM0099' }),
  ]);
  const asWarrior = filterHotbarSkillsForCharacter(mageBar, {
    classId: 'warrior',
    level: 20,
    equippedIds: ['warrior_power_strike'],
  });
  assert.strictEqual(asWarrior[0], null, 'mage fireball dropped for warrior');
  assert.strictEqual(asWarrior[1], null, 'mage frost shield dropped for warrior');
  assert.strictEqual(asWarrior[2]?.id, 'rest', 'action binds stay');
  assert.strictEqual(asWarrior[3]?.kind, 'item', 'item binds stay');

  const asMage = filterHotbarSkillsForCharacter(mageBar, {
    classId: 'wizard',
    level: 12,
    equippedIds: ['mage_fireball'],
  });
  assert.strictEqual(asMage[0]?.id, 'mage_fireball', 'equipped mage skill kept');
  assert.strictEqual(asMage[1], null, 'unequipped mage skill dropped even on a mage');

  const high = skillsForClass('mage').find((s) => s.levelRequired > 1);
  if (high) {
    const gated = filterHotbarSkillsForCharacter(
      [makeSkillBind(high.id)],
      { classId: 'mage', level: 1, equippedIds: [] }
    );
    assert.strictEqual(gated[0], null, 'skill not in the equipped set is dropped');
    const equippedLocked = filterHotbarSkillsForCharacter(
      [makeSkillBind(high.id)],
      { classId: 'mage', level: 1, equippedIds: [high.id] }
    );
    assert.strictEqual(equippedLocked[0]?.id, high.id, 'equipped skill kept even if over level gate');
  }

  const untouched = filterHotbarSkillsForCharacter(mageBar, {
    classId: '',
    level: 0,
    equippedIds: null,
  });
  assert.strictEqual(untouched[0]?.id, 'mage_fireball', 'no class and no equipped list does not wipe');
  const classOnly = filterHotbarSkillsForCharacter(mageBar, {
    classId: 'mage',
    level: 10,
    equippedIds: null,
  });
  assert.strictEqual(classOnly[0], null, 'legacy Fireball is not in the rune hand kit');
  assert.strictEqual(classOnly[1], null, 'legacy Frost Shield is not in the rune hand kit');
  const kitBar = filterHotbarSkillsForCharacter(
    [makeSkillBind('mage_inscribe'), makeSkillBind('mage_fireball'), makeSkillBind('mage_glyph')],
    { classId: 'runehand', level: 1, equippedIds: null }
  );
  assert.strictEqual(kitBar[0]?.id, 'mage_inscribe', 'level-available kit skill stays when equipped list is unknown');
  assert.strictEqual(kitBar[1], null, 'fireball dropped for rune hand');
  assert.strictEqual(kitBar[2], null, 'glyph is locked at L1');
  const wiped = filterHotbarSkillsForCharacter(mageBar, {
    classId: 'warrior',
    level: 10,
    equippedIds: [],
  });
  assert.strictEqual(wiped[0], null, 'empty equipped set clears skill slots');

  let state = reconcileHotbarForCharacter({
    activeCharacterId: '',
    map: {},
    activeBinds: mageBar,
    characterId: 'char-mage',
    classId: 'mage',
    level: 10,
    equippedIds: ['mage_fireball'],
  });
  assert.strictEqual(state.activeCharacterId, 'char-mage');
  assert.strictEqual(state.binds[0]?.id, 'mage_fireball');
  assert.ok(state.map['char-mage'], 'legacy bar claimed by the first character');

  state = reconcileHotbarForCharacter({
    activeCharacterId: state.activeCharacterId,
    map: state.map,
    activeBinds: state.binds,
    characterId: 'char-warrior',
    classId: 'warrior',
    level: 15,
    equippedIds: ['warrior_shield_bash'],
  });
  assert.strictEqual(state.map['char-mage'][0]?.id, 'mage_fireball', 'previous character bar preserved');
  assert.ok(
    state.binds.every((b) => !b || b.kind !== 'skill' || b.id.startsWith('warrior_')),
    'warrior bar has no mage spells'
  );
  assert.strictEqual(state.binds[DEFAULT_REST_SLOT]?.id, 'rest', 'new character gets Rest, not the other bar');

  const back = reconcileHotbarForCharacter({
    activeCharacterId: state.activeCharacterId,
    map: state.map,
    activeBinds: state.binds,
    characterId: 'char-mage',
    classId: 'mage',
    level: 10,
    equippedIds: ['mage_fireball'],
  });
  assert.strictEqual(back.binds[0]?.id, 'mage_fireball', 'switching back restores that character bar');
  assert.ok(
    !back.binds.some((b) => b && b.kind === 'skill' && String(b.id).startsWith('warrior_')),
    'mage bar does not show warrior skills'
  );
}
console.log('hudPrefs: per-character hotbar filter OK');
