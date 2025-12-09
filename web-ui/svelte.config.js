import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

export default {
  preprocess: vitePreprocess(),
  compilerOptions: {
    // Use legacy mode to support Svelte 4 syntax (on:event, createEventDispatcher)
    // This allows gradual migration to Svelte 5
    compatibility: {
      componentApi: 4,
    },
  },
};
