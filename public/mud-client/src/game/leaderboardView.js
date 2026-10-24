// Maps GET /api/leaderboard into the rows the panel renders.
// Private account fields are dropped even if a response includes them.

const PUBLIC_KEYS = ['rank', 'name', 'class', 'race', 'level', 'itemScore'];

export function leaderboardView(payload) {
  const rows = Array.isArray(payload?.rows) ? payload.rows.map(mapRow) : [];
  const you = payload?.you ? mapRow(payload.you) : null;
  const marked = rows.map((row) => ({
    ...row,
    yours: !!(you && row.rank === you.rank && row.name === you.name),
  }));
  const yourRank = you ? { ...you, outside: !marked.some((row) => row.yours) } : null;
  return {
    limit: Number(payload?.limit) || marked.length,
    rows: marked,
    yourRank,
  };
}

function mapRow(row) {
  const out = {};
  for (const key of PUBLIC_KEYS) {
    if (row && Object.prototype.hasOwnProperty.call(row, key)) out[key] = row[key];
  }
  return {
    rank: Number(out.rank) || 0,
    name: String(out.name || ''),
    className: String(out.class || ''),
    race: String(out.race || ''),
    level: Number(out.level) || 0,
    itemScore: Number(out.itemScore) || 0,
  };
}
