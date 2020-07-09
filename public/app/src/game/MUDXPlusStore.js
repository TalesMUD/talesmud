import { writable } from "svelte/store";

function createStore() {
  const { subscribe, set, update } = writable({
    exits: [],
    actions: [],
    background: "oldtown-griphon",
  });
  return {
    subscribe,
    update,
    set,
    setBackground:(background) => {
      update((state) => {
        state.background = background;
        return state;
      });
    },
    setExits: (exits) => {
      update((state) => {
        state.exits = exits;
        return state;
      });
    },
    setActions: (actions) => {
      update((state) => {
        state.actions = actions;
        return state;
      });
    },
  };
}
export { createStore };
