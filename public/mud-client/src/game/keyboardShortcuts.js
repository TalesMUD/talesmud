/** True when the event target is a command line or any text field. */
export function isTextEntry(target) {
  if (!target || typeof target !== 'object') return false;
  const tag = String(target.tagName || '').toLowerCase();
  if (tag === 'input' || tag === 'textarea' || tag === 'select') return true;
  if (target.isContentEditable) return true;
  if (typeof target.closest === 'function') {
    if (target.closest('input, textarea, select, [contenteditable="true"], .xterm')) return true;
  }
  return false;
}

export function rarityClass(quality) {
  const value = String(quality || 'normal').toLowerCase();
  if (value === 'magic' || value === 'rare' || value === 'legendary' || value === 'mythic') return value;
  return 'normal';
}

export function prefersReducedMotion() {
  if (typeof window === 'undefined' || !window.matchMedia) return false;
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}
