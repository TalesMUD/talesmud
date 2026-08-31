import { writable, derived } from "svelte/store";

// Cardinal direction names for compass filtering
const CARDINAL_DIRECTIONS = ["north", "south", "east", "west"];
const VERTICAL_DIRECTIONS = ["up", "down"];
const ALL_MAPPED_DIRECTIONS = [...CARDINAL_DIRECTIONS, ...VERTICAL_DIRECTIONS];

// Common aliases for direction names
const DIRECTION_ALIASES = {
  upward: "up",
  upwards: "up",
  ascend: "up",
  downward: "down",
  downwards: "down",
  descend: "down",
};

// Offsets for all mapped directions (x, y, z)
const DIRECTION_OFFSETS = {
  north: [0, -1, 0],
  south: [0, 1, 0],
  east: [1, 0, 0],
  west: [-1, 0, 0],
  up: [0, 0, 1],
  down: [0, 0, -1],
};

const VISITED_ROOMS_KEY = 'talesmud_visitedRooms';

function emptyAtlas() {
  return {
    characterId: "",
    currentRoomId: "",
    currentLayer: "overworld",
    layers: [],
    places: [],
    paths: [],
    regions: [],
  };
}

function loadVisitedRooms() {
  try {
    const data = localStorage.getItem(VISITED_ROOMS_KEY);
    return data ? JSON.parse(data) : {};
  } catch {
    return {};
  }
}

function saveVisitedRooms(rooms) {
  try {
    localStorage.setItem(VISITED_ROOMS_KEY, JSON.stringify(rooms));
  } catch { /* storage full or unavailable */ }
}

function createStore() {
  const { subscribe, set, update } = writable({
    // Room data
    exits: [],
    actions: [],
    npcs: [],
    players: [],
    background: "oldtown-griphon",
    roomName: "",
    roomDescription: "",

    // Connection/session state
    connectionStatus: "disconnected", // disconnected | connecting | connected | reconnecting
    connectionMessage: "Disconnected",
    reconnectAttempt: 0,

    // Dialog state
    dialogActive: false,
    dialogNpcName: "",
    dialogNpcText: "",
    dialogOptions: [],
    dialogConversationID: "",

    // Game context flags
    inCombat: false,
    combatEnemies: [],
    combatPlayers: [],
    hasItems: false,
    hasMerchant: false,
    groundItems: [],

    // Inventory state
    inventory: [],
    equippedItems: {},
    gold: 0,

    // Character state
    character: null, // Full character object from characterSelected
    characterStats: {
      currentHitPoints: 0,
      maxHitPoints: 0,
      currentMana: 0,
      maxMana: 0,
      xp: 0,
      xpForNextLevel: 0,
      level: 0,
      gold: 0,
      inCombat: false,
      attributes: [],
      equippedSkills: [],
      unspentAttributePoints: 0,
      spentAttributePoints: {},
      attackPower: 0,
      attackAttr: "STR",
      weaponDamage: 0,
      attackMod: 0,
      defense: 0,
      manaRegen: 0,
    },

    // Quest state
    quests: [], // Array of quest log entries
    questNotifications: [], // Recent quest events for toast notifications

    // Minimap / atlas state
    currentRoomId: null,
    visitedRooms: loadVisitedRooms(),
    atlas: emptyAtlas(),
    atlasLayer: null,
  });

  const store = {
    subscribe,
    update,
    set,

    // Room data methods
    setBackground: (background) => {
      update((state) => {
        state.background = background;
        return state;
      });
    },
    setExits: (exits) => {
      update((state) => {
        state.exits = exits || [];
        return state;
      });
    },
    setActions: (actions) => {
      update((state) => {
        state.actions = actions || [];
        return state;
      });
    },
    setNPCs: (npcs) => {
      update((state) => {
        state.npcs = npcs || [];
        // Derive hasMerchant from NPCs
        state.hasMerchant = (npcs || []).some(n => n.isMerchant);
        return state;
      });
    },
    setPlayers: (players) => {
      update((state) => {
        state.players = players || [];
        return state;
      });
    },
    setRoomInfo: (name, description) => {
      update((state) => {
        state.roomName = name || "";
        state.roomDescription = description || "";
        return state;
      });
    },
    setConnectionState: (status, message = "", reconnectAttempt = 0) => {
      update((state) => {
        state.connectionStatus = status || "disconnected";
        state.connectionMessage = message || "";
        state.reconnectAttempt = reconnectAttempt || 0;
        return state;
      });
    },

    // Dialog methods
    setDialog: (npcName, npcText, options, conversationID) => {
      update((state) => {
        state.dialogActive = true;
        state.dialogNpcName = npcName || "";
        state.dialogNpcText = npcText || "";
        state.dialogOptions = options || [];
        state.dialogConversationID = conversationID || "";
        return state;
      });
    },
    clearDialog: () => {
      update((state) => {
        state.dialogActive = false;
        state.dialogNpcName = "";
        state.dialogNpcText = "";
        state.dialogOptions = [];
        state.dialogConversationID = "";
        return state;
      });
    },

    // Inventory methods
    setInventory: (inventory, equippedItems, gold) => {
      update((state) => {
        state.inventory = (inventory && inventory.items) || [];
        state.equippedItems = equippedItems || {};
        state.gold = gold || 0;
        return state;
      });
    },

    // Character methods
    setCharacter: (character) => {
      update((state) => {
        state.character = character;
        if (character) {
          state.characterStats = {
            currentHitPoints: character.currentHitPoints || 0,
            maxHitPoints: character.maxHitPoints || 0,
            currentMana: character.currentMana || 0,
            maxMana: character.maxMana || 0,
            xp: character.xp || 0,
            xpForNextLevel: state.characterStats.xpForNextLevel || 0,
            level: character.level || 0,
            gold: character.gold || 0,
            inCombat: character.inCombat || false,
            attributes: character.attributes || [],
            equippedSkills: character.equippedSkills || [],
            unspentAttributePoints: character.unspentAttributePoints || 0,
            spentAttributePoints: character.spentAttributePoints || {},
            attackPower: character.attackPower || 0,
            attackAttr: character.attackAttr || "STR",
            weaponDamage: character.weaponDamage || 0,
            attackMod: character.attackMod || 0,
            defense: character.defense || 0,
            manaRegen: character.manaRegen || 0,
          };
          state.gold = character.gold || 0;
        }
        return state;
      });
    },
    updateCharacterStats: (stats) => {
      update((state) => {
        state.characterStats = {
          ...state.characterStats,
          currentHitPoints: stats.currentHitPoints ?? state.characterStats.currentHitPoints,
          maxHitPoints: stats.maxHitPoints ?? state.characterStats.maxHitPoints,
          currentMana: stats.currentMana ?? state.characterStats.currentMana,
          maxMana: stats.maxMana ?? state.characterStats.maxMana,
          xp: stats.xp ?? state.characterStats.xp,
          xpForNextLevel: stats.xpForNextLevel ?? state.characterStats.xpForNextLevel,
          level: stats.level ?? state.characterStats.level,
          gold: stats.gold ?? state.characterStats.gold,
          inCombat: stats.inCombat ?? state.characterStats.inCombat,
          attributes: stats.attributes || state.characterStats.attributes,
          equippedSkills: stats.equippedSkills || state.characterStats.equippedSkills,
          unspentAttributePoints: stats.unspentAttributePoints ?? state.characterStats.unspentAttributePoints,
          spentAttributePoints: stats.spentAttributePoints || state.characterStats.spentAttributePoints,
          attackPower: stats.attackPower ?? state.characterStats.attackPower,
          attackAttr: stats.attackAttr || state.characterStats.attackAttr,
          weaponDamage: stats.weaponDamage ?? state.characterStats.weaponDamage,
          attackMod: stats.attackMod ?? state.characterStats.attackMod,
          defense: stats.defense ?? state.characterStats.defense,
          manaRegen: stats.manaRegen ?? state.characterStats.manaRegen,
        };
        // Keep gold in sync
        if (stats.gold !== undefined) {
          state.gold = stats.gold;
        }
        return state;
      });
    },

    // Ground items methods
    setGroundItems: (items) => {
      update((state) => {
        state.groundItems = items || [];
        state.hasItems = (items || []).length > 0;
        return state;
      });
    },
    removeGroundItem: (itemId) => {
      update((state) => {
        state.groundItems = state.groundItems.filter(i => i.id !== itemId);
        state.hasItems = state.groundItems.length > 0;
        return state;
      });
    },

    // Game context methods
    setCombatants: (enemies, players) => {
      update((state) => {
        state.combatEnemies = enemies || [];
        state.combatPlayers = players || [];
        return state;
      });
    },

    setGameContext: ({ inCombat, hasItems, hasMerchant } = {}) => {
      update((state) => {
        if (inCombat !== undefined) state.inCombat = inCombat;
        if (hasItems !== undefined) state.hasItems = hasItems;
        if (hasMerchant !== undefined) state.hasMerchant = hasMerchant;
        return state;
      });
    },

    // Quest methods
    updateQuests: (questLog) => {
      update((state) => {
        state.quests = questLog || [];
        return state;
      });
    },
    addQuestNotification: (notification) => {
      update((state) => {
        state.questNotifications = [...state.questNotifications, notification].slice(-4);
        return state;
      });

      // Auto-remove after 5 seconds
      setTimeout(() => {
        update((state) => {
          state.questNotifications = state.questNotifications.filter(n => n !== notification);
          return state;
        });
      }, 5000);
    },

    // Minimap methods
    trackRoomVisit: (room) => {
      update((state) => {
        const prevRoomId = state.currentRoomId;
        state.currentRoomId = room.id;

        // Extract navigable exits (cardinal + vertical) for spatial tracking
        // Note: hidden exits are included — they still define spatial relationships
        const cardinalExits = (room.exits || [])
          .filter(e => {
            const name = (e.name || e.Name || "").toLowerCase();
            const normalized = DIRECTION_ALIASES[name] || name;
            return ALL_MAPPED_DIRECTIONS.includes(normalized);
          })
          .map(e => {
            const name = (e.name || e.Name || "").toLowerCase();
            return {
              dir: DIRECTION_ALIASES[name] || name,
              targetId: e.target || e.Target || "",
            };
          });

        // Determine coords
        let coords = null;
        if (room.coords) {
          coords = { x: room.coords.x, y: room.coords.y, z: room.coords.z || 0 };
        } else if (prevRoomId && state.visitedRooms[prevRoomId]) {
          const prevRoom = state.visitedRooms[prevRoomId];
          if (prevRoom.coords) {
            // Try mapped direction exit from prevRoom to this room
            const exitToHere = prevRoom.cardinalExits.find(e => e.targetId === room.id);
            if (exitToHere && DIRECTION_OFFSETS[exitToHere.dir]) {
              const [dx, dy, dz] = DIRECTION_OFFSETS[exitToHere.dir];
              coords = { x: prevRoom.coords.x + dx, y: prevRoom.coords.y + dy, z: prevRoom.coords.z + dz };
            } else {
              // Fallback: we navigated from prevRoom but exit name is non-standard
              // (e.g., "residence", "portal", etc.) — place adjacent to prevent coord chain break
              const occupiedKeys = new Set(
                Object.values(state.visitedRooms)
                  .filter(r => r.coords && r.coords.z === prevRoom.coords.z)
                  .map(r => `${r.coords.x},${r.coords.y}`)
              );
              const fallbackOffsets = [[1, 0], [0, -1], [-1, 0], [0, 1], [1, -1], [-1, -1], [1, 1], [-1, 1]];
              for (const [dx, dy] of fallbackOffsets) {
                const key = `${prevRoom.coords.x + dx},${prevRoom.coords.y + dy}`;
                if (!occupiedKeys.has(key)) {
                  coords = { x: prevRoom.coords.x + dx, y: prevRoom.coords.y + dy, z: prevRoom.coords.z };
                  break;
                }
              }
              // Last resort: place at offset even if occupied
              if (!coords) {
                coords = { x: prevRoom.coords.x + 2, y: prevRoom.coords.y, z: prevRoom.coords.z };
              }
            }
          }
        }

        // Default first room to origin
        if (!coords && Object.keys(state.visitedRooms).length === 0) {
          coords = { x: 0, y: 0, z: 0 };
        }

        // Store room data (only update if we have coords or room wasn't visited yet)
        if (coords || !state.visitedRooms[room.id]) {
          state.visitedRooms = {
            ...state.visitedRooms,
            [room.id]: {
              id: room.id,
              name: room.name || "",
              coords,
              cardinalExits,
            },
          };
        } else if (state.visitedRooms[room.id]) {
          // Update exits even if we can't determine coords
          state.visitedRooms = {
            ...state.visitedRooms,
            [room.id]: {
              ...state.visitedRooms[room.id],
              cardinalExits,
            },
          };
        }

        saveVisitedRooms(state.visitedRooms);
        return state;
      });
    },
    clearVisitedRooms: () => {
      update((state) => {
        state.visitedRooms = {};
        state.currentRoomId = null;
        saveVisitedRooms({});
        return state;
      });
    },

    setAtlas: (atlas) => {
      update((state) => {
        state.atlas = atlas && Array.isArray(atlas.places) ? atlas : emptyAtlas();
        if (state.atlas.currentRoomId) {
          state.currentRoomId = state.atlas.currentRoomId;
        }
        const here = (state.atlas.places || []).find(p => p.id === state.currentRoomId);
        if (here && here.layer) {
          state.atlasLayer = here.layer;
        } else if (state.atlas.currentLayer) {
          state.atlasLayer = state.atlas.currentLayer;
        }
        return state;
      });
    },

    setAtlasLayer: (layerId) => {
      update((state) => {
        state.atlasLayer = layerId;
        return state;
      });
    },
  };

  return store;
}

// Helper function to get cardinal exits from store value
function getCardinalExits(exits) {
  const exitArray = exits || [];

  // Debug: log what we're working with
  console.log("getCardinalExits input:", exitArray);

  return CARDINAL_DIRECTIONS.map(dir => {
    // Check multiple possible property names and formats
    const exit = exitArray.find(e => {
      // Try all possible property names for exit name
      const exitName = (
        e.name || e.Name || e.direction || e.Direction ||
        e.exit || e.Exit || e.id || e.ID || ""
      ).toLowerCase().trim();

      // Try all possible property names for hidden flag
      const isHidden = e.hidden || e.Hidden || e.isHidden || false;

      const matches = exitName === dir && !isHidden;
      if (matches) {
        console.log(`Found exit match for ${dir}:`, e);
      }
      return matches;
    });
    return { name: dir, available: !!exit };
  });
}

// Helper function to get special exits (non-cardinal, non-hidden)
function getSpecialExits(exits) {
  return (exits || []).filter(e => {
    const name = e.name?.toLowerCase();
    return !e.hidden &&
           !CARDINAL_DIRECTIONS.includes(name) &&
           !VERTICAL_DIRECTIONS.includes(name);
  });
}

// Helper function to get vertical exits (up/down)
function getVerticalExits(exits) {
  return (exits || []).filter(e => {
    const name = e.name?.toLowerCase();
    return !e.hidden && VERTICAL_DIRECTIONS.includes(name);
  });
}

// Helper to find NPC by name for dialog type detection
function findNpcByName(npcs, npcName) {
  return (npcs || []).find(n =>
    n.name === npcName ||
    n.displayName === npcName ||
    n.name?.toLowerCase() === npcName?.toLowerCase()
  );
}

export {
  createStore,
  getCardinalExits,
  getSpecialExits,
  getVerticalExits,
  findNpcByName
};
