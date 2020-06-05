<style>

</style>

<script>
  import CharacterCard from "./CharacterCard.svelte";
  import { onMount } from "svelte";

  import { createAuth, getAuth } from "../auth.js";
  import axios from "axios";
  import { onInterval } from "../utils.js";
  import { getCharacters } from "../api/characters";

  let data = [];
  let topTen = [];

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

  onMount(async () => {
    document.addEventListener("DOMContentLoaded", function () {
      var elems = document.querySelectorAll(".modal");
      var instances = M.Modal.init(elems, {});
    });
  });



</script>

<!-- Modal Structure -->
<div id="modal1" class="modal">
  <div class="modal-content">
    <h4>Create new Character</h4>
    <p>A bunch of text</p>
  </div>
  <div class="modal-footer">
    <a href="#!" class="modal-close waves-effect waves-red btn-flat">Cancel</a>
    <a href="#!" class="modal-close waves-effect waves-green btn-flat">
      Create
    </a>
  </div>
</div>
