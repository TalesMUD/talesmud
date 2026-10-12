import assert from 'assert';
import { readFileSync } from 'fs';
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';

const here = dirname(fileURLToPath(import.meta.url));
const overlay = readFileSync(join(here, 'DialogOverlay.svelte'), 'utf8');
const client = readFileSync(join(here, '../Client.js'), 'utf8');
const shop = readFileSync(join(here, 'ShopOverlay.svelte'), 'utf8');
const index = readFileSync(join(here, '../../../public/index.html'), 'utf8');

assert.equal(overlay.includes('Leave conversation'), false);
assert.equal(overlay.includes('dialogclose'), false);
assert.equal(/keydown|Escape/.test(overlay), false);
assert.equal(overlay.includes('npc-name'), true);
assert.equal(overlay.includes('dialog-text'), true);

const endAt = client.indexOf('messageHandlers["dialogEnd"]');
assert.notEqual(endAt, -1);
const endHandler = client.slice(endAt, client.indexOf('messageHandlers["combatStart"]'));
assert.equal(endHandler.includes('clearDialog'), true);
assert.equal(endHandler.includes('dialogclose'), false);

const roomAt = client.indexOf('if (roomChanged)');
assert.notEqual(roomAt, -1);
const roomClear = client.slice(roomAt, roomAt + 400);
assert.equal(roomClear.includes('clearDialog'), true);

assert.equal(shop.includes('shopclose'), true);
assert.equal(shop.includes("e.key !== 'Escape'"), true);
assert.match(index, /bundle\.js\?v=examine1/);
assert.equal(index.includes('pickup2'), false);
assert.equal(index.includes('pickup1'), false);
assert.equal(index.includes('wander2'), false);
assert.equal(index.includes('wander1'), false);

console.log('dialogOverlay_test ok');
