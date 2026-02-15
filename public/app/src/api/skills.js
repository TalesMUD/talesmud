import axios from "axios";
import { backend } from "./base.js";

function getSkill(token, id, cb, errorCb) {
  axios
    .get(`${backend}/skills/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getSkills(token, filters, cb, errorCb) {
  axios
    .get(`${backend}/skills`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function createSkill(token, skill, cb, errorCb) {
  axios
    .post(`${backend}/skills`, skill, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    })
    .then((r) => cb(r.data))
    .catch((err) => errorCb(err));
}

function updateSkill(token, id, skill, cb, errorCb) {
  axios
    .put(`${backend}/skills/${id}`, skill, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    })
    .then((r) => cb(r.data))
    .catch((err) => errorCb(err));
}

function deleteSkill(token, id, cb, errorCb) {
  axios
    .delete(`${backend}/skills/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

export { getSkill, getSkills, createSkill, updateSkill, deleteSkill };
