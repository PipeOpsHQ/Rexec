/**
 * Rexec Embed Widget
 * Embeddable terminal widget for Rexec - Cloud Shell anywhere
 *
 * Usage:
 * ```html
 * <script src="https://rexec.dev/embed/rexec.min.js"></script>
 * <div id="terminal"></div>
 * <script>
 *   const term = Rexec.embed('#terminal', {
 *     token: 'your-api-token',
 *     container: 'container-id',
 *     // or
 *     shareCode: 'abc123',
 *     // or
 *     role: 'ubuntu',
 *   });
 * </script>
 * ```
 */

import { RexecTerminal } from './terminal';
import { RexecApiClient, generateSessionId } from './api';
import { DARK_THEME, LIGHT_THEME, getTheme } from './themes';

// Re-export types
export type {
  RexecEmbedConfig,
  RexecTerminalInstance,
  ConnectionState,
  SessionInfo,
  ContainerStats,
  RexecError,
  RexecEventMap,
  TerminalTheme,
  WsMessage,
  CreateContainerResponse,
  JoinSessionResponse,
  ContainerInfoResponse,
} from './types';

// Re-export classes
export { RexecTerminal } from './terminal';
export { RexecApiClient, TerminalWebSocket, generateSessionId } from './api';
export { DARK_THEME, LIGHT_THEME, getTheme } from './themes';

/**
 * Version of the embed library
 */
export const VERSION = '1.0.0';

/**
 * Active terminal instances for management
 */
const instances: Map<string, RexecTerminal> = new Map();

/**
 * Generate a unique instance ID
 */
function generateInstanceId(): string {
  return `rexec-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

/**
 * Create and embed a terminal in the specified element
 *
 * @param element - CSS selector or HTMLElement to embed the terminal in
 * @param config - Configuration options for the terminal
 * @returns The RexecTerminal instance
 *
 * @example
 * ```js
 * // Connect to existing container
 * const term = Rexec.embed('#terminal', {
 *   token: 'your-api-token',
 *   container: 'container-id'
 * });
 *
 * // Join via share code
 * const term = Rexec.embed('#terminal', {
 *   shareCode: 'abc123'
 * });
 *
 * // Create new container
 * const term = Rexec.embed('#terminal', {
 *   token: 'your-api-token',
 *   role: 'ubuntu'
 * });
 * ```
 */
export function embed(
  element: HTMLElement | string,
  config: import('./types').RexecEmbedConfig = {}
): RexecTerminal {
  const terminal = new RexecTerminal(element, config);
  const instanceId = generateInstanceId();
  instances.set(instanceId, terminal);
  return terminal;
}

/**
 * Get all active terminal instances
 */
export function getInstances(): RexecTerminal[] {
  return Array.from(instances.values());
}

/**
 * Destroy all active terminal instances
 */
export function destroyAll(): void {
  instances.forEach((terminal) => terminal.destroy());
  instances.clear();
}

/**
 * Create an API client for direct API access
 *
 * @param baseUrl - Base URL for the Rexec API (default: https://rexec.dev)
 * @param token - API token for authentication
 * @returns RexecApiClient instance
 */
export function createClient(baseUrl?: string, token?: string): RexecApiClient {
  return new RexecApiClient(baseUrl, token);
}

/**
 * Available themes
 */
export const themes = {
  dark: DARK_THEME,
  light: LIGHT_THEME,
  get: getTheme,
};

/**
 * Utility functions
 */
export const utils = {
  generateSessionId,
};

// Create the global Rexec object for CDN usage
const Rexec = {
  // Main API
  embed,
  createClient,

  // Instance management
  getInstances,
  destroyAll,

  // Themes
  themes,
  DARK_THEME,
  LIGHT_THEME,

  // Classes for advanced usage
  Terminal: RexecTerminal,
  ApiClient: RexecApiClient,

  // Utilities
  utils,
  generateSessionId,

  // Version
  VERSION,
};

// Expose to global scope for CDN usage
if (typeof window !== 'undefined') {
  (window as unknown as { Rexec: typeof Rexec }).Rexec = Rexec;
}

// Default export for ES modules
export default Rexec;
