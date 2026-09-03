// Shared room-chat + presence helpers. Prefer existing WS payloads:
// roomPresence / enterRoom.players, and default "message" say lines
// (username set, or "Name says: text" from the say command).

const SAY_RE = /^(.+?)\s+says:\s*(.*)$/i;

export function isSelfName(name, character) {
  const n = String(name || "").trim().toLowerCase();
  const you = String(character?.name || "").trim().toLowerCase();
  return Boolean(n && you && n === you);
}

export function markPlayersYou(players, character) {
  const id = character?.id;
  const name = String(character?.name || "").trim().toLowerCase();
  return (players || []).map((player) => {
    const isYou = Boolean(
      player.isYou ||
        (id && player.id === id) ||
        (name && String(player.name || "").trim().toLowerCase() === name)
    );
    return { ...player, isYou };
  });
}

export function parseRoomChatLine(msg, character) {
  if (!msg || typeof msg !== "object") return null;
  const type = String(msg.type || "message");
  const username = String(msg.username || msg.speaker || "").trim();
  const raw = String(msg.message || "").trim();
  if (!raw) return null;

  const say = raw.match(SAY_RE);
  if (username) {
    return {
      name: username,
      text: raw,
      isYou: isSelfName(username, character),
    };
  }
  if (type === "say" || type === "emote" || type === "roomMessage") {
    const name = say ? say[1].trim() : "";
    const text = say ? say[2] : raw;
    return {
      name,
      text,
      isYou: isSelfName(name, character),
    };
  }
  if (say) {
    const name = say[1].trim();
    return {
      name,
      text: say[2],
      isYou: isSelfName(name, character),
    };
  }
  return null;
}
