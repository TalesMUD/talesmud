<script>
  import { settingsStore } from '../SettingsStore.js';

  let activeTab = 'general';

  const tabs = [
    { id: 'general', label: 'General', icon: 'settings' },
    { id: 'interface', label: 'Interface', icon: 'palette' }
  ];

  function closeModal() {
    settingsStore.closeModal();
  }

  function handleBackdropClick(e) {
    if (e.target === e.currentTarget) {
      closeModal();
    }
  }

  function toggleSetting(category, key) {
    const currentValue = $settingsStore[category]?.[key];
    settingsStore.setSetting(category, key, !currentValue);
  }
</script>

<style>
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.8);
    backdrop-filter: blur(4px);
    -webkit-backdrop-filter: blur(4px);
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    animation: fadeIn 0.2s ease-out;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .modal-container {
    background: #1a1a1a;
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    width: 90%;
    max-width: 600px;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    animation: slideIn 0.2s ease-out;
  }

  @keyframes slideIn {
    from {
      opacity: 0;
      transform: translateY(-20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1em 1.25em;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .modal-title {
    font-size: 1.25em;
    font-weight: 600;
    color: #e5e7eb;
    display: flex;
    align-items: center;
    gap: 0.5em;
  }

  .modal-title i {
    font-size: 1.2em;
    color: #9ca3af;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: #9ca3af;
    cursor: pointer;
    padding: 0.5em;
    border-radius: 6px;
    transition: all 0.15s ease;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .close-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: #e5e7eb;
  }

  .modal-body {
    display: flex;
    flex: 1;
    overflow: hidden;
  }

  .tabs-sidebar {
    width: 160px;
    background: rgba(0, 0, 0, 0.3);
    border-right: 1px solid rgba(255, 255, 255, 0.1);
    padding: 0.75em;
    flex-shrink: 0;
  }

  .tab-btn {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 0.6em;
    padding: 0.75em 1em;
    background: transparent;
    border: none;
    border-radius: 8px;
    color: #9ca3af;
    font-size: 0.95em;
    cursor: pointer;
    transition: all 0.15s ease;
    text-align: left;
  }

  .tab-btn:hover {
    background: rgba(255, 255, 255, 0.05);
    color: #d1d5db;
  }

  .tab-btn.active {
    background: rgba(59, 130, 246, 0.2);
    color: #93c5fd;
  }

  .tab-btn i {
    font-size: 1.2em;
  }

  .tab-content {
    flex: 1;
    padding: 1.25em;
    overflow-y: auto;
  }

  .settings-section {
    margin-bottom: 1.5em;
  }

  .section-title {
    font-size: 0.85em;
    font-weight: 600;
    color: #6b7280;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 1em;
  }

  .setting-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.85em 1em;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 8px;
    margin-bottom: 0.5em;
  }

  .setting-info {
    flex: 1;
  }

  .setting-label {
    font-size: 1em;
    color: #e5e7eb;
    margin-bottom: 0.2em;
  }

  .setting-desc {
    font-size: 0.85em;
    color: #6b7280;
  }

  /* Toggle switch */
  .toggle-switch {
    position: relative;
    width: 44px;
    height: 24px;
    flex-shrink: 0;
    margin-left: 1em;
  }

  .toggle-switch input {
    opacity: 0;
    width: 0;
    height: 0;
  }

  .toggle-slider {
    position: absolute;
    cursor: pointer;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: #374151;
    transition: 0.2s;
    border-radius: 24px;
  }

  .toggle-slider:before {
    position: absolute;
    content: "";
    height: 18px;
    width: 18px;
    left: 3px;
    bottom: 3px;
    background-color: #9ca3af;
    transition: 0.2s;
    border-radius: 50%;
  }

  input:checked + .toggle-slider {
    background-color: rgba(59, 130, 246, 0.6);
  }

  input:checked + .toggle-slider:before {
    transform: translateX(20px);
    background-color: #93c5fd;
  }

  /* Responsive */
  @media screen and (max-width: 500px) {
    .modal-container {
      width: 95%;
      max-height: 90vh;
    }

    .tabs-sidebar {
      width: 56px;
      padding: 0.5em;
    }

    .tab-btn {
      padding: 0.75em;
      justify-content: center;
    }

    .tab-btn span {
      display: none;
    }

    .tab-btn i {
      margin: 0;
    }
  }
</style>

{#if $settingsStore.modalOpen}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="modal-backdrop" on:click={handleBackdropClick}>
    <div class="modal-container">
      <div class="modal-header">
        <div class="modal-title">
          <i class="material-icons">settings</i>
          Settings
        </div>
        <button class="close-btn" on:click={closeModal}>
          <i class="material-icons">close</i>
        </button>
      </div>

      <div class="modal-body">
        <div class="tabs-sidebar">
          {#each tabs as tab}
            <button
              class="tab-btn"
              class:active={activeTab === tab.id}
              on:click={() => activeTab = tab.id}
            >
              <i class="material-icons">{tab.icon}</i>
              <span>{tab.label}</span>
            </button>
          {/each}
        </div>

        <div class="tab-content">
          {#if activeTab === 'general'}
            <div class="settings-section">
              <div class="section-title">Audio</div>

              <div class="setting-item">
                <div class="setting-info">
                  <div class="setting-label">Sound Effects</div>
                  <div class="setting-desc">Enable game sound effects</div>
                </div>
                <label class="toggle-switch">
                  <input
                    type="checkbox"
                    checked={$settingsStore.general?.soundEnabled}
                    on:change={() => toggleSetting('general', 'soundEnabled')}
                  />
                  <span class="toggle-slider"></span>
                </label>
              </div>
            </div>
          {/if}

          {#if activeTab === 'interface'}
            <div class="settings-section">
              <div class="section-title">Room Display</div>

              <div class="setting-item">
                <div class="setting-info">
                  <div class="setting-label">Parchment Style</div>
                  <div class="setting-desc">Use textured parchment background for room descriptions</div>
                </div>
                <label class="toggle-switch">
                  <input
                    type="checkbox"
                    checked={$settingsStore.interface?.parchmentBackground}
                    on:change={() => toggleSetting('interface', 'parchmentBackground')}
                  />
                  <span class="toggle-slider"></span>
                </label>
              </div>

              <div class="setting-item">
                <div class="setting-info">
                  <div class="setting-label">Compact Mode</div>
                  <div class="setting-desc">Reduce padding and spacing for smaller screens</div>
                </div>
                <label class="toggle-switch">
                  <input
                    type="checkbox"
                    checked={$settingsStore.interface?.compactMode}
                    on:change={() => toggleSetting('interface', 'compactMode')}
                  />
                  <span class="toggle-slider"></span>
                </label>
              </div>
            </div>
          {/if}
        </div>
      </div>
    </div>
  </div>
{/if}
