<style>
  .gameContainer {
    padding: 1em;
    margin: auto auto;
    max-width: 900px;
  }
  .roomImage {
    background: #000;
    border-width: 1px;
    border-style: solid;
    border-color: #ffffff33;
    border-radius: 0.5em;
    height: 100%;

  }

  .roomImageInner {
    background-image: linear-gradient(
        to bottom,
        rgba(0, 0, 0, 0),
        rgba(0, 0, 0, 0),
        rgba(0, 0, 0, 0.1),
        rgba(0, 0, 0, 0.4),
        rgba(0, 0, 0, 1)
      ),
      url("/img/bg/oldtown-griphon.png");
    width: 100%;
    height: 350px;
    background-repeat: no-repeat;
    background-size: 100% auto;
    border-radius: 0.5em;

    image-rendering: pixelated;
  }

  #terminalWindow {
    max-width: 640px;
    height: 400px;
    margin: -130px auto;

    background: #000;
    border-width: 1px;
    border-style: solid;
    border-color: #ffffff33;
    border-radius: 0.5em;
  }
  #terminal {
    background: #000;
    margin: 1em;
    padding-bottom: 1em;
    height: 380px;
  }
  #terminal2 {
    background: #000;
  }
</style>

<script>
  import { writable } from "svelte/store";
  import MUDXPlus from "./MUDXPlus.svelte";
  import { createStore } from "./MUDXPlusStore";

  import MediaQuery from "../MediaQuery.svelte";

  import CharacterCreator from "./../characters/CharacterCreator.svelte";
  import "../../node_modules/xterm/css/xterm.css";
  import { onMount, onDestroy } from "svelte";
  import { createAuth, getAuth } from "../auth.js";
  import axios from "axios";
  import xterm from "xterm";
  import LocalEchoController from "./echo/LocalEchoController";
  import fit from "xterm-addon-fit";
  import { createClient, getClient } from "./Client";
  import { wsbackend } from "../api/base.js";
  import UserMenu from "../UserMenu.svelte";

  let client;
  let term;
  let ws;

  const muxStore = createStore();
  const muxClient = writable({});
  let muxplus = true;

  const { isAuthenticated, authToken } = getAuth();
  $: state = {
    isAuthenticated: $isAuthenticated,
    authToken: $authToken.slice(0, 20),
  };

  $: {
    if (client && !ws) {
      // connect to websocket server
      const url = wsbackend + "?access_token=";
      ws = new WebSocket(url + $authToken);
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

  const characterCreator = () => {
    console.log("CREATE CHARACTER");

    var Modalelem = document.querySelector(".modal");
    var instance = M.Modal.init(Modalelem);
    instance.open();
  };

  async function setupTerminal() {
    term = new xterm.Terminal();
    var fitAddon = new fit.FitAddon();
    term.loadAddon(fitAddon);
    term.setOption("cursorBlink", true);
    term.setOption("convertEol", true);

    term.open(document.getElementById("terminal"));
    fitAddon.fit();

    const localEcho = new LocalEchoController(term);
    localEcho.addAutocompleteHandler(autocompleteCommonCommands);
    client = createClient(
      createRenderer(term, localEcho),
      characterCreator,
      muxStore
    );

    muxClient.set(client);

    readLine(localEcho, term);
  }

  onMount(async () => {
    // change global background<img src="/img/bg/oldtown-griphon.png"/>

    document.body.style.backgroundImage = "url('/img/bg/oldtown-griphon.png')";
    document.body.style.backdropFilter =
      "blur(10px) saturate(30%) brightness(50%)";

    var nav = document.querySelector("nav");
    if (nav) {
      nav.style.backgroundColor = "#00000000";
      setupTerminal();
    }
  });

  onDestroy(async () => {
    // change global background
    document.body.style.backgroundImage = "";
    document.body.style.backdropFilter = "";

    var nav = document.querySelector("nav");
    if (nav) {
      nav.style.backgroundColor = "#00000055";
      setupTerminal();
    }
  });

  function autocompleteCommonCommands(index, tokens) {
    if (index == 0) return ["north", "east", "south", "west", "say"];
    return [];
  }
</script>

<CharacterCreator />

<div class="gameContainer">
  <div class="roomImage center-align">
    <div class="roomImageInner center-align"></div>

    <div id="terminalWindow" class="z-depth-5">
      <div id="terminal"></div>
    </div>

    <MUDXPlus
      store="{muxStore}"
      term="{term}"
      sendMessage="{(msg) => {
        client.sendMessage(msg);
      }}"
    />
  </div>

</div>
