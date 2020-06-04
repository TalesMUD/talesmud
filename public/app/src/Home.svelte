<style>

</style>

<script>
  import CharacterCard from "./CharacterCard.svelte";
  import { onMount } from "svelte";
  import { user } from "./stores.js";
  import { createAuth, getAuth } from "./auth.js";
  import axios from "axios";
  import { onInterval } from "./utils.js";

  const apiURL = "";

  let data = [];

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

  let loggedIn = async (fnc) => {
    if ($isAuthenticated) fnc;
    else {
    }
  };

  onInterval(() => loggedIn(update($authToken)), 4000);

  async function update(accessToken) {
    if ($isLoading) return;

    axios(`http://localhost:8010/api/characters`, {
      method: "GET",
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${$authToken}`,
      },
    })
      .then((result) => (data = result.data))
      .catch((err) => console.log(err));
  }

</script>

<div class="row">
  <div class="col s12">
    {#each data as character}
      <CharacterCard
        name="{character.name}"
        description="{character.description}"
        created="{character.created}"
      />
    {/each}
  </div>
</div>
