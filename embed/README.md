# Rexec Embed Widget

Embeddable terminal widget for Rexec - Cloud Shell anywhere.

Add a fully-featured cloud terminal to any website with just a few lines of code.

## Quick Start

### Via CDN

```html
<!-- Include the embed script -->
<script src="https://rexec.dev/embed/rexec.min.js"></script>

<!-- Create a container element -->
<div id="terminal" style="width: 100%; height: 400px;"></div>

<!-- Initialize the terminal -->
<script>
  const term = Rexec.embed('#terminal', {
    shareCode: 'your-share-code'
  });
</script>
```

### Via NPM

```bash
npm install @rexec/embed
```

```javascript
import { embed } from '@rexec/embed';

const term = embed('#terminal', {
  token: 'your-api-token',
  container: 'container-id'
});
```

## Connection Methods

### 1. Join via Share Code (Guest Access)

Join an existing shared session. No authentication required.

```javascript
const term = Rexec.embed('#terminal', {
  shareCode: 'abc123'
});
```

### 2. Connect to Existing Container

Connect to a container you own. Requires API token.

```javascript
const term = Rexec.embed('#terminal', {
  token: 'your-api-token',
  container: 'container-id'
});
```

### 3. Create New Container

Create a new container on-demand. Requires API token.

```javascript
const term = Rexec.embed('#terminal', {
  token: 'your-api-token',
  role: 'ubuntu'  // or 'node', 'python', 'go', 'rust', etc.
});
```

## Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `token` | string | - | API token for authentication |
| `container` | string | - | Container ID to connect to |
| `shareCode` | string | - | Share code for joining sessions |
| `role` | string | - | Environment type for new containers |
| `baseUrl` | string | 'https://rexec.dev' | API base URL |
| `theme` | 'dark' \| 'light' \| object | 'dark' | Terminal color theme |
| `fontSize` | number | 14 | Font size in pixels |
| `fontFamily` | string | 'JetBrains Mono, ...' | Terminal font |
| `cursorStyle` | 'block' \| 'underline' \| 'bar' | 'block' | Cursor appearance |
| `cursorBlink` | boolean | true | Whether cursor blinks |
| `scrollback` | number | 5000 | Lines in scrollback buffer |
| `webgl` | boolean | true | Use WebGL renderer |
| `showStatus` | boolean | true | Show connection status overlay |
| `autoReconnect` | boolean | true | Auto-reconnect on disconnect |
| `maxReconnectAttempts` | number | 10 | Max reconnection attempts |
| `initialCommand` | string | - | Command to run after connect |
| `className` | string | - | Custom CSS class |
| `fitToContainer` | boolean | true | Auto-fit to container size |

## Event Callbacks

```javascript
const term = Rexec.embed('#terminal', {
  shareCode: 'abc123',

  onReady: (terminal) => {
    console.log('Terminal connected!');
    console.log('Session:', terminal.session);
  },

  onStateChange: (state) => {
    // 'idle', 'connecting', 'connected', 'reconnecting', 'disconnected', 'error'
    console.log('State:', state);
  },

  onData: (data) => {
    // Terminal output received
    console.log('Output:', data);
  },

  onResize: (cols, rows) => {
    console.log(`Terminal resized to ${cols}x${rows}`);
  },

  onError: (error) => {
    console.error('Error:', error.code, error.message);
  },

  onDisconnect: (reason) => {
    console.log('Disconnected:', reason);
  }
});
```

## Terminal API

### Properties

```javascript
term.state    // Current connection state
term.session  // Session info (id, containerId, containerName, etc.)
term.stats    // Container stats (cpu, memory, disk, network)
```

### Methods

```javascript
// Write to terminal
term.write('echo "Hello"');
term.writeln('ls -la');  // Adds newline

// Terminal controls
term.clear();             // Clear screen
term.fit();               // Fit to container
term.focus();             // Focus terminal
term.blur();              // Blur terminal

// Appearance
term.setFontSize(16);
term.setTheme('light');

// Clipboard
await term.copySelection();
await term.paste();
term.selectAll();

// Dimensions
const { cols, rows } = term.getDimensions();

// Connection
await term.reconnect();
term.disconnect();

// Cleanup
term.destroy();
```

### Events

```javascript
// Subscribe to events
const unsubscribe = term.on('data', (data) => {
  console.log('Received:', data);
});

// Unsubscribe
unsubscribe();
// or
term.off('data', callback);
```

## Custom Themes

```javascript
const term = Rexec.embed('#terminal', {
  shareCode: 'abc123',
  theme: {
    background: '#1a1b26',
    foreground: '#a9b1d6',
    cursor: '#c0caf5',
    black: '#15161e',
    red: '#f7768e',
    green: '#9ece6a',
    yellow: '#e0af68',
    blue: '#7aa2f7',
    magenta: '#bb9af7',
    cyan: '#7dcfff',
    white: '#a9b1d6',
    brightBlack: '#414868',
    brightRed: '#f7768e',
    brightGreen: '#9ece6a',
    brightYellow: '#e0af68',
    brightBlue: '#7aa2f7',
    brightMagenta: '#bb9af7',
    brightCyan: '#7dcfff',
    brightWhite: '#c0caf5'
  }
});
```

## Global API

```javascript
// Get all active instances
const instances = Rexec.getInstances();

// Destroy all instances
Rexec.destroyAll();

// Create API client for direct API access
const client = Rexec.createClient('https://rexec.dev', 'api-token');

// Built-in themes
Rexec.DARK_THEME
Rexec.LIGHT_THEME

// Version
console.log(Rexec.VERSION);
```

## Development

```bash
# Install dependencies
npm install

# Development build with watch
npm run dev

# Production build
npm run build

# Clean build artifacts
npm run clean
```

## Building for Production

The build outputs to `dist/`:

- `rexec.min.js` - UMD bundle for CDN/script tag usage
- `rexec.esm.js` - ES module for bundlers
- `rexec.min.js.map` - Source map

## Browser Support

- Chrome 80+
- Firefox 75+
- Safari 13+
- Edge 80+

## License

MIT
