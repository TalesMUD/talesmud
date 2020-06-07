import { writable } from "svelte/store";

function createStore() {
  const { subscribe, set, update } = writable({
    rooms: [],
    filteredRooms: [],
    selectedRoom: null,

    filter: "",
  });

  return {
    subscribe,
    setSelectedRoom: (room, cb) => {
      update((state) => {
        state.selectedRoom = room;
        return state;
      });
      if (cb) cb();
    },
    deleteExit: (exit) =>
      update((state) => {
        state.selectedRoom.exits = state.selectedRoom.exits.filter(
          (x) => x.name != exit.name
        );
        return state;
      }),
    createExit: () => {
      update((state) => {
        if (state.selectedRoom.exits == null) {
          state.selectedRoom.exits = [];
        }

        state.selectedRoom.exits.push({
          name: "New Exit",
          description: "todo",
          target: "select target",
        });
        return state;
      });
    },
    setFilter: (filter) => {
      update((state) => {
        state.filter = filter;
        state.filteredRooms = rooms.filter((x) => x.name.includes(filter));
        return state;
      });
    },
    setRooms: (rooms) => {
      update((state) => {
        console.log("SET ROOMS: COUNT " + rooms.length);
        state.rooms = rooms;
        state.filteredRooms = rooms;
        return state;
      });
    },
  };
}

const store = createStore();

export { store };
