// Device-code handoff for the play client. The query name is "activate"
// so it does not collide with Auth0's redirect "code" parameter.

export const ACTIVATE_KEY = "talesmud_ssh_activate";

// readActivateCode keeps the code across the Auth0 round trip.
// A query value wins and is stored. Otherwise the stored value is returned.
export function readActivateCode(search, storage) {
  const params = new URLSearchParams(search || "");
  const fromQuery = String(params.get("activate") || "").trim();
  if (fromQuery) {
    if (storage && typeof storage.setItem === "function") {
      storage.setItem(ACTIVATE_KEY, fromQuery);
    }
    return fromQuery;
  }
  if (storage && typeof storage.getItem === "function") {
    return String(storage.getItem(ACTIVATE_KEY) || "").trim();
  }
  return "";
}

export function clearActivateCode(storage) {
  if (storage && typeof storage.removeItem === "function") {
    storage.removeItem(ACTIVATE_KEY);
  }
}

// stripActivateQuery removes activate from the visible URL after it has been
// stored. Auth0's code parameter is left in place. replaceState is called
// only when activate is present.
export function stripActivateQuery(href, replaceState) {
  if (typeof href !== "string" || typeof replaceState !== "function") {
    return href;
  }
  let url;
  try {
    url = new URL(href, "http://localhost");
  } catch {
    return href;
  }
  if (!url.searchParams.has("activate")) {
    return href;
  }
  url.searchParams.delete("activate");
  const next = url.pathname + url.search + url.hash;
  replaceState(null, "", next);
  return next;
}
