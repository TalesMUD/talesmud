// Guest tokens live in this tab only. A real Auth0 login must win over one,
// and logout must drop it so the next load is the welcome choice.

export const GUEST_TOKEN_KEY = "talesmud_guest_token";
export const PICKER_SEEN_KEY = "talesmud_char_picker_seen";

function defaultStorage() {
  try {
    if (typeof sessionStorage === "undefined") return null;
    return sessionStorage;
  } catch (err) {
    return null;
  }
}

function storeOf(storage) {
  return storage === undefined ? defaultStorage() : storage;
}

export function readGuestToken(storage) {
  const store = storeOf(storage);
  if (!store) return "";
  try {
    return store.getItem(GUEST_TOKEN_KEY) || "";
  } catch (err) {
    return "";
  }
}

export function clearGuestToken(storage) {
  const store = storeOf(storage);
  if (!store) return;
  try {
    store.removeItem(GUEST_TOKEN_KEY);
  } catch (err) {
    /* ignore */
  }
}

// True only when this tab's token is the guest token. A leftover guest token
// must not mark an Auth0 session as a guest.
export function isGuestSession(authToken, storage) {
  const guest = readGuestToken(storage);
  return !!authToken && !!guest && authToken === guest;
}

// auth0 | guest | welcome
export function restoredSession({ auth0Authenticated, guestToken }) {
  if (auth0Authenticated) return "auth0";
  if (guestToken) return "guest";
  return "welcome";
}

// Open the character picker once per login when a signed-in player has a choice.
export function shouldAutoOpenCharacterPicker({ guest, seen, characterCount }) {
  if (guest || seen) return false;
  return characterCount > 1;
}

// Drop the guest token and the one-shot character picker flag.
export function clearLocalSession(storage) {
  const store = storeOf(storage);
  clearGuestToken(store);
  if (!store) return;
  try {
    store.removeItem(PICKER_SEEN_KEY);
  } catch (err) {
    /* ignore */
  }
}
