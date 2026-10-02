import { writable, get } from 'svelte/store';
import {
  BATTLE_LAYOUT_STORAGE_KEY,
  battleLayoutSettingsDefault,
  normalizeBattleLayoutB,
  writeBattleLayoutOverride,
} from './battleLayout.js';
import {
  ACTION_BAR_LAYOUT_REVISION,
  DEFAULT_ACTION_BAR_PINS,
  DEFAULT_HOTBAR_BINDS,
  DEFAULT_INVENTORY_OPEN_MODE,
  migrateActionBarPins,
  normalizeActionBarPins,
  normalizeHotbarBinds,
  normalizeInventoryOpenMode,
  scrubLegacySearchBinds,
  seedRestOnEmptyHotbar,
  HOTBAR_BY_CHARACTER_STORAGE_KEY,
  parseHotbarByCharacter,
  reconcileHotbarForCharacter,
} from './hudPrefs.js';

const STORAGE_KEY = 'talesmud_settings_v1';

export function normalizeReducedMotion(value) {
  return value === 'on' || value === 'off' ? value : 'system';
}

const DEFAULT_SETTINGS = {
  // General settings
  general: {
    soundEnabled: true,
    musicVolume: 50,
    sfxVolume: 50
  },
  // Interface settings
  interface: {
    theme: 'dark-fantasy',       // UI theme: 'dark-fantasy' or 'clean-hud'
    parchmentBackground: false,  // Room description parchment style (default off)
    compactMode: false,          // Legacy stored field; no active UI consumer
    roomTextOverlay: false,      // Legacy stored field; no active UI consumer
    actionBarPins: [...DEFAULT_ACTION_BAR_PINS],
    actionBarLayoutRevision: ACTION_BAR_LAYOUT_REVISION,
    inventoryOpenMode: DEFAULT_INVENTORY_OPEN_MODE, // 'overlay' | 'widget'
    reducedMotion: 'system', // 'system' | 'on' | 'off'
    combatAutoFocus: true,
    battleLayoutB: true, // Layout B default; Classic is the Settings opt-out
    hotbarBinds: [...DEFAULT_HOTBAR_BINDS],
  }
};

let activeHotbarCharacterId = '';

function readHotbarCharacterMap() {
  if (typeof localStorage === 'undefined') return {};
  try {
    const raw = localStorage.getItem(HOTBAR_BY_CHARACTER_STORAGE_KEY);
    return parseHotbarByCharacter(raw ? JSON.parse(raw) : {});
  } catch (_) {
    return {};
  }
}

function writeHotbarCharacterMap(map) {
  if (typeof localStorage === 'undefined') return;
  try {
    localStorage.setItem(HOTBAR_BY_CHARACTER_STORAGE_KEY, JSON.stringify(map || {}));
  } catch (_) {
    /* quota / private mode */
  }
}

function createSettingsStore() {
  const { subscribe, set, update } = writable({
    ...DEFAULT_SETTINGS,
    modalOpen: false
  });

  return {
    subscribe,

    // Open settings modal
    openModal() {
      update(state => ({ ...state, modalOpen: true }));
    },

    // Close settings modal
    closeModal() {
      update(state => ({ ...state, modalOpen: false }));
    },

    // Load settings from localStorage
    loadFromStorage() {
      if (typeof localStorage === 'undefined') return false;
      try {
        const stored = localStorage.getItem(STORAGE_KEY);
        if (stored) {
          const data = JSON.parse(stored);
          if (data.version === 1) {
            const prevRev = Number(data.interface?.actionBarLayoutRevision) || 0;
            const iface = { ...DEFAULT_SETTINGS.interface, ...data.interface };
            iface.actionBarPins = migrateActionBarPins(iface.actionBarPins, prevRev);
            iface.actionBarLayoutRevision = ACTION_BAR_LAYOUT_REVISION;
            iface.inventoryOpenMode = normalizeInventoryOpenMode(iface.inventoryOpenMode);
            iface.reducedMotion = normalizeReducedMotion(iface.reducedMotion);
            iface.combatAutoFocus = iface.combatAutoFocus !== false;
            // No talesmud_battle_layout_b key → Layout B, even if this blob
            // still stores the old default false. Explicit 0/1 is the preference.
            let layoutRaw = null;
            try {
              layoutRaw = localStorage.getItem(BATTLE_LAYOUT_STORAGE_KEY);
            } catch (_) {
              layoutRaw = null;
            }
            iface.battleLayoutB = battleLayoutSettingsDefault(
              data.interface?.battleLayoutB,
              layoutRaw
            );
            const beforeSeed = scrubLegacySearchBinds(
              normalizeHotbarBinds(iface.hotbarBinds)
            );
            iface.hotbarBinds = seedRestOnEmptyHotbar(beforeSeed);
            update(state => ({
              ...state,
              general: { ...DEFAULT_SETTINGS.general, ...data.general },
              interface: iface
            }));
            this.applyTheme(iface.theme);
            const seededEmpty =
              beforeSeed.every((b) => b == null) &&
              iface.hotbarBinds.some((b) => b && b.id === 'rest');
            // Persist pin revision and empty-bar Rest seed so guests see Rest next load
            if (prevRev < ACTION_BAR_LAYOUT_REVISION || seededEmpty) {
              this.saveToStorage();
            }
            return true;
          }
        }
      } catch (e) {
        console.warn('Failed to load settings from storage:', e);
      }
      return false;
    },

    // Save settings to localStorage
    saveToStorage() {
      const state = get({ subscribe });
      const data = {
        version: 1,
        savedAt: new Date().toISOString(),
        general: state.general,
        interface: {
          ...state.interface,
          actionBarPins: normalizeActionBarPins(state.interface.actionBarPins),
          actionBarLayoutRevision: ACTION_BAR_LAYOUT_REVISION,
          inventoryOpenMode: normalizeInventoryOpenMode(state.interface.inventoryOpenMode),
          reducedMotion: normalizeReducedMotion(state.interface.reducedMotion),
          hotbarBinds: scrubLegacySearchBinds(
            normalizeHotbarBinds(state.interface.hotbarBinds)
          ),
        }
      };
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
        if (activeHotbarCharacterId) {
          const book = readHotbarCharacterMap();
          book[activeHotbarCharacterId] = data.interface.hotbarBinds;
          writeHotbarCharacterMap(book);
        }
        return true;
      } catch (e) {
        console.error('Failed to save settings:', e);
        return false;
      }
    },

    // Apply the current theme to the document body
    applyTheme(theme) {
      if (typeof document !== 'undefined') {
        document.body.dataset.theme = theme || 'dark-fantasy';
      }
    },

    // Update a specific setting
    setSetting(category, key, value) {
      update(state => {
        let nextValue = value;
        if (category === 'interface' && key === 'actionBarPins') {
          nextValue = normalizeActionBarPins(value);
        }
        if (category === 'interface' && key === 'inventoryOpenMode') {
          nextValue = normalizeInventoryOpenMode(value);
        }
        if (category === 'interface' && key === 'reducedMotion') {
          nextValue = normalizeReducedMotion(value);
        }
        if (category === 'interface' && key === 'combatAutoFocus') {
          nextValue = value !== false;
        }
        if (category === 'interface' && key === 'battleLayoutB') {
          nextValue = normalizeBattleLayoutB(value);
          // Explicit 1 or 0. Clearing the key would look like "no preference"
          // and snap back to the Layout B default.
          writeBattleLayoutOverride(nextValue ? true : false);
        }
        if (category === 'interface' && key === 'hotbarBinds') {
          nextValue = scrubLegacySearchBinds(normalizeHotbarBinds(value));
        }
        return {
          ...state,
          [category]: {
            ...state[category],
            [key]: nextValue
          }
        };
      });
      // Apply theme immediately when changed
      if (category === 'interface' && key === 'theme') {
        this.applyTheme(value);
      }
      this.saveToStorage();
    },

    // Get a specific setting value
    getSetting(category, key) {
      const state = get({ subscribe });
      return state[category]?.[key];
    },

    // Reset all settings to defaults
    resetToDefaults() {
      update(state => ({
        ...state,
        general: { ...DEFAULT_SETTINGS.general },
        interface: {
          ...DEFAULT_SETTINGS.interface,
          actionBarPins: [...DEFAULT_ACTION_BAR_PINS],
          actionBarLayoutRevision: ACTION_BAR_LAYOUT_REVISION,
          inventoryOpenMode: DEFAULT_INVENTORY_OPEN_MODE,
          hotbarBinds: [...DEFAULT_HOTBAR_BINDS],
        }
      }));
      writeBattleLayoutOverride(null);
      this.saveToStorage();
    },

    /**
     * On login / character switch: load this character's bar and drop skills
     * that are not equipped or available for their class.
     */
    syncHotbarForCharacter({ characterId, classId, level, equippedIds } = {}) {
      const state = get({ subscribe });
      const result = reconcileHotbarForCharacter({
        activeCharacterId: activeHotbarCharacterId,
        map: readHotbarCharacterMap(),
        activeBinds: state.interface?.hotbarBinds,
        characterId,
        classId,
        level,
        equippedIds,
      });
      activeHotbarCharacterId = result.activeCharacterId;
      writeHotbarCharacterMap(result.map);
      if (result.changed) {
        this.setSetting('interface', 'hotbarBinds', result.binds);
      }
    }
  };
}

export const settingsStore = createSettingsStore();
settingsStore.loadFromStorage();
export default settingsStore;
