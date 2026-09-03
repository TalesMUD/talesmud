import { writable, get } from 'svelte/store';
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
} from './hudPrefs.js';

const STORAGE_KEY = 'talesmud_settings_v1';

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
    compactMode: false,
    roomTextOverlay: false,      // Show game text overlay on room image (always on for mobile)
    actionBarPins: [...DEFAULT_ACTION_BAR_PINS],
    actionBarLayoutRevision: ACTION_BAR_LAYOUT_REVISION,
    inventoryOpenMode: DEFAULT_INVENTORY_OPEN_MODE, // 'overlay' | 'widget'
    hotbarBinds: [...DEFAULT_HOTBAR_BINDS],
  }
};

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
            iface.hotbarBinds = scrubLegacySearchBinds(
              normalizeHotbarBinds(iface.hotbarBinds)
            );
            update(state => ({
              ...state,
              general: { ...DEFAULT_SETTINGS.general, ...data.general },
              interface: iface
            }));
            this.applyTheme(iface.theme);
            // Persist migration so pins stay clean on next load
            if (prevRev < ACTION_BAR_LAYOUT_REVISION) {
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
          hotbarBinds: scrubLegacySearchBinds(
            normalizeHotbarBinds(state.interface.hotbarBinds)
          ),
        }
      };
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
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
      this.saveToStorage();
    }
  };
}

export const settingsStore = createSettingsStore();
settingsStore.loadFromStorage();
export default settingsStore;
