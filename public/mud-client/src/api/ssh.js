import axios from "axios";
import { backend } from "./base.js";

function headers(token) {
  return {
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
  };
}

function lookupDevice(token, userCode) {
  return axios.post(`${backend}/ssh/device/lookup`, { user_code: userCode }, headers(token));
}

function confirmDevice(token, userCode, csrf) {
  return axios.post(`${backend}/ssh/device/confirm`, { user_code: userCode, csrf }, headers(token));
}

function denyDevice(token, userCode, csrf) {
  return axios.post(`${backend}/ssh/device/deny`, { user_code: userCode, csrf }, headers(token));
}

function listKeys(token) {
  return axios.get(`${backend}/ssh/keys`, headers(token));
}

function revokeKey(token, id) {
  return axios.delete(`${backend}/ssh/keys/${encodeURIComponent(id)}`, headers(token));
}

export { lookupDevice, confirmDevice, denyDevice, listKeys, revokeKey };
