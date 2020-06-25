<style>
  .editor {
    font-family: "Source Code Pro", monospace;
    font-size: 14px;
    font-weight: 400;
    height: 340px;
    letter-spacing: normal;
    line-height: 20px;
    padding: 10px;
    tab-size: 4;
    background-color: #121212;
  }
  .body {
    font-family: "Source Code Pro", monospace;
    font-size: 14px;
    font-weight: 400;
    height: 340px;
    letter-spacing: normal;
    line-height: 20px;
    padding: 10px;
    tab-size: 4;
    background-color: #121212;
  }
  .result {
    font-family: "Source Code Pro", monospace;
    font-size: 14px;
    font-weight: 400;
    height: 340px;
    letter-spacing: normal;
    line-height: 20px;
    padding: 10px;
    tab-size: 4;
    color: #333;
    background-color: #eee;
  }
</style>

<script>
  import { Route } from "svelte-routing";
  import { writable } from "svelte/store";
  import Toolbar from "./Toolbar.svelte";
  import Sprites from "./../game/Sprites.svelte";
  import { onMount } from "svelte";
  import CRUDEditor from "./CRUDEditor.svelte";
  import { createStore } from "./CRUDEditorStore.js";
  import { v4 as uuidv4 } from "uuid";

  import { CodeJar } from "codejar";
  import { withLineNumbers } from "codejar/linenumbers";
  import { createAuth, getAuth } from "../auth.js";
  const { isAuthenticated, authToken } = getAuth();
  $: state = {
    isAuthenticated: $isAuthenticated,
    authToken: $authToken.slice(0, 20),
  };

  import {
    getScript,
    deleteScript,
    runScript,
    getScripts,
    updateScript,
    createScript,
  } from "../api/scripts.js";

  let jar;
  let test;

  const testBody = writable('{}')
  const result = writable("result");

  const highlight = (editor) => {
    editor.textContent = editor.textContent;
    hljs.highlightBlock(editor);
  };
  const config = {
    title: "Manage Scripts",
    actions: [],
    get: getScripts,
    getElement: getScript,
    create: createScript,
    update: updateScript,
    delete: deleteScript,
    refreshUI: () => {
      if ($store.selectedElement) {
        if (jar === undefined) {
          jar = CodeJar(
            document.querySelector(".editor"),
            withLineNumbers(highlight)
          );
          jar.onUpdate((code) => {
            updateCode(code);
          });
          test = CodeJar(
            document.querySelector(".body"),
            withLineNumbers(highlight)
          );
          test.onUpdate((code) => {
            testBody.set(code)
          });
        }

        jar.updateCode($store.selectedElement.code);
      }
    },
    hideDetails: true,
    new: (select) => {
      select({
        id: uuidv4(),
        name: "New Script",
        description: "something",
        code: "console.log('Hello World from Script');",
        isNew: true,
      });
    },
  };
  // create store outside of the component to use it in the slot..
  const store = createStore();
  const runCode = () => {
    runScript(
      $authToken,
      $store.selectedElement.id,
      $testBody,
      (r) => {
        result.set(JSON.stringify(r.result, undefined, 2));
      },
      () => {
        console.log("update error.");
      }
    );
  };

  const updateCode = (code) => {
    store.update((state) => {
      state.selectedElement.code = code;

      return state;
    });
  };

  onMount(async () => {});
  /////////
</script>

<CRUDEditor store="{store}" config="{config}">
  <div slot="extensions">
    <div class="card-panel editor language-javascript"></div>

    <Toolbar
      toolbar="{{ title: 'Testrunner', small: true, actions: [{ name: 'Run', fnc: () => runCode() }] }}"
    />

    <div class="card z-depth-3" style="background-color: #121212;">
      <div class="card-content language-javascript body col s6">
        {$testBody}
      </div>
      <div class="card-content language-javascript result col s6">
        {$result}
      </div>
    </div>
  </div>
</CRUDEditor>
