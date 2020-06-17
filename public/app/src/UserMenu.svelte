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
    getUser(
      $authToken,
      (u) => {
        user.set(u);
      },
      (err) => console.log(err)
    );
  }

  async function signup() {
    await login();

    if ($isAuthenticated) {
      await loadUserData();
    }
  }

  onMount(() => {
    document.addEventListener("DOMContentLoaded", function () {
      var elems = document.querySelectorAll(".dropdown-trigger");
      var instances = M.Dropdown.init(elems);
    });
  });
</script>

<!-- Dropdown Structure -->
<ul id="dropdown1" class="dropdown-content">
  <li>
    <a href="#!" on:click="{() => logout()}">logout</a>
  </li>
  <li>
    <a href="account">profile</a>
  </li>
</ul>

{#if $isLoading}
  <li class="right-align">...</li>
{:else if !$isAuthenticated}
  <li class="right-align">
    <p on:click="{() => signup()}" class="btn-small userbutton green">Signup</p>
  </li>
  <li class="right-align">
    <button on:click="{() => login()}" class="btn-small userbutton green">
      Log in
    </button>
  </li>
{/if}
<li>

  <a class="dropdown-trigger" href="#" data-target="dropdown1">
    <span class="valign-wrapper">

      {#if $isAuthenticated}
        <img src="{$userInfo.picture}" alt="" class="circle img " />
        <i class="material-icons left">arrow_drop_down</i>
      {/if}
    </span>

  </a>
</li>
