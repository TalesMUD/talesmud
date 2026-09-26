import { writable } from "svelte/store";

export const characterPickerOpen = writable(false);

export function openCharacterPicker() {
  characterPickerOpen.set(true);
}

export function closeCharacterPicker() {
  characterPickerOpen.set(false);
}
