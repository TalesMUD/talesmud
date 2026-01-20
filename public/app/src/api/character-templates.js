import axios from "axios";
import { backend } from "./base.js";

function getCharacterTemplates(token, filters, cb, errorCb) {
  axios
    .get(`${backend}/character-templates`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getCharacterTemplate(token, id, cb, errorCb) {
  axios
    .get(`${backend}/character-templates/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function createCharacterTemplate(token, template, cb, errorCb) {
  axios
    .post(`${backend}/character-templates`, template, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((r) => cb(r.data))
    .catch((err) => errorCb(err));
}

function updateCharacterTemplate(token, id, template, cb, errorCb) {
  axios
    .put(`${backend}/character-templates/${id}`, template, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((r) => cb(r.data))
    .catch((err) => errorCb(err));
}

function deleteCharacterTemplate(token, id, cb, errorCb) {
  axios
    .delete(`${backend}/character-templates/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((r) => cb(r.data))
    .catch((err) => errorCb(err));
}

function seedCharacterTemplates(token, cb, errorCb) {
  axios
    .post(`${backend}/character-templates/seed`, {}, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((r) => cb(r.data))
    .catch((err) => errorCb(err));
}

function getCharacterTemplatePresets(token, cb, errorCb) {
  axios
    .get(`${backend}/character-templates/presets`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

export {
  getCharacterTemplates,
  getCharacterTemplate,
  createCharacterTemplate,
  updateCharacterTemplate,
  deleteCharacterTemplate,
  seedCharacterTemplates,
  getCharacterTemplatePresets,
};
