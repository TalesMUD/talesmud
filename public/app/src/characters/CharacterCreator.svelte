<script>
  import CharacterTemplate from "./CharacterTemplate.svelte";
  import { onMount } from "svelte";
  import { getAuth } from "../auth.js";
  import { getCharacterTemplates, createNewCharacter } from "../api/characters";

  const { isLoading, isAuthenticated, login, authToken } = getAuth();

  let templates = [];
  let selectedTemplateId = null;
  let selectedTemplate = null;

  let name = "";
  let description = "";

  let isLoadingTemplates = false;
  let isSubmitting = false;
  let errorMessage = "";
  let successCharacter = null;

  const selectTemplate = (id) => {
    selectedTemplateId = id;
    selectedTemplate = templates.find((t) => t.id === id) || null;
    errorMessage = "";
  };

  const create = async () => {
    errorMessage = "";
    successCharacter = null;

    if (!selectedTemplateId) {
      errorMessage = "Please select a class template first.";
      return;
    }
    if (!name || name.trim().length < 3) {
      errorMessage = "Name must be at least 3 characters.";
      return;
    }

    isSubmitting = true;
    const createDTO = {
      name: name.trim(),
      description: (description || "").trim(),
      templateId: selectedTemplateId,
    };

    createNewCharacter(
      $authToken,
      createDTO,
      (character) => {
        successCharacter = character;
        isSubmitting = false;
      },
      (err) => {
        isSubmitting = false;
        errorMessage =
          err?.response?.data?.error ||
          err?.response?.data?.message ||
          err?.message ||
          "Failed to create character.";
      }
    );
  };

  // Fetch templates on component init
  function loadTemplates() {
    isLoadingTemplates = true;
    getCharacterTemplates(
      (result) => {
        templates = result || [];
        isLoadingTemplates = false;
      },
      (err) => {
        isLoadingTemplates = false;
        errorMessage =
          err?.response?.data?.error ||
          err?.response?.data?.message ||
          err?.message ||
          "Failed to load character templates.";
      }
    );
  }

  // Call immediately - onMount may not fire with yrv router
  loadTemplates();
</script>

{#if !$isAuthenticated}
  <div class="flex items-center justify-between">
    <p class="text-sm text-slate-500 dark:text-slate-400">
      Log in to create and manage characters.
    </p>
    <button class="btn btn-primary" type="button" on:click={() => login()}>
      Log in
    </button>
  </div>
{:else}
  <div class="space-y-5">
    <div class="flex items-start justify-between gap-4">
      <div>
        <p class="text-sm text-slate-500 dark:text-slate-400">
          Choose a class template (stats/race/class), then personalize your name and description.
        </p>
      </div>
      <a class="text-xs text-primary hover:underline" href="/characters/new">
        Open full creator
      </a>
    </div>

    {#if errorMessage}
      <div class="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-200">
        {errorMessage}
      </div>
    {/if}

    {#if successCharacter}
      <div class="rounded-lg border border-emerald-500/30 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-100">
        <div class="font-medium">Character created: {successCharacter.name}</div>
        <div class="mt-1 text-xs text-emerald-200/80">
          You can now <a class="underline" href="/play">start playing</a>.
        </div>
      </div>
    {/if}

    <div class="grid grid-cols-1 lg:grid-cols-[1.6fr,1fr] gap-6">
      <div class="space-y-3">
        <div class="label-caps">Pick a class</div>
        {#if isLoadingTemplates}
          <div class="text-sm text-slate-500 dark:text-slate-400">
            Loading templates…
          </div>
        {:else if templates.length === 0}
          <div class="rounded-lg border border-slate-800 bg-slate-900/40 p-4 text-sm text-slate-400">
            No templates available.
            <div class="mt-2 text-xs text-slate-500">
              This usually means the API base URL is wrong. Check `VITE_API_BASE_URL` (or reload after updating).
            </div>
          </div>
        {:else}
          <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
            {#each templates as t}
              <CharacterTemplate
                name={t.name}
                description={t.description}
                attributes={t.attributes}
                templateId={t.id}
                race={t.race}
                cclass={t.class}
                selected={t.id === selectedTemplateId}
                callback={selectTemplate}
              />
            {/each}
          </div>
        {/if}
      </div>

      <div class="space-y-4">
        <div class="label-caps">Details</div>

        <div class="space-y-1.5">
          <label class="label-caps" for="char-name">Name</label>
          <input id="char-name" class="input-base" bind:value={name} placeholder="e.g. Gandalf the White" />
        </div>
        <div class="space-y-1.5">
          <label class="label-caps" for="char-desc">Description</label>
          <input id="char-desc" class="input-base" bind:value={description} placeholder="Short backstory or vibe…" />
        </div>

        <!-- Selected template preview card -->
        <div class="rounded-xl border border-slate-800 bg-slate-900/40 p-4 space-y-4">
          {#if selectedTemplate}
            <!-- Template header with avatar -->
            <div class="flex items-center gap-4">
              <img
                src={`/img/avatars/${1 + Math.abs((selectedTemplate.name || "").hashCode() % 12)}p.png`}
                alt={selectedTemplate.name}
                class="w-14 h-14 rounded-full border-2 border-primary"
              />
              <div class="flex-1 min-w-0">
                <div class="font-semibold text-slate-100">{selectedTemplate.name}</div>
                <div class="flex items-center gap-2 mt-1">
                  {#if selectedTemplate.class}
                    <span class="px-2 py-0.5 rounded-full bg-slate-800 border border-slate-700 text-[10px] font-medium text-slate-300">
                      {selectedTemplate.class.name}
                    </span>
                  {/if}
                  {#if selectedTemplate.race}
                    <span class="px-2 py-0.5 rounded-full bg-slate-800 border border-slate-700 text-[10px] font-medium text-slate-300">
                      {selectedTemplate.race.name}
                    </span>
                  {/if}
                </div>
              </div>
            </div>

            <!-- Stats grid -->
            <div class="space-y-3">
              <div class="label-caps">Base Stats</div>
              <div class="grid grid-cols-2 gap-2">
                {#if selectedTemplate.attributes}
                  {#each selectedTemplate.attributes as attr}
                    <div class="flex items-center justify-between bg-slate-800/50 rounded-lg px-3 py-2">
                      <span class="text-xs text-slate-400">{attr.name}</span>
                      <span class="text-sm font-mono font-semibold text-slate-200">{attr.value}</span>
                    </div>
                  {/each}
                {/if}
              </div>
            </div>

            <!-- Combat info -->
            <div class="grid grid-cols-2 gap-3 pt-2 border-t border-slate-800">
              <div>
                <div class="label-caps">HP</div>
                <div class="text-sm font-medium text-slate-200 mt-1">
                  {selectedTemplate.currentHitPoints}/{selectedTemplate.maxHitPoints}
                </div>
              </div>
              <div>
                <div class="label-caps">Combat Type</div>
                <div class="text-sm font-medium text-slate-200 mt-1">
                  {selectedTemplate.class?.combatType || "—"}
                </div>
              </div>
            </div>
          {:else}
            <div class="text-center py-6">
              <span class="material-symbols-outlined text-4xl text-slate-600 mb-2">person_add</span>
              <p class="text-sm text-slate-500">Select a class template to see stats</p>
            </div>
          {/if}
        </div>

        <div class="flex justify-end">
          <button
            class="btn btn-primary"
            type="button"
            disabled={isSubmitting || !selectedTemplate}
            on:click={create}
          >
            {#if isSubmitting}Creating…{:else}Create Character{/if}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
