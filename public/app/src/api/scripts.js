import axios from "axios";
import { backend } from "./base.js";

function getScript(token, id, cb, errorCb) {
  axios
    .get(`${backend}/scripts/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function deleteScript(token, id, cb, errorCb) {
  axios
    .delete(`${backend}/scripts/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getScripts(token, filters, cb, errorCb) {
  const path = `${backend}/scripts`;
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
function updateScript(token, id, script, cb, errorCb) {
  axios
    .put(`${backend}/scripts/${id}`, script, {
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
function createScript(token, script, cb, errorCb) {
  axios
    .post(`${backend}/scripts`, script, {
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

export { getScript, deleteScript, getScripts, updateScript, createScript };
