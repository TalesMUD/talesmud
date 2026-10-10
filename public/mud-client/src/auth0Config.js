// Auth0 settings shared by the play client and the standalone /activate page.
export const auth0Config = {
  domain: "owndnd.eu.auth0.com",
  client_id: "mxcEqTuAUOzrL798mbVTpqFxpGGVp3gI",
  audience: "http://talesofapirate.com/dnd/api",
};

// The Auth0 callback stays on /play, which is already an allowed callback URL.
export function auth0Callback() {
  return window.location.origin + "/play";
}
