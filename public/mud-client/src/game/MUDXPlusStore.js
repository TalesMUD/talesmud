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

function mergeAtlas(existing, incoming) {
  if (!incoming || !Array.isArray(incoming.places)) {
    return existing && Array.isArray(existing.places) ? existing : emptyAtlas();
  }
  const base = existing && Array.isArray(existing.places) ? existing : emptyAtlas();

  const placeMap = new Map();
  for (const place of base.places) {
    placeMap.set(place.id, { ...place });
  }
  for (const place of incoming.places) {
    const prev = placeMap.get(place.id);
    if (!prev) {
      placeMap.set(place.id, { ...place });
      continue;
    }
    placeMap.set(place.id, {
      ...prev,
      ...place,
      discovered: prev.discovered || place.discovered,
      name: place.name || prev.name,
      landmark: prev.landmark || place.landmark,
    });
  }

  const pathKeys = new Set();
  const paths = [];
  for (const path of [...(base.paths || []), ...(incoming.paths || [])]) {
    const key = `${path.from}|${path.to}|${path.dir}`;
    if (pathKeys.has(key)) continue;
    pathKeys.add(key);
    paths.push(path);
  }

  const layerMap = new Map();
  for (const layer of [...(base.layers || []), ...(incoming.layers || [])]) {
    layerMap.set(layer.id, layer);
  }

  const regionMap = new Map();
  for (const region of [...(base.regions || []), ...(incoming.regions || [])]) {
    regionMap.set(region.id, region);
  }

  const characterId = incoming.characterId || base.characterId;
  const currentRoomId = incoming.currentRoomId || base.currentRoomId;
  const currentLayer = incoming.currentLayer || base.currentLayer;

  // Recompute current flags from currentRoomId — JSON omitempty leaves stale
  // current:true on previously visited rooms after merge.
  const places = Array.from(placeMap.values()).map((place) => ({
    ...place,
    current: isAtlasCurrentPlace(place.id, currentRoomId),
  }));
  // Exactly one current marker (exact id wins over template match).
  const exact = places.find((p) => p.id === currentRoomId);
  if (exact) {
    for (const p of places) p.current = p.id === currentRoomId;
  } else {
    let marked = false;
    for (const p of places) {
      if (p.current && !marked) {
        marked = true;
      } else {
        p.current = false;
      }
    }
  }

  return {
    characterId,
    currentRoomId,
    currentLayer,
    layers: Array.from(layerMap.values()),
    places,
    paths,
    regions: Array.from(regionMap.values()),
  };
}

function isAtlasCurrentPlace(placeId, currentRoomId) {
  if (!placeId || !currentRoomId) return false;
  if (placeId === currentRoomId) return true;
  const i = String(currentRoomId).indexOf('~');
  if (i > 0 && placeId === currentRoomId.slice(0, i)) return true;
  return false;
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


let combatLogSeq = 0;
function nextCombatLogId() {
  combatLogSeq += 1;
  return `clog-${Date.now()}-${combatLogSeq}`;
}

function normalizeCombatant(raw) {
  if (!raw) return null;
  return {
    id: raw.id || raw.ID || "",
    name: raw.name || raw.Name || "?",
    portrait: raw.portrait || raw.Portrait || "",
    hp: raw.hp ?? raw.HP ?? raw.currentHp ?? 0,
    maxHp: raw.maxHp ?? raw.MaxHP ?? raw.maxHP ?? 1,
  };
}

function normalizeCombatantList(list) {
  return (list || []).map(normalizeCombatant).filter((c) => c && c.id);
}

function appendCombatLog(log, text) {
  const clean = String(text || "").trim();
  if (!clean) return log || [];
  const next = [...(log || []), { id: nextCombatLogId(), text: clean }];
  return next.slice(-8);
}

function patchCombatantHp(list, msg) {
  const tid = msg?.targetId;
  if (!tid) return list || [];
  return (list || []).map((c) => {
    if (c.id !== tid) return c;
    return {
      ...c,
      hp: msg.remainingHp ?? c.hp,
      maxHp: msg.maxHp || c.maxHp,
    };
  });
}

function mergeCombatantSnapshots(enemies, players, snapshots) {
  const enemyMap = new Map((enemies || []).map((e) => [e.id, e]));
  const playerMap = new Map((players || []).map((p) => [p.id, p]));
  for (const snap of snapshots) {
    if (enemyMap.has(snap.id)) {
      enemyMap.set(snap.id, { ...enemyMap.get(snap.id), ...snap });
    } else if (playerMap.has(snap.id)) {
      playerMap.set(snap.id, { ...playerMap.get(snap.id), ...snap });
    } else if ((enemies || []).length && !(players || []).some((p) => p.id === snap.id)) {
      // Unknown id after start — treat as enemy refresh
      enemyMap.set(snap.id, snap);
    } else {
      playerMap.set(snap.id, snap);
    }
  }
  return {
    enemies: Array.from(enemyMap.values()),
    players: Array.from(playerMap.values()),
  };
}

function createStore() {
  const { subscribe, set, update } = writable({
    // Room data
    exits: [],
    actions: [],
    npcs: [],
    players: [],
    roomChat: [],
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

    // Merchant shop overlay
    shop: null,
    shopError: "",

    // Game context flags
    inCombat: false,
    combatPhase: "idle", // idle | active | ending
    combatEnemies: [],
    combatPlayers: [],
    combatTargetId: null,
    combatTurn: null, // { actorId, actorName, round, deadlineMs }
    combatLog: [], // thin optional log [{id,text}]
    combatOutcome: null, // victory | defeat | fled | timeout
    combatFx: null, // { fxId, at, targetId, actorId, damage, heal, result, action }
    combatEndMessage: "",
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
    mapOverviewOpen: false,
    inventoryOverlayOpen: false,
  });

  const store = {
    subscribe,
    update,
    set,

    // Room data methods
    setBackground: (background) => {
      update((state) => {
        // Skip identical backgrounds — roomUpdate fires on NPC moves and would
        // otherwise retrigger the room-art crossfade (visible flicker).
        if (state.background === background) return state;
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
        const next = npcs || [];
        const prev = state.npcs || [];
        // Avoid store churn when roomUpdate repeats the same occupants.
        if (prev.length === next.length && prev.every((p, i) => {
          const n = next[i];
          return p && n && p.id === n.id
            && p.currentHp === n.currentHp
            && p.maxHp === n.maxHp
            && p.displayName === n.displayName
            && !!p.isEnemy === !!n.isEnemy
            && !!p.isMerchant === !!n.isMerchant
            && !!p.isQuestGiver === !!n.isQuestGiver
            && !!p.hasDialog === !!n.hasDialog
            && (p.portrait || '') === (n.portrait || '');
        })) {
          return state;
        }
        state.npcs = next;
        state.hasMerchant = next.some(n => n.isMerchant);
        return state;
      });
    },
    setPlayers: (players) => {
      update((state) => {
        state.players = players || [];
        return state;
      });
    },
    appendRoomChat: (line) => {
      if (!line || !line.text) return;
      update((state) => {
        const next = {
          id: line.id || `chat-${Date.now()}-${(state.roomChat || []).length}`,
          name: line.name || "",
          text: line.text,
          isYou: !!line.isYou,
        };
        state.roomChat = [...(state.roomChat || []), next].slice(-40);
        return state;
      });
    },
    clearRoomChat: () => {
      update((state) => {
        state.roomChat = [];
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
        const nextStatus = status || "disconnected";
        const nextMessage = message || "";
        const nextAttempt = reconnectAttempt || 0;
        // No-op when unchanged — avoids Svelte churn that flaps the status light.
        if (
          state.connectionStatus === nextStatus &&
          state.connectionMessage === nextMessage &&
          state.reconnectAttempt === nextAttempt
        ) {
          return state;
        }
        state.connectionStatus = nextStatus;
        state.connectionMessage = nextMessage;
        state.reconnectAttempt = nextAttempt;
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

    /** While Talk is open, flip [Quest] Name → [In Progress] Name after accept. */
    markDialogQuestAccepted: (questName) => {
      update((state) => {
        if (!state.dialogActive) return state;
        const needle = String(questName || "").trim().toLowerCase();
        if (!needle) return state;
        state.dialogOptions = (state.dialogOptions || []).map((opt) => {
          const text = String(opt?.text || "");
          const match = text.match(/^\[Quest\]\s*(.+)$/i);
          if (match && match[1].trim().toLowerCase() === needle) {
            return { ...opt, text: `[In Progress] ${match[1].trim()}` };
          }
          return opt;
        });
        return state;
      });
    },

    setShop: (shop) => {
      update((state) => {
        state.shop = shop || null;
        state.shopError = "";
        if (shop && typeof shop.gold === "number") {
          state.gold = shop.gold;
        }
        return state;
      });
    },
    clearShop: () => {
      update((state) => {
        state.shop = null;
        state.shopError = "";
        return state;
      });
    },
    setShopError: (message) => {
      update((state) => {
        state.shopError = message || "";
        return state;
      });
    },
    clearShopError: () => {
      update((state) => {
        state.shopError = "";
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
        const prevId = state.character?.id;
        state.character = character;
        if (character && character.id !== prevId) {
          state.atlas = emptyAtlas();
          state.atlasLayer = null;
          state.mapOverviewOpen = false;
          state.inventoryOverlayOpen = false;
          state.visitedRooms = {};
          saveVisitedRooms({});
        }
        if (character) {
          const prev = state.characterStats || {};
          // CharacterSelected JSON omits derived combat fields (attackPower,
          // defense, etc.) — those arrive only via characterUpdate. Keep prior
          // values instead of forcing 0 (HUD ATK/DEF flicker on reconnect).
          const keepDerived = (key, fallback) =>
            character[key] != null ? character[key] : (prev[key] ?? fallback);
          state.characterStats = {
            currentHitPoints: character.currentHitPoints || 0,
            maxHitPoints: character.maxHitPoints || 0,
            currentMana: character.currentMana || 0,
            maxMana: character.maxMana || 0,
            xp: character.xp || 0,
            xpForNextLevel: prev.xpForNextLevel || 0,
            level: character.level || 0,
            gold: character.gold || 0,
            inCombat: character.inCombat || false,
            attributes: character.attributes || [],
            equippedSkills: character.equippedSkills || [],
            unspentAttributePoints: character.unspentAttributePoints || 0,
            spentAttributePoints: character.spentAttributePoints || {},
            attackPower: keepDerived("attackPower", 0),
            attackAttr: keepDerived("attackAttr", "STR"),
            weaponDamage: keepDerived("weaponDamage", 0),
            attackMod: keepDerived("attackMod", 0),
            defense: keepDerived("defense", 0),
            manaRegen: keepDerived("manaRegen", 0),
          };
          state.gold = character.gold || 0;
        }
        return state;
      });
    },
    updateCharacterStats: (stats) => {
      update((state) => {
        const prev = state.characterStats;
        const next = {
          ...prev,
          currentHitPoints: stats.currentHitPoints ?? prev.currentHitPoints,
          maxHitPoints: stats.maxHitPoints ?? prev.maxHitPoints,
          currentMana: stats.currentMana ?? prev.currentMana,
          maxMana: stats.maxMana ?? prev.maxMana,
          xp: stats.xp ?? prev.xp,
          xpForNextLevel: stats.xpForNextLevel ?? prev.xpForNextLevel,
          level: stats.level ?? prev.level,
          gold: stats.gold ?? prev.gold,
          inCombat: stats.inCombat ?? prev.inCombat,
          attributes: stats.attributes || prev.attributes,
          equippedSkills: stats.equippedSkills || prev.equippedSkills,
          unspentAttributePoints: stats.unspentAttributePoints ?? prev.unspentAttributePoints,
          spentAttributePoints: stats.spentAttributePoints || prev.spentAttributePoints,
          attackPower: stats.attackPower ?? prev.attackPower,
          attackAttr: stats.attackAttr || prev.attackAttr,
          weaponDamage: stats.weaponDamage ?? prev.weaponDamage,
          attackMod: stats.attackMod ?? prev.attackMod,
          defense: stats.defense ?? prev.defense,
          manaRegen: stats.manaRegen ?? prev.manaRegen,
        };
        const goldNext = stats.gold !== undefined ? stats.gold : state.gold;
        // No-op when nothing changed — cuts needless Svelte churn / HUD flicker.
        const sameStats =
          next.currentHitPoints === prev.currentHitPoints &&
          next.maxHitPoints === prev.maxHitPoints &&
          next.currentMana === prev.currentMana &&
          next.maxMana === prev.maxMana &&
          next.xp === prev.xp &&
          next.xpForNextLevel === prev.xpForNextLevel &&
          next.level === prev.level &&
          next.gold === prev.gold &&
          next.inCombat === prev.inCombat &&
          next.attributes === prev.attributes &&
          next.equippedSkills === prev.equippedSkills &&
          next.unspentAttributePoints === prev.unspentAttributePoints &&
          next.spentAttributePoints === prev.spentAttributePoints &&
          next.attackPower === prev.attackPower &&
          next.attackAttr === prev.attackAttr &&
          next.weaponDamage === prev.weaponDamage &&
          next.attackMod === prev.attackMod &&
          next.defense === prev.defense &&
          next.manaRegen === prev.manaRegen &&
          goldNext === state.gold;
        if (sameStats) return state;
        state.characterStats = next;
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
        if (!state.combatTargetId && (enemies || []).length) {
          state.combatTargetId = enemies[0].id;
        }
        return state;
      });
    },

    setCombatTarget: (targetId) => {
      update((state) => {
        state.combatTargetId = targetId || null;
        return state;
      });
    },

    beginCombat: (enemies, players, message) => {
      update((state) => {
        const nextEnemies = normalizeCombatantList(enemies);
        const nextPlayers = normalizeCombatantList(players);
        state.inCombat = true;
        state.combatPhase = "active";
        state.combatOutcome = null;
        state.combatEndMessage = "";
        state.combatEnemies = nextEnemies;
        state.combatPlayers = nextPlayers;
        state.combatTargetId = nextEnemies[0]?.id || null;
        state.combatTurn = null;
        state.combatFx = null;
        state.combatLog = message ? [{ id: nextCombatLogId(), text: String(message).trim() }] : [];
        return state;
      });
    },

    setCombatTurn: (turn) => {
      update((state) => {
        state.inCombat = true;
        if (state.combatPhase === "idle") state.combatPhase = "active";
        state.combatTurn = turn
          ? {
              actorId: turn.actorId || "",
              actorName: turn.actorName || "",
              round: turn.round || 0,
              deadlineMs: turn.deadlineMs || 0,
            }
          : null;
        if (turn?.message) {
          state.combatLog = appendCombatLog(state.combatLog, turn.message);
        }
        return state;
      });
    },

    applyCombatAction: (msg) => {
      update((state) => {
        state.inCombat = true;
        if (state.combatPhase === "idle") state.combatPhase = "active";

        const snapshots = normalizeCombatantList(msg?.combatants);
        if (snapshots.length) {
          const { enemies, players } = mergeCombatantSnapshots(
            state.combatEnemies,
            state.combatPlayers,
            snapshots
          );
          state.combatEnemies = enemies;
          state.combatPlayers = players;
        } else if (msg?.targetId) {
          state.combatEnemies = patchCombatantHp(state.combatEnemies, msg);
          state.combatPlayers = patchCombatantHp(state.combatPlayers, msg);
        }

        // Drop dead enemies from target if needed
        const living = state.combatEnemies.filter((e) => (e.hp ?? 0) > 0);
        if (state.combatTargetId && !living.some((e) => e.id === state.combatTargetId)) {
          state.combatTargetId = living[0]?.id || null;
        }

        // Sync local character HP from player snapshot when present
        const selfId = state.character?.id;
        if (selfId) {
          const self = state.combatPlayers.find((p) => p.id === selfId);
          if (self && typeof self.hp === "number") {
            state.characterStats = {
              ...state.characterStats,
              currentHitPoints: self.hp,
              maxHitPoints: self.maxHp || state.characterStats.maxHitPoints,
            };
          }
        }

        // Always pulse FX when structured action arrives (fxId preferred; result fallback).
        if (msg?.fxId || msg?.result || msg?.damage || msg?.heal) {
          state.combatFx = {
            fxId: msg.fxId || "",
            at: Date.now(),
            targetId: msg.targetId || "",
            actorId: msg.actorId || "",
            damage: msg.damage || 0,
            heal: msg.heal || 0,
            result: msg.result || "",
            action: msg.action || "",
          };
        }

        if (msg?.message) {
          state.combatLog = appendCombatLog(state.combatLog, msg.message);
        }
        return state;
      });
    },

    endCombat: (outcome, message) => {
      update((state) => {
        state.inCombat = false;
        state.combatPhase = "ending";
        state.combatOutcome = outcome || "victory";
        state.combatEndMessage = message || "";
        state.combatTurn = null;
        if (message) {
          state.combatLog = appendCombatLog(state.combatLog, message);
        }
        state.characterStats = { ...state.characterStats, inCombat: false };
        return state;
      });

      // Brief outcome panel, then return to room UI
      setTimeout(() => {
        update((state) => {
          if (state.combatPhase !== "ending") return state;
          state.combatPhase = "idle";
          state.combatOutcome = null;
          state.combatEndMessage = "";
          state.combatEnemies = [];
          state.combatPlayers = [];
          state.combatTargetId = null;
          state.combatTurn = null;
          state.combatFx = null;
          state.combatLog = [];
          return state;
        });
      }, 2800);
    },

    clearCombat: () => {
      update((state) => {
        state.inCombat = false;
        state.combatPhase = "idle";
        state.combatOutcome = null;
        state.combatEndMessage = "";
        state.combatEnemies = [];
        state.combatPlayers = [];
        state.combatTargetId = null;
        state.combatTurn = null;
        state.combatFx = null;
        state.combatLog = [];
        state.characterStats = { ...state.characterStats, inCombat: false };
        return state;
      });
    },

    setGameContext: ({ inCombat, hasItems, hasMerchant } = {}) => {
      update((state) => {
        if (inCombat !== undefined) {
          state.inCombat = inCombat;
          // Battle stage lifecycle is owned by beginCombat/endCombat/clearCombat.
          // Only promote idle→active here; never tear down on a stray false flag.
          if (inCombat && state.combatPhase === "idle") state.combatPhase = "active";
        }
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

      // Accepted banner stays longer; corner toasts dismiss sooner
      const ttl = (notification?.type === 'accepted' || notification?.type === 'completed') ? 10000 : 5000;
      setTimeout(() => {
        update((state) => {
          state.questNotifications = state.questNotifications.filter(n => n.id !== notification.id && n !== notification);
          return state;
        });
      }, ttl);
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
        const incoming = atlas && Array.isArray(atlas.places) ? atlas : emptyAtlas();
        const sameCharacter =
          !state.atlas?.characterId ||
          !incoming.characterId ||
          state.atlas.characterId === incoming.characterId;
        state.atlas = sameCharacter ? mergeAtlas(state.atlas, incoming) : incoming;
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

    openMapOverview: () => {
      update((state) => {
        state.mapOverviewOpen = true;
        return state;
      });
    },
    closeMapOverview: () => {
      update((state) => {
        state.mapOverviewOpen = false;
        return state;
      });
    },
    setMapOverviewOpen: (open) => {
      update((state) => {
        state.mapOverviewOpen = !!open;
        return state;
      });
    },

    openInventoryOverlay: () => {
      update((state) => {
        state.inventoryOverlayOpen = true;
        return state;
      });
    },
    closeInventoryOverlay: () => {
      update((state) => {
        state.inventoryOverlayOpen = false;
        return state;
      });
    },
    setInventoryOverlayOpen: (open) => {
      update((state) => {
        state.inventoryOverlayOpen = !!open;
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
