<script>
  import { portraitSrc, onPortraitError } from '../portraitSrc.js';

  export let npcName = "";
  export let npcText = "";
  export let options = [];
  export let npcType = "npc";
  export let npc = null;
  export let sendMessage;

  function handleOption(index) {
    sendMessage(String(index));
  }

  function isQuestOption(option) {
    return option?.text?.startsWith("[Quest]") || option?.text?.startsWith("[Turn In]");
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
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .dialog-header {
    display: flex;
    align-items: flex-start;
    gap: 1em;
    margin-bottom: 1em;
  }

  .npc-portrait-wrap {
    width: clamp(100px, 22vw, 140px);
    aspect-ratio: 2 / 3;
    border-radius: 10px;
    overflow: hidden;
    flex-shrink: 0;
    position: relative;
    border: 2px solid rgba(255, 255, 255, 0.15);
    background: #111;
  }

  .npc-portrait-wrap.enemy {
    border-color: rgba(239, 68, 68, 0.5);
  }

  .npc-portrait-wrap.merchant {
    border-color: rgba(34, 197, 94, 0.5);
  }

  .npc-portrait-wrap.quest {
    border-color: rgba(245, 158, 11, 0.5);
  }

  .npc-portrait-wrap.npc {
    border-color: rgba(59, 130, 246, 0.5);
  }

  .npc-portrait {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: top center;
    transform: scale(1.6);
    transform-origin: top center;
    image-rendering: pixelated;
    display: block;
  }

  .dialog-header-text {
    flex: 1;
    min-width: 0;
  }

  .npc-name {
    font-size: 1.3em;
    font-weight: 600;
    color: #e5e7eb;
    display: block;
    margin-bottom: 0.35em;
  }

  .dialog-text {
    font-size: 1.1em;
    line-height: 1.6;
    color: #d1d5db;
    margin-bottom: 0;
    padding: 1em;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 8px;
    border-left: 3px solid rgba(59, 130, 246, 0.5);
  }

  .dialog-text.quest {
    border-left-color: rgba(245, 158, 11, 0.65);
  }

  .dialog-options {
    display: flex;
    flex-direction: column;
    gap: 0.5em;
    margin-top: 1em;
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
    transition: background 0.15s ease, border-color 0.15s ease;
  }

  .dialog-option-btn:hover,
  .dialog-option-btn:focus {
    background: rgba(59, 130, 246, 0.25);
    border-color: rgba(59, 130, 246, 0.5);
  }

  .dialog-option-btn.quest {
    background: rgba(245, 158, 11, 0.15);
    border-color: rgba(245, 158, 11, 0.35);
    color: #fcd34d;
  }

  .dialog-option-btn.quest:hover,
  .dialog-option-btn.quest:focus {
    background: rgba(245, 158, 11, 0.25);
    border-color: rgba(245, 158, 11, 0.55);
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

  .dialog-option-btn.quest .option-index {
    background: rgba(245, 158, 11, 0.28);
  }

  .option-text {
    flex: 1;
  }

  @media screen and (max-width: 600px) {
    .dialog-overlay {
      padding: 1em;
    }

    .npc-portrait-wrap {
      width: clamp(80px, 28vw, 110px);
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
    <div class="npc-portrait-wrap {npcType}">
      <img
        class="npc-portrait"
        src={portraitSrc(npc)}
        alt=""
        on:error={(e) => onPortraitError(e, npc)}
      />
    </div>
    <div class="dialog-header-text">
      <span class="npc-name">{npcName}</span>
      <div class="dialog-text" class:quest={npcType === 'quest'}>
        {npcText}
      </div>
    </div>
  </div>

  {#if options && options.length > 0}
    <div class="dialog-options">
      {#each options as option}
        <button
          class="dialog-option-btn"
          class:quest={isQuestOption(option)}
          on:click={() => handleOption(option.index)}
        >
          <span class="option-index">{option.index}</span>
          <span class="option-text">{option.text}</span>
        </button>
      {/each}
    </div>
  {/if}
</div>
