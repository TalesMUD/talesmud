import axios from "axios";
import { backend } from "./base.js";

function authHeaders(token) {
  return {
    Authorization: `Bearer ${token}`,
    "Content-Type": "application/json",
  };
}

function postPreview(token, path, payload, cb, errorCb) {
  axios
    .post(`${backend}/preview/${path}`, payload, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: authHeaders(token),
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function previewDialog(token, dialog, cb, errorCb) {
  postPreview(token, "dialog", dialog, cb, errorCb);
}

function previewQuest(token, quest, cb, errorCb) {
  postPreview(token, "quest", quest, cb, errorCb);
}

function previewRoom(token, room, cb, errorCb) {
  postPreview(token, "room", room, cb, errorCb);
}

function previewMerchant(token, npc, cb, errorCb) {
  postPreview(token, "merchant", npc, cb, errorCb);
}

export { previewDialog, previewQuest, previewRoom, previewMerchant };
