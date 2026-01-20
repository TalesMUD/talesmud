<script>
  const tabs = [
    { name: "Rooms", nav: "/creator/rooms" },
    { name: "Items", nav: "/creator/items" },
    { name: "Item Templates", nav: "/creator/item-templates" },
    { name: "Character Templates", nav: "/creator/character-templates" },
    { name: "NPCs", nav: "/creator/npcs" },
    { name: "Dialogs", nav: "/creator/dialogs" },
    { name: "Dialog Graph", nav: "/creator/dialog-graph" },
    { name: "Scripts", nav: "/creator/scripts" },
    { name: "World", nav: "/creator/world" },
  ];

  // Note: Creator is mounted at `/creator` and `yrv` matches nested routes relative
  // to the parent route. So children should be defined as `/rooms`, `/npcs`, etc.
  const isActive = (nav) => window.location.pathname.startsWith(nav);
</script>

<div class="bg-slate-50 dark:bg-slate-950 min-h-[calc(100vh-72px)]">
  <div class="bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-800 px-6 overflow-x-auto scrollbar-hide">
    <div class="flex items-center gap-8 py-3 text-xs font-bold uppercase tracking-widest text-slate-500 dark:text-slate-500 min-w-max">
      {#each tabs as tab}
        <a
          class={isActive(tab.nav) ? "text-primary" : "hover:text-primary transition-colors"}
          href={tab.nav}
        >
          {tab.name}
        </a>
      {/each}
    </div>
  </div>

  <slot />
</div>
