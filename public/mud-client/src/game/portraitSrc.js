const AVATAR_COUNT = 14;

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

export function portraitSrc(entity) {
  if (!entity) return hashedAvatar("npc");
  if (entity.portrait) return entity.portrait;
  const tid = stripInstance(entity.templateId || entity.id);
  if (tid) return `/api/portraits/${tid}.png`;
  return hashedAvatar(entity.name || "npc");
}

export function onPortraitError(ev, entity) {
  const img = ev && ev.currentTarget;
  if (!img || img.dataset.fallback === "1") return;
  img.dataset.fallback = "1";
  img.style.transform = "none";
  img.style.objectPosition = "50% 50%";
  img.src = hashedAvatar(entity && (entity.templateId || entity.id || entity.name));
}
