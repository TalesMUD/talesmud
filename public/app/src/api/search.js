import axios from "axios";
import { backend } from "./base.js";

function authHeaders(token) {
  return { Authorization: `Bearer ${token}` };
}

function searchContent(token, q, { types = "", limit = 25 } = {}) {
  const params = { q, limit };
  if (types) params.types = types;
  return axios
    .get(`${backend}/search`, { headers: authHeaders(token), params })
    .then((result) => result.data);
}

function getRefs(token, type, id) {
  const path = `${backend}/refs/${encodeURIComponent(type)}/${encodeURIComponent(id)}`;
  return axios.get(path, { headers: authHeaders(token) }).then((result) => result.data);
}

export { searchContent, getRefs };
