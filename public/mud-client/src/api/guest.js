import axios from "axios";
import { backend } from "./base.js";

/**
 * Creates a guest session by calling the backend API.
 * Returns { token, expiresIn } on success.
 */
function createGuestSession(cb, errorCb) {
  axios
    .post(`${backend}/guest`, {}, {
      mode: "no-cors",
      credentials: "same-origin",
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

export { createGuestSession };
