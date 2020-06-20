import axios from "axios";
import { backend } from "./base.js";

function getItemSlots(cb, errorCb) {
  axios
    .get(`${backend}/item-slots`)
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getItemQualities(cb, errorCb) {
  axios
    .get(`${backend}/item-qualities`)
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getItemTypes(cb, errorCb) {
  axios
    .get(`${backend}/item-types`)
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getItemSubTypes(cb, errorCb) {
  axios
    .get(`${backend}/item-subtypes`)
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getItemTemplate(token, id, cb, errorCb) {
  axios
    .get(`${backend}/item-templates/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function deleteItemTemplate(token, id, cb, errorCb) {
  axios
    .delete(`${backend}/item-templates/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getItemTemplates(token, filters, cb, errorCb) {
  const path = `${backend}/item-templates`;
  let filtered = createFilterRequest(path, filters);
  
  axios
    .get(filtered, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}
function createFilterRequest(path, filters) {
  let filtered = path;
  if (filters) {
    filtered += "?";
    let i = 0;
    filters.forEach((f) => {
      if (i > 0) {
        filtered += "&";
      }
      filtered += f.key + "=" + f.val;
      i++;
    });
  }
  return filtered;
}

function updateItemTemplate(token, id, ItemTemplate, cb, errorCb) {
  axios
    .put(`${backend}/item-templates/${id}`, ItemTemplate, {
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
function createItemTemplate(token, ItemTemplate, cb, errorCb) {
  axios
    .post(`${backend}/item-templates`, ItemTemplate, {
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

export {
  getItemTemplate,
  deleteItemTemplate,
  getItemTemplates,
  updateItemTemplate,
  createItemTemplate,
  getItemQualities,
  getItemTypes,
  getItemSubTypes,
  getItemSlots,
};
