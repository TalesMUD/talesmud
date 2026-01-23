<script>
  import { createEventDispatcher } from "svelte";
  import { isCardinalDirection } from "./WorldEditorStore.js";

  export let open = false;
  export let exits = [];
  export let roomsValueHelp = [];

  const dispatch = createEventDispatcher();

  // Local copy for editing
  let localExits = [];

  // Sync local exits when modal opens
  $: if (open) {
    localExits = exits
      .filter(e => !isCardinalDirection(e.name || ""))
      .map(e => ({ ...e }));
  }

  const close = () => {
    dispatch("close", localExits);
  };

  const handleKey = (event) => {
    if (!open) return;
    if (event.key === "Escape") {
      close();
    }
  };

  function addExit() {
    localExits = [
      ...localExits,
      {
        name: "",
        description: "",
        target: "",
        exitType: "normal",
        hidden: false,
      },
    ];
  }

  function removeExit(index) {
    localExits = localExits.filter((_, i) => i !== index);
  }

  function updateExit(index, field, value) {
    localExits = localExits.map((exit, i) =>
      i === index ? { ...exit, [field]: value } : exit
    );
  }

  const exitTypes = [
    { value: "normal", label: "Normal" },
    { value: "direction", label: "Direction" },
    { value: "teleport", label: "Teleport" },
    { value: "portal", label: "Portal" },
    { value: "climb", label: "Climb" },
    { value: "swim", label: "Swim" },
  ];
</script>

<svelte:window on:keydown={handleKey} />

{#if open}
  <div class="modal-backdrop" on:click|self={close}>
    <div class="modal-container">
      <div class="modal-header">
        <div class="header-text">
          <h2>Special Exits</h2>
          <p class="header-subtitle">
            Manage non-cardinal exits like portals, stairs, and custom directions.
          </p>
        </div>
        <button class="close-btn" type="button" on:click={close}>
          <span class="material-symbols-outlined">close</span>
        </button>
      </div>

      <div class="modal-content">
        {#if localExits.length === 0}
          <div class="empty-state">
            <span class="material-symbols-outlined empty-icon">door_front</span>
            <p>No special exits defined.</p>
            <button class="add-btn" type="button" on:click={addExit}>
              <span class="material-symbols-outlined">add</span>
              Add Special Exit
            </button>
          </div>
        {:else}
          <div class="exits-list">
            {#each localExits as exit, index}
              <div class="exit-row">
                <div class="exit-fields">
                  <div class="field-group">
                    <label>Exit Name</label>
                    <input
                      type="text"
                      placeholder="e.g., portal, stairs, ladder"
                      value={exit.name}
                      on:input={(e) => updateExit(index, "name", e.target.value)}
                    />
                  </div>

                  <div class="field-group">
                    <label>Type</label>
                    <select
                      value={exit.exitType || "normal"}
                      on:change={(e) => updateExit(index, "exitType", e.target.value)}
                    >
                      {#each exitTypes as type}
                        <option value={type.value}>{type.label}</option>
                      {/each}
                    </select>
                  </div>

                  <div class="field-group">
                    <label>Target Room</label>
                    <select
                      value={exit.target}
                      on:change={(e) => updateExit(index, "target", e.target.value)}
                    >
                      <option value="">Select target...</option>
                      {#each roomsValueHelp as room}
                        <option value={room.id}>{room.name}</option>
                      {/each}
                    </select>
                  </div>

                  <div class="field-group wide">
                    <label>Description</label>
                    <input
                      type="text"
                      placeholder="What the player sees when using this exit"
                      value={exit.description || ""}
                      on:input={(e) => updateExit(index, "description", e.target.value)}
                    />
                  </div>

                  <div class="field-group checkbox-group">
                    <label class="checkbox-label">
                      <input
                        type="checkbox"
                        checked={exit.hidden || false}
                        on:change={(e) => updateExit(index, "hidden", e.target.checked)}
                      />
                      <span>Hidden</span>
                    </label>
                  </div>
                </div>

                <button
                  class="remove-btn"
                  type="button"
                  on:click={() => removeExit(index)}
                  title="Remove exit"
                >
                  <span class="material-symbols-outlined">delete</span>
                </button>
              </div>
            {/each}
          </div>

          <button class="add-btn secondary" type="button" on:click={addExit}>
            <span class="material-symbols-outlined">add</span>
            Add Another Exit
          </button>
        {/if}
      </div>

      <div class="modal-footer">
        <button class="btn-cancel" type="button" on:click={close}>
          Done
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    z-index: 100;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(4px);
    padding: 40px 20px;
    overflow-y: auto;
  }

  .modal-container {
    width: 100%;
    max-width: 700px;
    background: #1e1e1e;
    border: 1px solid #3a3a3a;
    border-radius: 12px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
    overflow: hidden;
  }

  .modal-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    padding: 20px 24px;
    border-bottom: 1px solid #3a3a3a;
    background: #252525;
  }

  .header-text h2 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
    color: #fff;
  }

  .header-subtitle {
    margin: 4px 0 0;
    font-size: 12px;
    color: #888;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: #888;
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    transition: all 0.2s ease;
  }

  .close-btn:hover {
    background: #333;
    color: #fff;
  }

  .modal-content {
    padding: 20px 24px;
    max-height: 60vh;
    overflow-y: auto;
  }

  .empty-state {
    text-align: center;
    padding: 40px 20px;
    color: #888;
  }

  .empty-icon {
    font-size: 48px;
    margin-bottom: 12px;
    opacity: 0.5;
  }

  .empty-state p {
    margin: 0 0 20px;
    font-size: 14px;
  }

  .exits-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .exit-row {
    display: flex;
    gap: 12px;
    padding: 16px;
    background: #252525;
    border: 1px solid #3a3a3a;
    border-radius: 8px;
  }

  .exit-fields {
    flex: 1;
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
  }

  .field-group {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .field-group.wide {
    grid-column: span 2;
  }

  .field-group label {
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    color: #888;
  }

  .field-group input[type="text"],
  .field-group select {
    padding: 8px 10px;
    font-size: 13px;
    background: #1a1a1a;
    border: 1px solid #3a3a3a;
    border-radius: 4px;
    color: #fff;
  }

  .field-group input[type="text"]:focus,
  .field-group select:focus {
    outline: none;
    border-color: #00bcd4;
  }

  .checkbox-group {
    justify-content: flex-end;
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: #ccc;
    cursor: pointer;
  }

  .checkbox-label input[type="checkbox"] {
    width: 16px;
    height: 16px;
    cursor: pointer;
  }

  .remove-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    background: transparent;
    border: 1px solid #3a3a3a;
    border-radius: 6px;
    color: #888;
    cursor: pointer;
    transition: all 0.2s ease;
    align-self: center;
  }

  .remove-btn:hover {
    background: #f44336;
    border-color: #f44336;
    color: #fff;
  }

  .add-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 10px 20px;
    font-size: 13px;
    font-weight: 500;
    background: #00bcd4;
    border: none;
    border-radius: 6px;
    color: #fff;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .add-btn:hover {
    background: #00a5bb;
  }

  .add-btn.secondary {
    margin-top: 16px;
    background: transparent;
    border: 1px dashed #3a3a3a;
    color: #888;
  }

  .add-btn.secondary:hover {
    border-color: #00bcd4;
    color: #00bcd4;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    padding: 16px 24px;
    border-top: 1px solid #3a3a3a;
    background: #252525;
  }

  .btn-cancel {
    padding: 10px 24px;
    font-size: 13px;
    font-weight: 500;
    background: #00bcd4;
    border: none;
    border-radius: 6px;
    color: #fff;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .btn-cancel:hover {
    background: #00a5bb;
  }
</style>
