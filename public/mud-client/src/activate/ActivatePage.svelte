<script>
  import { onMount } from "svelte";
  import createAuth0Client from "@auth0/auth0-spa-js";
  import { auth0Config, auth0Callback } from "../auth0Config.js";
  import { lookupDevice, confirmDevice, denyDevice } from "../api/ssh.js";
  import { ACTIVATE_KEY, clearActivateCode } from "../sshActivate.js";
  import { normalizeCode, isCompleteCode, codeFromLocation, ACTIVATE_RETURN } from "./activateFlow.js";

  // loading | enter | signin | lookup | confirm | done | denied
  let phase = "loading";
  let code = "";
  let typed = "";
  let auth0 = null;
  let token = "";
  let who = "";
  let view = null;
  let busy = false;
  let errorText = "";

  function store(value) {
    try {
      if (value) sessionStorage.setItem(ACTIVATE_KEY, value);
      else sessionStorage.removeItem(ACTIVATE_KEY);
    } catch (e) { /* private mode */ }
  }

  function stored() {
    try { return normalizeCode(sessionStorage.getItem(ACTIVATE_KEY) || ""); } catch (e) { return ""; }
  }

  function apiError(err, fallback) {
    const text = err?.response?.data?.error;
    return typeof text === "string" && text ? text : fallback;
  }

  onMount(async () => {
    code = codeFromLocation(window.location.search) || stored();
    if (code) store(code);
    if (window.location.search) {
      window.history.replaceState(null, "", ACTIVATE_RETURN);
    }
    try {
      auth0 = await createAuth0Client({
        ...auth0Config,
        cacheLocation: "localstorage",
        useRefreshTokens: true,
      });
      if (await auth0.isAuthenticated()) {
        token = await auth0.getTokenSilently({ audience: auth0Config.audience });
        const user = await auth0.getUser();
        who = user?.name || user?.nickname || user?.email || "";
      }
    } catch (err) {
      console.warn("auth0", err);
      token = "";
    }
    next();
  });

  function next() {
    errorText = "";
    if (!isCompleteCode(code)) {
      typed = code;
      phase = "enter";
      return;
    }
    if (!token) {
      phase = "signin";
      return;
    }
    lookup();
  }

  function submitCode() {
    const value = normalizeCode(typed);
    if (!isCompleteCode(value)) {
      errorText = "Enter the 8-character code shown in your terminal.";
      return;
    }
    code = value;
    store(code);
    next();
  }

  function onType(e) {
    typed = normalizeCode(e.target.value);
    e.target.value = typed;
  }

  async function signIn(otherAccount = false) {
    if (!auth0) {
      errorText = "Sign-in is unavailable right now. Reload the page and try again.";
      return;
    }
    store(code);
    busy = true;
    const opts = { redirect_uri: auth0Callback(), appState: { returnTo: ACTIVATE_RETURN } };
    if (otherAccount) opts.prompt = "login";
    try {
      await auth0.loginWithRedirect(opts);
    } catch (err) {
      busy = false;
      errorText = "Could not open sign-in.";
    }
  }

  async function lookup() {
    phase = "lookup";
    errorText = "";
    view = null;
    try {
      const res = await lookupDevice(token, code);
      view = res.data || null;
      phase = "confirm";
    } catch (err) {
      const status = err?.response?.status;
      if (status === 401) {
        token = "";
        phase = "signin";
        errorText = "Your session expired. Sign in again.";
        return;
      }
      errorText = apiError(err, "That code was not found or has expired.");
      typed = "";
      phase = "enter";
    }
  }

  async function confirm() {
    if (!view?.csrf || busy) return;
    busy = true;
    errorText = "";
    try {
      await confirmDevice(token, code, view.csrf);
      clearActivateCode(sessionStorage);
      phase = "done";
    } catch (err) {
      errorText = apiError(err, "Could not confirm. Try again.");
    } finally {
      busy = false;
    }
  }

  async function deny() {
    if (!view?.csrf || busy) return;
    busy = true;
    errorText = "";
    try {
      await denyDevice(token, code, view.csrf);
      clearActivateCode(sessionStorage);
      phase = "denied";
    } catch (err) {
      errorText = apiError(err, "Could not reject the sign-in.");
    } finally {
      busy = false;
    }
  }

  function otherCode() {
    code = "";
    typed = "";
    store("");
    errorText = "";
    phase = "enter";
  }

  function since(value) {
    const t = new Date(value).getTime();
    if (!value || Number.isNaN(t)) return "just now";
    const s = Math.max(0, Math.round((Date.now() - t) / 1000));
    if (s < 60) return "just now";
    const m = Math.round(s / 60);
    return m === 1 ? "1 minute ago" : `${m} minutes ago`;
  }
</script>

<main class="wrap">
  <section class="card" aria-live="polite">
    <div class="brand">
      <div class="mark">&gt;_</div>
      <div class="word">Veilspan</div>
    </div>

    {#if phase === "loading" || phase === "lookup"}
      <p class="lead">Sign in to connect your terminal</p>
      <p class="muted center">{phase === "lookup" ? "Checking the code…" : "Loading…"}</p>

    {:else if phase === "enter"}
      <p class="lead">Sign in to connect your terminal</p>
      <form on:submit|preventDefault={submitCode}>
        <label for="code">Code from your terminal</label>
        <input
          id="code"
          class="codebox"
          value={typed}
          on:input={onType}
          placeholder="XXXX-XXXX"
          autocomplete="one-time-code"
          autocapitalize="characters"
          spellcheck="false"
          inputmode="text"
          maxlength="9"
          autofocus
        />
        <button class="primary" type="submit" disabled={!isCompleteCode(typed)}>Continue</button>
      </form>

    {:else if phase === "signin"}
      <p class="lead">Sign in to connect your terminal</p>
      <div class="code">{code}</div>
      <p class="muted center">Check this matches the code in your terminal.</p>
      <button class="primary" on:click={() => signIn(false)} disabled={busy}>Sign in</button>
      <button class="link" on:click={otherCode}>Use a different code</button>

    {:else if phase === "confirm"}
      <p class="lead">Connect this terminal to your account?</p>
      <div class="code">{code}</div>
      <dl>
        {#if who}<dt>Account</dt><dd>{who}</dd>{/if}
        <dt>From</dt><dd>{view?.ip || "unknown address"}</dd>
        <dt>Started</dt><dd>{since(view?.created)}</dd>
      </dl>
      <p class="warn">Only confirm if you started this from your own terminal just now.</p>
      <button class="primary" on:click={confirm} disabled={busy || !view?.csrf}>Confirm</button>
      <div class="row">
        <button class="link" on:click={deny} disabled={busy}>Not me — reject</button>
        <button class="link" on:click={() => signIn(true)} disabled={busy}>Use another account</button>
      </div>

    {:else if phase === "done"}
      <div class="tick" aria-hidden="true">✓</div>
      <p class="lead">Done — return to your terminal.</p>
      <p class="muted center">You can close this page.</p>

    {:else if phase === "denied"}
      <p class="lead">Sign-in rejected.</p>
      <p class="muted center">That terminal was not connected. You can close this page.</p>
    {/if}

    {#if errorText}<p class="err">{errorText}</p>{/if}
  </section>
  <p class="foot">ssh -p 2222 veilspan.com</p>
</main>

<style>
  :global(html), :global(body) {
    margin: 0;
    min-height: 100%;
    background: radial-gradient(ellipse at top, #0f1620 0%, #06080c 60%) fixed, #06080c;
    color: #b0aca6;
    font-family: "Cormorant Garamond", Georgia, serif;
    -webkit-text-size-adjust: 100%;
  }
  .wrap {
    min-height: 100vh;
    min-height: 100dvh;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: max(1.25rem, env(safe-area-inset-top)) 1rem max(1.25rem, env(safe-area-inset-bottom));
    box-sizing: border-box;
  }
  .card {
    width: 100%;
    max-width: 400px;
    box-sizing: border-box;
    padding: 2rem 1.5rem 1.6rem;
    background: #111820;
    border: 1px solid rgba(61, 220, 132, 0.18);
    border-radius: 12px;
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.4), 0 24px 80px rgba(0, 0, 0, 0.6);
    display: flex;
    flex-direction: column;
    gap: 0.9rem;
  }
  .brand { display: flex; flex-direction: column; align-items: center; gap: 0.35rem; margin-bottom: 0.2rem; }
  .mark {
    font-family: "Fira Code", ui-monospace, monospace;
    color: #3ddc84;
    font-size: 1rem;
    border: 1px solid rgba(61, 220, 132, 0.35);
    border-radius: 8px;
    padding: 0.25rem 0.55rem;
    text-shadow: 0 0 10px rgba(61, 220, 132, 0.55);
  }
  .word {
    font-family: "Cinzel Decorative", "Cinzel", Georgia, serif;
    font-size: 1.9rem;
    letter-spacing: 0.12em;
    color: #e8e6e3;
  }
  .lead {
    margin: 0;
    text-align: center;
    font-family: "Cinzel", Georgia, serif;
    font-size: 1.05rem;
    color: #e8e6e3;
    line-height: 1.4;
  }
  .muted { margin: 0; color: #6b6660; font-size: 1.05rem; }
  .center { text-align: center; }
  .code {
    text-align: center;
    font-family: "Fira Code", ui-monospace, monospace;
    font-size: 1.7rem;
    letter-spacing: 0.14em;
    color: #7ae8a4;
    background: #0a0e14;
    border: 1px dashed rgba(61, 220, 132, 0.35);
    border-radius: 8px;
    padding: 0.6rem 0.4rem;
  }
  form { display: flex; flex-direction: column; gap: 0.6rem; }
  label { font-family: "Fira Code", ui-monospace, monospace; font-size: 0.75rem; letter-spacing: 0.08em; text-transform: uppercase; color: #6b6660; }
  .codebox {
    width: 100%;
    box-sizing: border-box;
    font-family: "Fira Code", ui-monospace, monospace;
    font-size: 1.6rem;
    letter-spacing: 0.14em;
    text-align: center;
    text-transform: uppercase;
    color: #e8e6e3;
    background: #0a0e14;
    border: 1px solid rgba(61, 220, 132, 0.3);
    border-radius: 8px;
    padding: 0.65rem 0.4rem;
    outline: none;
  }
  .codebox:focus { border-color: #3ddc84; box-shadow: 0 0 0 3px rgba(61, 220, 132, 0.15); }
  button { font: inherit; cursor: pointer; }
  button:disabled { opacity: 0.5; cursor: default; }
  .primary {
    width: 100%;
    min-height: 48px;
    border: 1px solid #3ddc84;
    border-radius: 8px;
    background: #1a5c38;
    color: #eafff1;
    font-family: "Cinzel", Georgia, serif;
    font-weight: 600;
    font-size: 1.05rem;
    letter-spacing: 0.06em;
  }
  .primary:not(:disabled):hover { background: #217046; }
  .link {
    background: none;
    border: none;
    color: #8a857e;
    font-size: 1rem;
    padding: 0.4rem;
    text-decoration: underline;
    text-underline-offset: 3px;
  }
  .row { display: flex; justify-content: space-between; flex-wrap: wrap; gap: 0.4rem; }
  dl { display: grid; grid-template-columns: auto 1fr; gap: 0.3rem 0.8rem; margin: 0; font-size: 1.05rem; }
  dt { color: #6b6660; }
  dd { margin: 0; color: #e8e6e3; overflow-wrap: anywhere; }
  .warn { margin: 0; color: #f0c674; font-size: 1rem; text-align: center; }
  .err { margin: 0; color: #ff7a7a; font-size: 1rem; text-align: center; }
  .tick { text-align: center; font-size: 2.4rem; color: #3ddc84; text-shadow: 0 0 14px rgba(61, 220, 132, 0.5); }
  .foot { margin: 1rem 0 0; font-family: "Fira Code", ui-monospace, monospace; font-size: 0.75rem; color: #3d3a36; letter-spacing: 0.08em; }
  @media (max-width: 380px) {
    .card { padding: 1.5rem 1.1rem 1.3rem; }
    .code, .codebox { font-size: 1.4rem; }
    .word { font-size: 1.6rem; }
  }
</style>
