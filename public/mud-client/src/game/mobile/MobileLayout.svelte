<style>
  .mobile-layout {
    display: flex;
    flex-direction: column;
    height: 100vh;
    height: 100dvh;
    overflow: hidden;
    background: #000;
  }

  .mobile-content {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    padding-top: calc(48px + env(safe-area-inset-top, 0px));
    padding-bottom: calc(56px + env(safe-area-inset-bottom, 0px));
  }

  .tab-panel {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .tab-panel.hidden {
    display: none;
  }
</style>

<script>
  import { mobileStore } from './mobileStore.js';
  import MobileHeader from './MobileHeader.svelte';
  import MobileTabBar from './MobileTabBar.svelte';
  import MobileRoomView from './MobileRoomView.svelte';
  import MobileConsoleView from './MobileConsoleView.svelte';
  import BottomSheet from './BottomSheet.svelte';
  import CharacterWidget from '../widgets/CharacterWidget.svelte';
  import InventoryWidget from '../widgets/InventoryWidget.svelte';
  import EquipmentWidget from '../widgets/EquipmentWidget.svelte';

  export let store;
  export let sendMessage;
  export let onTerminalReady;
  export let onTerminalInput;

  const { activeTab, openSheet } = mobileStore;
</script>

<div class="mobile-layout">
  <MobileHeader {store} />

  <div class="mobile-content">
    <!-- Both tabs always mounted; inactive hidden via CSS to preserve terminal state -->
    <div class="tab-panel" class:hidden={$activeTab !== 'room'}>
      <MobileRoomView {store} {sendMessage} />
    </div>

    <div class="tab-panel" class:hidden={$activeTab !== 'console'}>
      <MobileConsoleView {onTerminalReady} onInput={onTerminalInput} />
    </div>
  </div>

  <MobileTabBar />

  {#if $openSheet === 'character'}
    <BottomSheet title="Character" icon="person" on:close={() => mobileStore.closeBottomSheet()}>
      <CharacterWidget {store} {sendMessage} />
    </BottomSheet>
  {/if}

  {#if $openSheet === 'inventory'}
    <BottomSheet title="Inventory" icon="inventory_2" on:close={() => mobileStore.closeBottomSheet()}>
      <InventoryWidget {store} {sendMessage} />
    </BottomSheet>
  {/if}

  {#if $openSheet === 'equipment'}
    <BottomSheet title="Equipment" icon="shield" on:close={() => mobileStore.closeBottomSheet()}>
      <EquipmentWidget {store} {sendMessage} />
    </BottomSheet>
  {/if}
</div>
