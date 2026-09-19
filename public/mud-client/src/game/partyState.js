/**
 * Pure helpers for party overlay store payloads (mirrors Friends structured msgs).
 */

export const DEFAULT_MAX_PARTY = 5;

export function normalizePartyMember(raw) {
  const m = raw && typeof raw === 'object' ? raw : {};
  return {
    id: String(m.id || ''),
    name: String(m.name || ''),
    online: !!m.online,
    level: Number(m.level) || 0,
    class: String(m.class || m.className || ''),
    portrait: String(m.portrait || m.picture || ''),
    isLeader: !!(m.isLeader || m.leader),
  };
}

export function normalizePartyState(raw) {
  const next = raw && typeof raw === 'object' ? raw : {};
  const members = Array.isArray(next.members)
    ? next.members.map(normalizePartyMember)
    : [];
  const maxMembers = Number(next.maxMembers) > 0 ? Number(next.maxMembers) : DEFAULT_MAX_PARTY;
  return {
    inParty: !!next.inParty,
    partyId: String(next.partyId || next.partyID || ''),
    partyName: String(next.partyName || ''),
    leaderId: String(next.leaderId || next.leaderCharacterId || next.LeaderCharacterID || ''),
    maxMembers,
    members,
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

/** Parse "[Party] Name: text" or "[Party] Name joined/left..." into a strip line. */
export function parsePartyChatLine(text, selfName) {
  const raw = String(text || '').trim();
  if (!raw.startsWith('[Party]')) return null;
  const body = raw.slice('[Party]'.length).trim();
  const say = body.match(/^([^:]+):\s*(.+)$/);
  if (say) {
    const name = say[1].trim();
    return {
      id: `psay-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
      name,
      text: say[2].trim(),
      isYou: !!(selfName && name.toLowerCase() === String(selfName).toLowerCase()),
      system: false,
    };
  }
  return {
    id: `psys-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
    name: '',
    text: body,
    isYou: false,
    system: true,
  };
}
