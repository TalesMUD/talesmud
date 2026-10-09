// Badges for the world-map overlay layers.
// The template passes the layer object in. A flag read only inside a
// closed-over function does not invalidate the tile.

const EMPTY = { levelBand: "", missingArt: false, badges: [], title: "" };

export function overlayView(room, layers, byId) {
  const row = byId?.get?.(room?.id);
  const flags = layers || {};
  if (!row) return { ...EMPTY, badges: [] };

  const badges = [];
  const lines = [];
  let levelBand = "";
  let missingArt = false;

  if (flags.level && row.levelBand) {
    levelBand = String(row.levelBand);
    const min = row.levelMin ?? 0;
    const max = row.levelMax ?? min;
    const where = row.levelSource === "zone" ? "zone" : "room";
    lines.push(min === max ? `Level ${min} (${where})` : `Levels ${min}–${max} (${where})`);
  }
  if (flags.aggro && row.aggro > 0) {
    const title = `${row.aggro} aggressive`;
    badges.push({ kind: "aggro", text: String(row.aggro), title });
    lines.push(title);
  }
  if (flags.spawners && row.spawners?.length) {
    const hover = row.spawners
      .map((spawner) => {
        const name = spawner.name || spawner.templateId || spawner.id;
        return spawner.respawnTime ? `${name} ${spawner.respawnTime}` : `${name} no respawn`;
      })
      .join(", ");
    badges.push({ kind: "spawner", text: "S", title: hover });
    lines.push(`Spawners: ${hover}`);
  }
  if (flags.players && row.players > 0) {
    const names = Array.isArray(row.playerNames) ? row.playerNames.filter(Boolean) : [];
    const title = names.length ? `${row.players} online: ${names.join(", ")}` : `${row.players} online`;
    badges.push({ kind: "players", text: String(row.players), title });
    lines.push(title);
  }
  if (flags.quests && row.quests?.length) {
    const ids = row.quests.filter(Boolean);
    const title = ids.join(", ");
    const text = ids.length > 1 ? `Q×${ids.length}` : "Q";
    badges.push({ kind: "quest", text, title });
    lines.push(`Quests: ${title}`);
  }
  if (flags.art && row.missingArt) {
    missingArt = true;
    lines.push("Missing art");
  }
  if (flags.copies && row.copies > 0) {
    const title = `${row.copies} live instance ${row.copies === 1 ? "copy" : "copies"}`;
    badges.push({ kind: "copies", text: `~${row.copies}`, title });
    lines.push(title);
  }

  return { levelBand, missingArt, badges, title: lines.join("\n") };
}
