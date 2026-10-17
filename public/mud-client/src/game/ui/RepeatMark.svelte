<script>
  import { formatAvailableAgain, repeatTag } from '../questRepeat.js';

  export let quest = null;
  export let repeats = false;
  export let repeatReset = '';
  export let nextAvailableAt = '';

  $: source = quest || { repeats, repeatReset, nextAvailableAt };
  $: tag = repeatTag(source);
  $: timer = formatAvailableAgain(source.nextAvailableAt);
</script>

{#if tag || timer}
  <span class="repeat-mark">
    {#if tag}<span class="repeat-tag">{tag}</span>{/if}
    {#if timer}<span class="repeat-timer">{timer}</span>{/if}
  </span>
{/if}

<style>
  .repeat-mark {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.35rem;
    max-width: 100%;
    min-width: 0;
  }

  .repeat-tag {
    display: inline-flex;
    align-items: center;
    padding: 0 0.4rem;
    border-radius: 3px;
    background: #c4a574;
    color: #3a2414;
    border: 1px solid #8a6232;
    font-size: 0.65rem;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    line-height: 1.5;
    font-weight: 700;
  }

  .repeat-timer {
    color: #d7b48a;
    font-size: 0.75rem;
    line-height: 1.3;
    overflow-wrap: anywhere;
  }
</style>
