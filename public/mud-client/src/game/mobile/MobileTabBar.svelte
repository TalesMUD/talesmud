<style>
  .mobile-tab-bar {
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    height: calc(56px + env(safe-area-inset-bottom, 0px));
    padding-bottom: env(safe-area-inset-bottom, 0px);
    background: rgba(0, 0, 0, 0.92);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    display: flex;
    z-index: 1001;
  }

  .tab {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 2px;
    cursor: pointer;
    border: none;
    background: none;
    color: #6b7280;
    padding: 6px 0;
    margin: 0;
    transition: color 0.15s ease;
    position: relative;
    -webkit-tap-highlight-color: transparent;
    min-height: 44px;
  }

  .tab.active {
    color: #f59e0b;
  }

  .tab.active::after {
    content: '';
    position: absolute;
    top: 0;
    left: 25%;
    right: 25%;
    height: 2px;
    background: #f59e0b;
    border-radius: 0 0 2px 2px;
  }

  .tab i {
    font-size: 22px;
  }

  .tab span {
    font-size: 10px;
    font-weight: 500;
    letter-spacing: 0.5px;
    text-transform: uppercase;
  }
</style>

<script>
  import { mobileStore } from './mobileStore.js';

  const { activeTab } = mobileStore;
</script>

<div class="mobile-tab-bar">
  <button
    class="tab"
    class:active={$activeTab === 'room'}
    on:click={() => mobileStore.switchTab('room')}
  >
    <i class="material-icons">landscape</i>
    <span>Room</span>
  </button>

  <button
    class="tab"
    class:active={$activeTab === 'console'}
    on:click={() => mobileStore.switchTab('console')}
  >
    <i class="material-icons">terminal</i>
    <span>Console</span>
  </button>
</div>
