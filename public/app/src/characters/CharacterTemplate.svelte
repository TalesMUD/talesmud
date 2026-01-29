<script>
  export let name;
  export let description;
  // These props are passed from parent but currently not displayed
  // svelte-ignore unused-export-let
  export let created;
  // svelte-ignore unused-export-let
  export let level;
  // svelte-ignore unused-export-let
  export let xp;
  export let templateId;
  // svelte-ignore unused-export-let
  export let attributes;
  export let race;
  // svelte-ignore unused-export-let
  export let cclass;
  export let callback;
  export let selected = false;

  function avatar() {
    let num = 1 + Math.abs(name.hashCode()%12)
    return "/img/avatars/" + num + "p.png";
  }
</script>

<div
  class="group relative rounded-xl border-2 p-4 cursor-pointer transition-all duration-200 hover:border-slate-600 hover:shadow-lg hover:-translate-y-0.5 {selected ? 'border-primary bg-slate-800' : 'border-slate-800 bg-slate-900/40 hover:bg-slate-800/60'}"
  on:click={() => callback(templateId)}
  role="button"
  tabindex="0"
  on:keydown={(e) => {
    if (e.key === "Enter" || e.key === " ") callback(templateId);
  }}
>
  <!-- Selection indicator -->
  {#if selected}
    <div class="absolute top-2 right-2">
      <span class="material-symbols-outlined text-primary text-lg">check_circle</span>
    </div>
  {/if}

  <div class="flex flex-col items-center text-center space-y-3">
    <!-- Avatar with glow on hover -->
    <div class="relative">
      <img
        src={avatar()}
        alt={name}
        class="w-16 h-16 rounded-full border-2 transition-all duration-200 {selected ? 'border-primary' : 'border-slate-700'}"
      />
    </div>

    <!-- Name -->
    <div class="font-semibold text-slate-100 group-hover:text-white transition-colors">
      {name}
    </div>

    <!-- Description -->
    <p class="text-xs text-slate-400 leading-relaxed line-clamp-2">
      {description}
    </p>

    <!-- Class/Race tags -->
    {#if cclass || race}
      <div class="flex flex-wrap items-center justify-center gap-1.5">
        {#if cclass}
          <span class="px-2 py-0.5 rounded-full bg-slate-800 border border-slate-700 text-[10px] font-medium text-slate-300">
            {cclass.name}
          </span>
        {/if}
        {#if race}
          <span class="px-2 py-0.5 rounded-full bg-slate-800 border border-slate-700 text-[10px] font-medium text-slate-300">
            {race.name}
          </span>
        {/if}
      </div>
    {/if}
  </div>
</div>
