<style>
  .sidelist {
    width: 20em;
  }
</style>

<script>
  import { writable } from "svelte/store";
  import { onMount } from "svelte";
  import { createAuth, getAuth } from "../auth.js";

  import axios from "axios";

  export let location;

  let items = [];
  let selectedItem = {};

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
  };

  $: {
  }

  onMount(async () => {});
</script>

<div class="row">
  <h4>Manage Items</h4>

</div>
