import axios from "axios";
import { backend } from "./base.js";

function getRoom(token, id, cb, errorCb) {
  axios
    .get(`${backend}/rooms/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getRoomOfTheDay(cb, errorCb) {
  axios
    .get(`${backend}/room-of-the-day`, {
      mode: "no-cors",
      credentials: "same-origin",
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function deleteRoom(token, id, cb, errorCb) {
  axios
    .delete(`${backend}/rooms/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}
function createFilterRequest(path, filters) {
  let filtered = path;
  if (filters) {
    filtered += "?";
    let i = 0;
    filters.forEach((f) => {
      if (i > 0) {
        filtered += "&";
      }
      filtered += f.key + "=" + f.val;
      i++;
    });
  }
  return filtered;
}
function getRooms(token, filters, cb, errorCb) {
  const path = `${backend}/rooms`;
  let filtered = createFilterRequest(path, filters);

  axios
    .get(filtered, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}
function getRoomsValueHelp(token, cb, errorCb) {
  const path = `${backend}/rooms-vh`;

  axios
    .get(path, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}
function updateRoom(token, id, room, cb, errorCb) {
  axios
    .put(`${backend}/rooms/${id}`, room, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        "Content-Type": "application/json",

        Authorization: `Bearer ${token}`,
      },
    })
    .then((r) => cb(r.data))
    .catch((err) => errorCb(err));
}
function createRoom(token, room, cb, errorCb) {
  axios
    .post(`${backend}/rooms`, room, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    })
    .then((r) => cb(r.data))
    .catch((err) => errorCb(err));
}

// Promise-based wrappers for async/await usage
function getRoomAsync(token, id) {
  return new Promise((resolve, reject) => {
    getRoom(token, id, resolve, reject);
  });
}

function getRoomsAsync(token, filters = []) {
  return new Promise((resolve, reject) => {
    getRooms(token, filters, resolve, reject);
  });
}

function getRoomsValueHelpAsync(token) {
  return new Promise((resolve, reject) => {
    getRoomsValueHelp(token, resolve, reject);
  });
}

function createRoomAsync(token, room) {
  return new Promise((resolve, reject) => {
    createRoom(token, room, resolve, reject);
  });
}

function updateRoomAsync(token, id, room) {
  return new Promise((resolve, reject) => {
    updateRoom(token, id, room, resolve, reject);
  });
}

function deleteRoomAsync(token, id) {
  return new Promise((resolve, reject) => {
    deleteRoom(token, id, resolve, reject);
  });
}

export {
  getRoom,
  getRoomOfTheDay,
  deleteRoom,
  getRoomsValueHelp,
  getRooms,
  updateRoom,
  createRoom,
  // Async versions
  getRoomAsync,
  getRoomsAsync,
  getRoomsValueHelpAsync,
  createRoomAsync,
  updateRoomAsync,
  deleteRoomAsync,
};
