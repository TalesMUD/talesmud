import axios from "axios";
import { backend } from "./base.js";

function getGuestStatsSummary(token, cb, errorCb) {
  axios
    .get(`${backend}/admin/stats/guests`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getGuestDailyStats(token, days, cb, errorCb) {
  axios
    .get(`${backend}/admin/stats/guests/daily`, {
      headers: { Authorization: `Bearer ${token}` },
      params: { days },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

export { getGuestStatsSummary, getGuestDailyStats };
