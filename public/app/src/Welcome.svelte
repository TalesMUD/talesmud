<style>
  .loginText {
    color: greenyellow;
  }

  .modal {
    background-color: darkslategrey;
  }
</style>

<script>
  import Sprites from "./game/Sprites.svelte";
  import NavLink from "./components/NavLink.svelte";
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

{#if $isLoading}
  <div class="progress">
    <div class="indeterminate"></div>
  </div>
{/if}
<div class="left">

  {#if $isAuthenticated}
    <div>

      <h5>Welcome back {$nickname}</h5>

      <div>

        <NavLink to="/play">Start playing</NavLink>
        or try to create some
        <NavLink to="/creator">own content.</NavLink>
      </div>

      <div>
        <p>
          <a class="modal-trigger" href="#modal1">Create a new Character</a>
        </p>

      </div>
    </div>
  {:else if !$isAuthenticated}
    <div>
      <h4>Welcome Stranger</h4>
      <p>
        Please
        <a href="/login" class="loginText">log</a>
        in or
        <a href="/signup" class="loginText">signup</a>
      </p>
    </div>
  {/if}

  <div>
    <h5>TalesMUD</h5>
    <p>
      TalesMUD is a MUD/MUX game engine/game development platform. Using
      TalesMud you can create your own MUD server, define your game content and
      either use the existing web client or build a new one from scratch.
    </p>
    <p>
      This is still a very early version with many more features planned. As
      development progresses you can expect several updates to the developer
      sandbox version running on this site.
    </p>
    Implemented
    <ul>
      <li>- Character creation</li>
      <li>- Room creation, movement between rooms</li>
    </ul>
    Planned
    <ul>
      <li>- Items, Inventory Management</li>
      <li>- NPCs, Enemies and Dialogs</li>
      <li>- Combat System</li>
      <li>- Quest System</li>
    </ul>
    <p>
      Head over to
      <NavLink to="/play">play</NavLink>
      create a character and try out the current set of commands by typing
      [help]. List of all global commands:
    </p>
    <ul>

      <li>[shrug] shrug emote</li>
      <li>[sc, selectcharacter] select a character, use: sc [charactername]</li>
      <li>[lc, listcharacters] list all your characters</li>
      <li>[h, help] are you really asking?</li>
      <li>[who] list all online players</li>
      <li>[inventory, i] Display your inventory</li>
      <li>[newcharacter, nc] Createa new character</li>
      <li>[scream] scream through the room</li>
    </ul>

  </div>

  <div>
    <h5>News</h5>
    <ul>
      <li>
        12.06.2020 - First minimalistic version live supporting room creation,
        character creation (template picks)
      </li>
    </ul>
  </div>
  <div>
    <h5>Credits</h5>
    The application uses several assets througout the app and the backend, here
    is a list of free and licensed art:
    <NavLink to="/credits">See Credits</NavLink>

  </div>

  <CharacterCreator />

</div>
