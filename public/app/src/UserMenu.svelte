<script>
  import { onMount, onDestroy } from "svelte";
  import { getAuth } from "./auth.js";

  const { isLoading, isAuthenticated, login, logout, userInfo } = getAuth();
  let open = false;
  let menuRef;

  const toggle = () => {
    open = !open;
  };

  const handleClickOutside = (event) => {
    if (menuRef && !menuRef.contains(event.target)) {
      open = false;
    }
  };

  onMount(() => {
    document.addEventListener("click", handleClickOutside);
  });

  onDestroy(() => {
    document.removeEventListener("click", handleClickOutside);
  });
</script>

{#if $isLoading}
  <span class="text-xs text-slate-400">Loading...</span>
{:else if !$isAuthenticated}
  <div class="flex items-center gap-2">
    <button class="btn btn-outline" type="button" on:click={() => login()}>
      Log in
    </button>
    <button class="btn btn-primary" type="button" on:click={() => login()}>
      Signup
    </button>
  </div>
{:else}
  <div class="relative" bind:this={menuRef}>
    <button
      class="flex items-center gap-3 pl-4 border-l border-slate-200 dark:border-slate-800"
      type="button"
      on:click={toggle}
    >
      <img
        alt="User Avatar"
        class="w-8 h-8 rounded-full border border-slate-200 dark:border-slate-700"
        src={$userInfo?.picture || "https://avatars.githubusercontent.com/u/0?v=4"}
      />
      <span class="text-sm font-medium hidden sm:inline">
        {$userInfo?.name || "User"}
      </span>
      <span class="material-symbols-outlined text-slate-400">expand_more</span>
    </button>
    {#if open}
      <div
        class="absolute right-0 mt-2 w-40 rounded-lg border border-slate-200 bg-white shadow-lg dark:border-slate-800 dark:bg-slate-900"
      >
        <a
          href="/account"
          class="block px-4 py-2 text-sm text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800"
          >Profile</a
        >
        <button
          class="w-full text-left px-4 py-2 text-sm text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800"
          type="button"
          on:click={() => logout()}
        >
          Logout
        </button>
      </div>
    {/if}
  </div>
{/if}
