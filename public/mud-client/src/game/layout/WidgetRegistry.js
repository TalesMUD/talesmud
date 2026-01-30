// Widget type definitions and configuration
// Components will be imported dynamically in WidgetGrid.svelte

export const WIDGET_TYPES = {
  room: {
    name: 'Room',
    description: 'Room image, name, description, and entities',
    defaultSize: { w: 12, h: 14 },
    minSize: { w: 6, h: 8 },
    maxInstances: 1,
    icon: 'landscape',
    category: 'core'
  },
  terminal: {
    name: 'Terminal',
    description: 'Classic MUD text terminal',
    defaultSize: { w: 12, h: 14 },
    minSize: { w: 6, h: 6 },
    maxInstances: 1,
    icon: 'terminal',
    category: 'core'
  },
  terminalx: {
    name: 'Terminal X',
    description: 'Veilspan-style MUD terminal with CRT effects',
    defaultSize: { w: 12, h: 14 },
    minSize: { w: 6, h: 6 },
    maxInstances: 1,
    icon: 'computer',
    category: 'core'
  },
  actionbar: {
    name: 'Action Bar',
    description: 'Commands, compass, and quick actions',
    defaultSize: { w: 24, h: 3 },
    minSize: { w: 12, h: 2 },
    maxInstances: 1,
    icon: 'gamepad',
    category: 'core'
  },
  character: {
    name: 'Character',
    description: 'Player stats and character sheet',
    defaultSize: { w: 6, h: 8 },
    minSize: { w: 4, h: 4 },
    maxInstances: 1,
    icon: 'person',
    category: 'player'
  },
  inventory: {
    name: 'Inventory',
    description: 'Player inventory items',
    defaultSize: { w: 6, h: 8 },
    minSize: { w: 4, h: 4 },
    maxInstances: 1,
    icon: 'inventory_2',
    category: 'player'
  },
  equipment: {
    name: 'Equipment',
    description: 'Worn and equipped items',
    defaultSize: { w: 6, h: 8 },
    minSize: { w: 4, h: 4 },
    maxInstances: 1,
    icon: 'shield',
    category: 'player'
  }
};

// Get widget configuration by type
export function getWidgetConfig(widgetType) {
  return WIDGET_TYPES[widgetType] || null;
}

// Get default size for a widget type
export function getDefaultSize(widgetType) {
  const config = WIDGET_TYPES[widgetType];
  return config?.defaultSize || { w: 6, h: 6 };
}

// Get minimum size for a widget type
export function getMinSize(widgetType) {
  const config = WIDGET_TYPES[widgetType];
  return config?.minSize || { w: 2, h: 2 };
}

// Get all widget types as array for UI rendering
export function getWidgetTypeList() {
  return Object.entries(WIDGET_TYPES).map(([key, config]) => ({
    type: key,
    ...config
  }));
}

// Get widget types by category
export function getWidgetsByCategory(category) {
  return Object.entries(WIDGET_TYPES)
    .filter(([_, config]) => config.category === category)
    .map(([key, config]) => ({
      type: key,
      ...config
    }));
}

// Categories for grouping in UI
export const WIDGET_CATEGORIES = {
  core: {
    name: 'Core',
    description: 'Essential game interface widgets'
  },
  player: {
    name: 'Player',
    description: 'Character and inventory widgets'
  }
};
