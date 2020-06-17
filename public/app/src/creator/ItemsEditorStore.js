import { writable } from "svelte/store";

function createStore() {
  const { subscribe, set, update } = writable({
    items: [],
    itemTemplates: [],
    selectedItemTemplate: null,
    selectedItem: null,

    filter: "",
  });

  return {
    subscribe,
    set,
    setSelectedItemTemplate: (itemTemplate, cb) => {
      update((state) => {
        state.selectedItemTemplate = itemTemplate;
        return state;
      });
      if (cb) cb();
    },
    setSelectedItem: (item, cb) => {
      update((state) => {
        state.selectedItem = item;
        return state;
      });
      if (cb) cb();
    },
    setItemTemplates: (itemTemplates) => {
      update((state) => {
        state.itemTemplates = itemTemplates;      
        return state;
      });
    },
    setItems: (items) => {
      update((state) => {
        state.items = items;      
        return state;
      });
    }
  };
}

const store = createStore();

export { store };
