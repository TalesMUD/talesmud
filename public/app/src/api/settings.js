import axios from "axios";
import { backend } from "./base.js";

function getSettings(token, cb, errorCb) {
  axios
    .get(`${backend}/settings`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function updateSettings(token, settings, cb, errorCb) {
  axios
    .put(`${backend}/settings`, settings, {
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

export { getSettings, updateSettings };
