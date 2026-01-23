import axios from "axios";
import { backend } from "./base.js";

function createFilterRequest(path, filters) {
  let filtered = path;
  if (filters && filters.length > 0) {
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

function getNPCSpawner(token, id, cb, errorCb) {
  axios
    .get(`${backend}/spawners/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getNPCSpawners(token, filters, cb, errorCb) {
  const path = `${backend}/spawners`;
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

function getNPCSpawnersByRoom(token, roomId, cb, errorCb) {
  getNPCSpawners(token, [{ key: "roomId", val: roomId }], cb, errorCb);
}

function createNPCSpawner(token, spawner, cb, errorCb) {
  axios
    .post(`${backend}/spawners`, spawner, {
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

function updateNPCSpawner(token, id, spawner, cb, errorCb) {
  axios
    .put(`${backend}/spawners/${id}`, spawner, {
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

function deleteNPCSpawner(token, id, cb, errorCb) {
  axios
    .delete(`${backend}/spawners/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

// Promise-based wrappers for async/await usage
function getNPCSpawnerAsync(token, id) {
  return new Promise((resolve, reject) => {
    getNPCSpawner(token, id, resolve, reject);
  });
}

function getNPCSpawnersAsync(token, filters = []) {
  return new Promise((resolve, reject) => {
    getNPCSpawners(token, filters, resolve, reject);
  });
}

function getNPCSpawnersByRoomAsync(token, roomId) {
  return new Promise((resolve, reject) => {
    getNPCSpawnersByRoom(token, roomId, resolve, reject);
  });
}

function createNPCSpawnerAsync(token, spawner) {
  return new Promise((resolve, reject) => {
    createNPCSpawner(token, spawner, resolve, reject);
  });
}

function updateNPCSpawnerAsync(token, id, spawner) {
  return new Promise((resolve, reject) => {
    updateNPCSpawner(token, id, spawner, resolve, reject);
  });
}

function deleteNPCSpawnerAsync(token, id) {
  return new Promise((resolve, reject) => {
    deleteNPCSpawner(token, id, resolve, reject);
  });
}

export {
  getNPCSpawner,
  getNPCSpawners,
  getNPCSpawnersByRoom,
  createNPCSpawner,
  updateNPCSpawner,
  deleteNPCSpawner,
  // Async versions
  getNPCSpawnerAsync,
  getNPCSpawnersAsync,
  getNPCSpawnersByRoomAsync,
  createNPCSpawnerAsync,
  updateNPCSpawnerAsync,
  deleteNPCSpawnerAsync,
};
