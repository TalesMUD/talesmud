import axios from "axios";
import { backend } from "./base.js";

function getWorldMap(token, cb, errorCb) {
  axios
    .get(`${backend}/world/map`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getWorldGraph(token, cb, errorCb) {
  axios
    .get(`${backend}/world/graph`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

// Promise-based version of getWorldGraph
function getWorldGraphAsync(token) {
  return axios
    .get(`${backend}/world/graph`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => result.data);
}

// Get minimal room data for the world map editor
function getMinimalRooms(token, cb, errorCb) {
  axios
    .get(`${backend}/world/rooms-minimal`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

// Promise-based version
function getMinimalRoomsAsync(token) {
  return axios
    .get(`${backend}/world/rooms-minimal`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => result.data);
}

export { getWorldMap, getWorldGraph, getWorldGraphAsync, getMinimalRooms, getMinimalRoomsAsync };
