<script>
  import CharacterCreator from "./characters/CharacterCreator.svelte";
  import { onMount } from "svelte";

  import { getRoomOfTheDay } from "./api/rooms";

  import { getAuth } from "./auth.js";
  import { getUser } from "./api/user.js";
  import { writable } from "svelte/store";
  import { onDestroy } from "svelte";

  const nickname = writable("stranger");
  const roomOfTheDay = writable({
    name: "RoomOfTheDay",
    meta: {
      background: "oldtown-griphon",
    },
  });
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

  let playMenuOpen = false;
  let playMenuEl;

  function togglePlayMenu(e) {
    e.preventDefault();
    e.stopPropagation();
    playMenuOpen = !playMenuOpen;
  }

  function onDocumentClick(e) {
    if (!playMenuEl) return;
    if (!playMenuEl.contains(e.target)) playMenuOpen = false;
  }

  onMount(async () => {
    document.addEventListener("click", onDocumentClick);
    getRoomOfTheDay(
      (room) => {
        roomOfTheDay.set(room);
      },
      (error) => console.log(error)
    );
    await loadUser();
  });

  onDestroy(() => {
    document.removeEventListener("click", onDocumentClick);
  });
</script>

{#if $isLoading}
  <div class="h-1 w-full bg-slate-200 dark:bg-slate-800">
    <div class="h-1 w-1/2 animate-pulse bg-primary"></div>
  </div>
{/if}

<div class="px-6 py-8">
  <div class="max-w-6xl mx-auto grid grid-cols-1 lg:grid-cols-[2fr,1fr] gap-8">
    <section class="space-y-6">
      <div class="card p-6">
        {#if $isAuthenticated}
          <div class="flex flex-col gap-3">
            <h1 class="text-2xl font-bold">Welcome back {$nickname}</h1>
            <p class="text-sm text-slate-500 dark:text-slate-400">
              Jump into the world or craft something new for your players.
            </p>
            <div class="flex flex-wrap gap-3">
              <div class="relative inline-flex" bind:this={playMenuEl}>
                <a class="btn btn-primary rounded-r-none" href="/play">
                  <span class="material-symbols-outlined text-sm">play_arrow</span>
                  Start playing
                </a>
                <button
                  class="btn btn-primary rounded-l-none px-2"
                  type="button"
                  aria-haspopup="menu"
                  aria-expanded={playMenuOpen}
                  on:click={togglePlayMenu}
                >
                  <span class="material-symbols-outlined text-sm">arrow_drop_down</span>
                </button>
                {#if playMenuOpen}
                  <div
                    class="absolute left-0 top-full mt-2 w-56 rounded-lg border border-slate-200 bg-white shadow-lg dark:border-slate-800 dark:bg-slate-900 overflow-hidden z-20"
                    role="menu"
                  >
                    <a
                      class="block px-3 py-2 text-sm hover:bg-slate-50 dark:hover:bg-slate-800/60"
                      href="/characters/new"
                      role="menuitem"
                    >
                      New Character
                    </a>
                  </div>
                {/if}
              </div>
              <a class="btn btn-outline" href="/creator/rooms">
                <span class="material-symbols-outlined text-sm">edit</span>
                Open Creator
              </a>
            </div>
          </div>
        {:else}
          <div class="flex flex-col gap-3">
            <h1 class="text-2xl font-bold">Welcome, Stranger</h1>
            <p class="text-sm text-slate-500 dark:text-slate-400">
              Sign in to save progress and access the Creator tools.
            </p>
            <div class="flex flex-wrap gap-3">
              <button class="btn btn-primary" type="button" on:click={() => login()}>
                Log in
              </button>
              <button class="btn btn-outline" type="button" on:click={() => login()}>
                Create account
              </button>
            </div>
          </div>
        {/if}
      </div>

      <div class="card p-6 space-y-4">
        <div>
          <h2 class="text-lg font-semibold">About TalesMUD</h2>
          <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">
            TalesMUD is a MUD/MUX engine and creator platform. Build your own world,
            define content, and run a live multiplayer text adventure.
          </p>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm text-slate-500 dark:text-slate-400">
          <div>
            <div class="label-caps text-primary">Implemented</div>
            <ul class="mt-2 space-y-1">
              <li>- Character creation</li>
              <li>- Room creation and movement</li>
            </ul>
          </div>
          <div>
            <div class="label-caps text-primary">Planned</div>
            <ul class="mt-2 space-y-1">
              <li>- Items and inventory management</li>
              <li>- NPCs, enemies, dialogs</li>
              <li>- Combat and quest systems</li>
            </ul>
          </div>
        </div>
        <div>
          <p class="text-sm text-slate-500 dark:text-slate-400">
            Head over to <a href="/play" class="text-primary">play</a> and
            try commands like <span class="font-mono">help</span>.
          </p>
        </div>
      </div>

      <div class="card p-6 space-y-3">
        <div class="flex items-center justify-between">
          <h2 class="text-lg font-semibold">Latest News</h2>
          <a class="text-sm text-primary hover:underline" href="/news">View all</a>
        </div>
        <ul class="text-sm text-slate-500 dark:text-slate-400 space-y-2">
          <li>
            <span class="font-mono text-xs text-slate-400">12.06.2020</span> · First
            minimalistic version live supporting room creation and character
            templates.
          </li>
        </ul>
      </div>

      <div class="card p-6 space-y-2">
        <h2 class="text-lg font-semibold">Credits</h2>
        <p class="text-sm text-slate-500 dark:text-slate-400">
          The application uses several assets throughout the app and backend.
        </p>
        <a class="text-sm text-primary hover:underline" href="/credits">
          See Credits
        </a>
      </div>

      <div class="card p-6 space-y-3">
        <h2 class="text-lg font-semibold">Create a Character</h2>
        <CharacterCreator />
        <div class="pt-2">
          <a class="text-sm text-primary hover:underline" href="/characters/new">
            Open full creator
          </a>
        </div>
      </div>
    </section>

    <aside class="space-y-6">
      <div class="card p-4">
        <div class="label-caps text-center">Room of the Day</div>
        <div class="mt-3 rounded-lg border border-slate-800 bg-slate-900/60 p-3 text-center">
          {#if $roomOfTheDay.meta}
            <img
              class="w-full rounded-md border border-slate-800"
              src={`/img/bg/${$roomOfTheDay.meta.background}.png`}
              alt="Room of the Day"
            />
          {/if}
          <div class="mt-3 text-sm font-medium">{$roomOfTheDay.name}</div>
          <p class="mt-2 text-xs text-slate-400">
            {$roomOfTheDay.description}
          </p>
        </div>
      </div>

      <div class="card p-4 space-y-2">
        <div class="label-caps">Quick Commands</div>
        <ul class="text-xs font-mono text-slate-400 space-y-1">
          <li>[help] list commands</li>
          <li>[sc] select character</li>
          <li>[lc] list characters</li>
          <li>[who] list online players</li>
          <li>[inventory] show inventory</li>
        </ul>
      </div>
    </aside>
  </div>
</div>
