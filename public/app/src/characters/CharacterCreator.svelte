<style>
  .modal {
    background-color: darkslategrey;
  }

  input {
    color: #eee;
  }
</style>

<script>
  import CharacterTemplate from "./CharacterTemplate.svelte";
  import { onMount } from "svelte";
  import { writable } from "svelte/store";

  import { createAuth, getAuth } from "../auth.js";
  import axios from "axios";
  import { onInterval } from "../utils.js";
  import { getCharacterTemplates, createNewCharacter } from "../api/characters";

  let data = [];
  let topTen = [];

  const store = writable({
    templateSelected: false,
    selectedTemplate: 0,
    character: {},
    name: "unnamed",
    description: "Describe your new character",
  });

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

  const create = () => {
    const createDTO = {
      name: $store.name,
      description: $store.description,
      templateId: $store.selectedTemplate,
    };

    createNewCharacter(
      $authToken,
      createDTO,
      (character) => console.log("Created character " + character.id),
      (err) => console.log(err)
    );
  };

  onMount(async () => {
    document.addEventListener("DOMContentLoaded", function () {
      var elems = document.querySelectorAll(".modal");
      var instances = M.Modal.init(elems, {});
    });

    getCharacterTemplates(
      (result) => {
        templates.set(result);
      },
      (err) => {
        console.log(err);
      }
    );
  });

  const templates = writable([]);
</script>

<!-- Modal Structure -->
<div id="modal1" class="modal">
  <div class="modal-content">
    <h4>Create new Character</h4>

    {#if !$store.templateSelected}
      <p>
        Select a template which fits you most, you can change starting
        attributes as well
      </p>
      <div class="row">
        {#each $templates as character}
          <div class="col s4">
            <CharacterTemplate
              name="{character.name}"
              description="{character.description}"
              created="{character.created}"
              attributes="{character.attributes}"
              templateId="{character.templateId}"
              callback="{(id) => {
                store.update((state) => {
                  state.templateSelected = true;
                  state.selectedTemplate = id;
                  state.character = character;
                  state.name = character.name;
                  state.description = character.description;
                  return state;
                });
              }}"
            />
          </div>
        {/each}
      </div>
    {/if}

    <div class="row">
      {#if $store.templateSelected}
        <div class="col s4">
          <CharacterTemplate
            name="{$store.name}"
            description="{$store.description}"
            created="{$store.character.created}"
            attributes="{$store.character.attributes}"
          />
        </div>
      {/if}
      <div class="col s8">

        <div class="input-field">
          <input
            bind:value="{$store.name}"
            id="name"
            type="text"
            class="validate"
          />
          <label for="name" class="active">Name</label>
        </div>

        <div class="input-field">
          <input
            bind:value="{$store.description}"
            id="description"
            type="text"
          />
          <label for="description" class="active">Description</label>
        </div>
      </div>
    </div>
  </div>
  <div class="modal-footer">
    <button
      href="#!"
      class="modal-close waves-effect waves-red btn"      
    >
      Cancel
    </button>
    <button
      href="#!"
      class="modal-close waves-effect waves-green btn"
      on:click="{create}"
    >
      Create
    </button>
  </div>
</div>
