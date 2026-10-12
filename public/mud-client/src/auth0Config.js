// Shared build-time Auth0 settings for play and device confirmation.
export { auth0Config, isAuth0Configured } from "./authConfig.js";

export function auth0Callback() {
  return window.location.origin + "/play";
}
