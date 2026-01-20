<script>
  import { writable } from "svelte/store";
  import { getAuth } from "./auth.js";

  import { getUser, updateUser } from "./api/user.js";

  let user = writable({});

  const { isLoading, isAuthenticated, authToken, authError, userInfo } =
    getAuth();

  $: state = {
    isLoading: $isLoading,
    isAuthenticated: $isAuthenticated,
    authError: $authError,
    userInfo: $userInfo ? $userInfo.name : null,
    authToken: $authToken.slice(0, 20),
    user,
  };

  let loaded = false;

  const loadUser = () => {
    getUser(
      $authToken,
      (u) => {
        user.set(u);
      },
      (err) => console.log(err)
    );
  };

  $: if (!loaded && $isAuthenticated && $authToken) {
    loaded = true;
    loadUser();
  }

  async function handleSubmit(event) {
    if ($isAuthenticated) {
      updateUser($authToken, $user, () => {
        console.log("user updated ");
      });
    }
  }
</script>

<div class="px-6 py-8">
  <div class="max-w-3xl mx-auto space-y-6">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">Account</h1>
      <p class="text-sm text-slate-500 dark:text-slate-400">
        Manage your profile details and preferences.
      </p>
    </div>

    <div class="card p-6 space-y-4">
      <div class="space-y-1.5">
        <label class="label-caps" for="refid">Reference ID</label>
        <input
          class="input-base font-mono text-slate-400"
          bind:value="{$user.refid}"
          id="refid"
          type="text"
          disabled
        />
      </div>

      <div class="space-y-1.5">
        <label class="label-caps" for="username">Name</label>
        <input class="input-base" bind:value="{$user.name}" id="username" type="text" />
      </div>

      <div class="space-y-1.5">
        <label class="label-caps" for="usernickname">Nickname</label>
        <input class="input-base" bind:value="{$user.nickname}" id="usernickname" type="text" />
      </div>

      <div class="space-y-1.5">
        <label class="label-caps" for="useremail">Email</label>
        <input class="input-base" bind:value="{$user.email}" id="useremail" type="email" />
      </div>

      <div class="flex justify-end">
        <button class="btn btn-primary" type="button" on:click={() => handleSubmit()}>
          <span class="material-symbols-outlined text-sm">save</span>
          Save changes
        </button>
      </div>
    </div>
  </div>
</div>
