import assert from 'assert';
import { readStageSize, shouldRepaintSize, applyCanvasBitmap, applyInFlowStyleSize } from './atlasLayout.js';

function simulateStableFrames() {
  const wrap = { clientWidth: 240, clientHeight: 180, y: 64 };
  const canvas = { width: 0, height: 0 };
  const frames = [];
  for (let i = 0; i < 2; i++) {
    const size = readStageSize(wrap);
    applyCanvasBitmap(canvas, size.w, size.h, 1);
    frames.push({
      y: wrap.y,
      wrapH: wrap.clientHeight,
      canvasH: canvas.height,
      empty: canvas.height < 1,
    });
  }
  return frames;
}

function simulateOldStyleLoop() {
  const wrap = { clientWidth: 240, clientHeight: 180, y: 64 };
  const frames = [];
  for (let i = 0; i < 2; i++) {
    applyInFlowStyleSize(wrap, wrap.clientHeight);
    frames.push({ y: wrap.y, wrapH: wrap.clientHeight });
  }
  return frames;
}

const stable = simulateStableFrames();
assert.strictEqual(stable.length, 2);
assert.strictEqual(stable[0].y, stable[1].y, 'widget Y must not move between frames');
assert.strictEqual(stable[0].wrapH, stable[1].wrapH, 'wrap height must not change between frames');
assert.strictEqual(stable[0].empty, false, 'canvas must not be empty after first fit');
assert.strictEqual(stable[1].empty, false, 'second frame must not empty the canvas');
assert.strictEqual(stable[0].canvasH, 180);
assert.strictEqual(stable[1].canvasH, 180);

assert.strictEqual(shouldRepaintSize({ w: 240, h: 180 }, { w: 240, h: 180 }), false);
assert.strictEqual(shouldRepaintSize({ w: 240, h: 180 }, { w: 240, h: 181 }), true);
assert.deepStrictEqual(readStageSize({ clientWidth: 240.9, clientHeight: 180.2 }), { w: 240, h: 180 });

const canvas = { width: 0, height: 0 };
applyCanvasBitmap(canvas, 240, 180, 1);
assert.strictEqual(canvas.style, undefined, 'bitmap apply must not write canvas.style');

const old = simulateOldStyleLoop();
assert.notStrictEqual(old[0].y, old[1].y, 'control: in-flow style size would slide widget Y');

console.log('atlasLayout: two frames keep widget Y and canvas height');
