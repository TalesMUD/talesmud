// Island reasons are the same semicolon-separated notes content health uses.
// Room ids and script ids in those notes become inspector links.

const ROOM_TABS = ["exits", "actions", "spawners", "items", "residents", "inspector"];

export function roomInspectorPath(id) {
  return `/creator/rooms?id=${encodeURIComponent(id)}&tab=inspector`;
}

export function scriptInspectorPath(id) {
  return `/creator/scripts?id=${encodeURIComponent(id)}`;
}

export function roomTab(search) {
  const tab = new URLSearchParams(search || "").get("tab");
  return ROOM_TABS.includes(tab) ? tab : "";
}

export function reasonChain(reason, roomIds) {
  const rooms = roomIds instanceof Set ? roomIds : new Set(roomIds || []);
  return String(reason || "")
    .split(";")
    .map((note) => note.trim())
    .filter(Boolean)
    .map((note) => ({ parts: linkifyNote(note, rooms) }));
}

function linkifyNote(note, rooms) {
  const missing = note.match(/^revealExit target '([^']*)' missing on (\S+) \((.+)\)$/);
  if (missing) {
    return [
      { text: `revealExit target '${missing[1]}' missing on ` },
      roomLink(missing[2]),
      { text: " (" },
      scriptLink(missing[3]),
      { text: ")" },
    ];
  }
  const absent = note.match(/^(\S+) no '([^']*)'$/);
  if (absent) {
    return [roomLink(absent[1]), { text: ` no '${absent[2]}'` }];
  }
  const idle = note.match(/^(\S+) never invoked$/);
  if (idle) {
    return [scriptLink(idle[1]), { text: " never invoked" }];
  }
  const hidden = note.match(/^hidden exit '([^']*)' on (\S+) has no revealer$/);
  if (hidden) {
    return [
      { text: `hidden exit '${hidden[1]}' on ` },
      roomLink(hidden[2]),
      { text: " has no revealer" },
    ];
  }
  return scanNote(note, rooms);
}

function scanNote(note, rooms) {
  const re = /[A-Za-z][A-Za-z0-9_-]*/g;
  const parts = [];
  let last = 0;
  let match;
  while ((match = re.exec(note))) {
    const token = match[0];
    const kind = tokenKind(token, rooms, note);
    if (!kind) continue;
    if (match.index > last) parts.push({ text: note.slice(last, match.index) });
    parts.push(kind === "room" ? roomLink(token) : scriptLink(token));
    last = match.index + token.length;
  }
  if (last < note.length) parts.push({ text: note.slice(last) });
  return parts.length ? parts : [{ text: note }];
}

function tokenKind(token, rooms, note) {
  if (rooms.has(token)) return "room";
  if (/^SCR/i.test(token)) return "script";
  if (note.includes(`(${token})`)) return "script";
  if (note === `${token} never invoked`) return "script";
  return "";
}

function roomLink(id) {
  return { text: id, href: roomInspectorPath(id) };
}

function scriptLink(id) {
  return { text: id, href: scriptInspectorPath(id) };
}
