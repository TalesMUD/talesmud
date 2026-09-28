import { writable } from 'svelte/store';

export const accountMenuOpen = writable(false);
export const battleDockOpen = writable(false);
export const cheatSheetOpen = writable(false);
export const layoutDialogOpen = writable(false);

let closeLayoutDialog = () => {};

export function setLayoutDialogCloser(fn) {
  closeLayoutDialog = typeof fn === 'function' ? fn : () => {};
}

export function requestCloseLayoutDialog() {
  closeLayoutDialog();
  layoutDialogOpen.set(false);
}
