// src/auth.js

import { onMount, setContext, getContext } from "svelte";
import { writable } from "svelte/store";

//const isLoading = writable(true);
const GAME_CLIENT = {};

function createClient(renderer, characterCreator) {
  let ws;
  let messageHandlers = new Map();

  let activeRoom = {};
  let currentCharacter = {};

  messageHandlers["enterRoom"] = (msg) => {
    activeRoom = msg.room;
    renderer(msg.message);
  };

  messageHandlers["createCharacter"] = (msg) => {
    renderer(msg.message);

    if (characterCreator) {
      characterCreator();
      //TODO: send select character
    }
  };
  messageHandlers["characterSelected"] = (msg) => {
    currentCharacter = msg.character;
    renderer(msg.message);
  };

  const setWSClient = async (wscl) => {
    ws = wscl;

    ws.addEventListener("message", function (e) {
      var msg = JSON.parse(e.data);

      if (messageHandlers[msg.type]) {
        messageHandlers[msg.type](msg);
      } else {
        let message = msg.message;

        if (msg.username) {
          message = msg.username + ":  " + msg.message;
        }
        renderer(message);
      }
    });
  };

  const onInput = async (data) => {
    const msg = await handleInput(data);
    sendMessage(msg);
    //renderer(msg);
  };

  const sendMessage = (msg) => {
    if (!ws) return;
    ws.send(
      JSON.stringify({
        message: msg,
        type: "message",
      })
    );
  };

  const renderRoom = async (room) => {
    renderer(room.description);
  };

  const handleInput = async (data) => {
    return `${data}`;
  };

  const client = {
    onInput,
    setWSClient,
  };

  // setInterval(function () {
  //   renderer("\n<The lights in front of you are flickering>")
  // }, 5000);

  setContext(GAME_CLIENT, client);
  return client;
}

function getClient() {
  return getContext(GAME_CLIENT);
}

export { createClient, getClient };
