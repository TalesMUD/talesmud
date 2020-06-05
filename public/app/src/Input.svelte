<style>
  textarea {
    color: white;
    font-size: smaller;
    font-family: "Courier New", Courier, monospace;


  }
</style>

<script>
  import { fade } from "svelte/transition";
  import { writable } from "svelte/store";

  let input = "";

  // Example POST method implementation:
  async function postData(url = "", data = {}) {
    // Default options are marked with *
    const response = await fetch(url, {
      method: "POST", // *GET, POST, PUT, DELETE, etc.
      mode: "no-cors", // no-cors, *cors, same-origin
      cache: "no-cache", // *default, no-cache, reload, force-cache, only-if-cached
      credentials: "same-origin", // include, *same-origin, omit
      headers: {
        "Content-Type": "application/json"
      },
      redirect: "follow", // manual, *follow, error
      referrerPolicy: "no-referrer", // no-referrer, *no-referrer-when-downgrade, origin, origin-when-cross-origin, same-origin, strict-origin, strict-origin-when-cross-origin, unsafe-url
      body: JSON.stringify(data) // body data type must match "Content-Type" header
    });
    return response.json(); // parses JSON response into native JavaScript objects
  }

  async function post() {
    if (input === "") return;

    let data = {
      text: input
    };

  }
</script>

<div class="row left-align">
  <form class="col s12 valign-wrapper left-align">

    <div class="input-field col s12 white-text">
      <textarea
        id="icon_prefix2"
        class="materialize-textarea"
        bind:value="{input}"
      ></textarea>
      <label for="icon_prefix2">leave a message</label>
    </div>
    <button class="waves-effect waves-light btn-small white-text grey darken-2" on:click="{post}">
      send
    </button>
  </form>

</div>
