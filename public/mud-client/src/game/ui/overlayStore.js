import { writable } from 'svelte/store';

const MAX_MESSAGES = 5;

let messageId = 0;

function createOverlayStore() {
  const { subscribe, update } = writable([]);

  return {
    subscribe,

    pushMessage(text) {
      if (!text || text.trim() === '') return;

      const id = ++messageId;
      const cleanText = text.trim();

      // Scale display duration by text length: 2s base + 0.5s per 50 chars, max 5s
      const displayDuration = Math.min(
        2000 + Math.floor(cleanText.length / 50) * 500,
        5000
      );
      // Scale fade duration by text length: 0.8s base + 0.2s per 50 chars, max 2s
      const fadeOutDuration = Math.min(
        800 + Math.floor(cleanText.length / 50) * 200,
        2000
      );

      update(messages => {
        const updated = [...messages, {
          id,
          text: cleanText,
          displayDuration,
          fadeOutDuration,
          fading: false
        }];
        // Drop oldest if over max
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
