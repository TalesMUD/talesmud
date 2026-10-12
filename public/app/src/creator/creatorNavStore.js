import { writable } from "svelte/store";

// Mobile drawer only. The icon-rail collapse lives in localStorage via navState.js.
export const creatorDrawerOpen = writable(false);

// True below the desktop rail breakpoint (1024px). App.svelte owns the listener.
export const creatorNavNarrow = writable(false);
