/**
 * Parse server examine dumps shaped like:
 *   === Item Name ===
 *   short blurb
 *   longer lore...
 *   --- Item Details ---
 *   Type: Quest Item (artifact_fragment)
 *   Quality: Normal
 * into a structured card model for RoomTextOverlay.
 *
 * Returns null when the text is not an examine dump.
 */

const HEADER_RE = /^===\s*(.+?)\s*===/;
const SECTION_RE = /^---\s*(.+?)\s*---\s*$/;
const DETAIL_LINE_RE = /^([^:\n]+):\s*(.+)$/;

/** Snake_case / camel internal ids that should not appear in player chrome. */
function isInternalId(value) {
  const s = String(value || '').trim();
  if (!s) return true;
  if (/^[a-z0-9]+(_[a-z0-9]+)+$/.test(s)) return true; // artifact_fragment
  if (/^[a-z]+[A-Z]/.test(s)) return true; // camelCase leftovers
  return false;
}

/**
 * Strip trailing "(internal_id)" from Type lines while keeping human labels.
 * "Quest Item (artifact_fragment)" → "Quest Item"
 * "Weapon (Sword)" → "Weapon (Sword)" when parenthetical is human.
 */
export function cleanDetailValue(key, value) {
  let v = String(value || '').trim();
  if (!v) return v;

  const paren = v.match(/^(.*?)(\s*)\(([^)]+)\)\s*$/);
  if (paren) {
    const outer = paren[1].trim();
    const inner = paren[3].trim();
    if (isInternalId(inner)) {
      v = outer;
    }
  }

  // Bare internal id as the whole value → drop (caller skips empty)
  if (isInternalId(v) && !/\s/.test(v) && !/^[A-Z]/.test(v)) {
    return '';
  }

  return v;
}

/**
 * @param {string} text
 * @returns {null | {
 *   title: string,
 *   blurb: string,
 *   lore: string,
 *   details: Array<{label: string, value: string}>,
 *   attributes: Array<{label: string, value: string}>,
 *   properties: Array<{label: string, value: string}>,
 * }}
 */
export function parseExamineText(text) {
  const raw = String(text || '').replace(/\r\n/g, '\n').trim();
  if (!raw) return null;

  const headerMatch = raw.match(HEADER_RE);
  if (!headerMatch) return null;

  // Must look like an examine dump (details section or Type:/Quality: lines)
  if (!/---\s*Item Details\s*---/i.test(raw) && !/^Type:\s*/m.test(raw)) {
    // Still accept === Title === dumps that are clearly examine-shaped
    // (multi-line after header). Single-line === banners are not examine cards.
    const afterHeader = raw.slice(headerMatch[0].length).trim();
    if (!afterHeader || !afterHeader.includes('\n')) return null;
  }

  const title = headerMatch[1].trim();
  const body = raw.slice(headerMatch[0].length).replace(/^\n+/, '');

  const lines = body.split('\n');
  /** @type {'prose'|'details'|'attributes'|'properties'} */
  let mode = 'prose';
  const proseLines = [];
  const details = [];
  const attributes = [];
  const properties = [];
  const effects = [];

  for (const line of lines) {
    const section = line.match(SECTION_RE);
    if (section) {
      const name = section[1].trim().toLowerCase();
      if (name === 'item details') mode = 'details';
      else if (name === 'attributes') mode = 'attributes';
      else if (name === 'properties') mode = 'properties';
      else if (name === 'effects') mode = 'effects';
      else mode = 'details';
      continue;
    }

    if (mode === 'prose') {
      proseLines.push(line);
      continue;
    }

    const trimmed = line.trim();
    if (!trimmed) continue;
    if (mode === 'effects') {
      effects.push(trimmed);
      continue;
    }
    const kv = trimmed.match(DETAIL_LINE_RE);
    if (!kv) continue;
    const label = kv[1].trim();
    const value = cleanDetailValue(label, kv[2]);
    if (!value) continue;
    const row = { label, value };
    if (mode === 'attributes') attributes.push(row);
    else if (mode === 'properties') properties.push(row);
    else details.push(row);
  }

  // Split prose: first non-empty paragraph = blurb, rest = lore
  const prose = proseLines.join('\n').trim();
  let blurb = '';
  let lore = '';
  if (prose) {
    const parts = prose.split(/\n\s*\n/).map((p) => p.trim()).filter(Boolean);
    if (parts.length === 0) {
      blurb = '';
      lore = '';
    } else if (parts.length === 1) {
      // Single block: if long, treat as lore; if short, blurb
      if (parts[0].length > 140 || parts[0].includes('\n')) {
        lore = parts[0];
      } else {
        blurb = parts[0];
      }
    } else {
      blurb = parts[0];
      lore = parts.slice(1).join('\n\n');
    }
  }

  return {
    title,
    blurb,
    lore,
    details,
    attributes,
    properties,
    effects,
  };
}

/** Title-case a snake_case / camelCase key for player chrome. */
function formatLabel(key) {
  const raw = String(key || '').trim();
  if (!raw) return '';
  return raw
    .replace(/_/g, ' ')
    .replace(/([a-z])([A-Z])/g, '$1 $2')
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

function formatStatValue(value) {
  if (value == null) return '';
  if (typeof value === 'boolean') return value ? 'Yes' : 'No';
  if (typeof value === 'number') {
    if (Number.isInteger(value)) return String(value);
    return String(Math.round(value * 10) / 10);
  }
  return String(value).trim();
}

/**
 * Build the examine card model from a structured `examineResult` websocket
 * payload. Prefer this over parseExamineText whenever the server sends it.
 *
 * @param {object} msg
 * @returns {null | {
 *   title: string,
 *   blurb: string,
 *   lore: string,
 *   details: Array<{label: string, value: string}>,
 *   attributes: Array<{label: string, value: string}>,
 *   properties: Array<{label: string, value: string}>,
 *   effects: string[],
 * }}
 */
export function examineResultToCard(msg) {
  if (!msg || typeof msg !== 'object') return null;
  const title = String(msg.name || '').trim();
  if (!title) return null;

  const description = String(msg.description || '').trim();
  const detail = String(msg.detail || '').trim();
  let blurb = '';
  let lore = '';
  if (description && detail) {
    blurb = description;
    lore = detail;
  } else if (description) {
    if (description.length > 140 || description.includes('\n')) lore = description;
    else blurb = description;
  } else if (detail) {
    lore = detail;
  }

  const details = [];
  const quality = String(msg.quality || '').trim();
  if (quality) {
    details.push({ label: 'Quality', value: formatLabel(quality) });
  }
  const slot = String(msg.slot || '').trim();
  if (slot && slot !== 'inventory') {
    details.push({ label: 'Equip Slot', value: formatLabel(slot) });
  }
  const value = Number(msg.value);
  if (Number.isFinite(value) && value > 0) {
    details.push({ label: 'Value', value: `${Math.trunc(value)} gold` });
  }

  const attributes = [];
  const stats = msg.stats;
  if (stats && typeof stats === 'object' && !Array.isArray(stats)) {
    for (const [key, raw] of Object.entries(stats)) {
      const label = formatLabel(key);
      const formatted = formatStatValue(raw);
      if (!label || !formatted) continue;
      attributes.push({ label, value: formatted });
    }
  }

  const effects = Array.isArray(msg.effects)
    ? msg.effects.map((line) => String(line || '').trim()).filter(Boolean)
    : [];

  return {
    title,
    blurb,
    lore,
    details,
    attributes,
    properties: [],
    effects,
  };
}

/** True when overlay text should render as an examine item card. */
export function isExamineOverlayText(text) {
  return parseExamineText(text) != null;
}
