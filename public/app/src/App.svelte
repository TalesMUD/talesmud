<script>
  import { Router } from "yrv";
  import AppContent from "./AppContent.svelte";
  import UserMenu from "./UserMenu.svelte";
  import { createAuth } from "./auth.js";
  import { onDestroy, onMount } from "svelte";

  const config = {
    domain: import.meta.env.VITE_AUTH0_DOMAIN || "owndnd.eu.auth0.com",
    client_id:
      import.meta.env.VITE_AUTH0_CLIENT_ID ||
      "mxcEqTuAUOzrL798mbVTpqFxpGGVp3gI",
    audience:
      import.meta.env.VITE_AUTH0_AUDIENCE ||
      "http://talesofapirate.com/dnd/api",
  };

  const { isAuthenticated, isLoading } = createAuth(config);

  let playMenuOpen = false;
  let playMenuEl;

  function togglePlayMenu(e) {
    e.preventDefault();
    e.stopPropagation();
    playMenuOpen = !playMenuOpen;
  }

  function onDocumentClick(e) {
    if (!playMenuEl) return;
    if (!playMenuEl.contains(e.target)) playMenuOpen = false;
  }

  onMount(() => {
    document.addEventListener("click", onDocumentClick);
  });
  onDestroy(() => {
    document.removeEventListener("click", onDocumentClick);
  });
</script>

<Router>
  <nav class="border-b border-slate-200 bg-white px-6 py-3 dark:border-slate-800 dark:bg-slate-900 sticky top-0 z-50">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-8">
        <a href="/" class="flex items-center gap-2 font-bold text-xl tracking-tight">
          <span class="material-symbols-outlined text-primary">auto_stories</span>
          <span>Tales</span>
        </a>
        <div class="hidden md:flex items-center gap-6 text-sm font-medium text-slate-500 dark:text-slate-400">
          <div class="relative flex items-center gap-1" bind:this={playMenuEl}>
            <a class="hover:text-primary transition-colors" href="/play">Play</a>
            <button
              class="p-1 rounded hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
              type="button"
              aria-haspopup="menu"
              aria-expanded={playMenuOpen}
              on:click={togglePlayMenu}
            >
              <span class="material-symbols-outlined text-base">arrow_drop_down</span>
            </button>
            {#if playMenuOpen}
              <div
                class="absolute left-0 top-full mt-2 w-48 rounded-lg border border-slate-200 bg-white shadow-lg dark:border-slate-800 dark:bg-slate-900 overflow-hidden"
                role="menu"
              >
                <a
                  class="block px-3 py-2 text-sm hover:bg-slate-50 dark:hover:bg-slate-800/60"
                  href="/play"
                  role="menuitem"
                >
                  Start playing
                </a>
                <a
                  class="block px-3 py-2 text-sm hover:bg-slate-50 dark:hover:bg-slate-800/60"
                  href="/characters/new"
                  role="menuitem"
                >
                  New Character
                </a>
              </div>
            {/if}
          </div>
          {#if $isAuthenticated}
            <a class="hover:text-primary transition-colors" href="/list">Top Characters</a>
            <a class="hover:text-primary transition-colors" href="/creator/rooms">Creator</a>
          {/if}
          <a class="hover:text-primary transition-colors" href="/news">News</a>
        </div>
      </div>
      <div class="flex items-center gap-4">
        <button class="p-2 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-full transition-colors" type="button">
          <span class="material-symbols-outlined">notifications</span>
        </button>
        <UserMenu />
      </div>
    </div>
  </nav>

  <AppContent />
</Router>
