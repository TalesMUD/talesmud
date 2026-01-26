import { writable, derived } from "svelte/store";

// Cardinal direction names for compass filtering
const CARDINAL_DIRECTIONS = ["north", "south", "east", "west"];
const VERTICAL_DIRECTIONS = ["up", "down"];

function createStore() {
  const { subscribe, set, update } = writable({
    // Room data
    exits: [],
    actions: [],
    npcs: [],
    background: "oldtown-griphon",
    roomName: "",
    roomDescription: "",

    // Dialog state
    dialogActive: false,
    dialogNpcName: "",
    dialogNpcText: "",
    dialogOptions: [],
    dialogConversationID: "",

    // Game context flags
    inCombat: false,
    hasItems: false,
    hasMerchant: false,
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
    setRoomInfo: (name, description) => {
      update((state) => {
        state.roomName = name || "";
        state.roomDescription = description || "";
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

    // Game context methods
    setGameContext: ({ inCombat, hasItems, hasMerchant } = {}) => {
      update((state) => {
        if (inCombat !== undefined) state.inCombat = inCombat;
        if (hasItems !== undefined) state.hasItems = hasItems;
        if (hasMerchant !== undefined) state.hasMerchant = hasMerchant;
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
