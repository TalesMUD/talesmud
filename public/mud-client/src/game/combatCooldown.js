/**
 * Client-side skill cooldown clock.
 *
 * Server only ships round-count SkillCooldowns. Convert to a wall-clock
 * seconds countdown using DecisionWindow (5s) + BeatBudget (~1.4s), and
 * recalibrate from observed player-turn intervals when we can.
 */

export const DEFAULT_DECISION_WINDOW_MS = 5000;
export const DEFAULT_BEAT_BUDGET_MS = 1400; // TurnBeatMs 1000 + ReactionMs 400

export function defaultSecPerRound() {
  return (DEFAULT_DECISION_WINDOW_MS + DEFAULT_BEAT_BUDGET_MS) / 1000;
}

export function cooldownRoundsForBind(bind, skillCooldowns) {
  if (!bind || bind.kind !== 'skill') return 0;
  const map = skillCooldowns && typeof skillCooldowns === 'object' ? skillCooldowns : {};
  const id = bind.id || '';
  if (id && map[id] > 0) return map[id];
  const byName = map[bind.name] || 0;
  return byName > 0 ? byName : 0;
}

export function cooldownKeysForBind(bind) {
  if (!bind) return [];
  const keys = [];
  if (bind.id) keys.push(bind.id);
  if (bind.name && bind.name !== bind.id) keys.push(bind.name);
  return keys;
}

export function displayCooldownSec(untilMs, nowMs, rounds, secPerRound) {
  if (!(rounds > 0)) return 0;
  const now = Number(nowMs) || 0;
  if (untilMs > now) return Math.max(1, Math.ceil((untilMs - now) / 1000));
  const rate = secPerRound > 0 ? secPerRound : defaultSecPerRound();
  return Math.max(1, Math.ceil(rounds * rate));
}

export function createSkillCooldownClock() {
  let untilById = Object.create(null);
  let roundsById = Object.create(null);
  let observedRoundMs = 0;
  let lastSelfTurnKey = '';
  let lastSelfTurnAt = 0;

  return {
    get secPerRound() {
      if (observedRoundMs >= 800) return observedRoundMs / 1000;
      return defaultSecPerRound();
    },

    noteSelfTurn(turnKey, nowMs) {
      const key = String(turnKey || '');
      if (!key || key === lastSelfTurnKey) return;
      const now = Number(nowMs) || Date.now();
      if (lastSelfTurnAt > 0) {
        const dt = now - lastSelfTurnAt;
        if (dt >= 800 && dt <= 60000) observedRoundMs = dt;
      }
      lastSelfTurnKey = key;
      lastSelfTurnAt = now;
    },

    sync(roundsMap, nowMs) {
      const next = roundsMap && typeof roundsMap === 'object' ? roundsMap : {};
      const now = Number(nowMs) || Date.now();
      const rate = this.secPerRound;
      const until = Object.create(null);
      const rounds = Object.create(null);
      for (const [id, raw] of Object.entries(next)) {
        const n = Number(raw) || 0;
        if (!(n > 0)) continue;
        rounds[id] = n;
        if (roundsById[id] !== n) {
          until[id] = now + n * rate * 1000;
        } else {
          until[id] = untilById[id] || now + n * rate * 1000;
        }
      }
      untilById = until;
      roundsById = rounds;
    },

    secondsFor(bind, nowMs) {
      const keys = cooldownKeysForBind(bind);
      const now = Number(nowMs) || Date.now();
      for (const key of keys) {
        const rounds = roundsById[key] || 0;
        if (!(rounds > 0)) continue;
        return displayCooldownSec(untilById[key] || 0, now, rounds, this.secPerRound);
      }
      return 0;
    },

    roundsFor(bind) {
      const keys = cooldownKeysForBind(bind);
      for (const key of keys) {
        if (roundsById[key] > 0) return roundsById[key];
      }
      return 0;
    },

    hasAny() {
      return Object.keys(roundsById).length > 0;
    },

    reset() {
      untilById = Object.create(null);
      roundsById = Object.create(null);
      observedRoundMs = 0;
      lastSelfTurnKey = '';
      lastSelfTurnAt = 0;
    },
  };
}
