import { defineConfig } from "vite";
import { resolve } from "path";

// JSON.stringify leaves <, >, / and line separators raw. Those characters can
// break out of a script when the string is concatenated into JavaScript.
const jsStringEscapes: Record<string, string> = {
  "<": "\\u003C",
  ">": "\\u003E",
  "/": "\\u002F",
  "\b": "\\b",
  "\f": "\\f",
  "\n": "\\n",
  "\r": "\\r",
  "\t": "\\t",
  "\0": "\\0",
  "\u2028": "\\u2028",
  "\u2029": "\\u2029",
};

function escapeUnsafeChars(value: string): string {
  return value.replace(
    /[<>/\b\f\n\r\t\0\u2028\u2029]/g,
    (ch) => jsStringEscapes[ch] ?? ch,
  );
}

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
    minify: "terser",
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
    // Inline all CSS into JS - cssCodeSplit false alone doesn't work for libraries
    cssCodeSplit: false,
    // Target modern browsers but keep compatibility
    target: "es2018",
    // Generate source maps for debugging
    sourcemap: true,
  },
  // Force CSS to be injected into JS at runtime
  css: {
    // Inject CSS into the document head at runtime
    devSourcemap: true,
  },
  plugins: [
    // Custom plugin to inject CSS into JS bundle
    {
      name: "css-inject",
      apply: "build",
      enforce: "post",
      generateBundle(options, bundle) {
        // Find the CSS file and JS files
        let cssContent = "";
        const cssFiles: string[] = [];

        for (const [fileName, chunk] of Object.entries(bundle)) {
          if (fileName.endsWith(".css") && chunk.type === "asset") {
            cssContent = chunk.source as string;
            cssFiles.push(fileName);
          }
        }

        if (cssContent) {
          // Remove CSS files from bundle
          for (const cssFile of cssFiles) {
            delete bundle[cssFile];
          }

          // Inject CSS into JS files
          const cssInjector = `(function(){if(typeof document!=="undefined"){var s=document.createElement("style");s.textContent=${escapeUnsafeChars(JSON.stringify(cssContent))};document.head.appendChild(s);}})();`;

          for (const [fileName, chunk] of Object.entries(bundle)) {
            if (chunk.type === "chunk" && fileName.endsWith(".js")) {
              chunk.code = cssInjector + chunk.code;
            }
          }
        }
      },
    },
  ],
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
