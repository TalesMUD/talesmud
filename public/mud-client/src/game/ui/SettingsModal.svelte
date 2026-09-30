<script>
  import { settingsStore } from '../SettingsStore.js';

  let activeTab = 'interface';
  const tabs = [
    { id: 'interface', label: 'Interface', icon: 'tune' },
    { id: 'general', label: 'General', icon: 'settings' },
  ];
  const set = (key, value) => settingsStore.setSetting('interface', key, value);
  const toggle = (key) => set(key, !$settingsStore.interface?.[key]);
  function closeModal() { settingsStore.closeModal(); }
  function handleBackdropClick(event) {
    if (event.target === event.currentTarget) closeModal();
  }
</script>

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    z-index: 1100;
    display: grid;
    place-items: center;
    padding: 1rem;
    background: rgba(5, 4, 3, 0.82);
    backdrop-filter: blur(5px);
    box-sizing: border-box;
  }
  .modal-container {
    width: min(100%, 720px);
    max-height: calc(100dvh - 2rem);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border: 1px solid #9b7435;
    border-radius: 10px;
    background: linear-gradient(150deg, #211b13, #12110e 65%);
    color: #e8dfca;
    box-shadow: 0 20px 80px #000c, inset 0 0 0 1px #e8c87822;
  }
  .modal-header {
    flex: none;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.8rem 1.1rem;
    border-bottom: 1px solid #8b692f88;
    background: linear-gradient(90deg, #392a16, #1b1813);
  }
  .modal-title { display: flex; align-items: center; gap: 0.5rem; color: #f5d78c; font: 700 1.25rem Georgia, serif; letter-spacing: 0.04em; }
  .modal-title i { font-size: 1.2rem; }
  .close-btn { display: grid; place-items: center; border: 1px solid #8b692f88; border-radius: 5px; padding: 0.3rem; background: #1c1812; color: #e8d5a7; cursor: pointer; }
  .close-btn:hover, .close-btn:focus-visible { border-color: #f5d78c; color: #fff2c8; }
  .modal-body { display: flex; min-height: 0; }
  .tabs-sidebar { flex: 0 0 145px; padding: 0.7rem; border-right: 1px solid #8b692f55; background: #100f0c88; }
  .tab-btn { width: 100%; display: flex; align-items: center; gap: 0.55rem; padding: 0.65rem; margin-bottom: 0.25rem; border: 1px solid transparent; border-radius: 5px; background: transparent; color: #bdb19b; text-align: left; cursor: pointer; }
  .tab-btn:hover, .tab-btn:focus-visible { color: #f5d78c; background: #d4a44a16; }
  .tab-btn.active { color: #f5d78c; border-color: #d4a44a77; background: #d4a44a25; }
  .tab-btn i { font-size: 1.15rem; }
  .tab-content { flex: 1; min-width: 0; overflow-y: auto; overscroll-behavior: contain; padding: 0.9rem 1.1rem 1.1rem; }
  .settings-section { margin-bottom: 1rem; }
  .section-title { margin: 0 0 0.5rem; color: #d4a44a; font: 700 0.78rem system-ui, sans-serif; letter-spacing: 0.12em; text-transform: uppercase; }
  .setting-item { display: flex; align-items: center; justify-content: space-between; gap: 0.8rem; padding: 0.65rem 0.8rem; margin-bottom: 0.4rem; border: 1px solid #ae8b5033; border-radius: 6px; background: #ffffff08; }
  .setting-info { min-width: 0; }
  .setting-label { color: #eee4ce; font-size: 0.92rem; font-weight: 600; }
  .setting-desc { color: #b6aa92; font-size: 0.78rem; line-height: 1.35; }
  .setting-desc { margin-top: 0.12rem; }
  .toggle-switch { position: relative; flex: none; width: 44px; height: 24px; cursor: pointer; }
  .toggle-switch input { position: absolute; inset: 0; opacity: 0; cursor: pointer; }
  .toggle-slider { display: block; width: 100%; height: 100%; border-radius: 20px; background: #51493d; box-shadow: inset 0 0 0 1px #c8b17a55; }
  .toggle-slider:before { content: ''; position: absolute; top: 3px; left: 3px; width: 18px; height: 18px; border-radius: 50%; background: #c4bba8; transition: transform 0.15s ease; }
  input:checked + .toggle-slider { background: #9a7131; box-shadow: inset 0 0 0 1px #e8c878; }
  input:checked + .toggle-slider:before { transform: translateX(20px); background: #fff0bb; }
  input:focus-visible + .toggle-slider { outline: 2px solid #f5d78c; outline-offset: 2px; }
  .choice-group { display: flex; gap: 0.35rem; flex-wrap: wrap; justify-content: flex-end; }
  .choice { border: 1px solid #8b692f88; border-radius: 5px; padding: 0.35rem 0.55rem; background: #18150f; color: #c5b89b; font-size: 0.8rem; cursor: pointer; white-space: nowrap; }
  .choice:hover, .choice:focus-visible { border-color: #e8c878; color: #fff0bb; }
  .choice.active { border-color: #d4a44a; background: #66491f; color: #fff0bb; }
  .theme-options { display: flex; gap: 0.5rem; }
  .theme-option { flex: 1; padding: 0.55rem; border: 1px solid #8b692f88; border-radius: 6px; background: #17140f; color: #c5b89b; cursor: pointer; text-align: left; }
  .theme-option.active { border-color: #d4a44a; color: #fff0bb; box-shadow: inset 0 0 0 1px #d4a44a55; }
  .theme-preview { display: block; height: 20px; margin-bottom: 0.35rem; border-radius: 3px; }
  .theme-preview.dark-fantasy { background: linear-gradient(90deg, #0e0b08, #8e6126); }
  .theme-preview.clean-hud { background: linear-gradient(90deg, #0c101b, #495671); }
  .note { padding: 0.8rem; border: 1px solid #ae8b5033; border-radius: 6px; background: #ffffff08; }
  @media (max-width: 600px) {
    .modal-backdrop { padding: 0.5rem; align-items: center; }
    .modal-container { max-height: calc(100dvh - 1rem); }
    .modal-header { padding: 0.6rem 0.8rem; }
    .modal-body { flex-direction: column; }
    .tabs-sidebar { flex: none; display: flex; gap: 0.4rem; padding: 0.45rem; border-right: 0; border-bottom: 1px solid #8b692f55; }
    .tab-btn { width: auto; margin: 0; padding: 0.45rem 0.65rem; }
    .tab-content { padding: 0.65rem; }
    .setting-item { align-items: flex-start; flex-wrap: wrap; }
    .choice-group { width: 100%; justify-content: flex-start; }
  }
</style>

{#if $settingsStore.modalOpen}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="modal-backdrop" on:click={handleBackdropClick}>
    <div class="modal-container" role="dialog" aria-modal="true" aria-label="Settings">
      <div class="modal-header">
        <div class="modal-title"><i class="material-icons" aria-hidden="true">settings</i> Settings</div>
        <button class="close-btn" aria-label="Close Settings" on:click={closeModal}><i class="material-icons" aria-hidden="true">close</i></button>
      </div>
      <div class="modal-body">
        <div class="tabs-sidebar" role="tablist" aria-label="Settings sections">
          {#each tabs as tab}
            <button class="tab-btn" class:active={activeTab === tab.id} role="tab" aria-selected={activeTab === tab.id} on:click={() => activeTab = tab.id}>
              <i class="material-icons" aria-hidden="true">{tab.icon}</i><span>{tab.label}</span>
            </button>
          {/each}
        </div>
        <div class="tab-content">
          {#if activeTab === 'interface'}
            <section class="settings-section">
              <h2 class="section-title">Appearance</h2>
              <div class="theme-options">
                <button class="theme-option" class:active={$settingsStore.interface?.theme === 'dark-fantasy'} aria-pressed={$settingsStore.interface?.theme === 'dark-fantasy'} on:click={() => set('theme', 'dark-fantasy')}><span class="theme-preview dark-fantasy"></span>Dark Fantasy</button>
                <button class="theme-option" class:active={$settingsStore.interface?.theme === 'clean-hud'} aria-pressed={$settingsStore.interface?.theme === 'clean-hud'} on:click={() => set('theme', 'clean-hud')}><span class="theme-preview clean-hud"></span>Clean HUD</button>
              </div>
            </section>
            <section class="settings-section">
              <h2 class="section-title">Accessibility</h2>
              <div class="setting-item">
                <div class="setting-info"><div class="setting-label">Reduced motion</div><div class="setting-desc">Control combat effects and map ambience.</div></div>
                <div class="choice-group" role="group" aria-label="Reduced motion">
                  {#each [{value:'system',label:'System'}, {value:'on',label:'On'}, {value:'off',label:'Off'}] as option}
                    <button class="choice" class:active={$settingsStore.interface?.reducedMotion === option.value} aria-pressed={$settingsStore.interface?.reducedMotion === option.value} on:click={() => set('reducedMotion', option.value)}>{option.label}</button>
                  {/each}
                </div>
              </div>
            </section>
            <section class="settings-section">
              <h2 class="section-title">Gameplay</h2>
              <div class="setting-item">
                <div class="setting-info"><div class="setting-label">Auto-focus BattleStage</div><div class="setting-desc">Open the combat view when a fight starts or you join one.</div></div>
                <label class="toggle-switch"><input type="checkbox" aria-label="Auto-focus BattleStage" checked={$settingsStore.interface?.combatAutoFocus !== false} on:change={() => toggle('combatAutoFocus')} /><span class="toggle-slider"></span></label>
              </div>
              <div class="setting-item">
                <div class="setting-info"><div class="setting-label">Inventory opens as</div><div class="setting-desc">Choose a full overlay or a layout widget.</div></div>
                <div class="choice-group" role="group" aria-label="Inventory opens as">
                  <button class="choice" class:active={$settingsStore.interface?.inventoryOpenMode === 'overlay'} aria-pressed={$settingsStore.interface?.inventoryOpenMode === 'overlay'} on:click={() => set('inventoryOpenMode', 'overlay')}>Overlay</button>
                  <button class="choice" class:active={$settingsStore.interface?.inventoryOpenMode === 'widget'} aria-pressed={$settingsStore.interface?.inventoryOpenMode === 'widget'} on:click={() => set('inventoryOpenMode', 'widget')}>Layout widget</button>
                </div>
              </div>
            </section>
            <section class="settings-section">
              <h2 class="section-title">Room display</h2>
              <div class="setting-item"><div class="setting-info"><div class="setting-label">Parchment style</div><div class="setting-desc">Textured room descriptions</div></div><label class="toggle-switch"><input type="checkbox" aria-label="Parchment style" checked={$settingsStore.interface?.parchmentBackground} on:change={() => toggle('parchmentBackground')} /><span class="toggle-slider"></span></label></div>
            </section>
          {:else}
            <section class="settings-section"><h2 class="section-title">Audio</h2><div class="note"><div class="setting-label">Game audio is coming soon</div><div class="setting-desc">There are no active sound effects or music controls yet.</div></div></section>
          {/if}
        </div>
      </div>
    </div>
  </div>
{/if}
