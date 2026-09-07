// Placeholders are replaced at Rollup build time from VITE_AUTH0_* env vars.
// Unreplaced placeholders mean Auth0 is not configured (guest play still works).
export const auth0Config = {
  domain: "__VITE_AUTH0_DOMAIN__",
  client_id: "__VITE_AUTH0_CLIENT_ID__",
  audience: "__VITE_AUTH0_AUDIENCE__",
};

export function isAuth0Configured() {
  return (
    Boolean(auth0Config.domain) &&
    Boolean(auth0Config.client_id) &&
    !auth0Config.domain.includes("__VITE_") &&
    !auth0Config.client_id.includes("__VITE_")
  );
}
