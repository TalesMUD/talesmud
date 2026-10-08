// Enemy-hook room lines use style "combatEvent" on a normal message.

export function isCombatEvent(msg) {
  return !!(msg && msg.style === "combatEvent" && msg.message);
}

export function combatEventIcon(hook) {
  switch (hook) {
    case "onAggro":
      return "music_note";
    case "onLowHealth":
      return "favorite";
    case "onDeath":
      return "dangerous";
    case "onFlee":
      return "directions_run";
    default:
      return "campaign";
  }
}

export function combatEventCard(msg) {
  if (!isCombatEvent(msg)) return null;
  return {
    text: String(msg.message),
    hook: msg.hook || "",
    source: msg.source || "",
    icon: combatEventIcon(msg.hook),
  };
}
