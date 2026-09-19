import assert from 'assert';
import { normalizePartyInvite, normalizePartyState } from './partyState.js';

assert.deepStrictEqual(
  normalizePartyState(null),
  { inParty: false, partyId: '', partyName: '', members: [] },
  'null party → empty'
);

assert.deepStrictEqual(
  normalizePartyState({
    inParty: true,
    partyID: 'p1',
    partyName: "Aster's Party",
    members: [{ id: 'c1', name: 'Aster', online: true }],
  }),
  {
    inParty: true,
    partyId: 'p1',
    partyName: "Aster's Party",
    members: [{ id: 'c1', name: 'Aster', online: true }],
  },
  'partyID alias + members preserved'
);

assert.strictEqual(normalizePartyInvite(null), null);
assert.strictEqual(normalizePartyInvite({ pending: false }), null);
assert.deepStrictEqual(
  normalizePartyInvite({ pending: true, inviterName: 'Aster', partyId: 'p1' }),
  { pending: true, inviterName: 'Aster', partyId: 'p1' },
  'pending invite normalized'
);

console.log('partyState: normalize helpers OK');
