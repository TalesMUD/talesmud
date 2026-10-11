# Structured websocket messages (protocol 1)

These additions preserve existing message types, fields, and prose. New fields
use `omitempty`; clients must ignore unknown fields and message types.

## Item inspection

Successful `examine <item>` and `inspect <item>` send the original `message`
reply, followed by `examineResult`. Item lookup still checks inventory, room,
and equipment. Errors and room actions keep their existing replies.

`examineResult` carries optional `name`, `slot`, `quality`, `stats` (the item's
attributes), `value` (base price in gold), `description`, `detail`, and `effects`
(strings formatted with `items.FormatEffect`). Script effects without authored
text get the same named "Special effect" fallback as prose. Empty/zero fields
are omitted. NPC inspection remains prose.

## Rejoining combat

Character selection (including the automatic selection on reconnect) sends one
`combatStatus` when `inCombat` is true and the character is a living, non-fled
participant in an active in-memory fight. A stale saved flag sends no snapshot.
It reads the fight under its pacing lock without starting it or changing turns.

The snapshot includes the full `combatants` roster, `round`, `actorId`,
`actorName`, `deadlineMs` (absolute Unix milliseconds, omitted if no decision
window), and the player's `targetId`, queue, cooldowns, and pacing clocks.
The web play client merges the roster and restores the battle, turn, and focus
immediately. Ordinary queue updates omit the new turn/target fields.
