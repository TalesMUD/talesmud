import assert from "node:assert/strict";
import test from "node:test";
import { reasonChain, roomInspectorPath, roomTab, scriptInspectorPath } from "./islandLinks.js";

const rooms = ["R0215", "R0217", "R0207"];

test("each reason note links the room and the script", () => {
  const chain = reasonChain(
    "revealExit target 'deeper' missing on R0215 (SCR0205); R0215 no 'deeper'; SCR0204 never invoked; hidden exit 'down' on R0217 has no revealer; no inbound exit",
    rooms,
  );
  assert.equal(chain.length, 5);
  assert.equal(chain[0].parts[1].text, "R0215");
  assert.equal(chain[0].parts[1].href, roomInspectorPath("R0215"));
  assert.equal(chain[0].parts[3].text, "SCR0205");
  assert.equal(chain[0].parts[3].href, scriptInspectorPath("SCR0205"));
  assert.equal(chain[1].parts[0].href, roomInspectorPath("R0215"));
  assert.equal(chain[2].parts[0].href, scriptInspectorPath("SCR0204"));
  assert.equal(chain[3].parts[1].href, roomInspectorPath("R0217"));
  assert.equal(chain[4].parts.length, 1);
  assert.equal(chain[4].parts[0].href, undefined);
  assert.equal(chain[4].parts[0].text, "no inbound exit");
});

test("a parenthetical script id still links when the note is free text", () => {
  const chain = reasonChain("wired from (S2) on R0207", rooms);
  const script = chain[0].parts.find((part) => part.text === "S2");
  const room = chain[0].parts.find((part) => part.text === "R0207");
  assert.equal(script.href, scriptInspectorPath("S2"));
  assert.equal(room.href, roomInspectorPath("R0207"));
});

test("the room editor tab query opens the inspector", () => {
  assert.equal(roomTab("?id=R1&tab=inspector"), "inspector");
  assert.equal(roomTab("?tab=exits"), "exits");
  assert.equal(roomTab("?tab=map"), "");
  assert.equal(roomTab(""), "");
});
