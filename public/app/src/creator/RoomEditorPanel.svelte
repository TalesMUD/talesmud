<script>
  import { createEventDispatcher } from "svelte";
  import CardinalExitsEditor from "./CardinalExitsEditor.svelte";
  import SpecialExitsModal from "./SpecialExitsModal.svelte";
  import { isCardinalDirection, CARDINAL_DIRECTIONS } from "./WorldEditorStore.js";

  export let room = null;
  export let isCreating = false;
  export let saving = false;
  export let roomsValueHelp = [];

  const dispatch = createEventDispatcher();

  let showSpecialExitsModal = false;

  // Local editable copy
  let editingRoom = null;
  let lastRoomId = null;

  // Sync local copy when room changes (only when ID changes, not on every re-render)
  $: if (room) {
    // Only reset editingRoom if the room ID has changed
    if (room.id !== lastRoomId) {
      lastRoomId = room.id;
      editingRoom = JSON.parse(JSON.stringify(room));
      // Ensure coords exist
      if (!editingRoom.coords) {
        editingRoom.coords = { x: 0, y: 0, z: 0 };
      }
      // Ensure exits exist
      if (!editingRoom.exits) {
        editingRoom.exits = [];
      }
      // Ensure meta exists
      if (!editingRoom.meta) {
        editingRoom.meta = { background: "" };
      }
    }
  } else {
    editingRoom = null;
    lastRoomId = null;
  }

  function handleSave() {
    dispatch("save", editingRoom);
  }

  function handleCancel() {
    dispatch("cancel");
  }

  function handleDelete() {
    if (confirm("Are you sure you want to delete this room?")) {
      dispatch("delete", editingRoom.id);
    }
  }

  function handleFullEditor() {
    dispatch("openFullEditor", editingRoom.id);
  }

  // Cardinal exits handling
  function handleAddCardinalExit(event) {
    const exit = event.detail;
    editingRoom.exits = [...editingRoom.exits, exit];
  }

  function handleRemoveCardinalExit(event) {
    const direction = event.detail;
    editingRoom.exits = editingRoom.exits.filter(
      e => e.name.toLowerCase() !== direction.toLowerCase()
    );
  }

  function handleUpdateCardinalExit(event) {
    const { direction, target } = event.detail;
    editingRoom.exits = editingRoom.exits.map(e =>
      e.name.toLowerCase() === direction.toLowerCase()
        ? { ...e, target }
        : e
    );
  }

  // Special exits modal handling
  function handleSpecialExitsClose(event) {
    const specialExits = event.detail;
    // Keep cardinal exits, replace special exits
    const cardinalExits = editingRoom.exits.filter(e => isCardinalDirection(e.name || ""));
    editingRoom.exits = [...cardinalExits, ...specialExits];
    showSpecialExitsModal = false;
  }

  // Count special exits
  $: specialExitsCount = editingRoom
    ? editingRoom.exits.filter(e => !isCardinalDirection(e.name || "")).length
    : 0;
</script>

<div class="editor-panel">
  {#if editingRoom}
    <div class="panel-header">
      <h3>
        {#if isCreating}
          Create New Room
        {:else}
          Edit Room
        {/if}
      </h3>
      <button class="close-btn" on:click={handleCancel}>
        <span class="material-symbols-outlined">close</span>
      </button>
    </div>

    <div class="panel-content">
      <div class="field-group">
        <label for="room-name">Name</label>
        <input
          id="room-name"
          type="text"
          placeholder="Room name"
          bind:value={editingRoom.name}
        />
      </div>

      <div class="field-group">
        <label for="room-description">Description</label>
        <textarea
          id="room-description"
          rows="3"
          placeholder="What the player sees when entering..."
          bind:value={editingRoom.description}
        ></textarea>
      </div>

      <div class="coords-grid">
        <div class="field-group">
          <label for="coord-x">X</label>
          <input
            id="coord-x"
            type="number"
            bind:value={editingRoom.coords.x}
          />
        </div>
        <div class="field-group">
          <label for="coord-y">Y</label>
          <input
            id="coord-y"
            type="number"
            bind:value={editingRoom.coords.y}
          />
        </div>
        <div class="field-group">
          <label for="coord-z">Z</label>
          <input
            id="coord-z"
            type="number"
            bind:value={editingRoom.coords.z}
          />
        </div>
      </div>

      <div class="field-row">
        <div class="field-group">
          <label for="room-area">Area</label>
          <input
            id="room-area"
            type="text"
            placeholder="e.g., Town Square"
            bind:value={editingRoom.area}
          />
        </div>
        <div class="field-group">
          <label for="room-area-type">Area Type</label>
          <input
            id="room-area-type"
            type="text"
            placeholder="e.g., town, forest"
            bind:value={editingRoom.areaType}
          />
        </div>
      </div>

      <div class="field-group">
        <label for="room-type">Room Type</label>
        <input
          id="room-type"
          type="text"
          placeholder="e.g., outdoor, indoor"
          bind:value={editingRoom.roomType}
        />
      </div>

      <div class="section-divider"></div>

      <div class="exits-section">
        <h4>Cardinal Exits</h4>
        <CardinalExitsEditor
          exits={editingRoom.exits}
          {roomsValueHelp}
          on:addExit={handleAddCardinalExit}
          on:removeExit={handleRemoveCardinalExit}
          on:updateExit={handleUpdateCardinalExit}
        />
      </div>

      <button
        class="special-exits-btn"
        type="button"
        on:click={() => showSpecialExitsModal = true}
      >
        <span class="material-symbols-outlined">door_sliding</span>
        Manage Special Exits
        {#if specialExitsCount > 0}
          <span class="badge">{specialExitsCount}</span>
        {/if}
      </button>
    </div>

    <div class="panel-footer">
      <button class="btn-link" type="button" on:click={handleFullEditor}>
        <span class="material-symbols-outlined">open_in_new</span>
        Full Editor
      </button>

      <div class="footer-actions">
        {#if !isCreating}
          <button class="btn-delete" type="button" on:click={handleDelete}>
            <span class="material-symbols-outlined">delete</span>
          </button>
        {/if}
        <button class="btn-cancel" type="button" on:click={handleCancel}>
          Cancel
        </button>
        <button class="btn-save" type="button" on:click={handleSave} disabled={saving}>
          {#if saving}
            Saving...
          {:else if isCreating}
            Create
          {:else}
            Save
          {/if}
        </button>
      </div>
    </div>
  {:else}
    <div class="no-selection">
      <span class="material-symbols-outlined">touch_app</span>
      <p>Select a room to edit</p>
      <p class="hint">Or drag from a room's direction handle to create a new adjacent room</p>
    </div>
  {/if}
</div>

<SpecialExitsModal
  open={showSpecialExitsModal}
  exits={editingRoom ? editingRoom.exits : []}
  {roomsValueHelp}
  on:close={handleSpecialExitsClose}
/>

<style>
  .editor-panel {
    width: 380px;
    min-width: 380px;
    height: 100%;
    background: #1e1e1e;
    border-left: 1px solid #3a3a3a;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
    border-bottom: 1px solid #3a3a3a;
    background: #252525;
  }

  .panel-header h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: #00bcd4;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: #888;
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .close-btn:hover {
    background: #333;
    color: #fff;
  }

  .panel-content {
    flex: 1;
    padding: 20px;
    overflow-y: auto;
  }

  .field-group {
    margin-bottom: 16px;
  }

  .field-group label {
    display: block;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    color: #888;
    margin-bottom: 6px;
  }

  .field-group input,
  .field-group textarea {
    width: 100%;
    padding: 10px 12px;
    font-size: 13px;
    background: #2a2a2a;
    border: 1px solid #3a3a3a;
    border-radius: 6px;
    color: #fff;
    resize: vertical;
  }

  .field-group input:focus,
  .field-group textarea:focus {
    outline: none;
    border-color: #00bcd4;
  }

  .field-group input[type="number"] {
    -moz-appearance: textfield;
  }

  .field-group input[type="number"]::-webkit-outer-spin-button,
  .field-group input[type="number"]::-webkit-inner-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }

  .coords-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
    margin-bottom: 16px;
  }

  .coords-grid .field-group {
    margin-bottom: 0;
  }

  .field-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }

  .field-row .field-group {
    margin-bottom: 16px;
  }

  .section-divider {
    height: 1px;
    background: #3a3a3a;
    margin: 20px 0;
  }

  .exits-section {
    margin-bottom: 16px;
  }

  .exits-section h4 {
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    color: #888;
    margin: 0 0 12px 0;
  }

  .special-exits-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    width: 100%;
    padding: 12px;
    font-size: 13px;
    font-weight: 500;
    background: transparent;
    border: 1px dashed #3a3a3a;
    border-radius: 6px;
    color: #888;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .special-exits-btn:hover {
    border-color: #00bcd4;
    color: #00bcd4;
  }

  .special-exits-btn .badge {
    background: #00bcd4;
    color: #000;
    font-size: 11px;
    font-weight: 600;
    padding: 2px 6px;
    border-radius: 10px;
  }

  .panel-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
    border-top: 1px solid #3a3a3a;
    background: #252525;
  }

  .btn-link {
    display: flex;
    align-items: center;
    gap: 4px;
    background: transparent;
    border: none;
    color: #00bcd4;
    font-size: 12px;
    cursor: pointer;
    padding: 0;
  }

  .btn-link:hover {
    text-decoration: underline;
  }

  .btn-link .material-symbols-outlined {
    font-size: 16px;
  }

  .footer-actions {
    display: flex;
    gap: 8px;
  }

  .btn-delete {
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
  }

  .btn-delete:hover {
    background: #f44336;
    border-color: #f44336;
    color: #fff;
  }

  .btn-cancel {
    padding: 8px 16px;
    font-size: 13px;
    background: transparent;
    border: 1px solid #3a3a3a;
    border-radius: 6px;
    color: #aaa;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .btn-cancel:hover {
    border-color: #555;
    color: #fff;
  }

  .btn-save {
    padding: 8px 20px;
    font-size: 13px;
    font-weight: 500;
    background: #00bcd4;
    border: none;
    border-radius: 6px;
    color: #000;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .btn-save:hover:not(:disabled) {
    background: #00a5bb;
  }

  .btn-save:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .no-selection {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 40px;
    color: #666;
  }

  .no-selection .material-symbols-outlined {
    font-size: 48px;
    margin-bottom: 16px;
    opacity: 0.5;
  }

  .no-selection p {
    margin: 0 0 8px;
    font-size: 14px;
  }

  .no-selection .hint {
    font-size: 12px;
    color: #555;
  }
</style>
