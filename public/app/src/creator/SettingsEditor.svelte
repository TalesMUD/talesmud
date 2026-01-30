<script>
  import { onMount } from "svelte";
  import { getAuth } from "../auth.js";
  import { getSettings, updateSettings } from "../api/settings.js";

  const { isAuthenticated, authToken } = getAuth();

  let settings = { serverName: "", about: "" };
  let saving = false;
  let saveMessage = "";
  let hasLoaded = false;

  function loadSettings() {
    if (!$isAuthenticated || !$authToken) return;
    getSettings(
      $authToken,
      (data) => {
        settings = {
          serverName: data.serverName || "",
          about: data.about || "",
        };
        hasLoaded = true;
      },
      (err) => console.error("Failed to load settings:", err)
    );
  }

  function save() {
    if (!$authToken) return;
    saving = true;
    saveMessage = "";
    updateSettings(
      $authToken,
      settings,
      () => {
        saving = false;
        saveMessage = "Settings saved.";
        setTimeout(() => (saveMessage = ""), 3000);
      },
      (err) => {
        saving = false;
        saveMessage = "Failed to save settings.";
        console.error("Failed to save settings:", err);
      }
    );
  }

  $: if ($isAuthenticated && $authToken && !hasLoaded) {
    loadSettings();
  }

  onMount(() => {
    loadSettings();
  });
</script>

{#if !$isAuthenticated}
  <div class="px-6 py-12 text-center text-sm text-slate-500 dark:text-slate-400">
    Please log in to access Creator tools.
  </div>
{:else}
  <div class="flex h-[calc(100vh-128px)]">
    <section class="flex-1 overflow-y-auto">
      <div class="p-8 pb-4">
        <div class="max-w-3xl mx-auto">
          <div class="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-8">
            <div class="space-y-1">
              <h1 class="text-3xl font-bold tracking-tight">Settings</h1>
              <p class="text-slate-500 dark:text-slate-400 text-sm">
                Configure general server settings.
              </p>
            </div>
            <div class="flex items-center gap-3">
              {#if saveMessage}
                <span class="text-sm text-slate-400">{saveMessage}</span>
              {/if}
              <button
                class="btn btn-primary"
                type="button"
                on:click={save}
                disabled={saving}
              >
                <span class="material-symbols-outlined text-sm">save</span>
                {saving ? "Saving..." : "Save"}
              </button>
            </div>
          </div>

          <div class="card p-6 space-y-6">
            <div class="space-y-1.5">
              <label class="label-caps" for="server_name">Server Name</label>
              <input
                id="server_name"
                type="text"
                class="input-base text-lg font-semibold"
                bind:value={settings.serverName}
                placeholder="My MUD Server"
              />
              <p class="text-xs text-slate-400 dark:text-slate-500">
                Displayed in the MUD client header and welcome message.
              </p>
            </div>

            <div class="space-y-1.5">
              <label class="label-caps" for="server_about">About</label>
              <textarea
                id="server_about"
                rows="8"
                class="input-base"
                bind:value={settings.about}
                placeholder="Describe your server, its world, and its story..."
              />
              <p class="text-xs text-slate-400 dark:text-slate-500">
                A longer description of your server and game world.
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
{/if}
