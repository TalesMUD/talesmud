import { writable } from 'svelte/store';

export const MOBILE_QUERY = '(max-width: 768px), (max-width: 1024px) and (max-height: 500px) and (orientation: landscape)';

function createMobileStore() {
  const isMobile = writable(false);
  const activeTab = writable('room'); // 'room' | 'console'
  const openSheet = writable(null);   // null | 'character' | 'inventory' | 'equipment'

  if (typeof window !== 'undefined') {
    const mql = window.matchMedia(MOBILE_QUERY);
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
