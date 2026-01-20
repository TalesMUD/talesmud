<script>
  export let toolbar = {
    title: "TODO",
    small: false,
    actions: [
      {
        name: "Change Me",
        color: "green",
        fnc: () => console.log("ACTION"),
      },
    ],
  };
</script>

<div class={toolbar.small ? "px-4 py-3" : "px-6 py-4"}>
  <div class="flex items-center justify-between gap-4">
    <h2 class={toolbar.small ? "text-sm font-semibold" : "text-lg font-semibold"}>
      {toolbar.title}
    </h2>
    <div class="flex items-center gap-2">
      {#each toolbar.actions as action}
        <button
          class="btn btn-ghost"
          type="button"
          title={action.name || ""}
          on:click={() => action.fnc()}
        >
          {#if action.icon}
            <span class="material-symbols-outlined text-lg">{action.icon}</span>
          {/if}
          {#if action.name}
            <span class="text-xs">{action.name}</span>
          {/if}
        </button>
      {/each}
    </div>
  </div>
</div>
