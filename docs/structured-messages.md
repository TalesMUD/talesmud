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
