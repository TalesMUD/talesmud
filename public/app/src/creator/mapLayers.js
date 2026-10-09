// World-map layer toggles. Multi-select, persisted beside the nav collapse flag.

export const MAP_LAYER_KEY = "tales.creator.mapLayers.v1";

export const MAP_LAYER_IDS = [
  "reachability",
  "level",
  "aggro",
  "spawners",
  "players",
  "quests",
  "art",
  "copies",
];

export function defaultLayers() {
  return {
    reachability: false,
    level: false,
    aggro: false,
    spawners: false,
    players: false,
    quests: false,
    art: false,
    copies: false,
  };
}

export function readLayers(storage) {
  const out = defaultLayers();
  if (!storage) return out;
  try {
    const parsed = JSON.parse(storage.getItem(MAP_LAYER_KEY) || "");
    if (!parsed || typeof parsed !== "object") return out;
    for (const id of MAP_LAYER_IDS) {
      if (parsed[id] === true) out[id] = true;
    }
  } catch {
    return defaultLayers();
  }
  return out;
}

export function writeLayers(storage, layers) {
  if (!storage) return;
  const out = {};
  for (const id of MAP_LAYER_IDS) {
    out[id] = layers?.[id] === true;
  }
  storage.setItem(MAP_LAYER_KEY, JSON.stringify(out));
}

export function overlayLayersOn(layers) {
  return !!(
    layers?.level ||
    layers?.aggro ||
    layers?.spawners ||
    layers?.players ||
    layers?.quests ||
    layers?.art ||
    layers?.copies
  );
}
