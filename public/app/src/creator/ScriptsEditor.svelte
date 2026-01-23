<script>
  import { onMount } from "svelte";
  import { writable, get } from "svelte/store";
  import { v4 as uuidv4 } from "uuid";
  import { CodeJar } from "codejar";
  import { withLineNumbers } from "codejar/linenumbers";
  import hljs from "highlight.js/lib/core";
  import lua from "highlight.js/lib/languages/lua";
  import "highlight.js/styles/atom-one-dark.css";

  import CRUDEditor from "./CRUDEditor.svelte";
  import ScriptsGuideModal from "./ScriptsGuideModal.svelte";
  import { createStore } from "./CRUDEditorStore.js";
  import { getAuth } from "../auth.js";

  import {
    getScript,
    deleteScript,
    runScript,
    getScripts,
    updateScript,
    createScript,
    getScriptTypes,
  } from "../api/scripts.js";

  hljs.registerLanguage("lua", lua);

  const { isAuthenticated, authToken } = getAuth();
  $: state = {
    isAuthenticated: $isAuthenticated,
    authToken: $authToken.slice(0, 20),
  };

  const store = createStore();
  let scriptTypes = [];
  let jar;
  let test;
  let result;
  let editorRef;
  let testRef;
  let resultRef;
  let deprecatedScripts = [];
  let showGuide = false;
  let activeScriptId = null;

  const testBody = writable("{}");

  const highlight = (editor) => {
    editor.textContent = editor.textContent;
    hljs.highlightElement(editor);
  };

  const getLuaScripts = (token, filters, cb, errorCb) => {
    getScripts(
      token,
      filters,
      (all) => {
        deprecatedScripts = all.filter((script) => script.language !== "lua");
        cb(all.filter((script) => script.language === "lua"));
      },
      errorCb
    );
  };

  const config = {
    title: "Manage Scripts",
    subtitle: "Configure logic for game events and objects.",
    listTitle: "Scripts",
    labels: {
      create: "Create Script",
      update: "Update Script",
      delete: "Delete",
    },
    get: getLuaScripts,
    getElement: getScript,
    create: createScript,
    update: updateScript,
    delete: deleteScript,
    hideDetails: true,
  };

  const formatResult = (payload) => {
    if (!payload) return "";
    if (payload.success === false) {
      return `ERROR: ${payload.error || "Unknown error"}`;
    }

    const resultValue =
      typeof payload.result === "string"
        ? payload.result
        : JSON.stringify(payload.result ?? null, null, 2);

    if (payload.durationMs !== undefined) {
      return `SUCCESS (${payload.durationMs}ms)\n${resultValue}`;
    }

    return resultValue;
  };

  const runCode = () => {
    runScript(
      $authToken,
      $store.selectedElement.id,
      $testBody,
      (r) => {
        result?.updateCode(formatResult(r));
      },
      () => {
        console.log("update error.");
      }
    );
  };

  const reloadScripts = () => {
    getLuaScripts(
      $authToken,
      $store.filters,
      (all) => {
        store.setElements(all);
        store.setSelectedElement(all[0]);
      },
      (err) => console.log(err)
    );
  };

  const deleteDeprecated = async () => {
    if (!deprecatedScripts.length) return;

    await Promise.all(
      deprecatedScripts.map(
        (script) =>
          new Promise((resolve) => {
            deleteScript(
              $authToken,
              script.id,
              () => resolve(true),
              () => resolve(false)
            );
          })
      )
    );

    deprecatedScripts = [];
    reloadScripts();
  };

  config.extraActions = [
    {
      label: "Scripting Guide",
      icon: "help",
      variant: "btn-ghost",
      onClick: () => (showGuide = true),
    },
    {
      label: "Run Test",
      icon: "play_arrow",
      variant: "btn-outline",
      onClick: runCode,
    },
  ];

  config.new = (select) => {
    select({
      id: uuidv4(),
      name: "New Script",
      description: "something",
      language: "lua",
      code: `local input = ctx.ctx
tales.game.log("info", "Hello from Lua script")
return input`,
      isNew: true,
    });
  };

  const updateCode = (code) => {
    store.update((state) => {
      state.selectedElement.code = code;
      return state;
    });
  };

  const setupEditors = () => {
    if (!editorRef || !testRef || !resultRef) return;
    if (!jar) {
      jar = CodeJar(editorRef, withLineNumbers(highlight));
      jar.onUpdate((code) => updateCode(code));
    }
    if (!test) {
      test = CodeJar(testRef, withLineNumbers(highlight));
      test.onUpdate((code) => testBody.set(code));
    }
    if (!result) {
      result = CodeJar(resultRef, withLineNumbers(highlight));
    }

    if ($store.selectedElement?.id !== activeScriptId) {
      activeScriptId = $store.selectedElement?.id || null;
      jar.updateCode($store.selectedElement?.code || "");
    }
    if (test) {
      test.updateCode(get(testBody) || "{}");
    }
  };

  onMount(async () => {
    getScriptTypes((t) => (scriptTypes = t));
    setupEditors();
  });

  $: if ($store.selectedElement) {
    setupEditors();
  }
</script>

<CRUDEditor store={store} config={config}>
  <div slot="content" class="space-y-4">
    <div class="space-y-1.5">
      <label class="label-caps">Script Type</label>
      <select class="input-base" bind:value={$store.selectedElement.type}>
        <option value="" disabled selected>Script Type</option>
        {#each scriptTypes as type}
          <option value={type}>{type}</option>
        {/each}
      </select>
    </div>
    {#if deprecatedScripts.length}
      <div class="rounded-lg border border-amber-500/40 bg-amber-500/10 px-4 py-3 text-xs text-amber-200 flex items-center justify-between">
        <div>
          {deprecatedScripts.length} deprecated JavaScript scripts found.
          This editor only supports Lua scripts.
        </div>
        <button class="btn btn-outline text-xs" type="button" on:click={deleteDeprecated}>
          Delete deprecated scripts ({deprecatedScripts.length})
        </button>
      </div>
    {/if}
  </div>

  <div slot="extensions" class="space-y-4">
    <div class="flex items-center justify-between text-[10px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 px-1">
      <div class="flex items-center gap-4">
        <span>Script Code (Lua)</span>
        <span class="flex items-center gap-1 text-primary">
          <span class="w-2 h-2 rounded-full bg-primary animate-pulse"></span>
          Auto-saved
        </span>
      </div>
    </div>

    <div class="flex-1 bg-editor-dark rounded-xl border border-slate-800 shadow-2xl overflow-hidden flex flex-col min-h-[500px]">
      <div class="flex items-center gap-1.5 px-4 py-2 bg-slate-900 border-b border-slate-800">
        <div class="w-3 h-3 rounded-full bg-red-500/50"></div>
        <div class="w-3 h-3 rounded-full bg-yellow-500/50"></div>
        <div class="w-3 h-3 rounded-full bg-green-500/50"></div>
        <span class="ml-4 text-[11px] font-mono text-slate-500">
          {$store.selectedElement?.name || "script"}.lua
        </span>
      </div>
      <div class="flex-1 flex font-mono text-sm leading-relaxed overflow-hidden">
        <div
          class="flex-1 p-4 overflow-y-auto text-slate-300 whitespace-pre scrollbar-hide"
          bind:this={editorRef}
        ></div>
      </div>
      <div class="border-t border-slate-800 bg-slate-900/80 p-4 space-y-3">
        <div class="flex items-center justify-between">
          <span class="text-[10px] font-bold uppercase text-slate-500">
            Test Runner Input
          </span>
          <button class="text-xs text-primary hover:underline flex items-center gap-1" type="button" on:click={runCode}>
            <span class="material-symbols-outlined text-xs">refresh</span>
            Re-run
          </button>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="bg-black/40 rounded p-3 font-mono text-xs text-slate-200 border border-slate-800/60">
            <div bind:this={testRef}></div>
          </div>
          <div class="bg-black/40 rounded p-3 font-mono text-xs text-green-400 border border-emerald-900/30">
            <div bind:this={resultRef}></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</CRUDEditor>

<ScriptsGuideModal open={showGuide} on:close={() => (showGuide = false)} />
