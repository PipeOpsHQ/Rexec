/**
 * Rexec Embed Widget Themes
 * Terminal color themes matching the main Rexec application
 */

import type { TerminalTheme } from './types';

/**
 * Dark terminal theme - matches Rexec dark mode
 */
export const DARK_THEME: TerminalTheme = {
  background: '#0d1117',
  foreground: '#c9d1d9',
  cursor: '#58a6ff',
  cursorAccent: '#0d1117',
  selectionBackground: 'rgba(56, 139, 253, 0.4)',
  selectionForeground: '#ffffff',
  selectionInactiveBackground: 'rgba(56, 139, 253, 0.2)',
  black: '#484f58',
  red: '#ff7b72',
  green: '#3fb950',
  yellow: '#d29922',
  blue: '#58a6ff',
  magenta: '#bc8cff',
  cyan: '#39c5cf',
  white: '#b1bac4',
  brightBlack: '#6e7681',
  brightRed: '#ffa198',
  brightGreen: '#56d364',
  brightYellow: '#e3b341',
  brightBlue: '#79c0ff',
  brightMagenta: '#d2a8ff',
  brightCyan: '#56d4dd',
  brightWhite: '#f0f6fc',
};

/**
 * Light terminal theme - matches Rexec light mode
 */
export const LIGHT_THEME: TerminalTheme = {
  background: '#ffffff',
  foreground: '#24292f',
  cursor: '#0969da',
  cursorAccent: '#ffffff',
  selectionBackground: 'rgba(9, 105, 218, 0.3)',
  selectionForeground: '#24292f',
  selectionInactiveBackground: 'rgba(9, 105, 218, 0.15)',
  black: '#24292f',
  red: '#cf222e',
  green: '#116329',
  yellow: '#4d2d00',
  blue: '#0969da',
  magenta: '#8250df',
  cyan: '#1b7c83',
  white: '#6e7781',
  brightBlack: '#57606a',
  brightRed: '#a40e26',
  brightGreen: '#1a7f37',
  brightYellow: '#633c01',
  brightBlue: '#218bff',
  brightMagenta: '#a475f9',
  brightCyan: '#3192aa',
  brightWhite: '#8c959f',
};

/**
 * Get a theme by name
 */
export function getTheme(theme: TerminalTheme | 'dark' | 'light'): TerminalTheme {
  if (typeof theme === 'string') {
    return theme === 'light' ? LIGHT_THEME : DARK_THEME;
  }
  return theme;
}

/**
 * Default theme
 */
export const DEFAULT_THEME = DARK_THEME;
