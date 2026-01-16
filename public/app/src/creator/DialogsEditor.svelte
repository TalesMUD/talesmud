<script>
  import { onMount } from "svelte";
  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";
  import { v4 as uuidv4 } from "uuid";
  import Toolbar from "./Toolbar.svelte";

  import { getAuth } from "../auth.js";
  const { isAuthenticated, authToken } = getAuth();

  import {
    getDialog,
    getDialogs,
    createDialog,
    updateDialog,
    deleteDialog,
  } from "../api/dialogs.js";

  const config = {
    title: "Manage Dialogs",
    actions: [],
    get: getDialogs,
    getElement: getDialog,
    create: createDialog,
    update: updateDialog,
    delete: deleteDialog,
    hideDetails: true,
    beforeSelect: (element) => {
      if (element.options === undefined) {
        element.options = [];
      }
      if (element.alternateTexts === undefined) {
        element.alternateTexts = [];
      }
      if (element.requiresVisitedDialogs === undefined) {
        element.requiresVisitedDialogs = [];
      }
    },
    refreshUI: () => {
      var elems = document.querySelectorAll("select");
      var instances = M.FormSelect.init(elems, {});

      setTimeout(function () {
        var elems = document.querySelectorAll("select");
        var instances = M.FormSelect.init(elems, {});
      }, 50);

      M.updateTextFields();
      var elems2 = document.querySelectorAll(".collapsible");
      if (elems2 != undefined) {
        var instances = M.Collapsible.init(elems2, {});
      }

      var textareas = document.querySelectorAll(".materialize-textarea");
      textareas.forEach((e) => {
        M.textareaAutoResize(e);
      });
    },

    new: (select) => {
      select({
        id: uuidv4(),
        name: "New Dialog",
        nodeId: "main",
        text: "Hello, traveler!",
        alternateTexts: [],
        options: [],
        answer: null,
        requiresVisitedDialogs: [],
        showOnlyOnce: false,
        isDialogExit: false,
        isNew: true,
      });
    },

    badge: (element) => {
      return element.nodeId || "main";
    },
  };

  const store = createStore();

  // Add a dialog option
  const addOption = () => {
    store.update((state) => {
      if (state.selectedElement.options == null) {
        state.selectedElement.options = [];
      }
      state.selectedElement.options.push({
        nodeId: "option_" + (state.selectedElement.options.length + 1),
        text: "New option",
        options: [],
        answer: null,
        showOnlyOnce: false,
        isDialogExit: false,
      });
      return state;
    });
    config.refreshUI();
  };

  // Delete a dialog option
  const deleteOption = (option) => {
    store.update((state) => {
      state.selectedElement.options = state.selectedElement.options.filter(
        (x) => x.nodeId !== option.nodeId
      );
      return state;
    });
  };

  // Add alternate text
  const addAlternateText = () => {
    store.update((state) => {
      if (state.selectedElement.alternateTexts == null) {
        state.selectedElement.alternateTexts = [];
      }
      state.selectedElement.alternateTexts.push("");
      return state;
    });
    config.refreshUI();
  };

  // Delete alternate text
  const deleteAlternateText = (index) => {
    store.update((state) => {
      state.selectedElement.alternateTexts.splice(index, 1);
      return state;
    });
  };

  // Toggle answer
  const toggleAnswer = () => {
    store.update((state) => {
      if (state.selectedElement.answer) {
        state.selectedElement.answer = null;
      } else {
        state.selectedElement.answer = {
          nodeId: "answer",
          text: "NPC response...",
          options: [],
        };
      }
      return state;
    });
    config.refreshUI();
  };

  const optionsToolbar = {
    title: "Dialog Options",
    small: true,
    actions: [
      {
        icon: "add",
        fnc: () => addOption(),
      },
    ],
  };

  const alternateTextsToolbar = {
    title: "Alternate Texts",
    small: true,
    actions: [
      {
        icon: "add",
        fnc: () => addAlternateText(),
      },
    ],
  };

  // Reactive toolbar for answer toggle
  $: answerToolbar = {
    title: "NPC Answer",
    small: true,
    actions: [
      {
        icon: ($store.selectedElement && $store.selectedElement.answer) ? "remove" : "add",
        fnc: () => toggleAnswer(),
      },
    ],
  };
</script>

<CRUDEditor store="{store}" config="{config}">
  <div slot="content">
    <div class="row">
      <div class="no-padding input-field col s6">
        <input
          placeholder="Node ID"
          id="nodeId"
          type="text"
          bind:value="{$store.selectedElement.nodeId}"
        />
        <label class="active first_label" for="nodeId">Node ID</label>
      </div>
    </div>

    <div class="row">
      <div class="input-field col s12">
        <textarea
          placeholder="Main dialog text"
          id="dialog_text"
          type="text"
          class="materialize-textarea"
          bind:value="{$store.selectedElement.text}"
        ></textarea>
        <label class="active" for="dialog_text">Text</label>
      </div>
    </div>

    <div class="row">
      <div class="col s4">
        <label>
          <input
            type="checkbox"
            bind:checked="{$store.selectedElement.showOnlyOnce}"
          />
          <span>Show Only Once</span>
        </label>
      </div>

      <div class="col s4">
        <label>
          <input
            type="checkbox"
            bind:checked="{$store.selectedElement.isDialogExit}"
          />
          <span>Is Dialog Exit</span>
        </label>
      </div>

      <div class="col s4">
        <label>
          <input
            type="checkbox"
            bind:checked="{$store.selectedElement.orderedTexts}"
          />
          <span>Ordered Texts</span>
        </label>
      </div>
    </div>
  </div>

  <div slot="extensions">
    <Toolbar toolbar="{alternateTextsToolbar}" />

    {#if $store.selectedElement && $store.selectedElement.alternateTexts && $store.selectedElement.alternateTexts.length > 0}
      <div class="card-panel blue-grey darken-3">
        {#each $store.selectedElement.alternateTexts as altText, index}
          <div class="row" style="margin-bottom: 0;">
            <div class="input-field col s10">
              <textarea
                placeholder="Alternate text {index + 1}"
                class="materialize-textarea"
                bind:value="{$store.selectedElement.alternateTexts[index]}"
              ></textarea>
              <label class="active">Alternate Text {index + 1}</label>
            </div>
            <div class="col s2">
              <button
                class="btn-small red"
                on:click="{() => deleteAlternateText(index)}"
              >
                <i class="material-icons">remove</i>
              </button>
            </div>
          </div>
        {/each}
      </div>
    {/if}

    <Toolbar toolbar="{answerToolbar}" />

    {#if $store.selectedElement && $store.selectedElement.answer}
      <div class="card-panel blue-grey darken-3">
        <div class="row">
          <div class="input-field col s4">
            <input
              placeholder="Answer Node ID"
              id="answer_nodeId"
              type="text"
              bind:value="{$store.selectedElement.answer.nodeId}"
            />
            <label class="active" for="answer_nodeId">Answer Node ID</label>
          </div>

          <div class="input-field col s8">
            <textarea
              placeholder="NPC's response"
              class="materialize-textarea"
              bind:value="{$store.selectedElement.answer.text}"
            ></textarea>
            <label class="active">Answer Text</label>
          </div>
        </div>
      </div>
    {/if}

    <Toolbar toolbar="{optionsToolbar}" />

    {#if $store.selectedElement && $store.selectedElement.options && $store.selectedElement.options.length > 0}
      <ul class="card-panel blue-grey darken-3 collapsible" style="padding: 0; border: none;">
        {#each $store.selectedElement.options as option}
          <li>
            <div class="collapsible-header blue-grey darken-2">
              <i class="material-icons">chat</i>
              <span style="color: #fff;">{option.nodeId}: {option.text.substring(0, 50)}{option.text.length > 50 ? '...' : ''}</span>
              <span class="badge" style="color: #ccc;">
                {#if option.isDialogExit}Exit{/if}
                {#if option.showOnlyOnce}Once{/if}
              </span>
            </div>
            <div class="collapsible-body blue-grey darken-3">
              <div class="row">
                <div class="input-field col s4">
                  <input
                    placeholder="Node ID"
                    type="text"
                    bind:value="{option.nodeId}"
                  />
                  <label class="active">Node ID</label>
                </div>

                <div class="input-field col s6">
                  <textarea
                    placeholder="Option text"
                    class="materialize-textarea"
                    bind:value="{option.text}"
                  ></textarea>
                  <label class="active">Option Text</label>
                </div>

                <div class="col s2">
                  <button
                    class="btn-small red"
                    on:click="{() => deleteOption(option)}"
                  >
                    <i class="material-icons">delete</i>
                  </button>
                </div>
              </div>

              <div class="row">
                <div class="col s4">
                  <label>
                    <input
                      type="checkbox"
                      bind:checked="{option.showOnlyOnce}"
                    />
                    <span>Show Only Once</span>
                  </label>
                </div>

                <div class="col s4">
                  <label>
                    <input
                      type="checkbox"
                      bind:checked="{option.isDialogExit}"
                    />
                    <span>Is Dialog Exit</span>
                  </label>
                </div>
              </div>

              {#if option.answer}
                <div class="row" style="margin-top: 1em;">
                  <div class="col s12">
                    <label style="color: #26a69a;">Answer:</label>
                  </div>
                  <div class="input-field col s4">
                    <input
                      placeholder="Answer Node ID"
                      type="text"
                      bind:value="{option.answer.nodeId}"
                    />
                    <label class="active">Answer Node ID</label>
                  </div>
                  <div class="input-field col s8">
                    <textarea
                      placeholder="Answer text"
                      class="materialize-textarea"
                      bind:value="{option.answer.text}"
                    ></textarea>
                    <label class="active">Answer Text</label>
                  </div>
                </div>
              {:else}
                <div class="row">
                  <div class="col s12">
                    <button
                      class="btn-small"
                      on:click="{() => {
                        option.answer = { nodeId: 'response', text: 'Response...' };
                        config.refreshUI();
                      }}"
                    >
                      <i class="material-icons left">add</i>Add Answer
                    </button>
                  </div>
                </div>
              {/if}
            </div>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</CRUDEditor>
