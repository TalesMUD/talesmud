<script>
  import CharacterScore from "./CharacterScore.svelte";
  import { onMount } from "svelte";

  import { getAuth } from "../auth.js";
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

<div class="px-6 py-8">
  <div class="max-w-6xl mx-auto space-y-8">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">Top Characters</h1>
      <p class="text-sm text-slate-500 dark:text-slate-400">
        Highest-ranked adventurers across the realms.
      </p>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
      {#each topTen as character}
        <CharacterScore
          name={character.name}
          level={character.level}
          xp={character.xp}
          cclass={character.class}
        />
      {/each}
    </div>

    <div class="card overflow-hidden">
      <div class="px-6 py-4 border-b border-slate-200 dark:border-slate-800">
        <h2 class="text-sm font-semibold uppercase tracking-wider text-slate-400">
          Leaderboard
        </h2>
      </div>
      <div class="overflow-x-auto">
        <table class="min-w-full text-sm">
          <thead class="bg-slate-50 dark:bg-slate-900/60 text-slate-500 dark:text-slate-400">
            <tr>
              <th class="px-6 py-3 text-left font-medium">Level</th>
              <th class="px-6 py-3 text-left font-medium">Name</th>
              <th class="px-6 py-3 text-left font-medium">Class</th>
              <th class="px-6 py-3 text-left font-medium">Description</th>
              <th class="px-6 py-3 text-left font-medium">XP</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200 dark:divide-slate-800">
            {#each data as character}
              <tr class="hover:bg-slate-50/60 dark:hover:bg-slate-900/40">
                <td class="px-6 py-3">{character.level}</td>
                <td class="px-6 py-3 font-medium">{character.name}</td>
                <td class="px-6 py-3">{character.class.name}</td>
                <td class="px-6 py-3 text-slate-500 dark:text-slate-400">
                  {character.description}
                </td>
                <td class="px-6 py-3 text-primary">{character.xp}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  </div>
</div>
