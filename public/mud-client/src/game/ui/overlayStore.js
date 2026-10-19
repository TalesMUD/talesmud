import { writable } from 'svelte/store';
import { parseExamineText } from './parseExamineOverlay.js';

let messageId = 0;

function createOverlayStore() {
  const { subscribe, update } = writable([]);

  return {
    subscribe,

    /**
     * Push a room-hero toast.
     * @param {string|{text:string, kind?:'ambiance'|'chat'|'examine', examine?:object}} payload
     *   string → auto-detect examine dumps; else kind 'chat'
     *   { text, kind: 'ambiance' } → mood toast (soft style)
     *   examine dumps → kind 'examine' (item card chrome)
     *   { examine } → prefer structured examineResult card over prose parse
     */
    pushMessage(payload) {
      let text;
      let kind = 'chat';
      let examineFromPayload = null;

      if (payload && typeof payload === 'object') {
        text = payload.text;
        if (payload.kind === 'ambiance' || payload.kind === 'chat' || payload.kind === 'examine') {
          kind = payload.kind;
        }
        if (payload.examine && typeof payload.examine === 'object') {
          examineFromPayload = payload.examine;
        }
      } else {
        text = payload;
      }

      // Structured examineResult may arrive with only a card model (no prose blob).
      if ((!text || String(text).trim() === '') && examineFromPayload) {
        text = [
          examineFromPayload.title,
          examineFromPayload.blurb,
          examineFromPayload.lore,
          ...(examineFromPayload.effects || []),
        ].filter(Boolean).join('\n');
      }

      if (!text || String(text).trim() === '') return;

      const id = ++messageId;
      const cleanText = String(text).trim();
      // Prefer a structured examineResult card; fall back to prose parse.
      const examine = examineFromPayload || parseExamineText(cleanText);
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
      update(messages => {
        // Replace any existing examine card so only one item card shows.
        let next = messages;
        if (kind === 'examine') {
          next = messages.filter(m => m.kind !== 'examine');
        }

        const message = { id, text: cleanText, kind, examine: examine || null, displayDuration };
        const updated = kind === 'examine' ? [message, ...next] : [...next, message];
        return updated;
      });

      return id;
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
