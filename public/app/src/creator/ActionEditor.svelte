<style>
  .collapsible-header {
    background-color: transparent;
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
  import { onMount } from "svelte";

  export let action;
  export let deleteAction;

  onMount(async () => {
    if (action.params === null) {
      action.params = new Map();
    }
  });
</script>

<li>
  <div class="collapsible-header">
    <div class="col s9 left valign-wrapper">
      <i class="material-icons left-align">flare</i>
      {action.name}
    </div>
    <div class="col s3 right-align">
      <button
        on:click="{() => deleteAction(action)}"
        class="btn-small red align-right"
      >
        Delete Action
      </button>
    </div>

  </div>

  <div class="collapsible-body">
    <div class="row">

      <div class="input-field">
        <input
          placeholder="Placeholder"
          id="name-{action.name}"
          type="text"
          bind:value="{action.name}"
        />
        <label for="name-{action.name}">Name</label>

        <div class="input-field">
          <input
            placeholder="Placeholder"
            id="desc-{action.description}"
            type="text"
            bind:value="{action.description}"
          />
          <label for="desc-{action.description}">Description</label>
          <div class="input-field">
            <input
              id="target-{action.response}"
              type="text"
              bind:value="{action.response}"
            />
            <label for="target-{action.response}">Response</label>
          </div>

          <div class="input-field">
            <select
              id="autocomplete-input-{action.type}"
              bind:value="{action.type}"
            >

              <option value="response">Respond to Player</option>
              <option value="response_room">Respond to Room</option>
              <option value="script">Run Script</option>
            </select>
            <label>Type</label>
          </div>
          <div class="input-field">
            <input
              id="scriptid-{action.response}-1"
              type="text"
            />
            <label for="target-{action.response}-1">
              Script ID(if script is selected)
            </label>
          </div>
        </div>

      </div>
    </div>
  </div>
</li>
