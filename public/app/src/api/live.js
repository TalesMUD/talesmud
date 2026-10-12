import axios from "axios";
import { backend } from "./base.js";

function auth(token) {
  return { headers: { Authorization: `Bearer ${token}` } };
}

export function getServerInfo() {
  return axios.get(`${backend}/server-info`).then((res) => res.data);
}

export function getLiveCharacters(token, params = {}) {
  return axios.get(`${backend}/live/characters`, { ...auth(token), params }).then((res) => res.data);
}

export function getLiveCharacter(token, id) {
  return axios.get(`${backend}/live/characters/${encodeURIComponent(id)}`, auth(token)).then((res) => res.data);
}

export function getLiveNPCs(token, params = {}) {
  return axios.get(`${backend}/live/npcs`, { ...auth(token), params }).then((res) => res.data);
}

export function getLiveInstances(token) {
  return axios.get(`${backend}/live/instances`, auth(token)).then((res) => res.data);
}
