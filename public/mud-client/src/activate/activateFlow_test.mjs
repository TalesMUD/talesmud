import test from "node:test";
import assert from "node:assert/strict";
import { normalizeCode, isCompleteCode, codeFromLocation, returnTarget, activateRedirect } from "./activateFlow.js";

test("normalizeCode formats user input", () => {
  assert.equal(normalizeCode("bcdf ghjk"), "BCDF-GHJK");
  assert.equal(normalizeCode("BCDF-GHJK"), "BCDF-GHJK");
  assert.equal(normalizeCode("bc"), "BC");
  assert.equal(normalizeCode("bcdfg"), "BCDF-G");
  assert.equal(normalizeCode("<b>cdf-ghjk-xx"), "BCDF-GHJK");
  assert.equal(normalizeCode(null), "");
});

test("isCompleteCode", () => {
  assert.ok(isCompleteCode("BCDF-GHJK"));
  assert.ok(!isCompleteCode("BCDF-GHJ"));
  assert.ok(!isCompleteCode("bcdf-ghjk"));
});

test("codeFromLocation reads code or activate", () => {
  assert.equal(codeFromLocation("?code=bcdf-ghjk"), "BCDF-GHJK");
  assert.equal(codeFromLocation("?activate=BCDFGHJK"), "BCDF-GHJK");
  assert.equal(codeFromLocation(""), "");
});

test("returnTarget only allows /activate", () => {
  assert.equal(returnTarget({ returnTo: "/activate" }), "/activate");
  assert.equal(returnTarget({ returnTo: "https://evil.example/activate" }), "");
  assert.equal(returnTarget({ returnTo: "//evil.example" }), "");
  assert.equal(returnTarget({ returnTo: "/activate?x=1" }), "");
  assert.equal(returnTarget(undefined), "");
});

test("activateRedirect maps old play links", () => {
  assert.equal(activateRedirect("?activate=bcdf-ghjk"), "/activate?code=BCDF-GHJK");
  assert.equal(activateRedirect("?activate="), "/activate");
  assert.equal(activateRedirect("?code=abc&state=x"), "");
});
