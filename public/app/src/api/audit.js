import axios from "axios";
import { backend } from "./base.js";

function auth(token) {
  return { headers: { Authorization: `Bearer ${token}` } };
}

export function listAudit(token, params = {}) {
  return axios.get(`${backend}/audit`, { ...auth(token), params }).then((res) => res.data);
}

export function undoAudit(token, id) {
  return axios
    .post(`${backend}/audit/${encodeURIComponent(id)}/undo`, {}, auth(token))
    .then((res) => res.data);
}
