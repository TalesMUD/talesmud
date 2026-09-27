import assert from 'assert';
import { itemArtFallbackSrc, itemArtGenericKey } from './itemArtSrc.js';
import { figureFallback, playerPortraitSrc, playerSilhouette, portraitSrc, SILHOUETTE } from './portraitSrc.js';

const fang = { templateId: 'ITM0087', type: 'weapon', name: 'Fang' };
assert.equal(itemArtGenericKey(fang), 'weapon');
assert.equal(itemArtFallbackSrc(fang, 0), '/api/item-art/generic-weapon.png');
assert.equal(itemArtFallbackSrc(fang, 1), '/api/item-art/generic-default.png');
assert.ok(itemArtFallbackSrc(fang, 2).startsWith('data:image/svg+xml'));
assert.equal(itemArtFallbackSrc(fang, 9), itemArtFallbackSrc(fang, 2));

const fox = { templateId: 'ENM0038', name: 'Shadow Fox', isEnemy: true };
assert.equal(portraitSrc(fox), '/api/portraits/ENM0038.png');
assert.equal(figureFallback(fox), SILHOUETTE.enemy);

const guest = { id: '11111111-1111-1111-1111-111111111111', class: { name: 'Wizard' } };
assert.equal(portraitSrc(guest), playerSilhouette('Wizard'));
assert.equal(playerSilhouette('Wizard'), SILHOUETTE.mage);
assert.equal(portraitSrc({ ...guest, race: { id: 'human' }, class: { id: 'warrior' } }), '/api/portraits/player-human-warrior.png');
assert.equal(playerPortraitSrc({ race: { id: 'elve' }, class: { id: 'wizard' } }), '/api/portraits/player-elf-mage.png');
assert.equal(playerPortraitSrc({ race: 'Dwarf', class: 'Hunter' }), '/api/portraits/player-dwarf-ranger.png');
assert.equal(playerPortraitSrc({ race: 'Orc', class: 'Warrior' }), '');
assert.equal(figureFallback({ name: 'Elder' }), SILHOUETTE.npc);

console.log('artFallback_test ok');
