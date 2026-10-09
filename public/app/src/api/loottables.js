import axios from "axios";
import { backend } from "./base.js";

function getLootTables(token, filters, cb, errorCb) {
  let path = `${backend}/loottables`;
  if (filters && filters.length) {
    const query = filters
      .map((f) => `${encodeURIComponent(f.key)}=${encodeURIComponent(f.val)}`)
      .join("&");
    path += `?${query}`;
  }
  axios
    .get(path, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getLootTable(token, id, cb, errorCb) {
  axios
    .get(`${backend}/loottables/${id}`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

export { getLootTables, getLootTable };
