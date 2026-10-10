import assert from "assert";
import { ACTIVATE_KEY, clearActivateCode, readActivateCode, stripActivateQuery } from "./sshActivate.js";

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
  assert.equal(readActivateCode("?code=auth0-code&activate=BCDF-GHJK", storage), "BCDF-GHJK");
  assert.equal(storage.getItem(ACTIVATE_KEY), "BCDF-GHJK");
  assert.equal(readActivateCode("", storage), "BCDF-GHJK");
  clearActivateCode(storage);
  assert.equal(readActivateCode("?code=auth0-code", storage), "");
}

{
  assert.equal(readActivateCode("", null), "");
}

{
  let replaced = "";
  const next = stripActivateQuery(
    "https://veilspan.example/play/?code=auth0-code&activate=BCDF-GHJK",
    (_data, _title, url) => { replaced = url; },
  );
  assert.equal(next, "/play/?code=auth0-code");
  assert.equal(replaced, "/play/?code=auth0-code");
  assert.equal(replaced.includes("BCDF-GHJK"), false);
  assert.equal(replaced.includes("activate="), false);
}

{
  let called = false;
  const href = "https://veilspan.example/play/?code=auth0-code";
  const out = stripActivateQuery(href, () => { called = true; });
  assert.equal(called, false);
  assert.equal(out, href);
}
