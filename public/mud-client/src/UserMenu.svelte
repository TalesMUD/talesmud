<style>
  .img {
    width: 42px;
  }

  /* ── Dark theme overrides for Materialize dropdown ── */
  :global(#dropdown1.dropdown-content) {
    background: #111820;
    border: 1px solid rgba(61, 220, 132, 0.12);
    border-radius: 6px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
    min-width: 180px;
  }

  :global(#dropdown1.dropdown-content li > a) {
    color: #ccc8c2;
    font-size: 0.9rem;
    display: flex;
    align-items: center;
    padding: 0.6rem 1rem;
  }

  :global(#dropdown1.dropdown-content li > a:hover) {
    background: rgba(61, 220, 132, 0.08);
    color: #e8e6e3;
  }

  :global(#dropdown1.dropdown-content li > a > i) {
    color: rgba(61, 220, 132, 0.5);
  }

  :global(#dropdown1.dropdown-content li > a:hover > i) {
    color: #7ae8a4;
  }

  :global(#dropdown1.dropdown-content li.divider) {
    background-color: rgba(255, 255, 255, 0.06);
  }
</style>

<script>
  import { onMount } from "svelte";

  import { getAuth } from "./auth.js";
  import { getUser } from "./api/user.js";
  import { user } from "./stores.js";
  import { layoutStore } from "./game/layout/LayoutStore.js";
  import { settingsStore } from "./game/SettingsStore.js";

  const {
    isLoading,
    isAuthenticated,
    login,
    logout,
    authToken,
    authError,
    userInfo,
  } = getAuth();

  async function loadUserData() {
    if (!$authToken) return;

    getUser(
      $authToken,
      (u) => {
        user.set(u);
      },
      (err) => console.error("Failed to load user data:", err)
    );
  }

  // Load user data whenever authToken changes and user is authenticated
  $: if ($isAuthenticated && $authToken) {
    loadUserData();
  }

  onMount(() => {
    // Initialize Materialize dropdown
    const elems = document.querySelectorAll(".dropdown-trigger");
    if (elems.length > 0 && typeof M !== "undefined") {
      M.Dropdown.init(elems);
    }
  });
</script>

<!-- Dropdown Structure -->
<ul id="dropdown1" class="dropdown-content">
  <li>
    <!-- svelte-ignore a11y-invalid-attribute -->
    <a href="#!" on:click="{() => layoutStore.enterEditMode()}">
      <i class="material-icons" style="font-size: 1.2em; vertical-align: middle; margin-right: 0.5em;">dashboard_customize</i>
      Edit Layout
    </a>
  </li>
  <li>
    <!-- svelte-ignore a11y-invalid-attribute -->
    <a href="#!" on:click="{() => settingsStore.openModal()}">
      <i class="material-icons" style="font-size: 1.2em; vertical-align: middle; margin-right: 0.5em;">settings</i>
      Settings
    </a>
  </li>
  {#if $user && ($user.role === "creator" || $user.role === "admin")}
    <li>
      <a href="/creator" target="_blank">
        <i class="material-icons" style="font-size: 1.2em; vertical-align: middle; margin-right: 0.5em;">public</i>
        World Builder
      </a>
    </li>
  {/if}
  <li class="divider"></li>
  <li>
    <!-- svelte-ignore a11y-invalid-attribute -->
    <a href="#!" on:click="{() => logout()}">
      <i class="material-icons" style="font-size: 1.2em; vertical-align: middle; margin-right: 0.5em;">logout</i>
      Logout
    </a>
  </li>
</ul>

{#if $isLoading}
  <li class="right-align">...</li>
{/if}
<li>

  <!-- svelte-ignore a11y-invalid-attribute -->
  <a class="dropdown-trigger" href="#" data-target="dropdown1">
    <span class="valign-wrapper">

      {#if $isAuthenticated}
        <img src="{$userInfo.picture}" alt="" class="circle img " />
        <i class="material-icons left">arrow_drop_down</i>
      {/if}
    </span>

  </a>
</li>
