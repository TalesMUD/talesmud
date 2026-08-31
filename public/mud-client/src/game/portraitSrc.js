const AVATAR_COUNT = 14;

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
  return hashedAvatar(entity.templateId || entity.id || entity.name);
}

export function onPortraitError(ev, entity) {
  const img = ev && ev.currentTarget;
  if (!img || img.dataset.fallback === "1") return;
  img.dataset.fallback = "1";
  img.src = hashedAvatar(entity && (entity.templateId || entity.id || entity.name));
}
