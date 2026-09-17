import { writable } from 'svelte/store';

const MAX_MESSAGES = 4;

let messageId = 0;

function createOverlayStore() {
  const { subscribe, update } = writable([]);

  return {
    subscribe,

    /**
     * Push a room-hero toast.
     * @param {string|{text:string, kind?:'ambiance'|'chat'}} payload
     *   string → kind 'chat' (default alert chrome)
     *   { text, kind: 'ambiance' } → mood toast (no System: prefix; soft style)
     */
    pushMessage(payload) {
      let text;
      let kind = 'chat';

      if (payload && typeof payload === 'object') {
        text = payload.text;
        if (payload.kind === 'ambiance' || payload.kind === 'chat') {
          kind = payload.kind;
        }
      } else {
        text = payload;
      }

      if (!text || String(text).trim() === '') return;

      const id = ++messageId;
      const cleanText = String(text).trim();

      // Give longer reactions enough on-screen time to be read (not a blink).
      // Ambiance mood lines get a touch more dwell so flavor sinks in.
      const base = kind === 'ambiance' ? 3200 : 2800;
      const displayDuration = Math.min(
        base + Math.floor(cleanText.length / 40) * 700,
        9000
      );
      const fadeOutDuration = Math.min(
        900 + Math.floor(cleanText.length / 50) * 250,
        2200
      );

      update(messages => {
        const updated = [...messages, {
          id,
          text: cleanText,
          kind,
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
