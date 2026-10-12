import assert from 'assert';
import { formatOverlayText } from './formatOverlayText.js';

const look = formatOverlayText('You look at Wolf.\n[HOSTILE]\nAggressive');
assert.match(look, /<span class="aggro-mark">Aggressive<\/span>$/);
assert.doesNotMatch(look, /<span class="aggro-mark">\[HOSTILE\]/);

const plain = formatOverlayText('A wolf is aggressive in the brush.');
assert.doesNotMatch(plain, /aggro-mark/);

const bold = formatOverlayText('**Hello** & <x>');
assert.equal(bold, '<strong>Hello</strong> &amp; &lt;x&gt;');

console.log('formatOverlayText_test: ok');
