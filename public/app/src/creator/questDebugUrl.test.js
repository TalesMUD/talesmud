import test from "node:test";
import assert from "node:assert/strict";
import { questIdFromQuery } from "./questDebugUrl.js";

test("missing query has no quest id", () => {
  assert.equal(questIdFromQuery(undefined), "");
  assert.equal(questIdFromQuery({}), "");
  assert.equal(questIdFromQuery({ id: "" }), "");
});

test("a string id is kept", () => {
  assert.equal(questIdFromQuery({ id: "QST0217" }), "QST0217");
});

test("duplicate ids use the first value", () => {
  assert.equal(questIdFromQuery({ id: ["QST0217", "QST1"] }), "QST0217");
  assert.equal(questIdFromQuery({ id: ["", "QST1"] }), "");
});
