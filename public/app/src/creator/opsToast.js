import { writable } from "svelte/store";

export const opsToast = writable(null);

export function showOpsToast(toast) {
  opsToast.set({ ...toast, id: Date.now() });
}

export function clearOpsToast() {
  opsToast.set(null);
}
