import assert from 'assert';
import {
  normalizePartyInvite,
  normalizePartyState,
  parsePartyChatLine,
  DEFAULT_MAX_PARTY,
} from './partyState.js';

assert.deepStrictEqual(
  normalizePartyState(null),
  { inParty: false, partyId: '', partyName: '', leaderId: '', maxMembers: DEFAULT_MAX_PARTY, members: [] },
  'null party → empty'
);

assert.deepStrictEqual(
  normalizePartyState({
    inParty: true,
    partyID: 'p1',
    partyName: "Aster's Party",
    leaderId: 'c1',
    maxMembers: 5,
    members: [{ id: 'c1', name: 'Aster', online: true, level: 3, class: 'Warrior', isLeader: true }],
  }),
  {
    inParty: true,
    partyId: 'p1',
    partyName: "Aster's Party",
    leaderId: 'c1',
    maxMembers: 5,
    members: [{ id: 'c1', name: 'Aster', online: true, level: 3, class: 'Warrior', portrait: '', isLeader: true }],
  },
  'partyID alias + rich members preserved'
);

assert.strictEqual(normalizePartyInvite(null), null);
assert.strictEqual(normalizePartyInvite({ pending: false }), null);
assert.deepStrictEqual(
  normalizePartyInvite({ pending: true, inviterName: 'Aster', partyId: 'p1' }),
  { pending: true, inviterName: 'Aster', partyId: 'p1' },
  'pending invite normalized'
);

const say = parsePartyChatLine('[Party] Aster: regroup', 'Aster');
assert.ok(say && say.name === 'Aster' && say.text === 'regroup' && say.isYou);
const sys = parsePartyChatLine('[Party] Bryn joined the party.', 'Aster');
assert.ok(sys && sys.system && sys.text.includes('joined'));

console.log('partyState: normalize helpers OK');
