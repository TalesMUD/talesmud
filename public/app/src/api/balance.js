import axios from "axios";
import { backend } from "./base.js";

function getEnemyScaling(token, cb, errorCb) {
  axios
    .get(`${backend}/balance/enemy-scaling`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

export { getEnemyScaling };
