<style>

</style>

<script>
  import CharacterCard from "./CharacterCard.svelte";
  import { onMount } from "svelte";

  import { createAuth, getAuth } from "../auth.js";
  import axios from "axios";
  import { onInterval } from "../utils.js";
  import { getCharacters } from "../api/characters";

  let data = [];
  let topTen = [];

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

  const loadData = async () => {
    if ($isLoading && !$isAuthenticated) return;
    getCharacters(
      $authToken,
      (characters) => {
        data = characters;
        topTen = characters.slice(0, 9);
      },
      (err) => console.log(err)
    );
  };

  onMount(async () => {
    await loadData();
    console.log("Finished loading characters");
  });
</script>

<div>

  <div class="row">

    {#each topTen as character}
      <div class="col s3">
        <CharacterCard
          name="{character.name}"
          level="{character.level}"
          xp="{character.xp}"
          description="{character.description}"
          created="{character.created}"
        />
      </div>
    {/each}

  </div>
  <div>

    <table>
      <thead>
        <tr>
          <th>Name</th>
          <th>Description</th>
          <th>Created</th>
        </tr>
      </thead>

      <tbody>
        {#each data as character}
          <tr>
            <td>{character.name}</td>
            <td>{character.description}</td>
            <td>{character.created}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
