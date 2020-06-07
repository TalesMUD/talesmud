<style>
  .collapsible-header {
    color: #333;
  }
  .collapsible-body {
    padding: 1em;
    margin: 1em;
  }
  label {
    color: #eee;
  }
  input {
    color: white;
  }
  input:disabled {
    color: white;
  }
</style>

<script>
  export let exit;
  export let deleteExit;
</script>

<li>
  <div class="collapsible-header">

    <i class="material-icons">exit_to_app</i>
    {exit.name}
  </div>

  <div class="collapsible-body">
    <div class="row">
      <button
        on:click="{() => deleteExit(exit)}"
        class="waves-effect waves-light btn-small red right"
      >
        <i class="material-icons">delete_forever</i>
        Delete Exit
      </button>
    </div>
    <div class="row">
      <div class="input-field">
        <input
          placeholder="Placeholder"
          id="name-{exit.name}"
          type="text"
          class="validate"
          bind:value="{exit.name}"
        />
        <label for="name-{exit.name}">Name</label>

        <div class="input-field">
          <input
            placeholder="Placeholder"
            id="desc-{exit.description}"
            type="text"
            class="validate"
            bind:value="{exit.description}"
          />
          <label for="desc-{exit.description}">Description</label>
          <div class="input-field">
            <input
              id="target-{exit.target}"
              type="text"
              class="validate"
              bind:value="{exit.target}"
            />
            <label for="target-{exit.target}">Target Room</label>
          </div>
          <div class="input-field">
            <input type="text" id="autocomplete-input" class="autocomplete" />
            <label for="autocomplete-input">Target Room</label>
          </div>

        </div>

      </div>
    </div>
  </div>
</li>
