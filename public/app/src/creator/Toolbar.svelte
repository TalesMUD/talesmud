<style>
  ul {
    display: flex;
    justify-content: flex-end;
    flex-direction: row;
    align-items: center;
    /* width: 100%; */
  }

  ul li {
    display: inline-block;
    list-style: none;
    color: #eee;
  }

  .btn-flat {
    border: 1px #aaaaaa88 solid;
    color: #eee;
    border-radius: 4px;
    margin: 5px;
  }

  .header {
    font-weight: 500;
    font-size: 200%;
  }

  .padding {
    padding-right: 0.5em;
    padding-left: 0.5em;
    margin: 0;
  }
  .padding-small {
    padding-left: 5px;
  }

  .header-small {
    font-weight: 300;
    font-size: 120%;

    padding-right: 5px;
  }

  ul li:first-child {
    margin-right: auto;
  }
</style>

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

  const headerClass = () => {
    if (toolbar.small) {
      return "header-small";
    }
    return "header";
  };
  const headerPadding = () => {
    if (toolbar.small) {
      return "padding-small";
    }
    return "padding";
  };
  const headerDiv = () => {
    if (toolbar.small) {
      return "";
    }
    return "row";
  };
</script>

<div class="{headerDiv()}">
  <ul class="{headerPadding()}">
    <li class="{headerClass()}">{toolbar.title}</li>
    {#each toolbar.actions as action}
      <li>
        <button class="btn-flat" on:click="{action.fnc()}">

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
