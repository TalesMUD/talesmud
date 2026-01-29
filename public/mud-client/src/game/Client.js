// src/game/Client.js

import { onMount } from "svelte";
import { writable, get } from "svelte/store";

const GAME_CLIENT = writable(null);

function createClient(renderer, characterCreator, muxStore) {
  let ws;
  let messageHandlers = new Map();
  let wsurl = "";

  let mux = muxStore;

  let activeRoom = {};
  let currentCharacter = {};


  messageHandlers["enterRoom"] = (msg) => {
    activeRoom = msg.room;
    renderer(msg.message);

    // Debug: log raw room data from server
    console.log("=== enterRoom message received ===");
    console.log("Full room object:", JSON.stringify(activeRoom, null, 2));
    console.log("Raw exits from server:", JSON.stringify(activeRoom.exits, null, 2));

    if (mux) {
      mux.setExits(activeRoom.exits);
      mux.setRoomInfo(activeRoom.name, activeRoom.description);

      //TODO: set default?
      if (activeRoom.meta != undefined && activeRoom.meta.background != undefined){
        mux.setBackground(activeRoom.meta.background)
      }

      if (activeRoom.actions != undefined) {
        mux.setActions(activeRoom.actions);
      } else {
        mux.setActions([]);
      }

      // Set NPCs in the room for UI rendering
      if (msg.npcs != undefined) {
        mux.setNPCs(msg.npcs);
      } else {
        mux.setNPCs([]);
      }

      // Set game context flags
      const hasItems = activeRoom.items && activeRoom.items.length > 0;
      mux.setGameContext({ hasItems });

      // Clear any active dialog when entering a new room
      mux.clearDialog();
    }
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

  messageHandlers["inventoryUpdate"] = (msg) => {
    if (mux) {
      mux.setInventory(msg.inventory, msg.equippedItems, msg.gold);
    }
  };

  // Dialog message handler - renders NPC dialog with numbered options
  messageHandlers["dialog"] = (msg) => {
    let output = "";

    // Show NPC name and text
    output += `[${msg.npcName}] ${msg.npcText}\n`;

    // Show options if any
    if (msg.options && msg.options.length > 0) {
      output += "\n";
      for (const opt of msg.options) {
        output += `${opt.index}. ${opt.text}\n`;
      }
      output += "\nEnter a number to respond:";
    }

    renderer(output);

    // Update store for UI overlay
    if (mux) {
      mux.setDialog(msg.npcName, msg.npcText, msg.options || [], msg.conversationID || "");
    }
  };

  // Dialog end message handler
  messageHandlers["dialogEnd"] = (msg) => {
    let output = "";
    if (msg.username) {
      output += `[${msg.username}] `;
    }
    output += msg.message;
    output += "\n[The conversation has ended]";
    renderer(output);

    // Clear dialog state in store
    if (mux) {
      mux.clearDialog();
    }
  };

  // Combat message handlers
  messageHandlers["combatStart"] = (msg) => {
    renderer(msg.message);
    if (mux) {
      mux.setGameContext({ inCombat: true });
    }
  };

  messageHandlers["combatTurn"] = (msg) => {
    renderer(msg.message);
    // Keep combat mode active
    if (mux) {
      mux.setGameContext({ inCombat: true });
    }
  };

  messageHandlers["combatAction"] = (msg) => {
    renderer(msg.message);
  };

  messageHandlers["combatStatus"] = (msg) => {
    renderer(msg.message);
    if (mux) {
      mux.setGameContext({ inCombat: true });
    }
  };

  messageHandlers["combatEnd"] = (msg) => {
    renderer(msg.message);
    if (mux) {
      mux.setGameContext({ inCombat: false });
    }
  };

  const setWSClient = async (wscl) => {
    ws = wscl;
    wsurl = ws.url;

    updateClient(ws);
  };

  const updateClient = (ws) => {
    ws.addEventListener("message", function (e) {
      var msg = JSON.parse(e.data);

      if (messageHandlers[msg.type]) {
        messageHandlers[msg.type](msg);
      } else {
        let message = msg.message;

        if (message === "" || message === "\n") {
          console.log("RECEIVED EMPTY MESSAGE")
          return;
        }

        if (msg.username) {
          message = msg.username + ":  " + msg.message;
        }
        renderer(message);
      }
    });

    ws.addEventListener("close", function (e) {
      renderer("Connection Closed.");
    });
  };

  const onInput = async (data) => {
    const msg = await handleInput(data);
    sendMessage(msg);
    //renderer(msg);
  };

  const sendMessage = (msg) => {
    if (!ws) return;

    if (
      ws.readyState == WebSocket.CLOSING ||
      ws.readyState == WebSocket.CLOSED
    ) {
      ws = new WebSocket(wsurl);
      updateClient(ws);
      renderer("reconnecting ...\n");
    }

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
    sendMessage,
  };

  // Set the client object in the store
  GAME_CLIENT.set(client);

  return client;
}

function getClient() {
  // Return the client object from the store
  return get(GAME_CLIENT);
}

export { createClient, getClient };
