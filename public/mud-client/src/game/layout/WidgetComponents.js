// Shared widget component map and prop-building logic
// Used by both WidgetGrid.svelte and TabContainerWidget.svelte

import RoomWidget from '../widgets/RoomWidget.svelte';
import TerminalWidget from '../widgets/TerminalWidget.svelte';
import TerminalXWidget from '../widgets/TerminalXWidget.svelte';
import ActionBarWidget from '../widgets/ActionBarWidget.svelte';
import CharacterWidget from '../widgets/CharacterWidget.svelte';
import InventoryWidget from '../widgets/InventoryWidget.svelte';
import EquipmentWidget from '../widgets/EquipmentWidget.svelte';
import QuestLogWidget from '../widgets/QuestLogWidget.svelte';
import MinimapWidget from '../widgets/MinimapWidget.svelte';

// Map widget types to components (excludes tabcontainer to avoid circular imports)
export const childWidgetComponents = {
  room: RoomWidget,
  terminal: TerminalWidget,
  terminalx: TerminalXWidget,
  actionbar: ActionBarWidget,
  character: CharacterWidget,
  inventory: InventoryWidget,
  equipment: EquipmentWidget,
  questlog: QuestLogWidget,
  minimap: MinimapWidget
};

// Get props for a child widget type
// deps: { store, sendMessage, onTerminalReady, onTerminalInput }
export function getChildWidgetProps(widgetType, deps) {
  const { store, sendMessage, onTerminalReady, onTerminalInput } = deps;

  switch (widgetType) {
    case 'room':
      return { store, sendMessage };
    case 'terminal':
      return { onTerminalReady, onInput: onTerminalInput };
    case 'terminalx':
      return { onTerminalReady, onInput: onTerminalInput };
    case 'actionbar':
      return { store, sendMessage, term: null };
    case 'inventory':
      return { store, sendMessage };
    case 'equipment':
      return { store, sendMessage };
    case 'character':
      return { store, sendMessage };
    case 'minimap':
      return { store, sendMessage };
    case 'questlog':
      return { store, sendMessage };
    default:
      return {};
  }
}
