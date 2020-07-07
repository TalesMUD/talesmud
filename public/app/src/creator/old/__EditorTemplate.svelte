<style>

</style>

<script>
  import Sprites from "./../game/Sprites.svelte";
  import { onMount } from "svelte";
  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";
  import { v4 as uuidv4 } from "uuid";

  import {
    getScript,
    deleteScript,
    getScripts,
    updateScript,
    createScript,
  } from "../api/scripts.js";

  const config = {
    title: "Manage Scripts",
    actions: [],
    get: getScripts,
    getElement: getScript,
    create: createScript,
    update: updateScript,
    delete: deleteScript,
    refreshUI: () => {},

    new: (select) => {
      select({
        id: uuidv4(),
        name: "New Script",
        description: "something",
        isNew: true,
      });
    },

    badge: (element) => {
      return element.id;
    },
  };
  // create store outside of the component to use it in the slot..
  const store = createStore();

  onMount(async () => {});
  /////////
</script>

<CRUDEditor store="{store}" config="{config}" />
