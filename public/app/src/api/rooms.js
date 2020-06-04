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

function getRooms(token, cb, errorCb) {
  axios
    .get(`${backend}/rooms`, {
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

export { getRoom, deleteRoom, getRooms, updateRoom, createRoom };
