<script>
  export let npcName = "";
  export let npcText = "";
  export let options = [];
  export let npcType = "npc"; // "enemy", "merchant", or "npc"
  export let sendMessage;

  function handleOption(index) {
    sendMessage(String(index));
  }

  // Get icon based on NPC type
  function getNpcIcon(type) {
    switch (type) {
      case "enemy":
        return "swords";
      case "merchant":
        return "store";
      default:
        return "person";
    }
  }

  // Get icon color class based on NPC type
  function getNpcIconClass(type) {
    switch (type) {
      case "enemy":
        return "icon-enemy";
      case "merchant":
        return "icon-merchant";
      default:
        return "icon-npc";
    }
  }
</script>

<style>
  .dialog-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.85);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    z-index: 100;
    display: flex;
    flex-direction: column;
    padding: 1.5em;
    animation: fadeIn 0.3s ease-out;
    overflow-y: auto;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  .dialog-header {
    display: flex;
    align-items: center;
    gap: 0.75em;
    margin-bottom: 1em;
  }

  .npc-icon {
    width: 48px;
    height: 48px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .npc-icon i {
    font-size: 28px;
  }

  .icon-enemy {
    background: rgba(239, 68, 68, 0.2);
    border: 2px solid rgba(239, 68, 68, 0.5);
    color: #fca5a5;
  }

  .icon-merchant {
    background: rgba(34, 197, 94, 0.2);
    border: 2px solid rgba(34, 197, 94, 0.5);
    color: #86efac;
  }

  .icon-npc {
    background: rgba(59, 130, 246, 0.2);
    border: 2px solid rgba(59, 130, 246, 0.5);
    color: #93c5fd;
  }

  .npc-name {
    font-size: 1.3em;
    font-weight: 600;
    color: #e5e7eb;
  }

  .dialog-text {
    font-size: 1.1em;
    line-height: 1.6;
    color: #d1d5db;
    margin-bottom: 1.5em;
    padding: 1em;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 8px;
    border-left: 3px solid rgba(59, 130, 246, 0.5);
  }

  .dialog-options {
    display: flex;
    flex-direction: column;
    gap: 0.5em;
  }

  .dialog-option-btn {
    display: flex;
    align-items: center;
    gap: 0.75em;
    text-align: left;
    padding: 0.85em 1em;
    background: rgba(59, 130, 246, 0.15);
    border: 1px solid rgba(59, 130, 246, 0.3);
    border-radius: 8px;
    color: #93c5fd;
    font-size: 1em;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .dialog-option-btn:hover {
    background: rgba(59, 130, 246, 0.25);
    border-color: rgba(59, 130, 246, 0.5);
    transform: translateX(4px);
  }

  .dialog-option-btn:active {
    transform: translateX(2px);
  }

  .option-index {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    background: rgba(59, 130, 246, 0.3);
    border-radius: 4px;
    font-weight: 600;
    font-size: 0.9em;
    flex-shrink: 0;
  }

  .option-text {
    flex: 1;
  }

  /* Responsive adjustments */
  @media screen and (max-width: 600px) {
    .dialog-overlay {
      padding: 1em;
    }

    .npc-icon {
      width: 40px;
      height: 40px;
    }

    .npc-icon i {
      font-size: 24px;
    }

    .npc-name {
      font-size: 1.1em;
    }

    .dialog-text {
      font-size: 1em;
      padding: 0.75em;
    }

    .dialog-option-btn {
      padding: 0.7em 0.85em;
      font-size: 0.95em;
    }
  }
</style>

<div class="dialog-overlay">
  <div class="dialog-header">
    <div class="npc-icon {getNpcIconClass(npcType)}">
      <i class="material-icons">{getNpcIcon(npcType)}</i>
    </div>
    <span class="npc-name">{npcName}</span>
  </div>

  <div class="dialog-text">
    {npcText}
  </div>

  {#if options && options.length > 0}
    <div class="dialog-options">
      {#each options as option}
        <button
          class="dialog-option-btn"
          on:click={() => handleOption(option.index)}
        >
          <span class="option-index">{option.index}</span>
          <span class="option-text">{option.text}</span>
        </button>
      {/each}
    </div>
  {/if}
</div>
