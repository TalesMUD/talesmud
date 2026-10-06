import assert from "assert";
import { installClassCatalog } from "./classCatalog.js";
import { packPayload } from "../../test/packFixture.mjs";
import {
  FALLBACK_TEMPLATES,
  classBlurb,
  racesForTemplate,
  raceAllowed,
  guestPickerEnabled,
  guestCreateBody,
  originPortraitSrc,
} from "./raceAllow.js";

assert.ok(FALLBACK_TEMPLATES.some((t) => t.id === "tpl-warrior" && t.name === "Warrior"));
assert.ok(!FALLBACK_TEMPLATES.some((t) => t.id === "tpl-ward" || t.name === "Ward"));
assert.ok(!/sentinel|cutpurse|runecaster|rigger/i.test(JSON.stringify(FALLBACK_TEMPLATES)));
installClassCatalog(packPayload());
globalThis.TALESMUD_GUEST_PICKER_HOSTS = "play.example.com, www.play.example.com";

const ids = (template) => racesForTemplate(template).map((race) => race.id);

{
  assert.deepEqual(ids({ id: "tpl-sentinel" }), ["human", "dwarf"]);
  assert.equal(raceAllowed({ id: "tpl-sentinel", class: { id: "warrior" } }, "elf"), false);
  assert.equal(raceAllowed({ id: "tpl-sentinel" }, "construct"), false);
  assert.equal(raceAllowed({ class: { id: "warrior" } }, "dwarf"), true);
}

{
  assert.deepEqual(ids({ id: "tpl-runecaster", class: { id: "wizard" } }), ["human", "elf"]);
  assert.equal(raceAllowed({ id: "tpl-runecaster" }, "dwarf"), false);
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
  assert.ok(!FALLBACK_TEMPLATES.some((t) => /hitch/i.test(`${t.id} ${t.name}`)));
  assert.equal(classBlurb("ward").includes("Warrior"), false);
  assert.equal(classBlurb("ward"), "Heavy plate. You start slow. Hits you take stack Grit, and Slam and the hit you throw back get heavier.");
  assert.equal(raceAllowed({ id: "tpl-cutpurse" }, "elf"), true);
}

{
  assert.equal(guestPickerEnabled("play.example.com"), true);
  assert.equal(guestPickerEnabled("www.play.example.com"), true);
  assert.equal(guestPickerEnabled("talesmud.io"), false);
  assert.equal(guestPickerEnabled("www.talesmud.io"), false);
  assert.equal(guestPickerEnabled("localhost"), false);
}

{
  assert.deepEqual(guestCreateBody("talesmud.io", "tpl-rigger", "construct"), {});
  assert.deepEqual(guestCreateBody("www.talesmud.io", "tpl-sentinel", "elf"), {});
  assert.deepEqual(guestCreateBody("play.example.com", "", ""), {});
  assert.deepEqual(guestCreateBody("play.example.com", "tpl-rigger", "construct"), {
    templateId: "tpl-rigger",
    race: "construct",
  });
  assert.equal(guestCreateBody("www.play.example.com", "tpl-rigger", "human"), null);
  assert.equal(guestCreateBody("play.example.com", "tpl-sentinel", "elf"), null);
  assert.equal(guestCreateBody("play.example.com", "tpl-rigger", ""), null);
  assert.equal(classBlurb("rigger").includes("Rig drops a turret"), true);
  const construct = racesForTemplate({ id: "tpl-rigger" })[0];
  assert.equal(construct.name, "Construct");
  assert.equal(construct.blurb.includes("Poison never sticks"), true);
}

{
  assert.equal(originPortraitSrc({ id: "tpl-ward" }, "human"), "/api/portraits/player-human-ward.png");
  assert.equal(originPortraitSrc({ id: "tpl-ward", class: { id: "ward" } }, "Dwarf"), "/api/portraits/player-dwarf-ward.png");
  assert.equal(originPortraitSrc({ id: "tpl-ward" }, ""), "/api/portraits/player-human-ward.png");
  assert.equal(originPortraitSrc({ id: "tpl-ward" }, "elf"), "");
  assert.equal(originPortraitSrc({ id: "tpl-rigger" }, "construct"), "");
  assert.equal(originPortraitSrc({ id: "tpl-sentinel" }, "dwarf"), "/api/portraits/player-dwarf-warrior.png");
  assert.equal(originPortraitSrc({ id: "tpl-runecaster" }, ""), "/api/portraits/player-human-mage.png");
  assert.equal(originPortraitSrc({ id: "tpl-cutpurse" }, "elve"), "/api/portraits/player-elf-rogue.png");
}

console.log("raceAllow_test ok");
