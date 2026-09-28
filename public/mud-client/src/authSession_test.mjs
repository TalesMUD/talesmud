import assert from "assert";
import {
  GUEST_TOKEN_KEY,
  PICKER_SEEN_KEY,
  clearLocalSession,
  isGuestSession,
  readGuestToken,
  restoredSession,
  shouldAutoOpenCharacterPicker,
} from "./authSession.js";

function mem() {
  const data = new Map();
  return {
    getItem(key) { return data.has(key) ? data.get(key) : null; },
    setItem(key, value) { data.set(key, String(value)); },
    removeItem(key) { data.delete(key); },
  };
}

{
  const storage = mem();
  assert.equal(readGuestToken(storage), "");
  assert.equal(isGuestSession("", storage), false);
  assert.equal(isGuestSession("auth0-token", storage), false);
  assert.equal(restoredSession({ auth0Authenticated: false, guestToken: "" }), "welcome");
}

{
  const storage = mem();
  storage.setItem(GUEST_TOKEN_KEY, "guest-token");
  storage.setItem(PICKER_SEEN_KEY, "1");
  assert.equal(isGuestSession("guest-token", storage), true);
  assert.equal(isGuestSession("auth0-token", storage), false);
  assert.equal(restoredSession({ auth0Authenticated: true, guestToken: "guest-token" }), "auth0");
  assert.equal(restoredSession({ auth0Authenticated: false, guestToken: "guest-token" }), "guest");
  clearLocalSession(storage);
  assert.equal(readGuestToken(storage), "");
  assert.equal(storage.getItem(PICKER_SEEN_KEY), null);
  assert.equal(restoredSession({ auth0Authenticated: false, guestToken: readGuestToken(storage) }), "welcome");
}

assert.equal(shouldAutoOpenCharacterPicker({ guest: false, seen: false, characterCount: 9 }), true);
assert.equal(shouldAutoOpenCharacterPicker({ guest: false, seen: false, characterCount: 1 }), false);
assert.equal(shouldAutoOpenCharacterPicker({ guest: true, seen: false, characterCount: 9 }), false);
assert.equal(shouldAutoOpenCharacterPicker({ guest: false, seen: true, characterCount: 9 }), false);

console.log("authSession_test ok");
