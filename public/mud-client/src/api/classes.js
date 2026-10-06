import axios from "axios";
import { backend } from "./base.js";

export function fetchClasses() {
  return axios.get(`${backend}/classes`).then((result) => result.data);
}
