<script>
  import { onDestroy, onMount, tick } from "svelte";
  import { userRole } from "../stores.js";
  import OpsToast from "./OpsToast.svelte";
  import SearchPalette from "./SearchPalette.svelte";
  import CreatorNavList from "./CreatorNavList.svelte";
  import { creatorDrawerOpen, creatorNavNarrow } from "./creatorNavStore.js";
  import {
    CREATOR_NAV,
    loadNavState,
    saveNavState,
    shouldToggleRail,
    toggleCollapsed,
    toggleGroup,
    visibleNav,
    withActiveGroupOpen,
  } from "./navState.js";

  function readInitial() {
    try {
      if (typeof localStorage === "undefined") return { collapsed: false, groups: {} };
      return loadNavState(localStorage);
    } catch {
      return { collapsed: false, groups: {} };
    }
  }

  let pathname = typeof window !== "undefined" ? window.location.pathname : "/";
  let navState = readInitial();
  let lastExpandedPath = "";
  let drawerEl;

  $: visible = visibleNav(CREATOR_NAV, $userRole === "admin");

  $: if (pathname !== lastExpandedPath) {
    lastExpandedPath = pathname;
    const next = withActiveGroupOpen(navState, CREATOR_NAV, pathname);
    if (next !== navState) {
      navState = next;
      persist(navState);
    }
  }

  function persist(state) {
    try {
      if (typeof localStorage === "undefined") return;
      saveNavState(localStorage, state);
    } catch {
      // Ignore private-mode and quota failures. The rail still works for this visit.
    }
  }

  function syncPath() {
    pathname = window.location.pathname || "/";
  }

  function toggleRail() {
    navState = toggleCollapsed(navState);
    persist(navState);
  }

  function onToggleGroup(event) {
    navState = toggleGroup(navState, event.detail);
    persist(navState);
  }

  function closeDrawer(restoreFocus) {
    creatorDrawerOpen.set(false);
    if (!restoreFocus) return;
    tick().then(() => {
      document.getElementById("creator-nav-toggle")?.focus();
    });
  }

  function onKeydown(event) {
    if ($creatorDrawerOpen && event.key === "Tab") {
      trapDrawerTab(event);
      return;
    }
    if (event.key === "Escape" && $creatorDrawerOpen) {
      event.preventDefault();
      closeDrawer(true);
      return;
    }
    if (!shouldToggleRail(event)) return;
    event.preventDefault();
    if ($creatorNavNarrow) creatorDrawerOpen.update((open) => !open);
    else toggleRail();
  }

  function trapDrawerTab(event) {
    if (event.key !== "Tab" || !drawerEl) return;
    const nodes = drawerEl.querySelectorAll("a, button, input, select, textarea");
    const list = [...nodes].filter((el) => !el.disabled && el.getAttribute("tabindex") !== "-1");
    if (!list.length) return;
    const first = list[0];
    const last = list[list.length - 1];
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  }

  $: if ($creatorDrawerOpen && $creatorNavNarrow) {
    tick().then(() => {
      drawerEl?.querySelector("[data-drawer-close]")?.focus();
    });
  }

  onMount(() => {
    syncPath();
    window.addEventListener("popstate", syncPath);
  });

  onDestroy(() => {
    if (typeof window === "undefined") return;
    window.removeEventListener("popstate", syncPath);
  });
</script>

<svelte:window on:keydown={onKeydown} />

<div class="flex h-[calc(100vh-var(--app-nav-height,72px))] min-h-0 overflow-hidden bg-slate-50 dark:bg-slate-950">
  <nav
    id="creator-nav-rail"
    class="hidden h-full min-h-0 shrink-0 flex-col border-r border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-950 lg:flex {navState.collapsed
      ? 'w-14'
      : 'w-56'}"
    aria-label="Creator"
  >
    <div class="flex shrink-0 items-center px-2 py-2 {navState.collapsed ? 'justify-center' : 'justify-between'}">
      {#if !navState.collapsed}
        <span class="px-1 text-[10px] font-bold uppercase tracking-widest text-slate-500">Creator</span>
      {/if}
      <button
        type="button"
        class="rounded-md p-1.5 text-slate-500 hover:bg-slate-100 hover:text-slate-800 dark:hover:bg-slate-800 dark:hover:text-slate-100"
        aria-label={navState.collapsed ? "Expand sidebar" : "Collapse sidebar"}
        aria-expanded={!navState.collapsed}
        aria-controls="creator-nav-rail"
        aria-keyshortcuts="Control+B Meta+B"
        title={navState.collapsed ? "Expand sidebar (Ctrl+B)" : "Collapse sidebar (Ctrl+B)"}
        on:click={toggleRail}
      >
        <span class="material-symbols-outlined text-[20px]" aria-hidden="true">
          {navState.collapsed ? "menu" : "menu_open"}
        </span>
      </button>
    </div>
    <div class="min-h-0 flex-1 overflow-y-auto">
      <CreatorNavList
        groups={visible}
        {pathname}
        collapsed={navState.collapsed}
        folded={navState.groups}
        idPrefix="rail-group"
        on:toggleGroup={onToggleGroup}
      />
    </div>
  </nav>

  {#if $creatorDrawerOpen && $creatorNavNarrow}
    <button
      type="button"
      class="fixed bottom-0 left-0 right-0 z-40 bg-slate-950/60"
      style="top: var(--app-nav-height, 72px)"
      aria-label="Close navigation"
      on:click={() => closeDrawer(true)}
    ></button>
    <aside
      id="creator-nav-drawer"
      class="fixed bottom-0 left-0 z-50 flex w-72 max-w-[86vw] flex-col border-r border-slate-200 bg-white shadow-xl dark:border-slate-800 dark:bg-slate-950"
      style="top: var(--app-nav-height, 72px)"
      role="dialog"
      aria-modal="true"
      aria-label="Creator"
      bind:this={drawerEl}
    >
      <div class="flex shrink-0 items-center justify-between px-3 py-2">
        <span class="text-[10px] font-bold uppercase tracking-widest text-slate-500">Creator</span>
        <button
          type="button"
          class="rounded-md p-1.5 text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800"
          data-drawer-close
          aria-label="Close navigation"
          on:click={() => closeDrawer(true)}
        >
          <span class="material-symbols-outlined text-[20px]" aria-hidden="true">close</span>
        </button>
      </div>
      <div class="min-h-0 flex-1 overflow-y-auto">
        <CreatorNavList
          groups={visible}
          {pathname}
          collapsed={false}
          folded={navState.groups}
          idPrefix="drawer-group"
          on:toggleGroup={onToggleGroup}
          on:navigate={() => closeDrawer(false)}
        />
      </div>
    </aside>
  {/if}

  <div class="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
    <!-- Ctrl+K opens search. The top-bar input is wired after the nav rebase. -->
    <SearchPalette />
    <OpsToast />
    <div class="h-full min-h-0 min-w-0 flex-1 overflow-auto">
      <slot />
    </div>
  </div>
</div>
