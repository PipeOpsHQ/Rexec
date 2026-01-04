import { defineConfig } from "vite";
import { resolve } from "path";

export default defineConfig({
  build: {
    lib: {
      entry: resolve(__dirname, "src/index.ts"),
      name: "Rexec",
      formats: ["umd", "es"],
      fileName: (format) => {
        if (format === "umd") return "rexec.min.js";
        if (format === "es") return "rexec.esm.js";
        return `rexec.${format}.js`;
      },
    },
    outDir: "dist",
    emptyOutDir: true,
    minify: "esbuild",
    rollupOptions: {
      output: {
        // Use named exports to avoid the default export warning
        exports: "named",
        // Ensure CSS is inlined
        inlineDynamicImports: true,
        // Global variable name for UMD build
        name: "Rexec",
        // Ensure all assets are bundled
        assetFileNames: "rexec.[ext]",
      },
    },
    // Inline all CSS into JS
    cssCodeSplit: false,
    // Target modern browsers but keep compatibility
    target: "es2018",
    // Generate source maps for debugging
    sourcemap: true,
  },
  resolve: {
    alias: {
      "@": resolve(__dirname, "src"),
    },
  },
  define: {
    // Prevent issues with process.env in browser
    "process.env.NODE_ENV": JSON.stringify("production"),
  },
});
