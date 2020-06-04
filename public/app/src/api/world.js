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
export { getWorldMap };
