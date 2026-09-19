/**
 * Pure helpers for party overlay store payloads (mirrors Friends structured msgs).
 */
export function normalizePartyState(raw) {
  const next = raw && typeof raw === 'object' ? raw : {};
  return {
    inParty: !!next.inParty,
    partyId: String(next.partyId || next.partyID || ''),
    partyName: String(next.partyName || ''),
    members: Array.isArray(next.members) ? next.members : [],
  };
}

export function normalizePartyInvite(raw) {
  if (!raw || !raw.pending) return null;
  return {
    pending: true,
    inviterName: String(raw.inviterName || ''),
    partyId: String(raw.partyId || raw.partyID || ''),
  };
}
