<style>

</style>

<script>
  import CharacterScore from "./CharacterScore.svelte";
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
        data = characters.slice(0, 100);
        topTen = characters.slice(0, 8);
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
        <CharacterScore
          name="{character.name}"
          level="{character.level}"
          xp="{character.xp}"
          cclass="{character.class}"
        />
      </div>
    {/each}

  </div>
  <div>

    <table>
      <thead>
        <tr>
          <th>Level</th>
          <th>Name</th>
          <th>Class</th>
          <th>Description</th>
          <th>XP</th>
        </tr>
      </thead>

      <tbody>
        {#each data as character}
          <tr>
            <td>{character.level}</td>
            <td>{character.name}</td>
            <td>{character.class.name}</td>
            <td>{character.description}</td>
            <td>{character.xp}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
