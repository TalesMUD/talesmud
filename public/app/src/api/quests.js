import axios from "axios";
import { backend } from "./base.js";

function getQuest(token, id, cb, errorCb) {
  axios
    .get(`${backend}/quests/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getQuests(token, filters, cb, errorCb) {
  axios
    .get(`${backend}/quests`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function createQuest(token, quest, cb, errorCb) {
  axios
    .post(`${backend}/quests`, quest, {
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

function updateQuest(token, id, quest, cb, errorCb) {
  axios
    .put(`${backend}/quests/${id}`, quest, {
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

function deleteQuest(token, id, cb, errorCb) {
  axios
    .delete(`${backend}/quests/${id}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getQuestProgress(token, characterId, cb, errorCb) {
  axios
    .get(`${backend}/quest-progress/${characterId}`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function completeQuest(token, characterId, questId, cb, errorCb) {
  axios
    .post(`${backend}/quest-progress/${characterId}/complete/${questId}`, {}, {
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

export { getQuest, getQuests, createQuest, updateQuest, deleteQuest, getQuestProgress, completeQuest };
