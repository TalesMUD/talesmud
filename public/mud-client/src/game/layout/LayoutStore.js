import { writable, get } from 'svelte/store';

const STORAGE_KEY = 'talesmud_layout_v1';

// Default layout: Room + Terminal side by side, ActionBar at bottom
const DEFAULT_LAYOUT = [
  {
    id: 'room-1',
    widgetType: 'room',
    x: 0,
    y: 0,
    w: 12,
    h: 14,
    visible: true
  },
  {
    id: 'terminal-1',
    widgetType: 'terminal',
    x: 12,
    y: 0,
    w: 12,
    h: 14,
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

function createLayoutStore() {
  const { subscribe, set, update } = writable({
    widgets: toGridItems(DEFAULT_LAYOUT, false),
    editMode: false,
    pendingWidgets: null
  });

  return {
    subscribe,

    // Load layout from localStorage
    loadFromStorage() {
      try {
        const stored = localStorage.getItem(STORAGE_KEY);
        if (stored) {
          const data = JSON.parse(stored);
          if (data.version === 1 && Array.isArray(data.widgets)) {
            update(state => ({
              ...state,
              widgets: toGridItems(data.widgets, state.editMode)
            }));
            return true;
          }
        }
      } catch (e) {
        console.warn('Failed to load layout from storage:', e);
      }
      return false;
    },

    // Save current layout to localStorage
    saveToStorage() {
      const state = get({ subscribe });
      const data = {
        version: 1,
        savedAt: new Date().toISOString(),
        widgets: fromGridItems(state.widgets)
      };
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
            pendingWidgets: null
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
        widgets: toGridItems(DEFAULT_LAYOUT, state.editMode)
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
    }
  };
}

export const layoutStore = createLayoutStore();

// Initialize on load
if (typeof window !== 'undefined') {
  layoutStore.loadFromStorage();
}
