<script>
  import CharacterCreator from "./CharacterCreator.svelte";
</script>

<div class="px-6 py-8">
  <div class="max-w-6xl mx-auto space-y-6">
    <div class="flex items-start justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">Create a Character</h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">
          Pick a class template, review starting stats, then personalize your name and description.
        </p>
      </div>
      <a class="btn btn-outline" href="/">
        <span class="material-symbols-outlined text-sm">arrow_back</span>
        Back
      </a>
    </div>

    <div class="card p-6">
      <CharacterCreator />
    </div>
  </div>
</div>

