import { mount } from "svelte";
import posthog from "posthog-js";
import App from "./App.svelte";
import "./styles/global.css";
import { registerSW } from "virtual:pwa-register";
import { initInstallPrompt } from "./lib/stores/pwa";

// Initialize PostHog analytics
posthog.init("phc_bXCnSXoq52IpajVPLj2ddwOKhyOUhxmqw984yuc8XfR", {
  api_host: "https://eu.i.posthog.com",
  person_profiles: "identified_only",
  capture_pageview: true,
  capture_pageleave: true,
});

const app = mount(App, {
  target: document.getElementById("app")!,
});

// Initialize PWA install prompt handling
initInstallPrompt();

// Listen for service worker update messages (important for Safari)
if ("serviceWorker" in navigator) {
  navigator.serviceWorker.addEventListener("message", (event) => {
    if (event.data?.type === "SW_UPDATED") {
      console.log("[App] Service worker updated to:", event.data.version);
      // Auto-reload to get fresh content (silent for Safari)
      window.location.reload();
    }
  });
}

// Register service worker with auto-update
const updateSW = registerSW({
  onNeedRefresh() {
    // Show a prompt to update when new content is available
    if (confirm("New content available. Reload to update?")) {
      updateSW(true);
    }
  },
  onOfflineReady() {
    console.log("App ready to work offline");
  },
  onRegistered(r) {
    // Check for updates every hour
    if (r) {
      setInterval(
        () => {
          r.update();
        },
        60 * 60 * 1000,
      );
    }
  },
  onRegisterError(error) {
    console.error("SW registration error:", error);
  },
});

export default app;
