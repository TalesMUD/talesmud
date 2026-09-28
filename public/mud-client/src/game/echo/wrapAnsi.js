/** Visible columns of a string. ANSI CSI/OSC sequences take none. */
export function visibleWidth(text) {
  const src = String(text ?? '');
  let col = 0;
  let i = 0;
  while (i < src.length) {
    if (src[i] === '\u001b') {
      i += consumeAnsi(src, i).length;
      continue;
    }
    const cp = src.codePointAt(i);
    const ch = String.fromCodePoint(cp);
    if (ch !== '\n' && ch !== '\r') col += glyphWidth(cp);
    i += ch.length;
  }
  return col;
}

/**
 * Soft-wrap on spaces so a terminal column does not split a word.
 * Existing newlines stay. A token longer than `cols` is broken.
 * ANSI sequences are copied through and do not count as columns.
 */
export function wrapAnsi(text, cols) {
  const width = Math.max(1, Number(cols) || 1);
  const src = String(text ?? '').replace(/\r\n/g, '\n').replace(/\r/g, '\n');
  let out = '';
  let col = 0;
  let breakAt = -1;
  let i = 0;

  while (i < src.length) {
    if (src[i] === '\u001b') {
      const seq = consumeAnsi(src, i);
      out += seq;
      i += seq.length;
      continue;
    }
    const cp = src.codePointAt(i);
    const ch = String.fromCodePoint(cp);
    if (ch === '\n') {
      out += '\n';
      col = 0;
      breakAt = -1;
      i += ch.length;
      continue;
    }
    const w = glyphWidth(cp);
    if (w > 0 && col + w > width) {
      if (ch === ' ' || ch === '\t') {
        out += '\n';
        col = 0;
        breakAt = -1;
        i += ch.length;
        continue;
      }
      if (breakAt >= 0) {
        const rest = out.slice(breakAt + 1);
        out = out.slice(0, breakAt) + '\n' + rest;
        col = visibleWidth(rest);
        breakAt = -1;
      }
      if (col + w > width && col > 0) {
        out += '\n';
        col = 0;
      }
    }
    out += ch;
    if (ch === ' ' || ch === '\t') breakAt = out.length - 1;
    col += w;
    i += ch.length;
  }
  return out;
}

function glyphWidth(cp) {
  if (cp === 0x09) return 1;
  if (cp < 32 || cp === 0x7f) return 0;
  if (cp >= 0x300 && cp <= 0x36f) return 0;
  if (
    (cp >= 0x1100 && cp <= 0x115f) ||
    cp === 0x2329 || cp === 0x232a ||
    (cp >= 0x2e80 && cp <= 0xa4cf) ||
    (cp >= 0xac00 && cp <= 0xd7a3) ||
    (cp >= 0xf900 && cp <= 0xfaff) ||
    (cp >= 0xfe10 && cp <= 0xfe19) ||
    (cp >= 0xfe30 && cp <= 0xfe6f) ||
    (cp >= 0xff00 && cp <= 0xff60) ||
    (cp >= 0xffe0 && cp <= 0xffe6) ||
    (cp >= 0x1f300 && cp <= 0x1f9ff)
  ) {
    return 2;
  }
  return 1;
}

function consumeAnsi(src, i) {
  if (src[i] !== '\u001b' || i + 1 >= src.length) return src[i] || '';
  const next = src[i + 1];
  if (next === '[') {
    let j = i + 2;
    while (j < src.length) {
      const c = src.charCodeAt(j);
      if (c >= 0x40 && c <= 0x7e) return src.slice(i, j + 1);
      j += 1;
    }
    return src.slice(i);
  }
  if (next === ']') {
    let j = i + 2;
    while (j < src.length) {
      if (src[j] === '\u0007') return src.slice(i, j + 1);
      if (src[j] === '\u001b' && src[j + 1] === '\\') return src.slice(i, j + 2);
      j += 1;
    }
    return src.slice(i);
  }
  return src.slice(i, i + 2);
}
