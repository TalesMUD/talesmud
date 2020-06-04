// src/auth.js

import { onMount, setContext, getContext } from "svelte";
import { writable } from "svelte/store";

const room = {
  id: "Dungeon001_Room1",
  name: "Main Chamber",
  description:
    "You reach the Main Chamber of the Catacomb. The noise increases but you can't make out the origin of it.",
  detail:
    "You look closer to all sides of the room. After a thorough investigation you can see that parts of a wall are made up of loose rocks. You might be able to [move] these rocks.",
  exits: [
    {
      exit: "north",
      description: "Follow the door to the left",
      target: "Dungeon001_Entrance",
    },
    {
      exit: "hidden path",
      hidden: true,
      description: "You follow the hidden path on the east wall",
      target: "Dungeon001_End",
    },
  ],
  actions: [
    {
      action: "move rocks",
      description:
        "You try to move one of the medium sized rocks. Parts of the wall start to crumble and a hidden path opens up.",
    },
  ],
};

//const isLoading = writable(true);
const GAME_CLIENT = {};

function createClient(renderer) {
  let ws;
  let messageHandlers = new Map();

  let activeRoom = room;
  let currentCharacter = {};

  messageHandlers["enterRoom"] = (msg) => {
    activeRoom = msg.room;
    renderer(msg.message);
  };
  messageHandlers["characterSelected"] = (msg) => {
    currentCharacter = msg.character;
    renderer(msg.message);
  };

  const setWSClient = async (wscl) => {
    ws = wscl;

    ws.addEventListener("message", function (e) {
      var msg = JSON.parse(e.data);

      if (messageHandlers.has(msg.type)) {
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
