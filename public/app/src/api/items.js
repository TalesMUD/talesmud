import axios from "axios";
import { backend } from "./base.js";

// Helper to build query string from filters array
function createFilterRequest(path, filters) {
  let filtered = path;
  if (filters && filters.length > 0) {
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

// Get single item by ID (template or instance)
function getItem(token, id, cb, errorCb) {
  axios
    .get(`${backend}/items/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

// Delete item by ID (template or instance)
function deleteItem(token, id, cb, errorCb) {
  axios
    .delete(`${backend}/items/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

// Get all items with optional filters
// Use filters: [{key: "isTemplate", val: "true"}] for templates only
function getItems(token, filters, cb, errorCb) {
  const path = `${backend}/items`;
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

// Get only item templates (convenience wrapper)
function getItemTemplates(token, filters, cb, errorCb) {
  const templateFilters = filters ? [...filters] : [];
  templateFilters.push({ key: "isTemplate", val: "true" });
  return getItems(token, templateFilters, cb, errorCb);
}

// Update item by ID (template or instance)
function updateItem(token, id, item, cb, errorCb) {
  axios
    .put(`${backend}/items/${id}`, item, {
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

// Create new item (template or instance based on isTemplate field)
function createItem(token, item, cb, errorCb) {
  axios
    .post(`${backend}/items`, item, {
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

// Create item instance from template
function createItemFromTemplate(token, templateId, cb, errorCb) {
  axios
    .post(`${backend}/items/from-template/${templateId}`, {}, {
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

// Metadata endpoints (public)
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

export {
  getItem,
  deleteItem,
  getItems,
  getItemTemplates,
  updateItem,
  createItem,
  createItemFromTemplate,
  getItemSlots,
  getItemQualities,
  getItemTypes,
  getItemSubTypes,
};
