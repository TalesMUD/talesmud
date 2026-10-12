import axios from "axios";
import { backend } from "./base.js";

function getLootTables(token, filters, cb, errorCb) {
  let path = `${backend}/loottables`;
  if (filters && filters.length) {
    const query = filters
      .map((f) => `${encodeURIComponent(f.key)}=${encodeURIComponent(f.val)}`)
      .join("&");
    path += `?${query}`;
  }
  axios
    .get(path, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getLootTable(token, id, cb, errorCb) {
  axios
    .get(`${backend}/loottables/${id}`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function createLootTable(token, lootTable, cb, errorCb) {
  axios
    .post(`${backend}/loottables`, lootTable, {
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function updateLootTable(token, id, lootTable, cb, errorCb) {
  axios
    .put(`${backend}/loottables/${id}`, lootTable, {
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function deleteLootTable(token, id, cb, errorCb) {
  axios
    .delete(`${backend}/loottables/${id}`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function rollLootTable(token, id, params = {}) {
  return axios
    .post(`${backend}/loot-tables/${encodeURIComponent(id)}/roll`, null, {
      params,
      headers: { Authorization: `Bearer ${token}` },
    })
    .then((result) => result.data);
}

export { getLootTables, getLootTable, createLootTable, updateLootTable, deleteLootTable, rollLootTable };
