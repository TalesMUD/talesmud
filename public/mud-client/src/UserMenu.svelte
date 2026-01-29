<style>
  .userbutton {
    margin-right: 1em;
  }

  .btn-small {
    color: #eee;
  }

  .img {
    width: 42px;
  }
</style>

<script>
  import { onMount } from "svelte";
  import { UserIcon } from "svelte-feather-icons";

  import { getAuth } from "./auth.js";
  import { getUser, updateUser } from "./api/user.js";
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

  async function signup() {
    await login();
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
{:else if !$isAuthenticated}
  <li class="right-align">
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-noninteractive-element-interactions -->
    <p on:click="{() => signup()}" class="btn-small userbutton green">Signup</p>
  </li>
  <li class="right-align">
    <button on:click="{() => login()}" class="btn-small userbutton green">
      Log in
    </button>
  </li>
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
