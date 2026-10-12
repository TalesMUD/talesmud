/**
 * Escape overlay text, keep **bold**, and mark a lone Aggressive line in red.
 * The line is the look/examine marker for an aggroOnSight enemy.
 */
export function formatOverlayText(text) {
  const escaped = String(text || '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
  return escaped
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/(^|\n)Aggressive(?=\n|$)/g, '$1<span class="aggro-mark">Aggressive</span>');
}
