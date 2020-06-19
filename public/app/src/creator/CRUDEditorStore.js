import { writable } from "svelte/store";

function createStore() {
  const { subscribe, set, update } = writable({
    elements: [],
    selectedElement: null,
    filters: [],
    filterActive: false,
  });
  return {
    subscribe,
    set,
    addFilter: (key, val) => {
      update((state) => {
        let newFilters = state.filters;
        for (var i = 0; i < newFilters.length; i++) {
          if (newFilters[i].key === key) {
            newFilters.splice(i, 1);
          }
        }
        newFilters.push({
          key: key,
          val: val,
        });

        console.log("Added filter " + key + " " + val);

        state.filters = newFilters;
        return state;
      });
    },
    removeFilter: (key) => {
      update((state) => {
        let newFilters = state.filters;
        for (var i = 0; i < newFilters.length; i++) {
          if (newFilters[i].key === key) {
            newFilters.splice(i, 1);
          }
        }
        console.log("Removed filter " + key);
        state.filters = newFilters;
        return state;
      });
    },
    toggleFilter: () => {
      update((state) => {
        state.filterActive = !state.filterActive;
        return state;
      });
    },
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
