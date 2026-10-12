import axios from "axios";
import { backend } from "./base.js";

function authHeaders(token) {
  return {
    Authorization: `Bearer ${token}`,
    "Content-Type": "application/json",
  };
}

function getHealth(token) {
  return axios
    .get(`${backend}/health`, { headers: authHeaders(token) })
    .then((result) => result.data);
}

function getDrift(token) {
  return axios
    .get(`${backend}/health/drift`, { headers: authHeaders(token) })
    .then((result) => result.data);
}

function putMute(token, ruleId, muted) {
  return axios
    .put(`${backend}/health/mute`, { ruleId, muted }, { headers: authHeaders(token) })
    .then((result) => result.data);
}

function exportDrift(token, entityType, id) {
  return axios
    .get(`${backend}/health/drift/export`, {
      headers: { Authorization: `Bearer ${token}` },
      params: { type: entityType, id },
      responseType: "blob",
    })
    .then((result) => result.data);
}

export { getHealth, getDrift, putMute, exportDrift };
