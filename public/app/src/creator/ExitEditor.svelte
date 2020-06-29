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
  export let exit;
  export let store;
  export let valueHelp;
  export let deleteExit;

  const initial = () => {
    exit.target;
  };
  onMount(async () => {
    setTimeout(function () {
      // set exit target autocomplete values
      let options = {
        data: {},
        onAutocomplete: (e) => {
          console.log("ON AUTO " + e);
        },
      };
      valueHelp.forEach((element) => {
        const key = element.name + " (" + element.id + ")";
        options.data[key] = element.id;
      });

      const selector = "#autocomplete-input-" + exit.name;
      var elems = document.querySelectorAll(selector);
      var instances = M.Autocomplete.init(elems, options);
    }, 50);
  });
</script>

<li>
  <div class="collapsible-header">
    <div class="col s9 left valign-wrapper">
      <i class="material-icons left-align">exit_to_app</i>
      {exit.name}
    </div>
    <div class="col s3 right-align">
      <button
        on:click="{() => deleteExit(exit)}"
        class="btn-small red align-right"
      >
        Delete Exit
      </button>
    </div>

  </div>

  <div class="collapsible-body">
    <div class="row">

      <label>
        <input type="checkbox" bind:checked="{exit.hidden}" />
        <span>Hidden</span>
      </label>

    </div>

    <div class="row">

      <div class="input-field">
        <input
          placeholder="Placeholder"
          id="name-{exit.name}"
          type="text"
          bind:value="{exit.name}"
        />
        <label for="name-{exit.name}">Name</label>

        <div class="input-field">
          <input
            placeholder="Placeholder"
            id="desc-{exit.description}"
            type="text"
            bind:value="{exit.description}"
          />
          <label for="desc-{exit.description}">Description</label>

          <!-- <div class="input-field">
            <input
              id="target-{exit.target}"
              type="text"
              bind:value="{exit.target}"
            />
            <label for="target-{exit.target}">Target Room</label>
          </div> -->
          <div class="input-field">
            <input
              type="text"
              id="autocomplete-input-{exit.name}"
              class="autocomplete targets"
              value="{initial()}"
            />
            <label for="autocomplete-input">Target Room</label>
          </div>

        </div>

      </div>
    </div>
  </div>
</li>
