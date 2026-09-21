import assert from 'assert';
import {
  parseLayoutStorage,
  buildLayoutStoragePayload,
  normalizeTemplates,
  normalizeTemplateName,
  widgetsEqual,
  serializeWidgets,
  LAYOUT_STORAGE_VERSION,
} from './layoutTemplates.js';

// v1 migrate: no templates field
{
  const parsed = parseLayoutStorage(
    JSON.stringify({
      version: 1,
      widgets: [{ id: 'a', widgetType: 'room', x: 0, y: 0, w: 12, h: 12, visible: true }],
    })
  );
  assert.ok(parsed);
  assert.equal(parsed.version, LAYOUT_STORAGE_VERSION);
  assert.equal(parsed.templates.length, 0);
  assert.equal(parsed.activeTemplateId, null);
}

// v2 round-trip
{
  const widgets = [
    { id: 'room-1', widgetType: 'room', x: 0, y: 0, w: 12, h: 10, visible: true },
  ];
  const templates = normalizeTemplates([
    { id: 't1', name: ' Raid ', widgets, savedAt: '2026-01-01T00:00:00.000Z' },
  ]);
  assert.equal(templates[0].name, 'Raid');
  const payload = buildLayoutStoragePayload({
    widgets,
    templates,
    activeTemplateId: 't1',
  });
  assert.equal(payload.version, 2);
  assert.equal(payload.activeTemplateId, 't1');
  assert.equal(payload.templates[0].name, 'Raid');
  // shareId reserved but not required
  assert.equal(payload.templates[0].shareId, undefined);

  const again = parseLayoutStorage(JSON.stringify(payload));
  assert.equal(again.templates.length, 1);
  assert.ok(widgetsEqual(again.widgets, widgets));
}

// dirty detection
{
  const a = [{ id: '1', widgetType: 'room', x: 0, y: 0, w: 12, h: 12, visible: true }];
  const b = [{ id: '1', widgetType: 'room', x: 1, y: 0, w: 12, h: 12, visible: true }];
  assert.ok(widgetsEqual(a, a));
  assert.ok(!widgetsEqual(a, b));
  assert.ok(serializeWidgets(a).includes('room'));
}

assert.equal(normalizeTemplateName('  x  '), 'x');
assert.equal(normalizeTemplateName(''), '');

console.log('layoutTemplates_test: ok');
