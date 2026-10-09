<script>
  import { onMount } from "svelte";
  import { getServerInfo } from "../api/live.js";
  import { userRole } from "../stores.js";
  import OpsToast from "./OpsToast.svelte";

  const tabs = [
    { name: "Rooms", nav: "/creator/rooms" },
    // Items editor hidden - items are now managed via Rooms editor
    // { name: "Items", nav: "/creator/items" },
    { name: "Item Templates", nav: "/creator/item-templates" },
    { name: "Character Templates", nav: "/creator/character-templates" },
    { name: "NPCs", nav: "/creator/npcs" },
    { name: "Dialogs", nav: "/creator/dialogs" },
    { name: "Dialog Graph", nav: "/creator/dialog-graph" },
    { name: "Quests", nav: "/creator/quests" },
    { name: "Skills", nav: "/creator/skills" },
    { name: "Scripts", nav: "/creator/scripts" },
    { name: "World", nav: "/creator/world" },
    { name: "Health", nav: "/creator/health" },
    { name: "Settings", nav: "/creator/settings" },
    { name: "Players", nav: "/creator/players", admin: true },
    { name: "Audit log", nav: "/creator/audit" },
  ];

  $: visibleTabs = tabs.filter((tab) => !tab.admin || $userRole === "admin");

  let envLabel = "";
  let envHost = "";

  // Note: Creator is mounted at `/creator` and `yrv` matches nested routes relative
  // to the parent route. So children should be defined as `/rooms`, `/npcs`, etc.
  const isActive = (nav) => window.location.pathname.startsWith(nav);

  onMount(async () => {
    try {
      const info = await getServerInfo();
      envLabel = String(info?.envLabel || "").trim();
      envHost = String(info?.host || "").trim();
    } catch {
      envLabel = "";
    }
  });
</script>

<div class="bg-slate-50 dark:bg-slate-950 min-h-[calc(100vh-72px)]">
  <div class="bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-800 px-6 overflow-x-auto scrollbar-hide">
    <div class="flex items-center gap-4 py-3">
      <div class="flex items-center gap-8 text-xs font-bold uppercase tracking-widest text-slate-500 dark:text-slate-500 min-w-max">
        {#each visibleTabs as tab}
          <a
            class={isActive(tab.nav) ? "text-primary" : "hover:text-primary transition-colors"}
            href={tab.nav}
          >
            {tab.name}
          </a>
        {/each}
      </div>
      {#if envLabel}
        <span class="ml-auto shrink-0 rounded bg-red-600 px-2 py-1 text-[10px] font-bold uppercase tracking-widest text-white">
          {envLabel}{envHost ? ` · ${envHost}` : ""}
        </span>
      {/if}
    </div>
  </div>

  <OpsToast />
  <slot />
</div>
