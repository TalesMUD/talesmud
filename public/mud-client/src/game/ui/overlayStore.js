import { writable } from 'svelte/store';
import { parseExamineText } from './parseExamineOverlay.js';

const MAX_MESSAGES = 4;

let messageId = 0;

function createOverlayStore() {
  const { subscribe, update } = writable([]);

  return {
    subscribe,

    /**
     * Push a room-hero toast.
     * @param {string|{text:string, kind?:'ambiance'|'chat'|'examine'}} payload
     *   string → auto-detect examine dumps; else kind 'chat'
     *   { text, kind: 'ambiance' } → mood toast (soft style)
     *   examine dumps → kind 'examine' (item card chrome)
     */
    pushMessage(payload) {
      let text;
      let kind = 'chat';

      if (payload && typeof payload === 'object') {
        text = payload.text;
        if (payload.kind === 'ambiance' || payload.kind === 'chat' || payload.kind === 'examine') {
          kind = payload.kind;
        }
      } else {
        text = payload;
      }

      if (!text || String(text).trim() === '') return;

      const id = ++messageId;
      const cleanText = String(text).trim();
      const examine = parseExamineText(cleanText);
      if (examine && kind !== 'ambiance') {
        kind = 'examine';
      }

      // Examine cards linger so lore can be read; ambiance a touch longer than chat.
      let base = 2800;
      if (kind === 'ambiance') base = 3200;
      if (kind === 'examine') base = 8000;
      const displayDuration = Math.min(
        base + Math.floor(cleanText.length / 40) * (kind === 'examine' ? 900 : 700),
        kind === 'examine' ? 22000 : 9000
      );
      const fadeOutDuration = Math.min(
        900 + Math.floor(cleanText.length / 50) * 250,
        2200
      );

      update(messages => {
        // Replace any existing examine card so only one item card shows.
        let next = messages;
        if (kind === 'examine') {
          next = messages.filter(m => m.kind !== 'examine');
        }
        const updated = [...next, {
          id,
          text: cleanText,
          kind,
          examine: examine || null,
          displayDuration,
          fadeOutDuration,
          fading: false
        }];
        if (updated.length > MAX_MESSAGES) {
          updated.splice(0, updated.length - MAX_MESSAGES);
        }
        return updated;
      });

      return id;
    },

    startFade(id) {
      update(messages =>
        messages.map(m => m.id === id ? { ...m, fading: true } : m)
      );
    },

    removeMessage(id) {
      update(messages => messages.filter(m => m.id !== id));
    },

    clearAll() {
      update(() => []);
    }
  };
}

export const overlayStore = createOverlayStore();
