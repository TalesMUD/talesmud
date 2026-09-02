function stripInstance(id) {
  const key = String(id || "").trim();
  const i = key.lastIndexOf("~");
  return i > 0 ? key.slice(0, i) : key;
}

export function itemArtGenericKey(item) {
  if (!item) return "default";
  const type = String(item.type || "").toLowerCase();
  const sub = String(item.subType || "").toLowerCase();
  if (sub.includes("torch")) return "torch";
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
  if (item.image) return item.image;
  const metaImg = item.meta && item.meta.img;
  if (metaImg) return metaImg;
  const tid = stripInstance(item.templateId || item.id);
  if (tid) return `/api/item-art/${tid}.png`;
  return `sprites/items/generic-${itemArtGenericKey(item)}.svg`;
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
  }
}
