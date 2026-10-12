import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";

// vite-plugin-svelte sets resolve.conditions to ["svelte"]. Vite 6 replaces
// the default conditions with that list, so the client build resolves the
// svelte package to its SSR entry. onMount is a no-op there and Rollup then
// drops the callbacks. Put browser back after the Svelte plugin.
function svelteClientConditions() {
  return {
    name: "svelte-client-conditions",
    config() {
      return {
        resolve: {
          conditions: ["module", "browser", "svelte", "development|production"],
        },
      };
    },
  };
}

export default defineConfig({
  plugins: [svelte(), svelteClientConditions()],
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8010",
        changeOrigin: true,
      },
      "/ws": {
        target: "ws://localhost:8010",
        ws: true,
      },
      "/admin": {
        target: "http://localhost:8010",
        changeOrigin: true,
      },
    },
  },
});
