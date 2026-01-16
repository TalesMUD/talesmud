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

function getDialog(token, id, cb, errorCb) {
  axios
    .get(`${backend}/dialogs/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getDialogs(token, filters, cb, errorCb) {
  const path = `${backend}/dialogs`;
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

function createDialog(token, dialog, cb, errorCb) {
  axios
    .post(`${backend}/dialogs`, dialog, {
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

function updateDialog(token, id, dialog, cb, errorCb) {
  axios
    .put(`${backend}/dialogs/${id}`, dialog, {
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

function deleteDialog(token, id, cb, errorCb) {
  axios
    .delete(`${backend}/dialogs/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

export { getDialog, getDialogs, createDialog, updateDialog, deleteDialog };
