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
  import { onMount, tick } from 'svelte';
  import { mobileStore } from './mobileStore.js';
  import MobileHeader from './MobileHeader.svelte';
  import MobileTabBar from './MobileTabBar.svelte';
  import MobileRoomView from './MobileRoomView.svelte';
  import MobileConsoleView from './MobileConsoleView.svelte';
  import BottomSheet from './BottomSheet.svelte';
  import CharacterWidget from '../widgets/CharacterWidget.svelte';
  import InventoryWidget from '../widgets/InventoryWidget.svelte';
  import EquipmentWidget from '../widgets/EquipmentWidget.svelte';
  import QuestLogWidget from '../widgets/QuestLogWidget.svelte';
  import LeaderboardPanel from '../ui/LeaderboardPanel.svelte';

  export let store;
  export let sendMessage;
  export let onTerminalReady;
  export let onTerminalInput;
  export let authToken = "";

  const { activeTab, openSheet } = mobileStore;
  let layoutEl;
  let sheetTop = 48;
  let sheetBottom = 56;

  // The sheet sits between the header and the hotbar so those controls stay hittable.
  function measureChrome() {
    if (!layoutEl) return;
    const header = layoutEl.querySelector('.mobile-header');
    const tab = layoutEl.querySelector('.mobile-tab-bar');
    const hotbar = layoutEl.querySelector('.tab-panel:not(.hidden) .hotbar-widget');
    const headerBottom = header ? Math.round(header.getBoundingClientRect().bottom) : 48;
    const tabTop = tab ? Math.round(tab.getBoundingClientRect().top) : window.innerHeight - 56;
    let limit = tabTop;
    if (hotbar) {
      const rect = hotbar.getBoundingClientRect();
      if (rect.height > 8 && rect.top > headerBottom && rect.top < tabTop) limit = Math.round(rect.top);
    }
    sheetTop = headerBottom;
    sheetBottom = Math.max(0, window.innerHeight - limit);
  }

  onMount(() => {
    measureChrome();
    const observer = new ResizeObserver(() => measureChrome());
    observer.observe(layoutEl);
    window.addEventListener('resize', measureChrome);
    layoutEl.addEventListener('scroll', measureChrome, true);
    return () => {
      observer.disconnect();
      window.removeEventListener('resize', measureChrome);
      layoutEl.removeEventListener('scroll', measureChrome, true);
    };
  });

  $: if ($openSheet || $activeTab) tick().then(measureChrome);
</script>

<div class="mobile-layout" bind:this={layoutEl} style="--sheet-top: {sheetTop}px; --sheet-bottom: {sheetBottom}px;">
  <MobileHeader {store} {authToken} />

  <div class="mobile-content">
    <!-- Both tabs always mounted; inactive hidden via CSS to preserve terminal state -->
    <div class="tab-panel" class:hidden={$activeTab !== 'room'}>
      <MobileRoomView {store} {sendMessage} active={$activeTab === 'room'} />
    </div>

    <div class="tab-panel" class:hidden={$activeTab !== 'console'}>
      <MobileConsoleView {onTerminalReady} onInput={onTerminalInput} />
    </div>
  </div>

  <MobileTabBar />

  {#if $openSheet === 'character'}
    <BottomSheet title="Character" icon="person" roomy={true} on:close={() => mobileStore.closeBottomSheet()}>
      <CharacterWidget {store} {sendMessage} embedded={true} initialTab="stats" />
    </BottomSheet>
  {/if}

  {#if $openSheet === 'skills'}
    <BottomSheet title="Skills" icon="auto_awesome" roomy={true} on:close={() => mobileStore.closeBottomSheet()}>
      <CharacterWidget {store} {sendMessage} embedded={true} initialTab="skills" />
    </BottomSheet>
  {/if}

  {#if $openSheet === 'inventory'}
    <BottomSheet title="Inventory" icon="inventory_2" on:close={() => mobileStore.closeBottomSheet()}>
      <InventoryWidget {store} {sendMessage} embedded={true} />
    </BottomSheet>
  {/if}

  {#if $openSheet === 'equipment'}
    <BottomSheet title="Equipment" icon="shield" roomy={true} on:close={() => mobileStore.closeBottomSheet()}>
      <EquipmentWidget {store} {sendMessage} embedded={true} />
    </BottomSheet>
  {/if}

  {#if $openSheet === 'quests'}
    <BottomSheet title="Quests" icon="assignment" roomy={true} on:close={() => mobileStore.closeBottomSheet()}>
      <QuestLogWidget {store} {sendMessage} embedded={true} />
    </BottomSheet>
  {/if}

  {#if $openSheet === 'rank'}
    <BottomSheet title="Leaderboard" icon="format_list_numbered" roomy={true} on:close={() => mobileStore.closeBottomSheet()}>
      <LeaderboardPanel {store} embedded={true} />
    </BottomSheet>
  {/if}
</div>
