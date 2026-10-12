import test from "node:test";
import assert from "node:assert/strict";
import { creatorSearchField, groupHits, isTypingTarget, referencedByHint } from "./searchGroups.js";

test("groups keep the ranked order", () => {
  const groups = groupHits([
    { type: "quest", id: "Q1" },
    { type: "item", id: "I1" },
    { type: "quest", id: "Q2" },
  ]);
  assert.deepEqual(groups.map((group) => group.type), ["quest", "item"]);
  assert.deepEqual(groups[0].hits.map((hit) => hit.id), ["Q1", "Q2"]);
  assert.equal(groups[0].hits[1].index, 2);
});

test("the top-bar search field is recognized", () => {
  assert.equal(creatorSearchField(null), null);
  assert.equal(creatorSearchField({ closest: () => null }), null);
  const field = {};
  assert.equal(
    creatorSearchField({
      closest: (selector) => (selector === "[data-creator-search]" ? field : null),
    }),
    field
  );
});

test("typing targets are ignored", () => {
  assert.equal(isTypingTarget(null), false);
  assert.equal(isTypingTarget({ closest: () => null, isContentEditable: false }), false);
  assert.equal(isTypingTarget({ closest: () => ({}), isContentEditable: false }), true);
  assert.equal(isTypingTarget({ closest: () => null, isContentEditable: true }), true);
});

test("delete hint names the first inbound refs", () => {
  const hint = referencedByHint({
    inbound: [
      {
        type: "quest",
        refs: [
          { type: "quest", id: "QST0217", name: "The Cooked Book" },
          { type: "quest", id: "Q2", name: "" },
        ],
      },
      {
        type: "script",
        refs: [
          { type: "script", id: "SCR0204", name: "Hatch" },
          { type: "script", id: "S2", name: "Two" },
          { type: "script", id: "S3", name: "Three" },
        ],
      },
    ],
  });
  assert.equal(
    hint,
    "Referenced by 5: quest The Cooked Book, quest Q2, script Hatch, script Two, …"
  );
  assert.equal(referencedByHint({ inbound: [] }), "");
});
