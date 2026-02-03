import { writable } from 'svelte/store';

const MOBILE_BREAKPOINT = 768;

function createMobileStore() {
  const isMobile = writable(false);
  const activeTab = writable('room'); // 'room' | 'console'
  const openSheet = writable(null);   // null | 'character' | 'inventory' | 'equipment'

  if (typeof window !== 'undefined') {
    const mql = window.matchMedia(`(max-width: ${MOBILE_BREAKPOINT}px)`);
    isMobile.set(mql.matches);
    mql.addEventListener('change', (e) => {
      isMobile.set(e.matches);
    });
  }

  return {
    isMobile,
    activeTab,
    openSheet,

    switchTab(tab) {
      activeTab.set(tab);
    },

    openBottomSheet(sheet) {
      openSheet.set(sheet);
    },

    closeBottomSheet() {
      openSheet.set(null);
    }
  };
}

export const mobileStore = createMobileStore();
