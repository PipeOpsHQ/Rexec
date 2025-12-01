import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

export default {
  preprocess: vitePreprocess(),
  compilerOptions: {
    // Enable runes mode for Svelte 5 (optional, for future)
    // runes: true,
  },
};
