<script>
  import { onMount, onDestroy, tick } from "svelte";
  import { portal } from "../portal.js";
  import { hintBlocked, hintRect } from "../guideHintLayout.js";

  export let store = null;
  export let sendMessage = null;

  let root;
  let box = null;
  let frame = 0;
  let timer = 0;

  $: hint = store ? ($store.guideHint || null) : null;
  $: blocked = store ? hintBlocked($store) : true;
  $: show = !!(hint && hint.text && !blocked);

  $: if (show) schedule();
  $: if (!show) box = null;

  function schedule() {
    cancelAnimationFrame(frame);
    frame = requestAnimationFrame(() => { place(); });
  }

  async function place() {
    await tick();
    if (!root || typeof window === "undefined") {
      box = null;
      return;
    }
    const width = Math.min(340, window.innerWidth - 16);
    const height = root.offsetHeight || 48;
    const bars = [...document.querySelectorAll(".hotbar-widget")].map(rectOf);
    const names = [...document.querySelectorAll(".room-name, .roomName")].map(rectOf);
    box = hintRect({
      vw: window.innerWidth,
      vh: window.innerHeight,
      hotbars: bars,
      roomNames: names,
      width,
      height,
    });
  }

  function rectOf(el) {
    if (!el || !el.getBoundingClientRect) return null;
    const r = el.getBoundingClientRect();
    if (r.width < 1 || r.height < 1) return null;
    return { top: r.top, left: r.left, width: r.width, height: r.height };
  }

  function act() {
    if (hint && hint.command && sendMessage) sendMessage(hint.command);
  }

  function dismiss() {
    if (sendMessage) sendMessage("guide dismiss");
    if (store && store.setGuideHint) store.setGuideHint(null);
  }

  onMount(() => {
    window.addEventListener("resize", schedule);
    timer = setInterval(() => { if (show) place(); }, 400);
  });

  onDestroy(() => {
    cancelAnimationFrame(frame);
    if (timer) clearInterval(timer);
    if (typeof window !== "undefined") window.removeEventListener("resize", schedule);
  });

  $: style = box
    ? `top:${box.top}px;left:${box.left}px;width:${box.width}px;visibility:visible;`
    : "visibility:hidden;top:0;left:8px;width:min(340px, calc(100vw - 16px));";
</script>

{#if show}
  <div
    class="guide-hint"
    use:portal
    bind:this={root}
    style={style}
    role="status"
    aria-live="polite"
  >
    <p>{hint.text}</p>
    {#if hint.command}
      <button type="button" class="act" on:click={act}>{hint.label || "Go"}</button>
    {/if}
    <button type="button" class="x" aria-label="Dismiss hint" on:click={dismiss}>×</button>
  </div>
{/if}

<style>
  .guide-hint {
    position: fixed;
    z-index: 8900;
    display: flex;
    align-items: center;
    gap: 8px;
    max-width: calc(100vw - 16px);
    box-sizing: border-box;
    margin: 0;
    padding: 8px 8px 8px 12px;
    pointer-events: auto;
    color: #f3e6d0;
    background: linear-gradient(180deg, #3a2a1c 0%, #24180f 100%);
    border: 1px solid rgba(211, 173, 99, 0.55);
    border-radius: 10px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.45);
    font-family: Georgia, "Iowan Old Style", Palatino, serif;
    font-size: 14px;
    line-height: 1.3;
  }

  p {
    margin: 0;
    flex: 1 1 auto;
    min-width: 0;
  }

  button {
    flex: 0 0 auto;
    cursor: pointer;
    font: inherit;
  }

  .act {
    color: #f8efe2;
    background: #8a5a2a;
    border: 1px solid rgba(211, 173, 99, 0.7);
    border-radius: 7px;
    padding: 6px 10px;
    font-weight: 600;
  }

  .x {
    color: #d7c4a3;
    background: transparent;
    border: 0;
    padding: 0 4px;
    font-size: 18px;
    line-height: 1;
  }
</style>
