import axios from "axios";
import { backend } from "./base.js";

function getAllUsers(token, cb, errorCb) {
  axios
    .get(`${backend}/admin/users`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function updateUserRole(token, userId, role, cb, errorCb) {
  axios
    .put(
      `${backend}/admin/users/${userId}/role`,
      { role },
      {
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
      }
    )
    .then((r) => cb(r.data))
    .catch((err) => errorCb(err));
}

function banUser(token, userId, cb, errorCb) {
  axios
    .post(
      `${backend}/admin/users/${userId}/ban`,
      {},
      {
        headers: { Authorization: `Bearer ${token}` },
      }
    )
    .then((r) => cb(r.data))
    .catch((err) => errorCb(err));
}

function unbanUser(token, userId, cb, errorCb) {
  axios
    .post(
      `${backend}/admin/users/${userId}/unban`,
      {},
      {
        headers: { Authorization: `Bearer ${token}` },
      }
    )
    .then((r) => cb(r.data))
    .catch((err) => errorCb(err));
}

function deleteUser(token, userId, cb, errorCb) {
  axios
    .delete(`${backend}/admin/users/${userId}`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .then((r) => cb(r.data))
    .catch((err) => errorCb(err));
}

export { getAllUsers, updateUserRole, banUser, unbanUser, deleteUser };
