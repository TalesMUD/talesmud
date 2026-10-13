/**
 * Pure helpers for Party Need/Greed loot roll popup.
 */

export function normalizeNeedGreed(raw) {
  if (!raw || !raw.pending) return null;
  const expiresAt = Number(raw.expiresAt) || 0;
  const quantity = Number(raw.quantity) || 0;
  const name = String(raw.itemName || raw.name || '').trim();
  const quality = String(raw.quality || raw.rarity || 'normal').toLowerCase() || 'normal';
  let itemLabel = name || 'an item';
  if (quantity > 1) {
    itemLabel = `${itemLabel} (x${quantity})`;
  }
  return {
    pending: true,
    roundId: String(raw.roundId || raw.roundID || ''),
    itemName: name,
    quality,
    quantity: quantity > 0 ? quantity : 1,
    itemLabel,
    expiresAt,
  };
}
