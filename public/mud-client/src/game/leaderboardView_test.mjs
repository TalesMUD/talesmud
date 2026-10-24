import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { leaderboardView } from './leaderboardView.js';

const read = (path) => readFileSync(new URL(path, import.meta.url), 'utf8');

test('leaderboard rows keep public fields and mark the viewer', () => {
  const view = leaderboardView({
    limit: 1,
    rows: [
      { rank: 1, name: 'Ced', class: 'Rogue', race: 'Human', level: 9, itemScore: 0, email: 'secret@example.com', id: 'u-play' },
    ],
    you: { rank: 3, name: 'Bea', class: 'Rogue', race: 'Human', level: 5, itemScore: 0, belongsUser: 'u-play' },
  });
  assert.equal(view.limit, 1);
  assert.equal(view.rows.length, 1);
  assert.equal(view.rows[0].className, 'Rogue');
  assert.equal(view.rows[0].yours, false);
  assert.equal(view.rows[0].email, undefined);
  assert.equal(view.rows[0].id, undefined);
  assert.equal(view.yourRank.rank, 3);
  assert.equal(view.yourRank.outside, true);
  assert.equal(view.yourRank.belongsUser, undefined);
  assert.deepEqual(Object.keys(view.rows[0]).sort(), ['className', 'itemScore', 'level', 'name', 'race', 'rank', 'yours']);
});

test('a viewer inside the page is highlighted and not marked outside', () => {
  const view = leaderboardView({
    rows: [{ rank: 2, name: 'Ada', class: 'Warrior', race: 'Dwarf', level: 5, itemScore: 80 }],
    you: { rank: 2, name: 'Ada', class: 'Warrior', race: 'Dwarf', level: 5, itemScore: 80 },
  });
  assert.equal(view.rows[0].yours, true);
  assert.equal(view.yourRank.outside, false);
});

test('leaderboard controls declare 44px targets', () => {
  const panel = read('ui/LeaderboardPanel.svelte');
  const menu = read('ui/CharacterSwitcher.svelte');
  const phone = read('mobile/MobileHeader.svelte');
  const sheet = read('mobile/MobileLayout.svelte');
  assert.match(panel, /min-height:\s*44px/);
  assert.match(panel, /min-width:\s*44px/);
  assert.match(menu, /class="menu-item rank-launch"/);
  assert.match(menu, /min-height:\s*44px/);
  assert.match(phone, /Leaderboard/);
  assert.match(phone, /min-height:\s*44px/);
  assert.match(sheet, /openSheet === 'rank'/);
  assert.doesNotMatch(panel, /#[dD]4af37|gold|gf-plaque/);
});
