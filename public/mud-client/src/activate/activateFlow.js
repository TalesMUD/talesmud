// Helpers for the standalone /activate page. Pure so node --test can cover them.

export const ACTIVATE_RETURN = "/activate";

// normalizeCode uppercases, drops anything that is not a letter or digit,
// and puts the dash back after four characters. "bcdf ghjk" -> "BCDF-GHJK".
export function normalizeCode(raw) {
  const clean = String(raw || "").toUpperCase().replace(/[^A-Z0-9]/g, "").slice(0, 8);
  if (clean.length <= 4) return clean;
  return clean.slice(0, 4) + "-" + clean.slice(4);
}

// A complete code is eight characters with the dash.
export function isCompleteCode(code) {
  return /^[A-Z0-9]{4}-[A-Z0-9]{4}$/.test(String(code || ""));
}

// codeFromLocation reads ?code= (or the older ?activate=) from the page URL.
// Auth0 never redirects back to /activate, so code here is always ours.
export function codeFromLocation(search) {
  const params = new URLSearchParams(search || "");
  return normalizeCode(params.get("code") || params.get("activate") || "");
}

// returnTarget is where the play client sends the browser after the Auth0
// callback. Only the exact activate path is accepted, so appState cannot
// become an open redirect.
export function returnTarget(appState) {
  const target = appState && typeof appState.returnTo === "string" ? appState.returnTo : "";
  return target === ACTIVATE_RETURN ? ACTIVATE_RETURN : "";
}

// activateRedirect turns an old /play/?activate=CODE link into /activate?code=CODE.
export function activateRedirect(search) {
  const params = new URLSearchParams(search || "");
  if (!params.has("activate")) return "";
  const code = normalizeCode(params.get("activate"));
  return code ? `${ACTIVATE_RETURN}?code=${encodeURIComponent(code)}` : ACTIVATE_RETURN;
}
