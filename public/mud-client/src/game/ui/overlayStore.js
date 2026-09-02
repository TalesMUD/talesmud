import { writable } from 'svelte/store';

const MAX_MESSAGES = 4;

let messageId = 0;

function createOverlayStore() {
  const { subscribe, update } = writable([]);

  return {
    subscribe,

    pushMessage(text) {
      if (!text || text.trim() === '') return;

      const id = ++messageId;
      const cleanText = text.trim();

      // Give longer reactions enough on-screen time to be read (not a blink).
      const displayDuration = Math.min(
        2800 + Math.floor(cleanText.length / 40) * 700,
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
