export type XtermCoreModules = {
  Terminal: typeof import("@xterm/xterm").Terminal;
  FitAddon: typeof import("@xterm/addon-fit").FitAddon;
  Unicode11Addon: typeof import("@xterm/addon-unicode11").Unicode11Addon;
  WebLinksAddon: typeof import("@xterm/addon-web-links").WebLinksAddon;
};

export type XtermWebglModule = {
  WebglAddon: typeof import("@xterm/addon-webgl").WebglAddon;
};

let xtermCssPromise: Promise<unknown> | null = null;
let xtermCorePromise: Promise<XtermCoreModules> | null = null;
let xtermWebglPromise: Promise<XtermWebglModule> | null = null;

export async function loadXtermCore(): Promise<XtermCoreModules> {
  if (!xtermCorePromise) {
    xtermCorePromise = (async () => {
      if (!xtermCssPromise) {
        xtermCssPromise = import("@xterm/xterm/css/xterm.css");
      }
      await xtermCssPromise;

      const [xterm, fit, unicode11, webLinks] = await Promise.all([
        import("@xterm/xterm"),
        import("@xterm/addon-fit"),
        import("@xterm/addon-unicode11"),
        import("@xterm/addon-web-links"),
      ]);

      return {
        Terminal: xterm.Terminal,
        FitAddon: fit.FitAddon,
        Unicode11Addon: unicode11.Unicode11Addon,
        WebLinksAddon: webLinks.WebLinksAddon,
      };
    })().catch((err) => {
      // Allow retry on transient failures
      xtermCorePromise = null;
      throw err;
    });
  }

  return xtermCorePromise;
}

export async function loadXtermWebgl(): Promise<XtermWebglModule> {
  if (!xtermWebglPromise) {
    xtermWebglPromise = import("@xterm/addon-webgl")
      .then((mod) => ({ WebglAddon: mod.WebglAddon }))
      .catch((err) => {
        xtermWebglPromise = null;
        throw err;
      });
  }

  return xtermWebglPromise;
}

/**
 * Preload xterm modules eagerly without waiting.
 * Call this when user initiates container creation to reduce
 * perceived latency when terminal becomes ready.
 */
export function preloadXterm(): void {
  // Start loading core modules immediately (fire-and-forget)
  loadXtermCore().catch(() => {
    // Retry once on failure (common on slow mobile networks)
    setTimeout(() => {
      loadXtermCore().catch(() => {
        // Second failure - will be handled when actually needed
      });
    }, 1000);
  });

  // Also preload WebGL addon
  loadXtermWebgl().catch(() => {
    // WebGL is optional, ignore errors
  });
}

/**
 * Check if xterm core modules are already loaded/cached.
 * Useful for showing loading indicators on slow connections.
 */
export function isXtermLoaded(): boolean {
  return xtermCorePromise !== null;
}

/**
 * Preload xterm with retry logic for mobile networks.
 * More aggressive retry for unreliable connections.
 */
export function preloadXtermWithRetry(maxRetries = 3): Promise<void> {
  let attempts = 0;

  const tryLoad = (): Promise<void> => {
    attempts++;
    return loadXtermCore()
      .then(() => {
        // Success - also try webgl
        loadXtermWebgl().catch(() => {});
      })
      .catch((err) => {
        if (attempts < maxRetries) {
          // Exponential backoff: 500ms, 1000ms, 2000ms
          const delay = 500 * Math.pow(2, attempts - 1);
          return new Promise((resolve) => setTimeout(resolve, delay)).then(
            tryLoad,
          );
        }
        // Final failure - don't throw, just log
        console.warn(
          "[xterm] Failed to preload after",
          maxRetries,
          "attempts:",
          err,
        );
      });
  };

  return tryLoad();
}
