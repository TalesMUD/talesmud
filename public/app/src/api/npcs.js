import axios from "axios";
import { backend } from "./base.js";

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

function getNPC(token, id, cb, errorCb) {
  axios
    .get(`${backend}/npcs/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getNPCs(token, filters, cb, errorCb) {
  const path = `${backend}/npcs`;
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

function createNPC(token, npc, cb, errorCb) {
  axios
    .post(`${backend}/npcs`, npc, {
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

function updateNPC(token, id, npc, cb, errorCb) {
  axios
    .put(`${backend}/npcs/${id}`, npc, {
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

function deleteNPC(token, id, cb, errorCb) {
  axios
    .delete(`${backend}/npcs/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

export { getNPC, getNPCs, createNPC, updateNPC, deleteNPC };
