import axios from "axios";
import { backend } from "./base.js";

export function runOp(token, action, body) {
  return axios
    .post(`${backend}/ops/${encodeURIComponent(action)}`, body, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .then((res) => res.data);
}
