import { writable, derived } from "svelte/store";
import { v4 as uuidv4 } from "uuid";

/**
 * WorldEditor state management store
 * Handles nodes, edges, selection, editing state, and room creation
 */
export function createWorldEditorStore() {
  const initialState = {
    nodes: [],
    edges: [],
    selectedRoomId: null,
    editingRoom: null,
    isCreating: false,
    saving: false,
    error: null,
    // Drag-to-create state
    dragSource: null,      // { nodeId, handleId (direction) }
    pendingRoom: null,     // Draft room being created
    roomsValueHelp: [],    // List of all rooms for dropdowns
  };

  const { subscribe, set, update } = writable(initialState);

  return {
    subscribe,
    set,
    update,

    // Set graph data from backend
    setGraphData: (nodes, edges) => {
      update(state => ({
        ...state,
        nodes,
        edges,
      }));
    },

    // Set rooms value help for dropdowns
    setRoomsValueHelp: (rooms) => {
      update(state => ({
        ...state,
        roomsValueHelp: rooms,
      }));
    },

    // Select a room for editing
    selectRoom: (roomId, roomData) => {
      update(state => ({
        ...state,
        selectedRoomId: roomId,
        editingRoom: roomData ? { ...roomData } : null,
        isCreating: false,
        pendingRoom: null,
        error: null,
      }));
    },

    // Clear selection
    clearSelection: () => {
      update(state => ({
        ...state,
        selectedRoomId: null,
        editingRoom: null,
        isCreating: false,
        pendingRoom: null,
        error: null,
      }));
    },

    // Start creating a new room (from drag or button)
    startCreating: (sourceRoomId, direction, sourceCoords) => {
      const offsets = {
        north: { x: 0, y: -1 },
        south: { x: 0, y: 1 },
        east: { x: 1, y: 0 },
        west: { x: -1, y: 0 },
      };

      const offset = offsets[direction] || { x: 0, y: 0 };
      const newCoords = sourceCoords ? {
        x: (sourceCoords.x || 0) + offset.x,
        y: (sourceCoords.y || 0) + offset.y,
        z: sourceCoords.z || 0,
      } : { x: 0, y: 0, z: 0 };

      const oppositeDirection = {
        north: "south",
        south: "north",
        east: "west",
        west: "east",
      };

      const newRoom = {
        id: uuidv4(),
        name: "New Room",
        description: "",
        detail: "",
        area: "",
        areaType: "",
        roomType: "",
        coords: newCoords,
        exits: [],
        actions: [],
        meta: { background: "" },
        isNew: true,
      };

      // If created via drag, add return exit to source
      if (sourceRoomId && direction) {
        newRoom.exits.push({
          name: oppositeDirection[direction],
          description: "",
          target: sourceRoomId,
          exitType: "direction",
          hidden: false,
        });
        newRoom._sourceRoomId = sourceRoomId;
        newRoom._sourceDirection = direction;
      }

      update(state => ({
        ...state,
        selectedRoomId: newRoom.id,
        editingRoom: newRoom,
        isCreating: true,
        pendingRoom: newRoom,
        error: null,
      }));
    },

    // Update a field on the editing room
    updateField: (field, value) => {
      update(state => {
        if (!state.editingRoom) return state;

        // Handle nested fields like coords.x
        if (field.includes(".")) {
          const [parent, child] = field.split(".");
          return {
            ...state,
            editingRoom: {
              ...state.editingRoom,
              [parent]: {
                ...state.editingRoom[parent],
                [child]: value,
              },
            },
          };
        }

        return {
          ...state,
          editingRoom: {
            ...state.editingRoom,
            [field]: value,
          },
        };
      });
    },

    // Update exits array
    updateExits: (exits) => {
      update(state => {
        if (!state.editingRoom) return state;
        return {
          ...state,
          editingRoom: {
            ...state.editingRoom,
            exits,
          },
        };
      });
    },

    // Add an exit
    addExit: (exit) => {
      update(state => {
        if (!state.editingRoom) return state;
        const exits = [...(state.editingRoom.exits || []), exit];
        return {
          ...state,
          editingRoom: {
            ...state.editingRoom,
            exits,
          },
        };
      });
    },

    // Remove an exit
    removeExit: (exitName) => {
      update(state => {
        if (!state.editingRoom) return state;
        const exits = (state.editingRoom.exits || []).filter(e => e.name !== exitName);
        return {
          ...state,
          editingRoom: {
            ...state.editingRoom,
            exits,
          },
        };
      });
    },

    // Update a specific exit
    updateExit: (exitName, updates) => {
      update(state => {
        if (!state.editingRoom) return state;
        const exits = (state.editingRoom.exits || []).map(e =>
          e.name === exitName ? { ...e, ...updates } : e
        );
        return {
          ...state,
          editingRoom: {
            ...state.editingRoom,
            exits,
          },
        };
      });
    },

    // Set saving state
    setSaving: (saving) => {
      update(state => ({ ...state, saving }));
    },

    // Set error
    setError: (error) => {
      update(state => ({ ...state, error }));
    },

    // Cancel editing
    cancel: () => {
      update(state => ({
        ...state,
        editingRoom: null,
        selectedRoomId: null,
        isCreating: false,
        pendingRoom: null,
        error: null,
      }));
    },

    // Track drag start for room creation
    setDragSource: (nodeId, handleId) => {
      update(state => ({
        ...state,
        dragSource: { nodeId, handleId },
      }));
    },

    // Clear drag source
    clearDragSource: () => {
      update(state => ({
        ...state,
        dragSource: null,
      }));
    },

    // Reset entire store
    reset: () => {
      set(initialState);
    },
  };
}

// Helper to get opposite direction
export function getOppositeDirection(direction) {
  const opposites = {
    north: "south",
    south: "north",
    east: "west",
    west: "east",
    up: "down",
    down: "up",
    northeast: "southwest",
    southwest: "northeast",
    northwest: "southeast",
    southeast: "northwest",
  };
  return opposites[direction.toLowerCase()] || null;
}

// Cardinal directions for quick editing
export const CARDINAL_DIRECTIONS = ["north", "east", "south", "west"];

// Check if an exit name is a cardinal direction
export function isCardinalDirection(name) {
  return CARDINAL_DIRECTIONS.includes(name.toLowerCase());
}
