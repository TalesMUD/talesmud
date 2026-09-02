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

export function onItemArtError(ev, item) {
  const img = ev && ev.currentTarget;
  if (!img) return;
  const stage = img.dataset.fallback || "0";
  const key = itemArtGenericKey(item);
  if (stage === "0") {
    img.dataset.fallback = "1";
    img.src = `/api/item-art/generic-${key}.png`;
    return;
  }
  if (stage === "1") {
    img.dataset.fallback = "2";
    img.src = `sprites/items/generic-${key}.svg`;
    return;
  }
  if (stage === "2") {
    img.dataset.fallback = "3";
    img.src = "sprites/items/generic-default.svg";
  }
}
