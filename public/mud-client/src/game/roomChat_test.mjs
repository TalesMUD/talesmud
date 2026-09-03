import assert from 'assert';
import {
  isSelfName,
  markPlayersYou,
  parseRoomChatLine,
} from './roomChat.js';

const you = { id: 'char-gimli', name: 'Gimli' };

assert.ok(isSelfName('Gimli', you));
assert.ok(isSelfName('gimli', you));
assert.ok(!isSelfName('Ranger', you));

const marked = markPlayersYou(
  [
    { id: 'char-gimli', name: 'Gimli', isYou: false },
    { id: 'char-ranger', name: 'Ranger', isYou: false },
  ],
  you
);
assert.strictEqual(marked[0].isYou, true, 'roomPresence marks self locally');
assert.strictEqual(marked[1].isYou, false);

const implicit = parseRoomChatLine(
  { type: 'message', username: 'Ranger', message: 'Careful, boars ahead.' },
  you
);
assert.deepStrictEqual(implicit, {
  name: 'Ranger',
  text: 'Careful, boars ahead.',
  isYou: false,
});

const selfSay = parseRoomChatLine(
  { type: 'message', message: 'Gimli says: Hold the line.' },
  you
);
assert.deepStrictEqual(selfSay, {
  name: 'Gimli',
  text: 'Hold the line.',
  isYou: true,
});

const look = parseRoomChatLine(
  { type: 'message', message: 'You look at the shelves.' },
  you
);
assert.strictEqual(look, null, 'look/system text is not room chat');

const empty = parseRoomChatLine({ type: 'message', message: '' }, you);
assert.strictEqual(empty, null);

console.log('roomChat_test.mjs ok');
