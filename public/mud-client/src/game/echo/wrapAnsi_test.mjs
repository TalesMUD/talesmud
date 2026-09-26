import assert from 'assert';
import { visibleWidth, wrapAnsi } from './wrapAnsi.js';

const CHAMBER = 'A circular chamber of ancient dark stone, impossibly old and heavy with silence. Pale sourceless light drifts down from above, illuminating a stone altar at the center where you lie. Symbols carved into every surface seem to writhe at the edge of vision, and the air tastes of dust and forgotten centuries.';

function words(text) {
  return String(text).replace(/\u001b\[[0-?]*[ -/]*[@-~]/g, '').split(/\s+/).filter(Boolean);
}

function assertFits(text, cols) {
  for (const line of wrapAnsi(text, cols).split('\n')) {
    assert.ok(visibleWidth(line) <= cols, `line wider than ${cols}: ${JSON.stringify(line)}`);
  }
}

{
  const wrapped = wrapAnsi('Pale sourceless light drifts', 16);
  assert.deepEqual(words(wrapped), ['Pale', 'sourceless', 'light', 'drifts']);
  assert.ok(!wrapped.split('\n').some((line) => line === 'lig' || line.endsWith('lig')));
  assertFits('Pale sourceless light drifts', 16);
}

for (const cols of [20, 28, 41, 48, 63]) {
  const wrapped = wrapAnsi(CHAMBER, cols);
  assert.deepEqual(words(wrapped), words(CHAMBER));
  assertFits(CHAMBER, cols);
  assert.ok(wrapped.split('\n').some((line) => line.includes('light')));
  assert.ok(wrapped.split('\n').some((line) => line.includes('centuries')));
}

{
  const wrapped = wrapAnsi('short\n\n- You can:', 40);
  assert.equal(wrapped, 'short\n\n- You can:');
}

{
  const wrapped = wrapAnsi('abcdefghijklmnopqrstuvwxyz', 10);
  assert.deepEqual(wrapped.split('\n'), ['abcdefghij', 'klmnopqrst', 'uvwxyz']);
  assertFits('abcdefghijklmnopqrstuvwxyz', 10);
}

{
  const src = '\u001b[31mhello world\u001b[0m';
  const wrapped = wrapAnsi(src, 8);
  assert.ok(wrapped.includes('\u001b[31m'));
  assert.ok(wrapped.includes('\u001b[0m'));
  assert.deepEqual(words(wrapped), ['hello', 'world']);
  assertFits(src, 8);
  assert.equal(visibleWidth('\u001b[31mhello\u001b[0m'), 5);
}

{
  assert.equal(wrapAnsi('', 10), '');
  assert.equal(wrapAnsi('hi', 0), 'h\ni');
}

console.log('wrapAnsi_test ok');
