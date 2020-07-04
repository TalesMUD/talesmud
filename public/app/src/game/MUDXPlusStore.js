import { writable } from "svelte/store";

function createStore() {
  const { subscribe, set, update } = writable({
    exits: [],
    actions: [],
  });
  return {
    subscribe,
    update,
    set,
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
