import { writable } from "svelte/store";

function createSubMenuStore() {
  const { subscribe, set, update} = writable({
    active: false,
    entries: [{ name: "Item1" }],
  });

  return {
    subscribe,
    show: () =>
      update((menu) => {
        menu.active = true;
        return menu;
      }),
    hide: () =>
      update((menu) => {
        menu.active = false;
        return menu;
      }),
    setItems: (items) => {
      update((menu) => {
        menu.entries = items;

        return menu;
      });
    },
  };
}

function createUserStore() {
  const { subscribe, set, update } = writable({
    name: "marcus",
    loggedIn: false,
  });

  return {
    subscribe,
    set,
    logIn: () =>
      update((user) => {
        user.loggedIn = true;
        return user;
      }),
    logOut: () =>
      update((user) => {
        user.loggedIn = false;
        return user;
      }),
  };
}

const user = createUserStore();
const subMenu = createSubMenuStore();

export { user, subMenu };
