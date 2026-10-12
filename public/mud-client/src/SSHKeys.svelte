<script>
  import { getAuth } from "./auth.js";
  import { listKeys, revokeKey } from "./api/ssh.js";
  import { sshKeysOpen } from "./sshStore.js";

  const { authToken } = getAuth();

  let keys = [];
  let errorText = "";
  let off = false;
  let loading = false;
  let busy = false;
  let loadedFor = "";

  function message(err, fallback) {
    const text = err?.response?.data?.error;
    return typeof text === "string" && text ? text : fallback;
  }

  async function load(token) {
    loading = true;
    errorText = "";
    off = false;
    try {
      const res = await listKeys(token);
      keys = Array.isArray(res.data) ? res.data : [];
    } catch (err) {
      keys = [];
      if (err?.response?.status === 404) {
        off = true;
      } else {
        errorText = message(err, "Could not load SSH keys.");
      }
    } finally {
      loading = false;
    }
  }

  $: if ($sshKeysOpen && $authToken && loadedFor !== $authToken) {
    loadedFor = $authToken;
    load($authToken);
  }

  function close() {
    sshKeysOpen.set(false);
    errorText = "";
    loadedFor = "";
  }

  async function revoke(id) {
    if (busy || !$authToken) return;
    busy = true;
    errorText = "";
    try {
      await revokeKey($authToken, id);
      await load($authToken);
    } catch (err) {
      errorText = message(err, "Could not revoke that key.");
    } finally {
      busy = false;
    }
  }

  function when(value) {
    if (!value) return "";
    const parsed = new Date(value);
    if (Number.isNaN(parsed.getTime()) || parsed.getTime() <= 0) return "";
    return parsed.toLocaleString();
  }
</script>

{#if $sshKeysOpen}
  <div class="ssh-overlay">
    <div class="ssh-panel" role="dialog" aria-modal="true" aria-label="SSH keys">
      <header>
        <h2>SSH keys</h2>
        <button type="button" class="text" on:click={close}>Close</button>
      </header>
      {#if off}
        <p>SSH keys are not enabled on this server.</p>
      {:else}
        <p class="hint">Keys are linked from an SSH sign-in. Press Y when the lobby asks to remember this computer, then reconnect and press Y again. A key pasted here does not sign in.</p>
        {#if loading}
          <p>Loading keys…</p>
        {:else if keys.length === 0}
          <p>No keys linked yet.</p>
        {:else}
          <ul>
            {#each keys as key (key.id)}
              <li>
                <div>
                  <strong>{key.label || key.key_type || "Key"}</strong>
                  <span>{key.fingerprint}</span>
                  {#if when(key.created_at)}
                    <span>Added {when(key.created_at)}{key.created_via ? ` via ${key.created_via}` : ""}</span>
                  {/if}
                </div>
                <button type="button" disabled={busy} on:click={() => revoke(key.id)}>Revoke</button>
              </li>
            {/each}
          </ul>
        {/if}
      {/if}
      {#if errorText}
        <p class="err">{errorText}</p>
      {/if}
    </div>
  </div>
{/if}

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
    width: min(100%, 520px);
    max-height: calc(100dvh - 2rem);
    overflow: auto;
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
  .hint { color: #b6aa92; font-size: 0.88rem; }
  ul { list-style: none; margin: 0 0 0.8rem; padding: 0; }
  li {
    display: flex;
    justify-content: space-between;
    gap: 0.75rem;
    padding: 0.55rem 0;
    border-bottom: 1px solid #8b692f55;
  }
  li div { display: flex; flex-direction: column; gap: 0.15rem; min-width: 0; }
  li span { overflow-wrap: anywhere; color: #b6aa92; font-size: 0.82rem; }
  button {
    border: 1px solid #8b692f;
    border-radius: 6px;
    padding: 0.4rem 0.7rem;
    background: #18150f;
    color: #f0e6d3;
    cursor: pointer;
  }
  button:disabled { opacity: 0.55; cursor: default; }
  button.text { background: transparent; }
  .err { color: #e7b1a4; }
</style>
