<style>
  .loginText {
    color: greenyellow;
  }

  .modal {
    background-color: darkslategrey;
  }
</style>

<script>
  import CharacterCreator from "./characters/CharacterCreator.svelte";
  import { onMount } from "svelte";

  import { createAuth, getAuth } from "./auth.js";
  import axios from "axios";
  import { onInterval } from "./utils.js";
  import { getUser } from "./api/user.js";
  import { writable } from "svelte/store";

  const nickname = writable("stranger");
  let loaded = false;

  const {
    isLoading,
    isAuthenticated,
    login,
    logout,
    authToken,
    authError,
    userInfo,
  } = getAuth();

  $: state = {
    isLoading: $isLoading,
    isAuthenticated: $isAuthenticated,
    authError: $authError,
    userInfo: $userInfo ? $userInfo.name : null,
    authToken: $authToken.slice(0, 20),
    nickname: $nickname,
  };

  $: {
    if (!loaded && !$isLoading && $isAuthenticated) {
      loaded = true;
      getUser(
        $authToken,
        (user) => {
          nickname.set(user.nickname);
        },
        (err) => console.log(err)
      );
    }
  }

  const loadUser = async () => {};

  onMount(async () => {
    document.addEventListener("DOMContentLoaded", function () {
      var elems = document.querySelectorAll(".modal");
      var instances = M.Modal.init(elems, {});
    });

    await loadUser();
  });
</script>

<div class="center-align">

  {#if $isAuthenticated}
    <h4>Welcome back {$nickname}</h4>
    <div>
      Start playing
      <button class="btn green">Play</button>
    </div>

    <div>
      <p>
        Or create a new Character
        <a class="waves-effect waves-light btn modal-trigger" href="#modal1">
          Modal
        </a>
      </p>

    </div>
  {:else if !$isAuthenticated}
    <h4>Welcome Stranger</h4>
    <p>
      Please
      <a href="/login" class="loginText">log</a>
      in or
      <a href="/signup" class="loginText">signup</a>
    </p>
  {/if}

  <h4>TalesMUD</h4>
  <p>What is TalesMUD about</p>

  <h4>News</h4>
  <ul>
    <li>Update #2</li>
  </ul>

  <CharacterCreator />

</div>
