<script>
  import { Router } from "yrv";
  import AppContent from "./AppContent.svelte";
  import UserMenu from "./UserMenu.svelte";
  import { createAuth } from "./auth.js";
  import { getUser } from "./api/user.js";
  import { getServerInfo } from "./api/live.js";
  import { userRole } from "./stores.js";
  import { creatorDrawerOpen, creatorNavNarrow } from "./creator/creatorNavStore.js";
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

  const { isAuthenticated, isLoading, authToken } = createAuth(config);

  // User role tracking. The store is shared so Creator can hide admin-only tools.
  let userRoleLoaded = false;

  $: isCreator = $userRole === "creator" || $userRole === "admin";
  $: isAdmin = $userRole === "admin";

  // Load user data (including role) after authentication
  $: if ($isAuthenticated && $authToken && !userRoleLoaded) {
    userRoleLoaded = true;
    getUser(
      $authToken,
      (u) => {
        userRole.set(u.role || "player");
      },
      (err) => console.error("Failed to load user role:", err)
    );
  }

  let playMenuOpen = false;
  let playMenuEl;
  let adminMenuOpen = false;
  let adminMenuEl;
  let navEl;
  let pathname = typeof window !== "undefined" ? window.location.pathname : "";

  if (typeof window !== "undefined" && window.matchMedia("(max-width: 1023px)").matches) {
    creatorNavNarrow.set(true);
  }
  let envLabel = "";
  let envHost = "";
  let removeNavWatch = () => {};

  function applyServerInfo(info) {
    let payload = info;
    if (typeof payload === "string") {
      try {
        payload = JSON.parse(payload);
      } catch {
        payload = null;
      }
    }
    envLabel = String(payload?.envLabel || "").trim();
    envHost = String(payload?.host || "").trim();
  }

  // Not inside onMount. A client build that resolves Svelte's SSR entry makes
  // onMount a no-op, and Rollup then deletes the callback — which removed this
  // request, and the badge, from the production bundle.
  if (typeof window !== "undefined") {
    getServerInfo().then(applyServerInfo).catch(() => {
      envLabel = "";
    });
  }

  $: onCreator = pathname.startsWith("/creator");

  $: if (!onCreator && $creatorDrawerOpen) creatorDrawerOpen.set(false);

  function togglePlayMenu(e) {
    e.preventDefault();
    e.stopPropagation();
    playMenuOpen = !playMenuOpen;
  }

  function toggleAdminMenu(e) {
    e.preventDefault();
    e.stopPropagation();
    adminMenuOpen = !adminMenuOpen;
  }

  function onDocumentClick(e) {
    if (!playMenuEl) return;
    if (!playMenuEl.contains(e.target)) playMenuOpen = false;
    if (adminMenuEl && !adminMenuEl.contains(e.target)) adminMenuOpen = false;
  }

  function syncPath() {
    const next = window.location.pathname || "/";
    if (next !== pathname) creatorDrawerOpen.set(false);
    pathname = next;
  }

  function syncNavHeight() {
    if (!navEl) return;
    document.documentElement.style.setProperty("--app-nav-height", `${navEl.offsetHeight}px`);
  }

  function toggleCreatorDrawer() {
    creatorDrawerOpen.update((open) => !open);
  }

  onMount(() => {
    document.addEventListener("click", onDocumentClick);
    syncPath();
    window.addEventListener("popstate", syncPath);
    const mq = window.matchMedia("(max-width: 1023px)");
    const applyNarrow = () => {
      creatorNavNarrow.set(mq.matches);
      if (!mq.matches) creatorDrawerOpen.set(false);
    };
    applyNarrow();
    mq.addEventListener("change", applyNarrow);
    syncNavHeight();
    const observer = new ResizeObserver(syncNavHeight);
    if (navEl) observer.observe(navEl);
    removeNavWatch = () => {
      window.removeEventListener("popstate", syncPath);
      mq.removeEventListener("change", applyNarrow);
      observer.disconnect();
    };
  });
  onDestroy(() => {
    document.removeEventListener("click", onDocumentClick);
    removeNavWatch();
  });
</script>

<Router>
  <nav
    class="sticky top-0 z-50 border-b border-slate-200 bg-white px-4 py-3 dark:border-slate-800 dark:bg-slate-900 sm:px-6"
    bind:this={navEl}
  >
    <div class="flex items-center gap-2 sm:gap-3">
      {#if onCreator && $creatorNavNarrow}
        <button
          id="creator-nav-toggle"
          class="rounded-md p-1.5 text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800 lg:hidden"
          type="button"
          aria-label={$creatorDrawerOpen ? "Close creator navigation" : "Open creator navigation"}
          aria-expanded={$creatorDrawerOpen}
          aria-controls="creator-nav-drawer"
          on:click={toggleCreatorDrawer}
        >
          <span class="material-symbols-outlined">menu</span>
        </button>
      {/if}
      <div class="flex min-w-0 items-center gap-8">
        <a href="/" class="flex shrink-0 items-center gap-2 text-xl font-bold tracking-tight">
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
            {#if isCreator}
              <a class="hover:text-primary transition-colors" href="/creator/rooms">Creator</a>
            {/if}
            {#if isAdmin}
              <div class="relative flex items-center gap-1" bind:this={adminMenuEl}>
                <a class="hover:text-primary transition-colors" href="/manage/users">Admin</a>
                <button
                  class="p-1 rounded hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                  type="button"
                  aria-haspopup="menu"
                  aria-expanded={adminMenuOpen}
                  on:click={toggleAdminMenu}
                >
                  <span class="material-symbols-outlined text-base">arrow_drop_down</span>
                </button>
                {#if adminMenuOpen}
                  <div
                    class="absolute left-0 top-full mt-2 w-48 rounded-lg border border-slate-200 bg-white shadow-lg dark:border-slate-800 dark:bg-slate-900 overflow-hidden"
                    role="menu"
                  >
                    <a
                      class="block px-3 py-2 text-sm hover:bg-slate-50 dark:hover:bg-slate-800/60"
                      href="/manage/users"
                      role="menuitem"
                    >
                      User Management
                    </a>
                    <a
                      class="block px-3 py-2 text-sm hover:bg-slate-50 dark:hover:bg-slate-800/60"
                      href="/manage/guest-stats"
                      role="menuitem"
                    >
                      Guest Statistics
                    </a>
                  </div>
                {/if}
              </div>
            {/if}
          {/if}
          <a class="hover:text-primary transition-colors" href="/news">News</a>
        </div>
      </div>
      {#if onCreator}
        <label class="ml-auto hidden min-w-0 max-w-md flex-1 items-center sm:flex">
          <span class="sr-only">Search</span>
          <span class="relative block w-full">
            <span class="material-symbols-outlined pointer-events-none absolute left-2 top-1/2 -translate-y-1/2 text-base text-slate-400" aria-hidden="true">search</span>
            <input
              class="h-8 w-full rounded-md border border-slate-200 bg-slate-50 pl-8 pr-3 text-xs text-slate-700 placeholder:text-slate-400 focus:border-primary focus:ring-1 focus:ring-primary dark:border-slate-700 dark:bg-slate-950 dark:text-slate-200"
              type="search"
              placeholder="Search rooms, NPCs, items…"
              aria-label="Search"
              autocomplete="off"
              data-creator-search
            />
          </span>
        </label>
      {/if}
      <div class="ml-auto flex shrink-0 items-center gap-2 whitespace-nowrap sm:gap-3 {onCreator ? 'sm:ml-0' : ''}">
        {#if envLabel}
          <span
            class="relative z-10 shrink-0 whitespace-nowrap rounded bg-red-600 px-2 py-1 font-display text-[10px] font-bold uppercase tracking-widest text-white"
            data-env-badge
            role="status"
          >
            {envLabel}{envHost ? ` · ${envHost}` : ""}
          </span>
        {/if}
        <button class="shrink-0 rounded-full p-2 transition-colors hover:bg-slate-100 dark:hover:bg-slate-800" type="button">
          <span class="material-symbols-outlined">notifications</span>
        </button>
        <UserMenu />
      </div>
    </div>
  </nav>

  <AppContent />
</Router>
