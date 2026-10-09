import assert from 'assert';
import { combatEventCard, combatEventIcon, isCombatEvent } from './combatEvent.js';

assert.strictEqual(isCombatEvent({ type: 'message', message: 'drip' }), false);
assert.strictEqual(combatEventCard({ type: 'message', username: 'SYSTEM', message: 'drip' }), null);

const card = combatEventCard({
  type: 'message',
  username: 'SYSTEM',
  message: 'The chanter hits the drum once.',
  style: 'combatEvent',
  hook: 'onAggro',
  source: 'The Barrow Warden',
});
assert.deepStrictEqual(card, {
  text: 'The chanter hits the drum once.',
  hook: 'onAggro',
  source: 'The Barrow Warden',
  icon: 'music_note',
});

assert.strictEqual(combatEventIcon('onLowHealth'), 'favorite');
assert.strictEqual(combatEventIcon('onDeath'), 'dangerous');
assert.strictEqual(combatEventIcon('onFlee'), 'directions_run');
assert.strictEqual(combatEventIcon('unique'), 'diamond');
assert.strictEqual(combatEventIcon('other'), 'campaign');

console.log('combatEvent_test: ok');
