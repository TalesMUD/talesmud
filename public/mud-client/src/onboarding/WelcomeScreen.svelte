<style>
  .welcome-screen {
    position: fixed;
    inset: 0;
    background: #0a0e14;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
  }

  .bg-image {
    position: absolute;
    inset: 0;
    background-image: url('img/bg/sample-tavern.png');
    background-size: cover;
    background-position: center;
    image-rendering: pixelated;
    opacity: 0.08;
    filter: blur(4px) saturate(0.3) brightness(0.7);
  }

  .bg-gradient {
    position: absolute;
    inset: 0;
    background: radial-gradient(ellipse 70% 60% at 50% 45%, transparent 0%, #0a0e14 100%);
  }

  .card {
    position: relative;
    z-index: 3;
    background: rgba(0, 0, 0, 0.75);
    backdrop-filter: blur(12px);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 12px;
    padding: 3rem 3.5rem;
    max-width: 440px;
    width: 90vw;
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.2rem;
    opacity: 0;
    animation: fadeSlideIn 0.5s ease forwards;
    animation-delay: 0.15s;
  }

  .card.picker-open {
    max-width: 760px;
    max-height: calc(100vh - 2rem);
    overflow: auto;
  }

  @keyframes fadeSlideIn {
    from {
      opacity: 0;
      transform: translateY(12px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .icon {
    color: #f59e0b;
    font-size: 2.4rem;
    margin-bottom: 0.25rem;
  }

  .title {
    font-family: 'Cinzel', serif;
    font-size: 1.6rem;
    font-weight: 600;
    color: #e5e7eb;
    letter-spacing: 0.06em;
    margin: 0;
    line-height: 1.3;
  }

  .subtitle {
    font-size: 0.9rem;
    color: #9ca3af;
    line-height: 1.5;
    max-width: 340px;
  }

  .divider {
    width: 60px;
    height: 1px;
    background: rgba(255, 255, 255, 0.1);
  }

  .buttons {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    width: 100%;
    max-width: 280px;
    margin-top: 0.5rem;
  }

  .btn-welcome {
    font-size: 0.85rem;
    font-weight: 500;
    letter-spacing: 0.03em;
    padding: 0.75rem 1.5rem;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s ease;
    text-decoration: none;
    width: 100%;
    box-sizing: border-box;
  }

  .btn-welcome.primary {
    border: none;
    color: #fff;
    background: #16a34a;
  }

  .btn-welcome.primary:hover {
    background: #15803d;
    box-shadow: 0 4px 12px rgba(22, 163, 74, 0.3);
    transform: translateY(-1px);
  }

  .btn-welcome.secondary {
    border: 1px solid rgba(255, 255, 255, 0.16);
    color: #9ca3af;
    background: transparent;
    font-size: 0.75rem;
    padding: 0.45rem 1rem;
  }

  .btn-welcome.secondary:hover {
    background: rgba(255, 255, 255, 0.06);
    border-color: rgba(255, 255, 255, 0.28);
    color: #e5e7eb;
  }

  .btn-welcome.guest {
    border: 1px solid rgba(245, 158, 11, 0.3);
    color: #f59e0b;
    background: rgba(245, 158, 11, 0.08);
  }

  .btn-welcome.guest:hover {
    background: rgba(245, 158, 11, 0.15);
    border-color: rgba(245, 158, 11, 0.5);
    transform: translateY(-1px);
  }

  .btn-welcome.guest:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
  }

  .guest-note {
    font-size: 0.7rem;
    color: #6b7280;
    text-align: center;
  }

  .error-banner {
    font-size: 0.8rem;
    color: #f87171;
    background: rgba(248, 113, 113, 0.1);
    border: 1px solid rgba(248, 113, 113, 0.2);
    padding: 0.5rem 1rem;
    border-radius: 6px;
    width: 100%;
    max-width: 280px;
    text-align: center;
  }

  .guest-picker {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 0.55rem;
    text-align: left;
  }

  .guest-picker-title {
    font-size: 0.75rem;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: #9ca3af;
    text-align: center;
  }

  .guest-class-grid,
  .guest-race-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 0.4rem;
  }

  .guest-portrait {
    width: 72px;
    height: 88px;
    object-fit: cover;
    object-position: center 12%;
    image-rendering: pixelated;
    border-radius: 4px;
    display: block;
    margin-bottom: 0.35rem;
    background: #0c1016;
  }

  .guest-choice {
    text-align: left;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 6px;
    color: #e5e7eb;
    padding: 0.45rem 0.55rem;
    cursor: pointer;
  }

  .guest-choice.selected {
    border-color: rgba(245, 158, 11, 0.55);
    background: rgba(245, 158, 11, 0.1);
  }

  .guest-choice strong {
    display: block;
    font-size: 0.82rem;
  }

  .guest-choice span {
    display: block;
    margin-top: 0.15rem;
    font-size: 0.72rem;
    color: #9ca3af;
    line-height: 1.3;
  }

  @media (max-width: 520px) {
    .welcome-screen {
      overflow-y: auto;
      align-items: flex-start;
      padding: 1.25rem 0 2rem;
    }

    .card {
      padding: 1.75rem 1.25rem;
      gap: 0.9rem;
    }

    .title {
      font-size: 1.35rem;
    }
  }
</style>

<script>
  import { onMount } from "svelte";
  import { getCharacterTemplates } from "../api/characters.js";
  import { ensureClassCatalog } from "./classCatalog.js";
  import { catalogFallbackTemplates, guestPickerEnabled, originPortraitSrc, racesForTemplate } from "./raceAllow.js";

  export let login;
  export let serverName = "Tales";
  export let authError = null;
  export let onGuestPlay = null;

  let guestLoading = false;
  let guestError = null;
  const showGuestPicker = guestPickerEnabled(typeof location !== "undefined" ? location.hostname : "");
  let guestTemplates = [];
  let guestTemplate = null;
  let guestRaceId = "";

  onMount(() => {
    if (!showGuestPicker) return;
    ensureClassCatalog().finally(() => {
      getCharacterTemplates(
        (result) => {
          guestTemplates = (result && result.length) ? result : catalogFallbackTemplates();
        },
        () => { guestTemplates = catalogFallbackTemplates(); }
      );
    });
  });

  function chooseGuestTemplate(template) {
    guestTemplate = template;
    const races = racesForTemplate(template);
    if (!races.some((race) => race.id === guestRaceId)) {
      guestRaceId = races[0]?.id || "";
    }
  }

  $: guestRaces = guestTemplate ? racesForTemplate(guestTemplate) : [];

  // A named connection skips Auth0 universal login. Email is the only
  // button that opens that page, where the browser can autofill a password.
  function loginWith(connection) {
    if (connection) login(undefined, { connection });
    else login();
  }

  function startGuest(pick) {
    if (!onGuestPlay) return;
    guestLoading = true;
    guestError = null;
    onGuestPlay(
      () => { guestLoading = false; },
      (err) => {
        guestLoading = false;
        guestError = err;
      },
      pick
    );
  }

  function handleGuest() {
    startGuest();
  }

  function handleGuestChosen() {
    if (!guestTemplate || !guestRaceId) return;
    startGuest({ templateId: guestTemplate.id, race: guestRaceId });
  }
</script>

<div class="welcome-screen">
  <div class="bg-image"></div>
  <div class="bg-gradient"></div>

  <div class="card" class:picker-open={showGuestPicker}>
    <i class="material-icons icon">auto_stories</i>

    <h1 class="title">{serverName}</h1>

    <p class="subtitle">
      A multiplayer text adventure. Create a character, explore the world, and forge your own legend.
    </p>

    <div class="divider"></div>

    {#if authError}
      <div class="error-banner">
        Authentication failed. Please try again.
      </div>
    {/if}

    {#if guestError}
      <div class="error-banner">
        {guestError}
      </div>
    {/if}

    <div class="buttons">
      <button class="btn-welcome primary" type="button" on:click={() => loginWith("twitter")}>
        Continue with X
      </button>
      <button class="btn-welcome primary" type="button" on:click={() => loginWith("google-oauth2")}>
        Continue with Google
      </button>
      <button class="btn-welcome secondary" type="button" on:click={() => loginWith()}>
        Email and password
      </button>

      <div class="divider" style="width: 100%; margin: 0.25rem 0;"></div>

      <button
        class="btn-welcome guest"
        on:click={handleGuest}
        disabled={guestLoading}
      >
        {guestLoading ? 'Starting...' : 'Play as guest'}
      </button>
      <span class="guest-note">
        30 min session, no login required
      </span>
    </div>

    {#if showGuestPicker}
      <div class="guest-picker">
        <div class="guest-picker-title">Or pick a class and race</div>
        {#if guestTemplates.length === 0}
          <div class="guest-note">Loading classes...</div>
        {/if}
        <div class="guest-class-grid">
          {#each guestTemplates as template}
            <button
              type="button"
              class="guest-choice"
              class:selected={guestTemplate && guestTemplate.id === template.id}
              on:click={() => chooseGuestTemplate(template)}
            >
              {#if originPortraitSrc(template, guestTemplate && guestTemplate.id === template.id ? guestRaceId : "")}
                <img
                  class="guest-portrait"
                  alt=""
                  src={originPortraitSrc(template, guestTemplate && guestTemplate.id === template.id ? guestRaceId : "")}
                />
              {/if}
              <strong>{template.name}</strong>
              <span>{template.description}</span>
            </button>
          {/each}
        </div>
        {#if guestTemplate}
          <div class="guest-race-grid">
            {#each guestRaces as race}
              <button
                type="button"
                class="guest-choice"
                class:selected={guestRaceId === race.id}
                on:click={() => guestRaceId = race.id}
              >
                <strong>{race.name}</strong>
                <span>{race.blurb}</span>
              </button>
            {/each}
          </div>
          <button
            class="btn-welcome guest"
            type="button"
            on:click={handleGuestChosen}
            disabled={guestLoading || !guestRaceId}
          >
            Play this guest
          </button>
        {/if}
      </div>
    {/if}
  </div>
</div>
