const AVATAR_COUNT = 14;

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

function svgUri(markup) {
  return "data:image/svg+xml;charset=utf-8," + encodeURIComponent(
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 96">${markup}</svg>`
  );
}

function figure(fill, accent) {
  return svgUri(
    `<rect width="64" height="96" fill="#14110e"/>` +
    `<circle cx="32" cy="26" r="12" fill="${fill}"/>` +
    `<path d="M14 46h36l8 42H6z" fill="${fill}"/>` +
    `<rect x="28" y="18" width="8" height="22" rx="2" fill="${accent}"/>`
  );
}

/** Built-in silhouettes. A data URI cannot 404, so the swap never loops. */
export const SILHOUETTE = {
  enemy: figure("#3a2428", "#7f1d1d"),
  npc: figure("#3a3228", "#d4a44a"),
  adventurer: figure("#2a3140", "#94a3b8"),
  warrior: figure("#2c333c", "#cbd5e1"),
  rogue: figure("#241c30", "#c4b5fd"),
  mage: figure("#1c2438", "#93c5fd"),
  ranger: figure("#1c2c22", "#86efac"),
  druid: figure("#1c2c22", "#4ade80"),
  cleric: figure("#2c2818", "#fde68a"),
};

export function classToken(entity) {
  if (!entity) return "";
  const raw = entity.class || entity.classId || entity.charClass || "";
  if (typeof raw === "string") return raw;
  return raw.id || raw.name || "";
}

/** Class/race stand-in for a player with no portrait file. */
export function playerSilhouette(className) {
  const c = String(className || "").toLowerCase();
  if (c.includes("warrior") || c.includes("fighter")) return SILHOUETTE.warrior;
  if (c.includes("rogue") || c.includes("thief")) return SILHOUETTE.rogue;
  if (c.includes("mage") || c.includes("wizard") || c.includes("sorc")) return SILHOUETTE.mage;
  if (c.includes("ranger") || c.includes("hunter")) return SILHOUETTE.ranger;
  if (c.includes("druid")) return SILHOUETTE.druid;
  if (c.includes("cleric") || c.includes("priest")) return SILHOUETTE.cleric;
  return SILHOUETTE.adventurer;
}

export function enemySilhouette() {
  return SILHOUETTE.enemy;
}

export function npcSilhouette() {
  return SILHOUETTE.npc;
}

/** Dark stand-in when a room painting 404s. Not an <img>, so it cannot show a broken icon. */
export const ROOM_PLACEHOLDER = svgUri('<rect width="64" height="96" fill="#1a1612"/>');

/** One replacement image after a portrait 404. */
export function figureFallback(entity) {
  if (!entity) return SILHOUETTE.npc;
  if (entity.isEnemy) return SILHOUETTE.enemy;
  const cls = classToken(entity);
  if (cls) return playerSilhouette(cls);
  return SILHOUETTE.npc;
}

function stripInstance(id) {
  const key = String(id || "").trim();
  const i = key.lastIndexOf("~");
  return i > 0 ? key.slice(0, i) : key;
}

export function hashedAvatar(id) {
  const key = String(id || "npc");
  let h = 0;
  for (let i = 0; i < key.length; i++) {
    h = (h * 31 + key.charCodeAt(i)) >>> 0;
  }
  return `/img/avatars/${(h % AVATAR_COUNT) + 1}p.png`;
}

/** Template key for portrait files — always prefer templateId over instance UUID. */
export function portraitTemplateKey(entity) {
  if (!entity) return "";
  const tid = stripInstance(entity.templateId);
  if (tid) return tid;
  const id = stripInstance(entity.id);
  if (id && !UUID_RE.test(id)) return id;
  return "";
}

/** Stable race/class portrait for playable characters. */
export function playerPortraitSrc(entity) {
  if (!entity) return '';
  const value = (field) => typeof field === 'string' ? field : (field?.id || field?.name || '');
  let race = String(value(entity.race)).toLowerCase();
  let cls = String(classToken(entity)).toLowerCase();
  if (race === 'elve' || race === 'elves') race = 'elf';
  if (cls === 'wizard') cls = 'mage';
  if (cls === 'hunter') cls = 'ranger';
  if (!['human', 'dwarf', 'elf'].includes(race) || !['warrior', 'rogue', 'mage', 'ranger', 'cleric', 'druid'].includes(cls)) return '';
  return `/api/portraits/player-${race}-${cls}.png`;
}

/**
 * Battle-stage sprite. Prefer the server portrait URL, then race/class art or
 * the enemy template file. A class or enemy silhouette is only the last resort
 * (and the image onerror target), so a missing file never stays a broken icon.
 * @param {object} combatant
 * @param {{ isPlayer?: boolean, selfId?: string, character?: object }} [opts]
 */
export function battleSpriteSrc(combatant, opts = {}) {
  const c = combatant || {};
  const p = String(c.portrait || '');
  if (p.startsWith('data:')) return p;
  if (p && !p.startsWith('img/')) {
    if (p.startsWith('/') || p.startsWith('http')) return p;
    return `/api/portraits/${p.replace(/\.png$/i, '')}.png`;
  }
  const isPlayer = opts.isPlayer === true || c.type === 'player';
  if (isPlayer) {
    const self = opts.selfId && c.id === opts.selfId ? opts.character : null;
    if (self) {
      const selfArt = portraitSrc(self);
      if (selfArt && !String(selfArt).startsWith('data:')) return selfArt;
    }
    const raced = playerPortraitSrc({
      race: c.race || c.raceId || (self && self.race),
      class: c.classId || c.class || c.charClass || (self && classToken(self)),
    });
    if (raced) return raced;
    if (self) return portraitSrc(self);
    return playerSilhouette(c.classId || c.class);
  }
  const templateKey = stripInstance(c.templateId || c.templateID || '');
  if (templateKey && !UUID_RE.test(templateKey)) {
    return `/api/portraits/${templateKey.replace(/\.png$/i, '')}.png`;
  }
  const idKey = stripInstance(c.id || '');
  if (idKey && !UUID_RE.test(idKey)) {
    return `/api/portraits/${idKey.replace(/\.png$/i, '')}.png`;
  }
  return enemySilhouette();
}

/** Full-body sprite URL for room cards (2:3). Players with no template use a class silhouette. */
export function portraitSrc(entity) {
  if (!entity) return figureFallback(null);
  const key = portraitTemplateKey(entity);
  if (key) return `/api/portraits/${key}.png`;
  const player = playerPortraitSrc(entity);
  if (player) return player;
  return figureFallback(entity);
}

/** 1:1 bust for dialog; falls back to full-body in onPortraitBustError. */
export function portraitBustSrc(entity) {
  const key = portraitTemplateKey(entity);
  if (key) return `/api/portraits/${key}-bust.png`;
  return portraitSrc(entity);
}

function swapOnce(img, next) {
  if (!img || img.dataset.fallback === "1") return;
  img.dataset.fallback = "1";
  if (!next || img.getAttribute("src") === next) return;
  img.src = next;
}

export function onPortraitError(ev, entity) {
  const img = ev && ev.currentTarget;
  swapOnce(img, figureFallback(entity));
}

export function onPortraitBustError(ev, entity) {
  const img = ev && ev.currentTarget;
  // Prefer full-body sprite; only silhouette if that is also missing.
  swapOnce(img, portraitSrc(entity));
}
