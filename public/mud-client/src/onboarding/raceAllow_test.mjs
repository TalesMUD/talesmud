import assert from "assert";
import {
  CLASS_BLURBS,
  FALLBACK_TEMPLATES,
  racesForTemplate,
  raceAllowed,
  guestPickerEnabled,
  guestCreateBody,
} from "./raceAllow.js";

const ids = (template) => racesForTemplate(template).map((race) => race.id);

{
  assert.deepEqual(ids({ id: "tpl-fenwatch" }), ["human", "dwarf"]);
  assert.equal(raceAllowed({ id: "tpl-fenwatch", class: { id: "warrior" } }, "elf"), false);
  assert.equal(raceAllowed({ id: "tpl-fenwatch" }, "construct"), false);
  assert.equal(raceAllowed({ class: { id: "warrior" } }, "dwarf"), true);
}

{
  assert.deepEqual(ids({ id: "tpl-runehand", class: { id: "wizard" } }), ["human", "elf"]);
  assert.equal(raceAllowed({ id: "tpl-runehand" }, "dwarf"), false);
  assert.equal(raceAllowed({ class: { id: "mage" } }, "elf"), true);
}

{
  assert.deepEqual(ids({ id: "tpl-rigger" }), ["construct"]);
  assert.equal(raceAllowed({ id: "tpl-rigger" }, "human"), false);
  assert.equal(raceAllowed({ class: { id: "rigger" } }, "construct"), true);
  assert.equal(raceAllowed({ id: "tpl-hitch" }, "elf"), false);
  assert.equal(raceAllowed({ id: "tpl-hitch" }, "human"), true);
  assert.deepEqual(ids({ id: "tpl-ward" }), ["human", "dwarf"]);
  assert.equal(raceAllowed({ id: "tpl-ward" }, "elf"), false);
  assert.ok(FALLBACK_TEMPLATES.some((t) => t.id === "tpl-ward" && t.name === "Ward"));
  assert.ok(!FALLBACK_TEMPLATES.some((t) => /hitch/i.test(`${t.id} ${t.name}`)));
  assert.equal(CLASS_BLURBS.ward.includes("Warrior"), false);
  assert.equal(CLASS_BLURBS.ward, "Heavy plate. You start slow. Hits you take stack Grit, and Slam and the hit you throw back get heavier.");
  assert.equal(raceAllowed({ id: "tpl-alley" }, "elf"), true);
}

{
  assert.equal(guestPickerEnabled("veilspan.com"), true);
  assert.equal(guestPickerEnabled("www.veilspan.com"), true);
  assert.equal(guestPickerEnabled("talesmud.io"), false);
  assert.equal(guestPickerEnabled("www.talesmud.io"), false);
  assert.equal(guestPickerEnabled("localhost"), false);
}

{
  assert.deepEqual(guestCreateBody("talesmud.io", "tpl-rigger", "construct"), {});
  assert.deepEqual(guestCreateBody("www.talesmud.io", "tpl-fenwatch", "elf"), {});
  assert.deepEqual(guestCreateBody("veilspan.com", "", ""), {});
  assert.deepEqual(guestCreateBody("veilspan.com", "tpl-rigger", "construct"), {
    templateId: "tpl-rigger",
    race: "construct",
  });
  assert.equal(guestCreateBody("www.veilspan.com", "tpl-rigger", "human"), null);
  assert.equal(guestCreateBody("veilspan.com", "tpl-fenwatch", "elf"), null);
  assert.equal(guestCreateBody("veilspan.com", "tpl-rigger", ""), null);
  assert.equal(CLASS_BLURBS.rigger.includes("Rig drops a turret"), true);
  const construct = racesForTemplate({ id: "tpl-rigger" })[0];
  assert.equal(construct.name, "Construct");
  assert.equal(construct.blurb.includes("Poison never sticks"), true);
}

console.log("raceAllow_test ok");
