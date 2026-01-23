<script>
  import { v4 as uuidv4 } from "uuid";
  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";

  import {
    getDialog,
    getDialogs,
    createDialog,
    updateDialog,
    deleteDialog,
  } from "../api/dialogs.js";

  const store = createStore();

  const config = {
    title: "Manage Dialogs",
    subtitle: "Author NPC dialog trees and alternate text responses.",
    listTitle: "Dialogs",
    labels: {
      create: "Create Dialog",
      update: "Update Dialog",
      delete: "Delete",
    },
    get: getDialogs,
    getElement: getDialog,
    create: createDialog,
    update: updateDialog,
    delete: deleteDialog,
    beforeSelect: (element) => {
      if (!element.options) element.options = [];
      if (!element.alternateTexts) element.alternateTexts = [];
    },
    new: (select) => {
      select({
        id: uuidv4(),
        name: "New Dialog",
        description: "",
        detail: "",
        options: [],
        alternateTexts: [],
        isNew: true,
      });
    },
  };

  const addOption = () => {
    store.update((state) => {
      state.selectedElement.options.push({
        text: "New option",
        nextDialogId: "",
      });
      return state;
    });
  };

  const removeOption = (index) => {
    store.update((state) => {
      state.selectedElement.options.splice(index, 1);
      return state;
    });
  };

  const addAlternateText = () => {
    store.update((state) => {
      state.selectedElement.alternateTexts.push("");
      return state;
    });
  };

  const removeAlternateText = (index) => {
    store.update((state) => {
      state.selectedElement.alternateTexts.splice(index, 1);
      return state;
    });
  };
</script>

<CRUDEditor store={store} config={config}>
  <div slot="content" class="space-y-6">
    <div class="card p-4 space-y-3">
      <div class="flex items-center justify-between">
        <div class="label-caps text-primary">Dialog Options</div>
        <button class="text-xs text-primary hover:underline" type="button" on:click={addOption}>
          + Add Option
        </button>
      </div>
      <div class="space-y-3">
        {#each $store.selectedElement.options as option, index}
          <div class="grid grid-cols-1 md:grid-cols-[2fr,1fr,auto] gap-3 items-center">
            <input class="input-base text-xs" bind:value={option.text} placeholder="Option text" />
            <input class="input-base text-xs" bind:value={option.nextDialogId} placeholder="Next dialog ID" />
            <button class="text-xs text-accent-red hover:underline" type="button" on:click={() => removeOption(index)}>
              Remove
            </button>
          </div>
        {/each}
      </div>
    </div>

    <div class="card p-4 space-y-3">
      <div class="flex items-center justify-between">
        <div class="label-caps text-primary">Alternate Texts</div>
        <button class="text-xs text-primary hover:underline" type="button" on:click={addAlternateText}>
          + Add Text
        </button>
      </div>
      <div class="space-y-2">
        {#each $store.selectedElement.alternateTexts as text, index}
          <div class="flex items-center gap-2">
            <input
              class="input-base text-xs flex-1"
              bind:value={$store.selectedElement.alternateTexts[index]}
              placeholder="Alternate response"
            />
            <button class="text-xs text-accent-red hover:underline" type="button" on:click={() => removeAlternateText(index)}>
              Remove
            </button>
          </div>
        {/each}
      </div>
    </div>
  </div>
</CRUDEditor>
