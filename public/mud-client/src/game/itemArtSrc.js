function stripInstance(id) {
  const key = String(id || "").trim();
  const i = key.lastIndexOf("~");
  return i > 0 ? key.slice(0, i) : key;
}

/** Meta.img is often an art-generation prompt — never treat prose as an <img src>. */
export function looksLikeArtPath(value) {
  if (value == null) return false;
  const s = String(value).trim();
  if (!s || /\s/.test(s)) return false;
  if (s.startsWith("/") || s.startsWith("sprites/") || s.startsWith("./")) return true;
  if (/^https?:\/\//i.test(s)) return true;
  if (/\.(png|jpe?g|webp|svg|gif)(\?|#|$)/i.test(s)) return true;
  return false;
}

export function itemArtGenericKey(item) {
  if (!item) return "default";
  const type = String(item.type || "").toLowerCase();
  const sub = String(item.subType || "").toLowerCase();
  const name = String(item.name || "").toLowerCase();
  if (sub.includes("torch") || name.includes("torch")) return "torch";
  switch (type) {
    case "weapon":
      return "weapon";
    case "armor":
      return "armor";
    case "quest":
      return "quest";
    case "currency":
      return "currency";
    case "consumable":
      return "consumable";
    case "collectible":
    case "crafting_material":
      return "junk";
    default:
      return "default";
  }
}

export function itemArtSrc(item) {
  if (!item) return "sprites/items/generic-default.svg";
  // Prefer explicit URL fields only when they look like real art paths.
  if (looksLikeArtPath(item.image)) return item.image;
  const metaImg = item.meta && item.meta.img;
  if (looksLikeArtPath(metaImg)) return metaImg;
  const tid = stripInstance(item.templateId || item.id);
  if (tid) return `/api/item-art/${tid}.png`;
  // No template id — start on generic PNG (SVG is last resort via onItemArtError).
  return `/api/item-art/generic-${itemArtGenericKey(item)}.png`;
}

const ITEM_SILHOUETTE =
  "data:image/svg+xml;charset=utf-8," +
  encodeURIComponent(
    '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><rect width="64" height="64" fill="#1c1814"/><rect x="18" y="14" width="28" height="36" rx="4" fill="#d4a44a"/></svg>'
  );

/** Next src after a failed item image. Stage 2 is a data URI and cannot 404. */
export function itemArtFallbackSrc(item, stage) {
  const key = itemArtGenericKey(item);
  const n = Number(stage) || 0;
  if (n <= 0) return `/api/item-art/generic-${key}.png`;
  if (n === 1) return "/api/item-art/generic-default.png";
  return ITEM_SILHOUETTE;
}

export function onItemArtError(ev, item) {
  const img = ev && ev.currentTarget;
  if (!img || img.dataset.fallback === "done") return;
  let stage = Number(img.dataset.fallback || "0");
  if (!Number.isFinite(stage) || stage < 0) stage = 0;
  let next = itemArtFallbackSrc(item, stage);
  const attr = img.getAttribute("src") || "";
  if (attr === next || (next.startsWith("/") && String(img.src || "").endsWith(next))) {
    stage += 1;
    next = itemArtFallbackSrc(item, stage);
  }
  img.dataset.fallback = stage >= 2 ? "done" : String(stage + 1);
  img.src = next;
}
