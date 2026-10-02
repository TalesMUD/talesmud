/**
 * Battle layout B — room sprites on gold markers (party left / enemies right),
 * edge-docked player (TL) + target (TR) frames, slim over-sprite HP (no name labels).
 * Layout B is the default when the player has no explicit preference.
 *
 * Classic opt-out:
 *   Settings → Gameplay → turn off "Battle layout B"
 *   localStorage.setItem('talesmud_battle_layout_b', '0')
 *   /play/?battleLayout=classic   or   /play/?battlepoc=0
 * Force Layout B:
 *   Settings toggle on
 *   localStorage.setItem('talesmud_battle_layout_b', '1')
 *   /play/?battleLayout=b   or   /play/?battlepoc=1
 *
 * URL wins over localStorage; an explicit localStorage 0/1 wins over Settings.
 * No URL and no localStorage key → Layout B, even if an older settings blob
 * still has battleLayoutB: false (that was the previous default, not a choice).
 */

export const BATTLE_LAYOUT_STORAGE_KEY = 'talesmud_battle_layout_b';

/** Fresh guests and anyone with no stored preference use Layout B. */
export const BATTLE_LAYOUT_DEFAULT = true;

/**
 * @param {unknown} value
 * @returns {boolean}
 */
export function normalizeBattleLayoutB(value) {
  return value === true || value === 'true' || value === 1 || value === '1';
}

/**
 * Parse a raw URL/localStorage token into on/off/null (unset).
 * @param {string|null|undefined} raw
 * @returns {boolean|null}
 */
export function parseBattleLayoutOverride(raw) {
  if (raw == null || raw === '') return null;
  const v = String(raw).trim().toLowerCase();
  if (v === 'b' || v === '1' || v === 'true' || v === 'on' || v === 'yes' || v === 'layoutb') {
    return true;
  }
  if (
    v === '0' ||
    v === 'false' ||
    v === 'off' ||
    v === 'no' ||
    v === 'classic' ||
    v === 'a' ||
    v === 'default'
  ) {
    return false;
  }
  return null;
}

/**
 * Resolve layout-B from URL search params (battleLayout / battlepoc).
 * @param {string} [search]
 * @returns {boolean|null}
 */
export function battleLayoutFromSearch(search) {
  if (typeof search !== 'string' || !search) return null;
  try {
    const q = new URLSearchParams(search.startsWith('?') ? search.slice(1) : search);
    const fromLayout = parseBattleLayoutOverride(q.get('battleLayout'));
    if (fromLayout !== null) return fromLayout;
    return parseBattleLayoutOverride(q.get('battlepoc'));
  } catch (_) {
    return null;
  }
}

/**
 * Resolve layout-B from localStorage key.
 * @param {Storage|null|undefined} [storage]
 * @returns {boolean|null}
 */
export function battleLayoutFromStorage(storage) {
  if (!storage || typeof storage.getItem !== 'function') return null;
  try {
    return parseBattleLayoutOverride(storage.getItem(BATTLE_LAYOUT_STORAGE_KEY));
  } catch (_) {
    return null;
  }
}

/**
 * Settings checkbox for a stored blob.
 * Absent localStorage is "no preference" and stays on Layout B, including when
 * the blob still has the old default `battleLayoutB: false`.
 * @param {unknown} _storedFlag previous settings value (ignored when unset)
 * @param {string|null|undefined} storageValue raw talesmud_battle_layout_b value
 * @returns {boolean}
 */
export function battleLayoutSettingsDefault(_storedFlag, storageValue) {
  const fromLs = parseBattleLayoutOverride(storageValue);
  if (fromLs !== null) return fromLs;
  return BATTLE_LAYOUT_DEFAULT;
}

/**
 * Final resolution: URL > localStorage > settings flag > Layout B.
 * Pass settingsFlag false only for an explicit Classic choice already in memory.
 * Omit it (fresh guest / no blob) and the result is Layout B.
 * @param {{
 *   search?: string,
 *   storage?: Storage|null,
 *   settingsFlag?: unknown,
 * }} [opts]
 * @returns {boolean}
 */
export function resolveBattleLayoutB(opts = {}) {
  const fromUrl = battleLayoutFromSearch(
    opts.search != null
      ? opts.search
      : typeof location !== 'undefined'
        ? location.search
        : ''
  );
  if (fromUrl !== null) return fromUrl;

  const storage =
    opts.storage !== undefined
      ? opts.storage
      : typeof localStorage !== 'undefined'
        ? localStorage
        : null;
  const fromLs = battleLayoutFromStorage(storage);
  if (fromLs !== null) return fromLs;

  if (opts.settingsFlag === undefined || opts.settingsFlag === null) {
    return BATTLE_LAYOUT_DEFAULT;
  }
  return normalizeBattleLayoutB(opts.settingsFlag);
}

/**
 * Persist explicit localStorage override (and clear when null).
 * @param {boolean|null} value
 * @param {Storage|null|undefined} [storage]
 */
export function writeBattleLayoutOverride(value, storage) {
  const store =
    storage !== undefined
      ? storage
      : typeof localStorage !== 'undefined'
        ? localStorage
        : null;
  if (!store) return;
  try {
    if (value === null) store.removeItem(BATTLE_LAYOUT_STORAGE_KEY);
    else store.setItem(BATTLE_LAYOUT_STORAGE_KEY, value ? '1' : '0');
  } catch (_) {
    /* ignore quota / private mode */
  }
}
