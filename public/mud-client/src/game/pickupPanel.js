import { itemEffectLines } from './itemEffects.js';
import { hotbarSlotFromKey, isTextEntry, rarityClass } from './keyboardShortcuts.js';

/** Ground rows the player can take. Room fixtures with noPickup stay out. */
export function pickableGroundItems(items) {
  if (!Array.isArray(items)) return [];
  return items.filter((item) => item && String(item.name || '').trim() && !item.noPickup);
}

export function pickupEntryCount(items) {
  return pickableGroundItems(items).length;
}

/** Hide the primary control when a single tap already takes everything. */
export function showPickupAll(items) {
  return pickupEntryCount(items) > 1;
}

export function shouldClosePickup(items) {
  return pickupEntryCount(items) === 0;
}

function hasCurrencyTag(item) {
  const tags = item && item.tags;
  if (!Array.isArray(tags)) return false;
  return tags.some((tag) => String(tag || '').trim().toLowerCase() === 'currency');
}

/**
 * Currency from type or a currency tag. A coin name matches only when the
 * payload has no type at all. Ordinary names that merely contain those letters do not match.
 */
export function isCurrencyItem(item) {
  if (!item) return false;
  const type = String(item.type || '').trim().toLowerCase();
  if (type === 'currency') return true;
  if (hasCurrencyTag(item)) return true;
  if (type) return false;
  const name = String(item.name || '').trim().toLowerCase();
  if (!name) return false;
  if (/\b(silver|gold|copper|platinum)\s+marks?\b/.test(name)) return true;
  if (/\b(gold|silver|copper|platinum)\s+coins?\b/.test(name)) return true;
  if (/^(gold|silver|copper|platinum|coins?|marks?)$/.test(name)) return true;
  return false;
}

export function groundItemCount(item) {
  const raw = item && (item.quantity != null ? item.quantity : item.count);
  const n = Number(raw);
  if (!Number.isFinite(n) || n < 1) return 1;
  return Math.floor(n);
}

function titleCaseType(value) {
  const raw = String(value || '').trim();
  if (!raw) return '';
  const key = raw.toLowerCase();
  if (key === 'crafting_material' || key === 'material') return 'Material';
  if (key === 'consumable') return 'Consumable';
  if (key === 'currency') return 'Currency';
  return raw.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

function healText(item) {
  const attrs = item && item.attributes;
  if (!attrs || typeof attrs !== 'object') return '';
  const raw = attrs.healthRestore != null ? attrs.healthRestore : attrs.healAmount;
  if (raw == null || raw === '') return '';
  const n = Number(raw);
  if (!Number.isFinite(n)) return '';
  return `Heals ${n} HP`;
}

/** Short type · effect line. The server effect wins; older payloads still derive one. */
export function groundItemHint(item) {
  if (!item) return '';
  const parts = [];
  let typeLabel = titleCaseType(item.type);
  if (!typeLabel && isCurrencyItem(item)) typeLabel = 'Currency';
  if (typeLabel) parts.push(typeLabel);
  const effect = String(item.effect || '').trim();
  if (effect) {
    if (effect !== typeLabel) parts.push(effect);
    return parts.join(' · ');
  }
  const heal = healText(item);
  if (heal) {
    parts.push(heal);
  } else {
    const lines = itemEffectLines(item);
    const text = lines.length ? String(lines[0].text || lines[0].name || '').trim() : '';
    if (text && text !== typeLabel) parts.push(text);
  }
  return parts.join(' · ');
}

export function groundItemRarity(item) {
  if (item && item.unique) return 'unique';
  return rarityClass(item && (item.quality || item.rarity));
}

function looksLikeTemplateKey(value) {
  const key = String(value || '').trim();
  if (!key) return false;
  if (key.includes('~')) return true;
  return /^ITM/i.test(key);
}

/**
 * Art input for itemArtSrc. Instance ids are not template files; leave them
 * blank so the helper starts on the type generic instead of a 404.
 */
export function groundArtItem(item) {
  const source = item || {};
  const currency = isCurrencyItem(source);
  const type = source.type || (currency ? 'currency' : '');
  const templateId = String(source.templateId || '').trim();
  const id = String(source.id || '').trim();
  const templateKeyed = looksLikeTemplateKey(templateId);
  const idKeyed = looksLikeTemplateKey(id);
  return {
    name: source.name,
    type,
    subType: source.subType,
    image: source.image,
    meta: source.meta,
    templateId: templateKeyed ? templateId : '',
    id: templateKeyed ? '' : (idKeyed ? id : ''),
  };
}

function rowId(item, index) {
  const id = item && item.id != null ? String(item.id).trim() : '';
  if (id) return id;
  const name = String(item && item.name || '').trim().toLowerCase();
  return `anon:${index}:${name}`;
}

/** One row per item name. Duplicate names and explicit quantities become the count. */
export function groupGroundItems(items) {
  const list = pickableGroundItems(items);
  const order = [];
  const map = new Map();
  list.forEach((item, index) => {
    const name = String(item.name).trim();
    const key = name.toLowerCase();
    let group = map.get(key);
    if (!group) {
      group = { key, name, entries: [] };
      map.set(key, group);
      order.push(group);
    }
    group.entries.push({ item, index });
  });
  return order.map((group, rowIndex) => {
    const sample = group.entries[0].item;
    const ids = group.entries.map(({ item, index }) => rowId(item, index));
    const count = group.entries.reduce((sum, { item }) => sum + groundItemCount(item), 0);
    const currency = group.entries.some(({ item }) => isCurrencyItem(item));
    const unique = group.entries.some(({ item }) => !!item.unique);
    return {
      key: group.key,
      name: group.name,
      ids,
      count,
      currency,
      unique,
      rarity: unique ? 'unique' : groundItemRarity(sample),
      hint: groundItemHint(sample),
      badge: rowIndex < 9 ? String(rowIndex + 1) : '',
      artItem: groundArtItem(sample),
    };
  });
}

/**
 * One `pickup <name>` per still-present id. Taken ids and ids no longer on
 * the ground are skipped, so a second click cannot send the same take twice.
 */
export function pickupCommands(items, { ids, takenIds } = {}) {
  const taken = new Set(takenIds || []);
  const wanted = ids ? new Set(ids) : null;
  const seen = new Set();
  const commands = [];
  pickableGroundItems(items).forEach((item, index) => {
    const id = rowId(item, index);
    if (seen.has(id) || taken.has(id)) return;
    if (wanted && !wanted.has(id)) return;
    seen.add(id);
    commands.push({
      id,
      command: `pickup ${String(item.name).trim()}`,
    });
  });
  return commands;
}

export function pickupRowCommands(items, row, takenIds) {
  const taken = new Set(takenIds || []);
  const id = (row && row.ids || []).find((value) => value && !taken.has(value));
  if (!id) return [];
  return pickupCommands(items, { ids: [id], takenIds });
}

export function pickupAllCommands(items, takenIds) {
  return pickupCommands(items, { takenIds });
}

/** 1–9 take that row, Enter takes all, Esc closes. Text entry and modifiers do nothing. */
export function pickupKeyAction(event, rowCount) {
  if (!event) return null;
  if (event.metaKey || event.ctrlKey || event.altKey) return null;
  if (isTextEntry(event.target)) return null;
  if (event.key === 'Escape') return { type: 'close' };
  if (event.key === 'Enter') return { type: 'all' };
  const slot = hotbarSlotFromKey(event.key);
  const rows = Number(rowCount) || 0;
  if (slot >= 0 && slot < rows && slot < 9) return { type: 'row', index: slot };
  return null;
}
