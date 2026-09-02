const AVATAR_COUNT = 14;

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

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
  return `img/avatars/${(h % AVATAR_COUNT) + 1}p.png`;
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

/** Full-body sprite URL for room cards (2:3). */
export function portraitSrc(entity) {
  if (!entity) return hashedAvatar("npc");
  const key = portraitTemplateKey(entity);
  if (key) return `/api/portraits/${key}.png`;
  return hashedAvatar(entity.name || "npc");
}

/** 1:1 bust for dialog; falls back to full-body in onPortraitBustError. */
export function portraitBustSrc(entity) {
  const key = portraitTemplateKey(entity);
  if (key) return `/api/portraits/${key}-bust.png`;
  return portraitSrc(entity);
}

export function onPortraitError(ev, entity) {
  const img = ev && ev.currentTarget;
  if (!img || img.dataset.fallback === "1") return;
  img.dataset.fallback = "1";
  const key = portraitTemplateKey(entity);
  img.src = hashedAvatar(key || (entity && entity.name) || "npc");
}

export function onPortraitBustError(ev, entity) {
  const img = ev && ev.currentTarget;
  if (!img || img.dataset.fallback === "1") return;
  img.dataset.fallback = "1";
  img.src = portraitSrc(entity);
}
