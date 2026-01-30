import axios from "axios";
import { backend } from "./base.js";

function getServerInfo(cb, errorCb) {
  axios
    .get(`${backend}/server-info`, {
      mode: "no-cors",
      credentials: "same-origin",
    })
    .then((result) => cb(result.data))
    .catch((err) => {
      if (errorCb) errorCb(err);
    });
}

export { getServerInfo };
