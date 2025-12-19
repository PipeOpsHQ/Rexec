export type XtermCoreModules = {
  Terminal: typeof import("@xterm/xterm").Terminal;
  FitAddon: typeof import("@xterm/addon-fit").FitAddon;
  Unicode11Addon: typeof import("@xterm/addon-unicode11").Unicode11Addon;
  WebLinksAddon: typeof import("@xterm/addon-web-links").WebLinksAddon;
};

export type XtermWebglModule = {
  WebglAddon: typeof import("@xterm/addon-webgl").WebglAddon;
};

let xtermCorePromise: Promise<XtermCoreModules> | null = null;
let xtermWebglPromise: Promise<XtermWebglModule> | null = null;

// Track if we're on a slow connection
function isSlowConnection(): boolean {
  const nav = navigator as Navigator & {
    connection?: {
      effectiveType?: string;
      saveData?: boolean;
      downlink?: number;
    };
  };
  if (nav.connection) {
    // Slow if 2G, slow-2g, or saveData enabled, or downlink < 1.5 Mbps
    const type = nav.connection.effectiveType;
    if (type === "slow-2g" || type === "2g" || nav.connection.saveData) {
      return true;
    }
    if (nav.connection.downlink && nav.connection.downlink < 1.5) {
      return true;
    }
  }
  return false;
}

// Check if running on mobile
function isMobileDevice(): boolean {
  return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
    navigator.userAgent,
  );
}

export async function loadXtermCore(): Promise<XtermCoreModules> {
  if (!xtermCorePromise) {
    xtermCorePromise = (async () => {
      // Load CSS and all JS modules in parallel for faster loading
      const [, xterm, fit, unicode11, webLinks] = await Promise.all([
        import("@xterm/xterm/css/xterm.css"),
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

  // Also preload WebGL addon (skip on slow connections to save bandwidth)
  if (!isSlowConnection()) {
    loadXtermWebgl().catch(() => {
      // WebGL is optional, ignore errors
    });
  }
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
  const mobile = isMobileDevice();
  const slow = isSlowConnection();

  // On slow mobile connections, use shorter timeouts but more retries
  const getDelay = () => {
    if (slow) {
      // Faster retries on slow connections: 300ms, 600ms, 1200ms
      return 300 * Math.pow(2, attempts - 1);
    }
    // Normal: 500ms, 1000ms, 2000ms
    return 500 * Math.pow(2, attempts - 1);
  };

  const tryLoad = (): Promise<void> => {
    attempts++;
    return loadXtermCore()
      .then(() => {
        // Success - also try webgl (skip on slow connections)
        if (!slow) {
          loadXtermWebgl().catch(() => {});
        }
      })
      .catch((err) => {
        // More retries for mobile on slow networks
        const effectiveMaxRetries =
          mobile && slow ? maxRetries + 2 : maxRetries;
        if (attempts < effectiveMaxRetries) {
          const delay = getDelay();
          return new Promise((resolve) => setTimeout(resolve, delay)).then(
            tryLoad,
          );
        }
        // Final failure - don't throw, just log
        console.warn(
          "[xterm] Failed to preload after",
          attempts,
          "attempts:",
          err,
        );
      });
  };

  return tryLoad();
}

/**
 * Get loading status info for UI feedback
 */
export function getLoadingInfo(): {
  isMobile: boolean;
  isSlowConnection: boolean;
  isLoaded: boolean;
} {
  return {
    isMobile: isMobileDevice(),
    isSlowConnection: isSlowConnection(),
    isLoaded: isXtermLoaded(),
  };
}
