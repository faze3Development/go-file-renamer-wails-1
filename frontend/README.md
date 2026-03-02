# Go File Renamer Pro - Frontend

Modern Svelte-based frontend for the Go File Renamer Pro application, built with Vite and integrated with Wails for desktop application development.

## 🎨 Tech Stack

- **Svelte 4**: Reactive UI framework
- **Vite 5**: Fast build tool and dev server
- **Wails v2**: Go + Web UI integration
- **Vitest**: Unit testing framework
- **Lucide Svelte**: Icon library
- **date-fns**: Date utility library
- **Svelte French Toast**: Toast notifications

## 📁 Project Structure

```
frontend/
├── src/
│   ├── components/           # Reusable UI components
│   │   ├── views/           # Main application views
│   │   │   ├── AdvancedView.svelte           # Advanced operations panel
│   │   │   ├── AdvancedOperationsView.svelte # Bulk processing UI
│   │   │   └── AdvancedOperationsPanel.svelte
│   │   ├── ConfigurationView.svelte    # Main configuration panel
│   │   ├── MonitoringView.svelte       # Real-time monitoring
│   │   ├── WatcherControls.svelte      # Start/stop controls
│   │   ├── FileRenamer.svelte          # Main application container
│   │   ├── Header.svelte               # Application header
│   │   ├── Sidebar.svelte              # Navigation sidebar
│   │   ├── SettingsModal.svelte        # Settings dialog
│   │   ├── ErrorBoundary.svelte        # Error handling
│   │   └── LoadingBoundary.svelte      # Loading states
│   ├── assets/              # Static assets (images, fonts)
│   ├── types/               # TypeScript definitions
│   ├── App.svelte           # Root component
│   ├── stores.js            # Svelte stores for state management
│   └── style.css            # Global styles
├── wailsjs/                 # Auto-generated Wails bindings
│   └── go/                  # Go method bindings
├── dist/                    # Build output (generated)
├── package.json             # Dependencies and scripts
└── vite.config.js           # Vite configuration
```

## 🚀 Development

### Prerequisites

- Node.js (v18 or higher)
- npm or pnpm

### Installation

```bash
# Install dependencies
npm install
```

### Development Server

```bash
# Start Vite dev server (standalone)
npm run dev

# Or use Wails dev mode from project root (recommended)
cd ..
wails dev
```

When using `wails dev`, the frontend runs with hot module replacement (HMR) and automatically reloads on changes.

### Testing

```bash
# Run tests
npm test

# Run tests in watch mode
npm test -- --watch

# Run tests with coverage
npm test -- --coverage
```

## 🏗️ Building

### Development Build

```bash
npm run build
```

This creates an optimized production build in the `dist/` directory.

### Production Build (via Wails)

```bash
cd ..
wails build
```

The Wails build process automatically builds the frontend and bundles it into the desktop application.

## 📦 Key Components

### Main Application Components

- **FileRenamer.svelte**: Main application container that orchestrates all views
- **ConfigurationView.svelte**: Configuration panel for watch settings, patterns, and naming schemes
- **MonitoringView.svelte**: Real-time log viewer and statistics display
- **AdvancedOperationsView.svelte**: Bulk file processing interface with drag-and-drop

### Utility Components

- **ErrorBoundary.svelte**: Catches and displays component errors gracefully
- **LoadingBoundary.svelte**: Shows loading states during async operations
- **WatcherControls.svelte**: Start/stop buttons with status indicators

### State Management

The application uses Svelte stores for reactive state management:

```javascript
// stores.js
import { writable } from 'svelte/store';

export const logs = writable([]);
export const stats = writable({ processed: 0, failed: 0, renamed: 0 });
export const watcherRunning = writable(false);
```

## 🎯 Features

### Configuration Management
- Directory selection with native file dialogs
- Pattern-based file filtering with regex support
- Multiple naming scheme options
- Profile save/load functionality

### Real-Time Monitoring
- Live log streaming from backend
- Statistics updates (files processed, renamed, failed)
- Scrollable log viewer with auto-scroll
- Color-coded log levels

### Advanced Operations
- Drag-and-drop file upload
- Bulk file processing with progress tracking
- Metadata extraction and removal
- File optimization and compression
- Download processed files

### User Experience
- Toast notifications for user feedback
- Modal dialogs for settings
- Responsive layout
- Loading states and error boundaries
- Keyboard shortcuts support

## 🔌 Wails Integration

### Go Method Bindings

The frontend communicates with the Go backend through auto-generated bindings:

```javascript
import * as App from '../wailsjs/go/main/App';

// Start watching directory
await App.StartWatching(config);

// Select directory
const dir = await App.SelectDirectory();

// Load profiles
const profiles = await App.LoadProfiles();
```

### Event Handling

Listen to backend events using Wails event system:

```javascript
import { EventsOn, EventsOff } from '../wailsjs/runtime';

// Subscribe to events
EventsOn('log', (entry) => {
  logs.update(l => [...l, entry]);
});

EventsOn('stats', (newStats) => {
  stats.set(newStats);
});

// Cleanup on component destroy
onDestroy(() => {
  EventsOff('log');
  EventsOff('stats');
});
```

## 🎨 Styling

The application uses:
- Custom CSS with CSS variables for theming
- Nunito font family for modern typography
- Responsive design with flexbox and grid
- Dark color scheme optimized for desktop use

### CSS Variables

```css
:root {
  --primary-color: #646cff;
  --background-color: #0a0a0a;
  --text-color: #e0e0e0;
  --border-color: #333333;
  /* ... more variables */
}
```

## 🧪 Testing

Tests are written using Vitest and Testing Library:

```javascript
import { render, fireEvent } from '@testing-library/svelte';
import { describe, it, expect } from 'vitest';
import Component from './Component.svelte';

describe('Component', () => {
  it('renders correctly', () => {
    const { getByText } = render(Component);
    expect(getByText('Expected Text')).toBeInTheDocument();
  });
});
```

## 📚 Resources

### Svelte
- [Svelte Documentation](https://svelte.dev/docs)
- [Svelte Tutorial](https://svelte.dev/tutorial)

### Wails
- [Wails Documentation](https://wails.io/docs/introduction)
- [Wails Frontend Guide](https://wails.io/docs/guides/frontend)

### Vite
- [Vite Documentation](https://vitejs.dev/guide/)
- [Vite + Svelte Plugin](https://github.com/sveltejs/vite-plugin-svelte)

## 🐛 Troubleshooting

### Hot Module Replacement (HMR) Issues

If HMR isn't working:
1. Make sure you're running `wails dev` from the project root
2. Check that the frontend dev server is running on the correct port (default: 5174)
3. Clear Vite cache: `rm -rf node_modules/.vite`

### Build Errors

If you encounter build errors:
1. Delete `node_modules` and reinstall: `rm -rf node_modules && npm install`
2. Clear Vite cache: `rm -rf node_modules/.vite`
3. Ensure all peer dependencies are installed

### Wails Binding Issues

If Go method bindings aren't working:
1. Rebuild the application: `wails dev` regenerates bindings automatically
2. Check that methods are exported (capitalized) in Go code
3. Verify the struct is bound in `main.go`

## 🤝 Contributing

When contributing to the frontend:

1. Follow Svelte best practices and component patterns
2. Write tests for new features
3. Update this README for significant changes
4. Use consistent code formatting (Prettier recommended)
5. Keep components small and focused on single responsibilities

## 📄 License

Copyright (c) 2025 FAZE3 DEVELOPMENT LLC. All rights reserved.
