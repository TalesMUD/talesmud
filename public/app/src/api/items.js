import axios from "axios";
import { backend } from "./base.js";

function getItem(token, id, cb, errorCb) {
  axios
    .get(`${backend}/items/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function deleteItem(token, id, cb, errorCb) {
  axios
    .delete(`${backend}/items/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getItems(token, filters, cb, errorCb) {
  const path = `${backend}/items`;
  let filtered = path;
  if (filters) {
    filtered += "?";
    let i = 0;
    filters.forEach((f) => {
      if (i > 0) {
        filtered += "&";
      }
      filtered += f.key + "=" + f.val;
    });
  }

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
function updateItem(token, id, item, cb, errorCb) {
  axios
    .put(`${backend}/items/${id}`, item, {
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
function createItem(token, item, cb, errorCb) {
  axios
    .post(`${backend}/items`, item, {
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

function createItemFromTemplate(token, templateId, cb, errorCb) {
  axios
    .put(`${backend}/item-create/${templateId}`, {
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

export {
  getItem,
  deleteItem,
  getItems,
  updateItem,
  createItem,
  createItemFromTemplate,
};
