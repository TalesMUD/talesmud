import { writable, get } from 'svelte/store';
import {
  LAYOUT_STORAGE_KEY,
  parseLayoutStorage,
  buildLayoutStoragePayload,
  normalizeTemplateName,
  makeTemplateId,
  widgetsEqual,
  normalizeTemplates,
} from './layoutTemplates.js';

const STORAGE_KEY = LAYOUT_STORAGE_KEY;

// Default layout: Room + Terminal, Spell Bar between room and Action Bar
const DEFAULT_LAYOUT = [
  {
    id: 'room-1',
    widgetType: 'room',
    x: 0,
    y: 0,
    w: 12,
    h: 12,
    visible: true
  },
  {
    id: 'terminal-1',
    widgetType: 'terminal',
    x: 12,
    y: 0,
    w: 12,
    h: 12,
    visible: true
  },
  {
    id: 'hotbar-1',
    widgetType: 'hotbar',
    x: 0,
    y: 12,
    w: 24,
    h: 2,
    visible: true
  },
  {
    id: 'actionbar-1',
    widgetType: 'actionbar',
    x: 0,
    y: 14,
    w: 24,
    h: 3,
    visible: true
  }
];

/** If a saved layout has no hotbar, insert one above the action bar. */
function ensureHotbarInLayout(widgets) {
  if (!Array.isArray(widgets) || widgets.some((w) => w.widgetType === 'hotbar')) {
    return widgets;
  }
  const actionbar = widgets.find((w) => w.widgetType === 'actionbar');
  const hotbarY = actionbar ? actionbar.y : 12;
  const next = widgets.map((w) => {
    if (w.widgetType === 'actionbar') {
      return { ...w, y: hotbarY + 2 };
    }
    if (
      (w.widgetType === 'room' || w.widgetType === 'terminal') &&
      w.y + w.h > hotbarY
    ) {
      return { ...w, h: Math.max(2, hotbarY - w.y) };
    }
    return w;
  });
  next.push({
    id: 'hotbar-1',
    widgetType: 'hotbar',
    x: 0,
    y: hotbarY,
    w: 24,
    h: 2,
    visible: true
  });
  return next;
}

// Convert layout items to svelte-grid format (with editing disabled by default)
function toGridItems(widgets, editable = false) {
  return widgets.map(widget => ({
    ...widget,
    [24]: {
      x: widget.x,
      y: widget.y,
      w: widget.w,
      h: widget.h,
      draggable: editable,
      resizable: editable,
      customResizer: editable,
      min: { w: 2, h: 2 },
      max: {}
    }
  }));
}

// Convert svelte-grid items back to our format
function fromGridItems(items) {
  return items.map(item => {
    const base = {
      id: item.id,
      widgetType: item.widgetType,
      x: item[24]?.x ?? item.x,
      y: item[24]?.y ?? item.y,
      w: item[24]?.w ?? item.w,
      h: item[24]?.h ?? item.h,
      visible: item.visible ?? true
    };
    // Preserve tab container data
    if (item.widgetType === 'tabcontainer') {
      base.tabs = item.tabs || [];
      base.activeTabIndex = item.activeTabIndex || 0;
    }
    return base;
  });
}

// Set draggable/resizable on all widgets
function setWidgetsEditable(widgets, editable) {
  return widgets.map(widget => ({
    ...widget,
    [24]: {
      ...widget[24],
      draggable: editable,
      resizable: editable
    }
  }));
}

function bumpEpoch(state) {
  return (state.layoutEpoch || 0) + 1;
}

function createLayoutStore() {
  const { subscribe, set, update } = writable({
    widgets: toGridItems(DEFAULT_LAYOUT, false),
    editMode: false,
    pendingWidgets: null,
    templates: [],
    activeTemplateId: null,
    /** Bumps when widgets must force-sync into WidgetGrid (apply/reset/load). */
    layoutEpoch: 0,
  });

  const api = {
    subscribe,

    // Load layout from localStorage
    loadFromStorage() {
      try {
        const stored = localStorage.getItem(STORAGE_KEY);
        const parsed = parseLayoutStorage(stored);
        if (parsed) {
          const widgets = ensureHotbarInLayout(parsed.widgets);
          update(state => ({
            ...state,
            widgets: toGridItems(widgets, state.editMode),
            templates: parsed.templates,
            activeTemplateId: parsed.activeTemplateId,
            layoutEpoch: bumpEpoch(state),
          }));
          return true;
        }
      } catch (e) {
        console.warn('Failed to load layout from storage:', e);
      }
      return false;
    },

    // Save current layout (+ templates metadata) to localStorage
    saveToStorage() {
      const state = get({ subscribe });
      const data = buildLayoutStoragePayload({
        widgets: fromGridItems(state.widgets),
        templates: state.templates,
        activeTemplateId: state.activeTemplateId,
      });
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
        return true;
      } catch (e) {
        console.error('Failed to save layout:', e);
        return false;
      }
    },

    // Enter edit mode - enable dragging/resizing
    enterEditMode() {
      update(state => ({
        ...state,
        editMode: true,
        widgets: setWidgetsEditable(state.widgets, true),
        pendingWidgets: JSON.parse(JSON.stringify(state.widgets))
      }));
    },

    // Exit edit mode - disable dragging/resizing
    exitEditMode(save = true) {
      update(state => {
        if (save) {
          // Keep current widgets, disable editing
          return {
            ...state,
            editMode: false,
            widgets: setWidgetsEditable(state.widgets, false),
            pendingWidgets: null
          };
        } else {
          // Restore from pending, disable editing
          const restored = state.pendingWidgets || state.widgets;
          return {
            ...state,
            editMode: false,
            widgets: setWidgetsEditable(restored, false),
            pendingWidgets: null,
            layoutEpoch: bumpEpoch(state),
          };
        }
      });

      if (save) {
        this.saveToStorage();
      }
    },

    // Update widget position/size
    updateWidgets(newWidgets) {
      update(state => ({
        ...state,
        widgets: newWidgets
      }));
    },

    // Update a single widget
    updateWidget(id, changes) {
      update(state => ({
        ...state,
        widgets: state.widgets.map(w =>
          w.id === id ? { ...w, ...changes } : w
        )
      }));
    },

    // Add a new widget (only in edit mode, so editable=true)
    addWidget(widgetType, config = {}) {
      const id = `${widgetType}-${Date.now()}`;
      const newWidget = {
        id,
        widgetType,
        x: 0,
        y: 0,
        w: config.defaultSize?.w || 6,
        h: config.defaultSize?.h || 6,
        visible: true,
        // Initialize tab container data
        ...(widgetType === 'tabcontainer' ? { tabs: [], activeTabIndex: 0 } : {}),
        [24]: {
          x: 0,
          y: 0,
          w: config.defaultSize?.w || 6,
          h: config.defaultSize?.h || 6,
          draggable: true,
          resizable: true,
          customResizer: true,
          min: { w: 2, h: 2 },
          max: {}
        }
      };

      update(state => {
        // Find first available position (simple: place at bottom)
        const maxY = Math.max(...state.widgets.map(w => (w[24]?.y || w.y) + (w[24]?.h || w.h)), 0);
        newWidget.y = maxY;
        newWidget[24].y = maxY;

        return {
          ...state,
          widgets: [...state.widgets, newWidget]
        };
      });

      return id;
    },

    // Remove a widget
    removeWidget(id) {
      update(state => ({
        ...state,
        widgets: state.widgets.filter(w => w.id !== id)
      }));
    },

    // Reset to default layout (in edit mode, so editable=true)
    resetToDefault() {
      update(state => ({
        ...state,
        widgets: toGridItems(DEFAULT_LAYOUT, state.editMode),
        activeTemplateId: null,
        layoutEpoch: bumpEpoch(state),
      }));
    },

    // Get widget by id
    getWidget(id) {
      const state = get({ subscribe });
      return state.widgets.find(w => w.id === id);
    },

    // Check if widget type can be added (respects maxInstances)
    // Counts both top-level widgets and widgets inside tab containers
    canAddWidget(widgetType, registry) {
      const state = get({ subscribe });
      const config = registry[widgetType];
      if (!config) return false;

      // Count top-level widgets of this type
      let currentCount = state.widgets.filter(w => w.widgetType === widgetType).length;

      // Also count widgets inside tab containers
      for (const w of state.widgets) {
        if (w.widgetType === 'tabcontainer' && w.tabs) {
          currentCount += w.tabs.filter(t => t.widgetType === widgetType).length;
        }
      }

      return currentCount < (config.maxInstances || Infinity);
    },

    // Add a widget tab to a tab container
    addTabToContainer(containerId, widgetType) {
      const tabId = `tab-${widgetType}-${Date.now()}`;
      update(state => ({
        ...state,
        widgets: state.widgets.map(w => {
          if (w.id === containerId && w.widgetType === 'tabcontainer') {
            const newTabs = [...(w.tabs || []), { widgetType, id: tabId }];
            return { ...w, tabs: newTabs, activeTabIndex: newTabs.length - 1 };
          }
          return w;
        })
      }));
      return tabId;
    },

    // Remove a widget tab from a tab container
    removeTabFromContainer(containerId, tabIndex) {
      update(state => ({
        ...state,
        widgets: state.widgets.map(w => {
          if (w.id === containerId && w.widgetType === 'tabcontainer') {
            const newTabs = w.tabs.filter((_, i) => i !== tabIndex);
            const newActive = Math.min(w.activeTabIndex || 0, Math.max(0, newTabs.length - 1));
            return { ...w, tabs: newTabs, activeTabIndex: newActive };
          }
          return w;
        })
      }));
    },

    // Set active tab index on a tab container
    setActiveTab(containerId, tabIndex) {
      update(state => ({
        ...state,
        widgets: state.widgets.map(w => {
          if (w.id === containerId && w.widgetType === 'tabcontainer') {
            return { ...w, activeTabIndex: tabIndex };
          }
          return w;
        })
      }));
    },

    /**
     * Snapshot current layout as a named personal template.
     * Same name (case-insensitive) overwrites that template.
     * @returns {{ id: string, name: string }|null}
     */
    saveAsTemplate(name) {
      const trimmed = normalizeTemplateName(name);
      if (!trimmed) return null;

      const state = get({ subscribe });
      const snapshot = fromGridItems(state.widgets);
      const existing = state.templates.find(
        (t) => t.name.toLowerCase() === trimmed.toLowerCase()
      );
      const id = existing?.id || makeTemplateId();
      const entry = {
        id,
        name: trimmed,
        savedAt: new Date().toISOString(),
        widgets: snapshot,
        // shareId reserved for future sharing — omit in v1
      };

      update((s) => {
        const templates = existing
          ? s.templates.map((t) => (t.id === id ? entry : t))
          : [...s.templates, entry];
        return {
          ...s,
          templates: normalizeTemplates(templates),
          activeTemplateId: id,
        };
      });

      this.saveToStorage();
      return { id, name: trimmed };
    },

    /**
     * Apply a named template as the active layout (persists immediately).
     * Stays in edit mode if already editing; Cancel baseline updates to applied layout.
     */
    applyTemplate(templateId) {
      const state = get({ subscribe });
      const tpl = state.templates.find((t) => t.id === templateId);
      if (!tpl) return false;

      const widgets = ensureHotbarInLayout(
        JSON.parse(JSON.stringify(tpl.widgets))
      );
      const gridItems = toGridItems(widgets, state.editMode);

      update((s) => ({
        ...s,
        widgets: gridItems,
        activeTemplateId: tpl.id,
        pendingWidgets: s.editMode
          ? JSON.parse(JSON.stringify(gridItems))
          : s.pendingWidgets,
        layoutEpoch: bumpEpoch(s),
      }));

      this.saveToStorage();
      return true;
    },

    renameTemplate(templateId, newName) {
      const trimmed = normalizeTemplateName(newName);
      if (!trimmed) return false;
      let ok = false;
      update((s) => {
        const templates = s.templates.map((t) => {
          if (t.id !== templateId) return t;
          ok = true;
          return { ...t, name: trimmed };
        });
        return { ...s, templates };
      });
      if (ok) this.saveToStorage();
      return ok;
    },

    deleteTemplate(templateId) {
      let ok = false;
      update((s) => {
        const next = s.templates.filter((t) => t.id !== templateId);
        if (next.length === s.templates.length) return s;
        ok = true;
        return {
          ...s,
          templates: next,
          activeTemplateId:
            s.activeTemplateId === templateId ? null : s.activeTemplateId,
        };
      });
      if (ok) this.saveToStorage();
      return ok;
    },

    /** True when current widgets differ from the active template snapshot. */
    isActiveTemplateDirty() {
      const state = get({ subscribe });
      if (!state.activeTemplateId) return false;
      const tpl = state.templates.find((t) => t.id === state.activeTemplateId);
      if (!tpl) return false;
      return !widgetsEqual(fromGridItems(state.widgets), tpl.widgets);
    },

    getActiveTemplate() {
      const state = get({ subscribe });
      if (!state.activeTemplateId) return null;
      return state.templates.find((t) => t.id === state.activeTemplateId) || null;
    },
  };

  return api;
}

export const layoutStore = createLayoutStore();

// Initialize on load
if (typeof window !== 'undefined') {
  layoutStore.loadFromStorage();
}
