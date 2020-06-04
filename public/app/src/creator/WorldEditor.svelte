<style>
  .sidelist {
    width: 20em;
  }
</style>

<script>
  import { writable } from "svelte/store";
  import { onMount } from "svelte";
  import { createAuth, getAuth } from "../auth.js";
  import { getWorldMap } from "../api/world.js";

  import axios from "axios";

  export let location;

  let img;

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
  };

  $: {
    if (!$isLoading && $isAuthenticated && authToken) {
      getWorldMap($authToken, (world) => {
        img = world;
      },
      (err) => console.log(err)
      );
    }
  }

  onMount(async () => {});
</script>

<div class="row">
  <h4>Map of Rooms</h4>
  {#if $isAuthenticated}
    <img class="responsive-img z-depth-3" src="{img}" alt="world map" />
  {/if}


</div>
