<script>
  import { confirmDevice, denyDevice, lookupDevice } from "./api/ssh.js";
  import { clearActivateCode } from "./sshActivate.js";

  export let code = "";
  export let token = "";
  export let onClose = () => {};

  let started = false;
  let loading = false;
  let busy = false;
  let view = null;
  let errorText = "";
  let done = "";

  $: if (token && code && !started) {
    started = true;
    lookup();
  }

  function message(err, fallback) {
    const text = err?.response?.data?.error;
    return typeof text === "string" && text ? text : fallback;
  }

  async function lookup() {
    loading = true;
    errorText = "";
    view = null;
    try {
      const res = await lookupDevice(token, code);
      view = res.data || null;
    } catch (err) {
      errorText = message(err, "Could not look up that code.");
    } finally {
      loading = false;
    }
  }

  function finish(text) {
    done = text;
    clearActivateCode(sessionStorage);
  }

  async function confirm() {
    if (!view?.csrf || busy || done) return;
    busy = true;
    errorText = "";
    try {
      await confirmDevice(token, code, view.csrf);
      finish("Confirmed. Return to your SSH window.");
    } catch (err) {
      errorText = message(err, "Could not confirm.");
    } finally {
      busy = false;
    }
  }

  async function deny() {
    if (!view?.csrf || busy || done) return;
    busy = true;
    errorText = "";
    try {
      await denyDevice(token, code, view.csrf);
      finish("Denied. That sign-in was rejected.");
    } catch (err) {
      errorText = message(err, "Could not deny.");
    } finally {
      busy = false;
    }
  }

  function close() {
    onClose();
  }

  function when(value) {
    if (!value) return "";
    const parsed = new Date(value);
    if (Number.isNaN(parsed.getTime())) return String(value);
    return parsed.toLocaleString();
  }
</script>

<div class="ssh-overlay">
  <div class="ssh-panel" role="dialog" aria-modal="true" aria-label="Confirm SSH sign-in">
    <header>
      <h2>Confirm SSH sign-in</h2>
      <button type="button" class="text" on:click={close}>Close</button>
    </header>
    <p class="warn">Only confirm if you started this sign-in from your own terminal.</p>
    <p class="code">{code}</p>
    {#if loading}
      <p>Looking up the code…</p>
    {:else if done}
      <p class="ok">{done}</p>
    {:else if view}
      <dl>
        <dt>Address</dt>
        <dd>{view.ip || "unknown"}</dd>
        <dt>Client</dt>
        <dd>{view.mode || "ssh"}{view.client_version ? ` · ${view.client_version}` : ""}</dd>
        <dt>Started</dt>
        <dd>{when(view.created) || "just now"}</dd>
        <dt>Key</dt>
        <dd>{view.key || "No key offered yet"}</dd>
      </dl>
      <div class="actions">
        <button type="button" class="yes" disabled={busy || !view.csrf} on:click={confirm}>Confirm</button>
        <button type="button" class="no" disabled={busy || !view.csrf} on:click={deny}>Deny</button>
      </div>
    {/if}
    {#if errorText}
      <p class="err">{errorText}</p>
      {#if !done}
        <button type="button" on:click={lookup}>Retry</button>
      {/if}
    {/if}
  </div>
</div>

<style>
  .ssh-overlay {
    position: fixed;
    inset: 0;
    z-index: 1200;
    display: grid;
    place-items: center;
    padding: 1rem;
    background: rgba(5, 4, 3, 0.82);
    box-sizing: border-box;
  }
  .ssh-panel {
    width: min(100%, 440px);
    padding: 1rem 1.1rem 1.15rem;
    border: 1px solid #9b7435;
    border-radius: 10px;
    background: linear-gradient(150deg, #211b13, #12110e 65%);
    color: #e8dfca;
    box-shadow: 0 20px 80px #000c;
    font-family: system-ui, sans-serif;
  }
  header { display: flex; align-items: center; justify-content: space-between; gap: 0.75rem; }
  h2 { margin: 0; color: #f5d78c; font: 700 1.15rem Georgia, serif; }
  .warn { color: #e7c98a; font-size: 0.9rem; }
  .code { margin: 0.2rem 0 0.8rem; font: 600 1.15rem ui-monospace, monospace; letter-spacing: 0.08em; }
  dl { display: grid; grid-template-columns: 6.5rem 1fr; gap: 0.35rem 0.6rem; margin: 0 0 0.9rem; }
  dt { color: #b6aa92; }
  dd { margin: 0; overflow-wrap: anywhere; }
  .actions { display: flex; gap: 0.5rem; }
  button {
    border: 1px solid #8b692f;
    border-radius: 6px;
    padding: 0.45rem 0.75rem;
    background: #18150f;
    color: #f0e6d3;
    cursor: pointer;
  }
  button:disabled { opacity: 0.55; cursor: default; }
  button.yes { background: #3d2a12; color: #fff0bb; }
  button.no { color: #e7b1a4; }
  button.text { background: transparent; }
  .ok { color: #d7e7b2; }
  .err { color: #e7b1a4; }
</style>
