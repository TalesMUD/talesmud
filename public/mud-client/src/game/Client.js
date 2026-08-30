// src/game/Client.js

import { onMount } from "svelte";
import { writable, get } from "svelte/store";
import { overlayStore } from "./ui/overlayStore.js";
import { getPlayerColor } from "./playerColors.js";

const GAME_CLIENT = writable(null);

function createClient(renderer, characterCreator, muxStore) {
  let ws;
  let messageHandlers = new Map();

  let mux = muxStore;

  let activeRoom = {};
  let currentCharacter = {};
  let authToken = "";

  const getAuthToken = () => {
    if (authToken) return authToken;
    try {
      const guest = sessionStorage.getItem("talesmud_guest_token");
      if (guest) return guest;
    } catch (e) { /* ignore */ }
    try {
      const stored = localStorage.getItem("token");
      if (stored) return stored;
    } catch (e) { /* ignore */ }
    return "";
  };


  messageHandlers["enterRoom"] = (msg) => {
    activeRoom = msg.room;
    renderer(msg.message);
    overlayStore.clearAll();

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

      // Set player characters in the room for UI rendering
      mux.setPlayers(msg.players || []);

      // Set ground items from server-resolved item details
      mux.setGroundItems(msg.items || []);

      // Clear any active dialog when entering a new room
      mux.clearDialog();

      // Track room visit for minimap fallback and refresh the atlas
      mux.trackRoomVisit(activeRoom);
      requestAtlas();
    }
  };

  // Silent room update: refresh exits/items/NPCs without re-rendering the room description.
  // Used by revealExit so narrative messages aren't buried by a full room re-display.
  messageHandlers["roomUpdate"] = (msg) => {
    activeRoom = msg.room;

    if (mux) {
      mux.setExits(activeRoom.exits);

      if (activeRoom.meta != undefined && activeRoom.meta.background != undefined){
        mux.setBackground(activeRoom.meta.background)
      }

      if (activeRoom.actions != undefined) {
        mux.setActions(activeRoom.actions);
      } else {
        mux.setActions([]);
      }

      if (msg.npcs != undefined) {
        mux.setNPCs(msg.npcs);
      } else {
        mux.setNPCs([]);
      }

      // Set player characters in the room
      mux.setPlayers(msg.players || []);

      mux.setGroundItems(msg.items || []);
      requestAtlas();
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
    if (mux && msg.character) {
      mux.setCharacter(msg.character);
      mux.updateCharacterStats({ xpForNextLevel: msg.xpForNextLevel });
    }
  };

  messageHandlers["atlas"] = (msg) => {
    if (mux && msg.atlas) {
      mux.setAtlas(msg.atlas);
    }
  };

  messageHandlers["inventoryUpdate"] = (msg) => {
    if (mux) {
      mux.setInventory(msg.inventory, msg.equippedItems, msg.gold);
      // Keep character stats gold in sync
      mux.updateCharacterStats({ gold: msg.gold });
    }
  };

  messageHandlers["characterUpdate"] = (msg) => {
    if (mux) {
      mux.updateCharacterStats({
        currentHitPoints: msg.currentHitPoints,
        maxHitPoints: msg.maxHitPoints,
        currentMana: msg.currentMana,
        maxMana: msg.maxMana,
        xp: msg.xp,
        xpForNextLevel: msg.xpForNextLevel,
        level: msg.level,
        gold: msg.gold,
        inCombat: msg.inCombat,
        attributes: msg.attributes,
        equippedSkills: msg.equippedSkills,
        unspentAttributePoints: msg.unspentAttributePoints,
        spentAttributePoints: msg.spentAttributePoints,
        attackPower: msg.attackPower,
        attackAttr: msg.attackAttr,
        weaponDamage: msg.weaponDamage,
        attackMod: msg.attackMod,
        defense: msg.defense,
        manaRegen: msg.manaRegen,
      });
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
    overlayStore.pushMessage(output);

    // Clear dialog state in store
    if (mux) {
      mux.clearDialog();
    }
  };

  // Combat message handlers
  messageHandlers["combatStart"] = (msg) => {
    renderer(msg.message);
    overlayStore.pushMessage(msg.message);
    if (mux) {
      mux.setGameContext({ inCombat: true });
      mux.updateCharacterStats({ inCombat: true });
    }
  };

  messageHandlers["combatTurn"] = (msg) => {
    renderer(msg.message);
    overlayStore.pushMessage(msg.message);
    // Keep combat mode active
    if (mux) {
      mux.setGameContext({ inCombat: true });
    }
  };

  messageHandlers["combatAction"] = (msg) => {
    renderer(msg.message);
    overlayStore.pushMessage(msg.message);
  };

  messageHandlers["combatStatus"] = (msg) => {
    renderer(msg.message);
    overlayStore.pushMessage(msg.message);
    if (mux) {
      mux.setGameContext({ inCombat: true });
    }
  };

  messageHandlers["combatEnd"] = (msg) => {
    renderer(msg.message);
    overlayStore.pushMessage(msg.message);
    if (mux) {
      mux.setGameContext({ inCombat: false });
      mux.updateCharacterStats({ inCombat: false });
    }
  };

  // Quest message handlers
  messageHandlers["questAccepted"] = (msg) => {
    renderer(msg.message);
    overlayStore.pushMessage(msg.message);

    // Update quest log
    requestQuestLog();

    // Show notification
    if (mux) {
      mux.addQuestNotification({
        id: `quest-accepted-${msg.questId || Date.now()}`,
        questId: msg.questId,
        type: 'accepted',
        questName: msg.questName || 'Quest',
        message: msg.message || 'Quest accepted',
      });
    }
  };

  messageHandlers["questProgress"] = (msg) => {
    renderer(msg.message);

    // Update quest log
    requestQuestLog();

    // Show objective progress notification
    if (mux) {
      const changedObjective = msg.changedObjective || (msg.objectives || []).find(o => o.completed);
      if (changedObjective) {
        const progressText = changedObjective.completed
          ? `Objective complete: ${changedObjective.description || 'Objective completed'}`
          : `${changedObjective.description || 'Objective'}: ${changedObjective.current}/${changedObjective.required}`;
        mux.addQuestNotification({
          id: `quest-progress-${msg.questId || Date.now()}-${changedObjective.objectiveId}-${Date.now()}`,
          questId: msg.questId,
          type: 'progress',
          questName: msg.questName || 'Quest',
          message: progressText,
        });
      }
    }
  };

  messageHandlers["questReady"] = (msg) => {
    renderer(msg.message);
    overlayStore.pushMessage(msg.message);

    requestQuestLog();

    if (mux) {
      const changedObjective = msg.changedObjective || (msg.objectives || []).find(o => o.completed);
      mux.addQuestNotification({
        id: `quest-ready-${msg.questId || Date.now()}`,
        questId: msg.questId,
        type: 'ready',
        questName: msg.questName || 'Quest',
        message: changedObjective?.description
          ? `Ready to turn in. Last objective: ${changedObjective.description}`
          : 'Ready to turn in.',
      });
    }
  };

  messageHandlers["questCompleted"] = (msg) => {
    renderer(msg.message);
    overlayStore.pushMessage(msg.message);

    // Refresh quest log
    requestQuestLog();

    // Show completion notification with rewards
    if (mux) {
      mux.addQuestNotification({
        id: `quest-completed-${msg.questId || Date.now()}`,
        questId: msg.questId,
        type: 'completed',
        questName: msg.questName || 'Quest',
        message: msg.message || 'Quest completed!',
      });
    }
  };

  messageHandlers["questAbandoned"] = (msg) => {
    renderer(msg.message);

    // Refresh quest log so the abandoned quest is removed from the active list
    requestQuestLog();
  };

  messageHandlers["questLog"] = (msg) => {
    // Update full quest log
    if (mux && msg.quests) {
      mux.updateQuests(msg.quests);
    }
  };

  const requestAtlas = () => {
    if (!currentCharacter || !currentCharacter.id) return;
    const token = getAuthToken();
    if (!token) return;

    fetch(`/api/characters/${currentCharacter.id}/map`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })
    .then(r => {
      if (!r.ok) throw new Error(`HTTP ${r.status}`);
      return r.json();
    })
    .then(data => {
      if (mux && data) {
        mux.setAtlas(data);
      }
    })
    .catch((err) => { console.warn("atlas fetch failed", err); });
  };

  // Helper function to fetch quest log from API
  const requestQuestLog = () => {
    if (!currentCharacter || !currentCharacter.id) return;

    const token = localStorage.getItem('token');
    if (!token) return;

    fetch(`/api/quest-progress/${currentCharacter.id}`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })
    .then(r => {
      if (!r.ok) throw new Error(`HTTP ${r.status}`);
      return r.json();
    })
    .then(data => {
      if (mux && Array.isArray(data)) {
        mux.updateQuests(data);
      }
    })
    .catch(err => {
      console.error('Failed to fetch quest log:', err);
    });
  };

  const setWSClient = async (wscl) => {
    ws = wscl;

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
          // Pass structured data so terminals can color the player name
          const color = getPlayerColor(msg.username);
          renderer({ username: msg.username, message: msg.message, color });
          overlayStore.pushMessage(msg.username + ":  " + msg.message);
        } else {
          renderer(message);
          overlayStore.pushMessage(message);
        }
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
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      renderer("Connection is reconnecting. Try again in a moment.");
      if (mux) {
        mux.setConnectionState("reconnecting", "Reconnecting to the game server...");
      }
      return;
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
    setAuthToken: (token) => { authToken = token || ""; },
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
