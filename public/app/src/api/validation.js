import axios from "axios";
import { backend } from "./base.js";

function authHeaders(token) {
  return {
    Authorization: `Bearer ${token}`,
    "Content-Type": "application/json",
  };
}

function validateEntity(token, entityType, entity, cb, errorCb) {
  axios
    .post(`${backend}/validate/${entityType}`, entity, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: authHeaders(token),
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getWorldDiagnostics(token, cb, errorCb) {
  axios
    .get(`${backend}/diagnostics/world`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function validateEntityAsync(token, entityType, entity) {
  return new Promise((resolve, reject) => {
    validateEntity(token, entityType, entity, resolve, reject);
  });
}

function getWorldDiagnosticsAsync(token) {
  return new Promise((resolve, reject) => {
    getWorldDiagnostics(token, resolve, reject);
  });
}

export {
  validateEntity,
  getWorldDiagnostics,
  validateEntityAsync,
  getWorldDiagnosticsAsync,
};
