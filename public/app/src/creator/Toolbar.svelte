<style>
  ul {
    display: flex;
    justify-content: flex-end;
    flex-direction: row;
    align-items: center;
    /* width: 100%; */

    padding: 15px;
  }

  ul li {
    display: inline-block;
    list-style: none;
    color: #eee;
  }

  .btn-flat {
    border: 1px #aaaaaa88 solid;
    color: #eee;
    margin: 5px;
  }

  .header {
    font-weight: 500;
    font-size: 200%;
  }

  ul li:first-child {
    margin-right: auto;
  }
</style>

<script>
  export let toolbar = {
    title: "TODO",
    actions: [
      {
        name: "Change Me",
        color: "green",
        fnc: () => console.log("ACTION"),
      },
    ],
  };
</script>

<div class="row">
  <ul>
    <li class="header">{toolbar.title}</li>
    {#each toolbar.actions as action}
      <li>
        <button class="btn-flat {action.color}" on:click="{action.fnc()}">

          {#if action.icon && action.name}
            <i class="material-icons left">{action.icon}</i>
          {:else if action.icon}
            <i class="material-icons center">{action.icon}</i>
          {/if}
          {#if action.name}{action.name}{/if}

        </button>
      </li>
    {/each}

  </ul>
</div>
