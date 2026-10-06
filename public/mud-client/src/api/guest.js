import axios from "axios";
import { backend } from "./base.js";

/**
 * Creates a guest session by calling the backend API.
 * Returns { token, expiresIn } on success.
 */
function createGuestSession(cb, errorCb, body) {
  const payload = body && typeof body === "object" ? body : {};
  axios
    .post(`${backend}/guest`, payload, {
      mode: "no-cors",
      credentials: "same-origin",
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

export { createGuestSession };
