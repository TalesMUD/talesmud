import { writable } from "svelte/store";

function createStore() {
  const { subscribe, set, update } = writable({
    elements: [],
    selectedElement: null,
    filters: [],
    filterActive: false,
    // Data table state
    detailOpen: false,
    tableFilterValues: {},
    sortKey: "",
    sortDir: "asc",
  });
  return {
    subscribe,
    update,
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
    // Data table methods
    openDetail: () => {
      update((state) => {
        state.detailOpen = true;
        return state;
      });
    },
    closeDetail: () => {
      update((state) => {
        state.detailOpen = false;
        return state;
      });
    },
    setTableFilter: (key, val) => {
      update((state) => {
        state.tableFilterValues = { ...state.tableFilterValues, [key]: val };
        return state;
      });
    },
    resetTableFilters: () => {
      update((state) => {
        state.tableFilterValues = {};
        return state;
      });
    },
    setSort: (key, dir) => {
      update((state) => {
        state.sortKey = key;
        state.sortDir = dir;
        return state;
      });
    },
  };
}
export { createStore };
