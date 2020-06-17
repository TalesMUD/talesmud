import { writable } from "svelte/store";

function createStore() {
  const { subscribe, set, update } = writable({
    elements: [],
    selectedElement: null,
    filter: "",
  });
  return {
    subscribe,
    set,
    setSelectedElement: (element, cb) => {
      update((state) => {
        state.selectedElement = element;
        return state;
      });
      if (cb) cb();
    },
    setElements: (newelements) => {
      update((state) => {
        state.elements = newelements;
        return state;
      });
    },
  };
}
export { createStore };
