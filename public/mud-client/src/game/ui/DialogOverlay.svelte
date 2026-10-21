<script>
  import { portraitBustSrc, onPortraitBustError } from '../portraitSrc.js';
  import { portal } from '../portal.js';
  import { optionLabel } from '../questRepeat.js';
  import RepeatMark from './RepeatMark.svelte';

  export let npcName = "";
  export let npcText = "";
  export let options = [];
  export let npcType = "npc";
  export let npc = null;
  export let sendMessage;

  const TRADE_TOKEN = "__trade__";

  $: displayOptions = (() => {
    const opts = Array.isArray(options) ? options.slice() : [];
    if (npc?.isMerchant) {
      const hasTrade = opts.some((o) => /^(trade|shop)\b/i.test(String(o?.text || "").trim()));
      if (!hasTrade) {
        opts.push({ index: TRADE_TOKEN, text: "Trade / Shop" });
      }
    }
    return opts;
  })();

  $: leaveOption = displayOptions.find((option) =>
    /\b(go|leave|goodbye|farewell|bye)\b/i.test(String(option?.text || ''))
  ) || null;

  function handleOption(option) {
    if (option?.index === TRADE_TOKEN || /^(trade|shop)\b/i.test(String(option?.text || "").trim())) {
      // ExactCommandKey: must be bare "list" / "trade" (no NPC name)
      sendMessage("list");
      return;
    }
    sendMessage(String(option.index));
  }

  function isQuestOption(option) {
    return option?.text?.startsWith("[Quest]") || option?.text?.startsWith("[Turn In]");
  }

  function isTradeOption(option) {
    return option?.index === TRADE_TOKEN || /^(trade|shop)\b/i.test(String(option?.text || "").trim());
  }
</script>

<style>
  .dialog-overlay {
    position: fixed;
    inset: 0;
    z-index: var(--layer-dialog);
    background: rgba(0, 0, 0, 0.72);
    display: flex;
    align-items: flex-end;
    justify-content: center;
    padding: 0;
    padding-top: max(0.5rem, env(safe-area-inset-top, 0px));
    box-sizing: border-box;
    animation: fadeIn 0.3s ease-out;
    overflow: hidden;
  }

  .dialog-sheet {
    width: 100%;
    max-height: min(88dvh, calc(100dvh - env(safe-area-inset-top, 0px) - 0.5rem));
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
    box-sizing: border-box;
    padding: 0.85rem 1rem max(1rem, env(safe-area-inset-bottom, 0px));
    background: rgba(8, 10, 14, 0.97);
    border: 1px solid rgba(255, 255, 255, 0.14);
    border-bottom: none;
    border-radius: 16px 16px 0 0;
    box-shadow: 0 -16px 40px rgba(0, 0, 0, 0.5);
  }

  .dialog-sheet-bar {
    position: sticky;
    top: 0;
    z-index: 2;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    min-height: 44px;
    margin: -0.15rem 0 0.35rem;
    background: rgba(8, 10, 14, 0.97);
  }

  .dialog-dismiss {
    appearance: none;
    min-width: 44px;
    min-height: 44px;
    border: none;
    border-radius: 8px;
    background: transparent;
    color: #cbd5e1;
    font-size: 1.6rem;
    line-height: 1;
    cursor: pointer;
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

  /* Large 1:1 bust in a rounded frame so the flat chest crop is clipped. */
  .npc-portrait-slot {
    width: clamp(100px, 22vw, 140px);
    aspect-ratio: 1 / 1;
    flex-shrink: 0;
    position: relative;
    overflow: hidden;
    border-radius: 12px;
    border: 1px solid rgba(255, 255, 255, 0.18);
    background: #111;
    box-sizing: border-box;
  }
  .npc-portrait-slot.enemy {
    border-color: rgba(239, 68, 68, 0.45);
  }
  .npc-portrait-slot.merchant {
    border-color: rgba(34, 197, 94, 0.45);
  }
  .npc-portrait-slot.quest {
    border-color: rgba(245, 158, 11, 0.45);
  }
  .npc-portrait-slot.npc {
    border-color: rgba(59, 130, 246, 0.4);
  }

  .npc-portrait {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: center center;
    image-rendering: pixelated;
    display: block;
    pointer-events: none;
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

  .dialog-option-btn.trade {
    background: rgba(34, 197, 94, 0.15);
    border-color: rgba(34, 197, 94, 0.35);
    color: #86efac;
  }

  .dialog-option-btn.trade:hover,
  .dialog-option-btn.trade:focus {
    background: rgba(34, 197, 94, 0.25);
    border-color: rgba(34, 197, 94, 0.55);
  }

  .dialog-option-btn.trade .option-index {
    background: rgba(34, 197, 94, 0.28);
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

  .option-body {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.2rem;
  }

  .option-text {
    flex: 1;
    min-width: 0;
    overflow-wrap: anywhere;
  }

  @media screen and (min-width: 769px) {
    .dialog-overlay {
      align-items: center;
      padding:
        max(1.25rem, env(safe-area-inset-top, 0px))
        max(1.25rem, env(safe-area-inset-right, 0px))
        max(1.25rem, env(safe-area-inset-bottom, 0px))
        max(1.25rem, env(safe-area-inset-left, 0px));
      overflow: hidden;
    }

    .dialog-sheet {
      width: min(720px, 100%);
      max-height: min(86dvh, 820px);
      min-height: 0;
      overflow-y: auto;
      border-radius: 12px;
      border-bottom: 1px solid rgba(255, 255, 255, 0.14);
      background: rgba(0, 0, 0, 0.9);
      backdrop-filter: blur(8px);
      -webkit-backdrop-filter: blur(8px);
    }

    .dialog-sheet-bar {
      background: transparent;
    }
  }

  /* Phone landscape is wider than 768px and shorter than the centered panel. */
  @media screen and (max-height: 500px) {
    .dialog-overlay {
      align-items: flex-end;
      padding: max(0.35rem, env(safe-area-inset-top, 0px)) 0 0;
      overflow: hidden;
    }

    .dialog-sheet {
      width: 100%;
      max-height: min(92dvh, calc(100dvh - env(safe-area-inset-top, 0px) - 0.35rem));
      min-height: 0;
      overflow-y: auto;
      border-radius: 16px 16px 0 0;
    }

    .npc-portrait-slot {
      width: 72px;
    }

    .dialog-text {
      font-size: 0.95em;
      padding: 0.6em;
    }

    .dialog-option-btn {
      padding: 0.55em 0.75em;
    }
  }

  @media screen and (max-width: 600px) {
    .dialog-sheet {
      padding-left: 0.85rem;
      padding-right: 0.85rem;
    }

    .npc-portrait-slot {
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

<div class="dialog-overlay" use:portal>
  <div class="dialog-sheet" role="dialog" aria-modal="true" aria-label={npcName || 'Conversation'}>
  {#if leaveOption}
    <div class="dialog-sheet-bar">
      <button class="dialog-dismiss" type="button" aria-label="Close" on:click={() => handleOption(leaveOption)}>×</button>
    </div>
  {/if}
  <div class="dialog-header">
    <div class="npc-portrait-slot {npcType}">
      <img
        class="npc-portrait"
        src={portraitBustSrc(npc)}
        alt=""
        on:error={(e) => onPortraitBustError(e, npc)}
      />
    </div>
    <div class="dialog-header-text">
      <span class="npc-name">{npcName}</span>
      <div class="dialog-text" class:quest={npcType === 'quest'}>
        {npcText}
      </div>
    </div>
  </div>

  {#if displayOptions && displayOptions.length > 0}
    <div class="dialog-options">
      {#each displayOptions as option}
        <button
          class="dialog-option-btn"
          class:quest={isQuestOption(option)}
          class:trade={isTradeOption(option)}
          on:click={() => handleOption(option)}
        >
          <span class="option-index">{isTradeOption(option) ? '★' : option.index}</span>
          <span class="option-body">
            <span class="option-text">{optionLabel(option)}</span>
            <RepeatMark repeats={option.repeats} repeatReset={option.repeatReset} nextAvailableAt={option.nextAvailableAt} />
          </span>
        </button>
      {/each}
    </div>
  {/if}
  </div>
</div>
