<style>
  #terminalWindow {
    padding: 1em;
    background: #000;
    border-width: 2px;
    border-style: solid;
    border-color: #33ff2266;
    border-radius: 0.5em;
    margin-top: 10em;
  }

  
</style>

<script>
  import "../../node_modules/xterm/css/xterm.css";
  import { onMount, onDestroy } from "svelte";
  import { createAuth, getAuth } from "../auth.js";
  import axios from "axios";
  import xterm from "xterm";
  import LocalEchoController from "./echo/LocalEchoController";
  import fit from "xterm-addon-fit";
  import { createClient, getClient } from "./Client";

  let client;
  let term;
  let ws;

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
    if ($isAuthenticated && $authToken && client && !ws) {
      ws = new WebSocket("ws://localhost:8010/ws?access_token=" + $authToken);
      client.setWSClient(ws);
    }
  }

  function sleep(ms) {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  function readLine(localEcho, term) {
    localEcho
      .read("~$ ")
      .then((input) => {
        client.onInput(input);
        readLine(localEcho, term);
      })
      .catch((error) => console.log(`Error reading: ${error}`));
  }

  const createRenderer = (term, localEcho) => {
    return (data) => {
      localEcho.clearInput();
      term.writeln(data);
    };
  };

  async function setupTerminal() {
    term = new xterm.Terminal();

    var fitAddon = new fit.FitAddon();

    term.loadAddon(fitAddon);

    // setup terminal
    term.setOption("cursorBlink", true);
    term.setOption("convertEol", true);

    term.open(document.getElementById("terminal"));
    fitAddon.fit();

    const localEcho = new LocalEchoController(term);
    localEcho.addAutocompleteHandler(autocompleteCommonCommands);
    client = createClient(createRenderer(term, localEcho));

    readLine(localEcho, term);
  }

  onMount(async () => {
    // change global background
    document.body.style.backgroundImage = "url('bg.jpg')";
    document.body.style.backdropFilter =
      "blur(10px) saturate(30%) brightness(50%)";

    var nav = document.querySelector("nav");
    nav.style.backgroundColor = "#00000000";
    setupTerminal();
  });

  onDestroy(async () => {
    // change global background
    document.body.style.backgroundImage = "";
    document.body.style.backdropFilter = "";

    var nav = document.querySelector("nav");
    nav.style.backgroundColor = "#00000055";
    setupTerminal();
  });

  function autocompleteCommonCommands(index, tokens) {
    if (index == 0) return ["north", "east", "south", "west", "say"];
    return [];
  }
</script>

<div id="terminalWindow" class="z-depth-5">
  <div id="terminal"></div>
</div>
