import axios from "axios";
import {backend} from "./base.js";

export default {
  getUser(token, cb, errorCb) {
    axios
      .get(`${backend}/user`, {
        mode: "no-cors",
        credentials: "same-origin",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })
      .then((result) => cb(result.data))
      .catch((err) => errorCb(err));
  },

  updateUser(token, user, cb) {
    axios
      .put(`${backend}/user`, user, {
        mode: "no-cors",
        credentials: "same-origin",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })
      .then((r) => cb(r.data));
  },
};
