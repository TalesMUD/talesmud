<style>
  .collapsible-header {
    background-color: transparent;
  }
  .collapsible-body {
    padding: 1em;
    margin: 1em;
    border: none;
    padding-bottom: 0;
    margin-bottom: 0;
  }

  td {
    padding: 5px;
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

  $: {
    if ($valueHelp) {
      updateUI();
    }
  }

  const updateUI = () => {
    // set exit target autocomplete values
    let options = {
      data: {},
      onAutocomplete: (e) => {
        const roomid = options.data[e];

        console.log("ON AUTO " + roomid);

        exit.target = roomid;
      },
    };

    let selected = "";

    $valueHelp.forEach((element) => {
      const key = element.name + " (" + element.id + ")";

      if (element.id === exit.target) {
        selected = key;
      }

      options.data[key] = element.id;
    });

    const selector = "#autocomplete-input-" + exit.name.replace(" ", "_");
    var elems = document.querySelectorAll(selector);
    if (elems.length > 0) {
      elems[0].value = selected;
    }

    var instances = M.Autocomplete.init(elems, options);
  };

  const initial = () => {
    exit.target;
  };
  onMount(async () => {});
</script>

<tr>

  <td style="width:2em;">
    <input
      placeholder="Name"
      id="name-{exit.name}"
      type="text"
      bind:value="{exit.name}"
      style="margin:0; height: 2em;font-size: 14px;border-color:#ffffff33;"
    />
  </td>
  <td>
    <input
      placeholder="Name"
      id="name-{exit.description}"
      type="text"
      bind:value="{exit.description}"
      style="margin:0; height: 2em;font-size: 14px;border-color:#ffffff33;"
    />
  </td>
  <td>

    <div class="input-field">

      <input
        type="text"
        id="autocomplete-input-{exit.name.replace(' ', '_')}"
        class="autocomplete targets"
        style="margin:0; height: 2em;font-size: 14px;border-color:#ffffff33;"
      />

    </div>

  </td>
  <td>

    <button
      class="btn-flat waves-effect waves-light right"
      on:click="{() => deleteExit(exit)}"
    >
      <i class="material-icons" style="color:red;">remove</i>
    </button>
  </td>
</tr>
