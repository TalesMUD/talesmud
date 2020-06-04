<style>
  .content {
    display: grid;
    grid-template-columns: 20% 80%;
    grid-column-gap: 10px;
  }

  input {
    color: #eee;
  }
  #refid {
    color: #999;
  }
  #refidlabel {
    color: #999;
  }
</style>

<script>
  import { writable } from "svelte/store";
  import CharacterCard from "./CharacterCard.svelte";
  import { onMount } from "svelte";

  import { createAuth, getAuth } from "./auth.js";
  import axios from "axios";
  import { onInterval } from "./utils.js";

  let user = writable({});

  const {
    isLoading,
    isAuthenticated,
    login,
    logout,
    authToken,
    authError,
    userInfo,
  } = getAuth();

  $: state = {
    isLoading: $isLoading,
    isAuthenticated: $isAuthenticated,
    authError: $authError,
    userInfo: $userInfo ? $userInfo.name : null,
    authToken: $authToken.slice(0, 20),
    user,
  };

  $: {
    axios(`http://localhost:8010/api/user`, {
      method: "GET",
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${$authToken}`,
      },
    }).then((r) => {
      user.update((u) => r.data);
      
    });
  }

  async function handleSubmit(event) {
    if ($isAuthenticated) {
      axios.put(`http://localhost:8010/api/user`, $user, {
        mode: "no-cors",
        credentials: "same-origin",
        headers: {
          Authorization: `Bearer ${$authToken}`,
        },
      });
    }
  }
</script>

<div class="row">
  <h3>Your Account data</h3>
  <form class="col s12">

    <div class="row">
      <div class="input-field col s12">
        <input
          bind:value="{$user.refid}"
          id="refid"
          type="text"
          disabled
        />
        <label id="refidlabel" for="refid" class="active">Reference ID</label>
      </div>
    </div>

    <div class="row">
      <div class="input-field col s12">
        <input
          bind:value="{$user.name}"
          id="username"
          type="text"
          class="validate"
        />
        <label for="username" class="active">Name</label>
      </div>
    </div>
    <div class="row">
      <div class="input-field col s12">
        <input
          bind:value="{$user.nickname}"
          id="usernickname"
          type="text"
          class="validate"
        />
        <label for="usernickname" class="active">Nickname</label>
      </div>
    </div>
    <div class="row">
      <div class="input-field col s12">
        <input
          bind:value="{$user.email}"
          id="useremail"
          type="text"
          class="validate"
        />
        <label for="useremail" class="active">E-Mail</label>
      </div>
    </div>
  </form>

  <div class="row">
    <div class="input-field col s12">
      <button
        class="btn waves-effect waves-light"
        on:click="{() => handleSubmit()}"
      >
        Submit
        <i class="material-icons right">send</i>
      </button>
    </div>
  </div>
</div>
