<script>
  import { onDestroy } from 'svelte';
  import { get } from 'svelte/store';
  import Grid from 'svelte-grid';
  import gridHelp from 'svelte-grid/build/helper/index.mjs';
  import { layoutStore } from './LayoutStore.js';
  import { WIDGET_TYPES, getMinSize } from './WidgetRegistry.js';
  import WidgetWrapper from './WidgetWrapper.svelte';

  // Widget components
  import RoomWidget from '../widgets/RoomWidget.svelte';
  import TerminalWidget from '../widgets/TerminalWidget.svelte';
  import TerminalXWidget from '../widgets/TerminalXWidget.svelte';
  import ActionBarWidget from '../widgets/ActionBarWidget.svelte';
  import CharacterWidget from '../widgets/CharacterWidget.svelte';
  import InventoryWidget from '../widgets/InventoryWidget.svelte';
  import EquipmentWidget from '../widgets/EquipmentWidget.svelte';

  export let store;
  export let sendMessage;
  export let onTerminalReady = () => {};
  export let onTerminalInput = () => {};

  // Map widget types to components
  const widgetComponents = {
    room: RoomWidget,
    terminal: TerminalWidget,
    terminalx: TerminalXWidget,
    actionbar: ActionBarWidget,
    character: CharacterWidget,
    inventory: InventoryWidget,
    equipment: EquipmentWidget
  };

  // Grid configuration
  const COLS = 24;
  const ROW_HEIGHT = 40;
  const GAP = 8;

  // Get initial state without creating reactive dependency
  let initialState = get(layoutStore);
  let widgets = initialState.widgets;
  let editMode = initialState.editMode;
  let widgetCount = widgets.length;

  // Subscribe manually to react to editMode changes and widget additions/removals
  // This avoids the reactive $: block which causes issues with svelte-grid binding
  const unsubscribe = layoutStore.subscribe(state => {
    const storeWidgetCount = state.widgets.length;

    // Sync when edit mode changes (updates draggable/resizable flags)
    // or when widgets are added/removed (count changed)
    if (state.editMode !== editMode || storeWidgetCount !== widgetCount) {
      editMode = state.editMode;
      widgets = state.widgets;
      widgetCount = storeWidgetCount;
    }
    // When widgets are updated via handleChange (move/resize), we do NOT sync back
    // because that would overwrite svelte-grid's changes
  });

  onDestroy(unsubscribe);

  // Handle widget changes from svelte-grid - save to store
  // Use the event data to explicitly update our widgets array
  // This ensures the resize persists even with svelte-grid's timing issues
  function handleChange(e) {
    const { unsafeItem, id, cols } = e.detail;

    // Force update our widgets array with the changed item from the event
    // This is necessary because svelte-grid's binding may not propagate in time
    widgets = widgets.map(w => {
      if (w.id === id && unsafeItem && unsafeItem[cols]) {
        return {
          ...w,
          [cols]: {
            ...w[cols],
            x: unsafeItem[cols].x,
            y: unsafeItem[cols].y,
            w: unsafeItem[cols].w,
            h: unsafeItem[cols].h,
            // Preserve min/max constraints and customResizer
            min: w[cols]?.min || { w: 4, h: 3 },
            max: w[cols]?.max || { w: 24, h: 20 },
            customResizer: w[cols]?.customResizer ?? true
          }
        };
      }
      return w;
    });

    layoutStore.updateWidgets(widgets);
  }

  // Handle widget removal
  function handleRemove(e) {
    layoutStore.removeWidget(e.detail.id);
    // Manually update local widgets after removal
    widgets = widgets.filter(w => w.id !== e.detail.id);
  }

  // Get the component for a widget type
  function getComponent(widgetType) {
    return widgetComponents[widgetType] || null;
  }

  // Get props for each widget type
  function getWidgetProps(widget) {
    const baseProps = {};

    switch (widget.widgetType) {
      case 'room':
        return { store, sendMessage };
      case 'terminal':
        return {
          onTerminalReady,
          onInput: onTerminalInput
        };
      case 'terminalx':
        return {
          onTerminalReady,
          onInput: onTerminalInput
        };
      case 'actionbar':
        return { store, sendMessage, term: null };
      case 'inventory':
        return { store, sendMessage };
      default:
        return baseProps;
    }
  }
</script>

<style>
  .widget-grid-container {
    width: 100%;
    height: 100%;
    position: relative;
  }

  .widget-grid-container :global(.svlt-grid-container) {
    height: 100% !important;
  }

  .widget-grid-container :global(.svlt-grid-item) {
    transition: none;
  }

  /* Style the built-in resizer - make it more visible and ensure it's above overlay */
  .widget-grid-container :global(.svlt-grid-resizer) {
    width: 24px;
    height: 24px;
    z-index: 150;
  }

  .widget-grid-container :global(.svlt-grid-resizer::after) {
    right: 4px;
    bottom: 4px;
    width: 8px;
    height: 8px;
    border-right: 3px solid #f59e0b;
    border-bottom: 3px solid #f59e0b;
  }

  /* Ensure the active resize handle has pointer events */
  .widget-grid-container.edit-mode :global(.svlt-grid-resizer) {
    pointer-events: auto;
    cursor: se-resize;
  }

  /* Shadow/placeholder styling */
  .widget-grid-container :global(.svlt-grid-shadow) {
    background: rgba(245, 158, 11, 0.3) !important;
    border: 2px dashed #f59e0b;
    border-radius: 12px;
  }

  .grid-item-content {
    width: 100%;
    height: 100%;
    overflow: hidden;
    position: relative;
  }
</style>

<div class="widget-grid-container" class:edit-mode={editMode}>
  <Grid
    bind:items={widgets}
    rowHeight={ROW_HEIGHT}
    cols={[[COLS, COLS]]}
    gap={[GAP, GAP]}
    let:dataItem
    let:resizePointerDown
    on:change={handleChange}
    fastStart={true}
    sensor={20}
    throttleResize={999999}
  >
    <div class="grid-item-content">
      <WidgetWrapper
        widget={dataItem}
        {editMode}
        {resizePointerDown}
        on:remove={handleRemove}
      >
        {#if getComponent(dataItem.widgetType)}
          <svelte:component
            this={getComponent(dataItem.widgetType)}
            {...getWidgetProps(dataItem)}
          />
        {:else}
          <div>Unknown widget: {dataItem.widgetType}</div>
        {/if}
      </WidgetWrapper>
    </div>
  </Grid>
</div>
