import { WIDGET_TYPES } from './WidgetRegistry.js';

/**
 * Personal HUD layout templates (named snapshots).
 *
 * Storage: same localStorage blob as the active layout (`talesmud_layout_v1`),
 * extended to version 2. Layout prefs are browser-local today — templates stay
 * local too. Future share: optional `shareId` on a template (not implemented).
 */

export const LAYOUT_STORAGE_KEY = 'talesmud_layout_v1';
export const LAYOUT_STORAGE_VERSION = 2;

const KNOWN_WIDGETS = new Set(Object.keys(WIDGET_TYPES));

/** Reject unusable saved grids before they can replace the viewport preset. */
export function isUsableLayout(widgets) {
  if (!Array.isArray(widgets) || widgets.length === 0) return false;
  const ids = new Set();
  let visible = false;
  for (const w of widgets) {
    if (!w || typeof w !== 'object' || typeof w.id !== 'string' || !w.id ||
        ids.has(w.id) || !KNOWN_WIDGETS.has(w.widgetType)) return false;
    ids.add(w.id);
    for (const key of ['x', 'y', 'w', 'h']) {
      if (!Number.isFinite(Number(w[key])) || Math.abs(Number(w[key])) > 10000) return false;
    }
    if (Number(w.w) <= 0 || Number(w.h) <= 0) return false;
    if (w.visible !== false) visible = true;
    if (w.widgetType === 'tabcontainer' &&
        (!Array.isArray(w.tabs) || w.tabs.length === 0 ||
          w.tabs.some((tab) => !tab || !KNOWN_WIDGETS.has(tab.widgetType) ||
            tab.widgetType === 'tabcontainer' || typeof tab.id !== 'string' || !tab.id))) return false;
  }
  return visible;
}

/** @typedef {{ id: string, name: string, savedAt: string, widgets: object[], shareId?: string|null }} LayoutTemplate */

export function makeTemplateId() {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID();
  }
  return `tpl-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

export function normalizeTemplateName(name) {
  return String(name || '').trim().slice(0, 64);
}

/** Stable compare of widget layout snapshots (positions / types / tabs). */
export function serializeWidgets(widgets) {
  if (!Array.isArray(widgets)) return '[]';
  const normalized = widgets.map((w) => {
    const base = {
      id: w.id,
      widgetType: w.widgetType,
      x: w.x,
      y: w.y,
      w: w.w,
      h: w.h,
      visible: w.visible !== false,
    };
    if (w.widgetType === 'tabcontainer') {
      base.tabs = Array.isArray(w.tabs) ? w.tabs : [];
      base.activeTabIndex = w.activeTabIndex || 0;
    }
    return base;
  });
  return JSON.stringify(normalized);
}

export function widgetsEqual(a, b) {
  return serializeWidgets(a) === serializeWidgets(b);
}

/**
 * Normalize templates array from storage.
 * @returns {LayoutTemplate[]}
 */
export function normalizeTemplates(raw) {
  if (!Array.isArray(raw)) return [];
  const out = [];
  const seen = new Set();
  for (const t of raw) {
    if (!t || typeof t !== 'object') continue;
    const id = String(t.id || '').trim() || makeTemplateId();
    if (seen.has(id)) continue;
    const name = normalizeTemplateName(t.name) || 'Untitled';
    if (!Array.isArray(t.widgets)) continue;
    seen.add(id);
    const entry = {
      id,
      name,
      savedAt: t.savedAt || new Date().toISOString(),
      widgets: t.widgets,
    };
    // Reserved for future share feature — never required in v1.
    if (t.shareId != null && t.shareId !== '') {
      entry.shareId = String(t.shareId);
    }
    out.push(entry);
  }
  return out;
}

/**
 * Parse storage JSON → { widgets, templates, activeTemplateId } or null.
 */
export function parseLayoutStorage(raw) {
  if (!raw) return null;
  let data;
  try {
    data = typeof raw === 'string' ? JSON.parse(raw) : raw;
  } catch {
    return null;
  }
  if (!data || !isUsableLayout(data.widgets)) return null;
  if (data.version !== 1 && data.version !== 2) return null;
  return {
    version: LAYOUT_STORAGE_VERSION,
    widgets: data.widgets,
    templates: normalizeTemplates(data.templates),
    activeTemplateId: data.activeTemplateId ? String(data.activeTemplateId) : null,
    savedAt: data.savedAt || null,
  };
}

export function buildLayoutStoragePayload({ widgets, templates, activeTemplateId }) {
  return {
    version: LAYOUT_STORAGE_VERSION,
    savedAt: new Date().toISOString(),
    widgets,
    templates: normalizeTemplates(templates),
    activeTemplateId: activeTemplateId || null,
  };
}
