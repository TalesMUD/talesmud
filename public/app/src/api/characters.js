import axios from "axios";
import {backend} from "./base.js";

function getCharacter(token, id, cb, errorCb) {
    axios
      .get(`${backend}/characters/${id}`, {
        mode: "no-cors",
        credentials: "same-origin",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })
      .then((result) => cb(result.data))
      .catch((err) => errorCb(err));
  };
function getCharacters(token, cb, errorCb) {
    console.log ("using backend " + backend)
    axios
      .get(`${backend}/characters`, {
        mode: "no-cors",
        credentials: "same-origin",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })
      .then((result) => cb(result.data))
      .catch((err) => errorCb(err));
  };
function updateCharacter(token, id, character, cb, errorCb) {
    axios
      .put(`${backend}/characters/${id}`, character, {
        mode: "no-cors",
        credentials: "same-origin",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })
      .then((r) => cb(r.data))
      .catch((err) => errorCb(err));
  };

export {getCharacter, getCharacters, updateCharacter};
