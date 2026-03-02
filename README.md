# Go File Renamer Pro (Wails)

A powerful, enterprise-grade file management application built with Go and Wails, featuring a modern Svelte frontend. This application provides comprehensive file processing capabilities including automated renaming, metadata operations, file optimization, and bulk file processing with advanced security features.

## 🚀 Features

### Core File Operations
*   **Directory Watching**: Real-time monitoring of directories for new files with recursive support
*   **Pattern-Based Renaming**: Advanced pattern matching using regular expressions
*   **Flexible Naming Schemes**: Multiple naming strategies including:
    *   Sequential numbering with custom patterns
    *   Timestamp-based naming (various formats)
    *   Random string generation
    *   Metadata-based renaming (EXIF data from images)
    *   Custom pattern templates
*   **Profile Management**: Save and load configuration profiles for different workflows

### Advanced File Processing
*   **Bulk File Processing**: Asynchronous processing of large file batches
*   **Metadata Operations**:
    *   Extract metadata from images, videos, PDFs, documents, and audio files
    *   Remove EXIF and other metadata for privacy
    *   Write custom metadata to files
    *   Rename files based on extracted metadata
*   **File Optimization**:
    *   Image compression and optimization
    *   Concurrent processing with goroutines
*   **Live Monitoring**: Real-time logs, statistics, and job status tracking
*   **Dry Run Mode**: Preview all operations without modifying files

### Security & Performance
*   **Rate Limiting**: IP-based request throttling to prevent abuse
*   **Input Validation**: Comprehensive sanitization and validation of all inputs
*   **File Type Validation**: Magic number detection and content-type verification
*   **Executable Detection**: Prevents processing of potentially malicious files
*   **Size Limits**: Configurable file size restrictions
*   **Secure File Handling**: Proper error handling and resource cleanup
*   **Cross-Platform**: Works on Windows, macOS, and Linux

## 📁 Project Structure

```
.
├── file_server/                    # Go backend
│   ├── action/                     # Post-processing actions
│   ├── advanced_file_operations/   # Advanced processing features
│   │   ├── infrastructure/         # Core services
│   │   │   ├── common/            # Shared types and utilities
│   │   │   ├── db/                # Database service (Firestore)
│   │   │   ├── errors/            # Error handling
│   │   │   ├── exiftool/          # ExifTool integration
│   │   │   ├── security/          # Security validation & rate limiting
│   │   │   └── user/              # User management
│   │   ├── operations/            # File processing operations
│   │   │   ├── file_operations/   # File manipulation
│   │   │   │   ├── bulk_file_processing/   # Async batch processing
│   │   │   │   ├── limited_file_processing/# Sync processing
│   │   │   │   ├── optimizer/     # File optimization
│   │   │   │   └── renamers/      # Naming strategies
│   │   │   └── metadata_Operations/  # Metadata features
│   │   │       ├── metadata_extraction/  # Extract metadata
│   │   │       ├── metadata_removal/     # Remove metadata
│   │   │       ├── metadata_renamer/     # Rename by metadata
│   │   │       └── metadata_writer/      # Write metadata
│   │   └── server/                # HTTP server setup
│   ├── config/                    # Configuration types
│   ├── logger/                    # Logging utilities
│   ├── processor/                 # File processing logic
│   ├── stats/                     # Statistics tracking
│   ├── users/                     # Profile management
│   └── watcher/                   # Directory monitoring
├── frontend/                      # Svelte frontend
│   ├── src/
│   │   ├── components/           # UI components
│   │   │   ├── views/           # Main application views
│   │   │   └── ...              # Reusable components
│   │   ├── assets/              # Static assets
│   │   ├── App.svelte           # Root component
│   │   └── stores.js            # State management
│   └── wailsjs/                 # Wails generated bindings
├── main.go                       # Application entry point
├── wails.json                    # Wails configuration
└── README.md
```

## 🛠️ Prerequisites

### Required Software
*   **Go** 1.24.0 or higher
*   **Node.js** and npm (for frontend development)
*   **Wails CLI** v2.10.2 or higher: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
*   **ExifTool** (for metadata operations):
    *   **Ubuntu/Debian**: `sudo apt-get install libimage-exiftool-perl`
    *   **macOS**: `brew install exiftool`
    *   **Windows**: Download from [exiftool.org](https://exiftool.org/)

## 🚀 Getting Started

### Installation

1. Clone the repository:
```bash
git clone https://github.com/faze3Development/go-file-renamer-wails.git
cd go-file-renamer-wails
```

2. Install dependencies:
```bash
# Install Go dependencies
go mod download

# Install frontend dependencies
cd frontend
npm install
cd ..
```

### Development

Run in live development mode with hot-reload:
```bash
wails dev
```

This starts:
- Vite development server for the frontend (hot module replacement)
- Go backend with automatic compilation
- Desktop application window

The frontend is also accessible in a browser at `http://localhost:5174` for debugging with browser developer tools.

### Building

Build a production-ready executable:
```bash
wails build
```

The executable will be created in the `build/bin/` directory and includes both frontend and backend in a single file.

#### Build Options
```bash
# Build for specific platform
wails build -platform windows/amd64
wails build -platform darwin/universal
wails build -platform linux/amd64

# Clean build
wails build -clean
```

## 📖 Usage

### Basic Workflow

1. **Configure Settings**: 
   - Select a directory to watch
   - Choose a naming scheme or pattern
   - Configure advanced options (metadata, optimization)

2. **Start Watching**:
   - Click "Start Watching" to begin monitoring
   - View real-time logs and statistics
   - Files are processed automatically when added

3. **Advanced Features**:
   - Use the Advanced Operations panel for bulk processing
   - Extract or remove metadata from files
   - Optimize images and documents

### Configuration Options

*   **Watch Directory**: Directory to monitor for new files
*   **Recursive**: Include subdirectories in monitoring
*   **Dry Run**: Preview operations without making changes
*   **Name Pattern**: Regex pattern to match files (e.g., `screenshot-.*\.png`)
*   **Naming Scheme**: Choose from multiple naming strategies:
    *   Sequential (e.g., `file_001.jpg`, `file_002.jpg`)
    *   Timestamp (various date/time formats)
    *   Random strings
    *   Metadata-based (extract from EXIF data)
*   **Action**: Post-processing actions (move, copy, delete)
*   **Advanced Options**:
    *   Remove metadata from files
    *   Optimize/compress files
    *   Custom metadata operations

### Profile Management

Save frequently used configurations as profiles:
1. Configure your settings
2. Click "Save Profile"
3. Give it a descriptive name
4. Load profiles instantly when needed

## 🔒 Security

This application implements multiple layers of security:

- **Input Sanitization**: All user inputs are sanitized to prevent XSS and injection attacks
- **File Validation**: Content-type verification and magic number detection
- **Rate Limiting**: Prevents API abuse with configurable limits
- **Secure File Permissions**: Configuration files use restricted permissions (0600)
- **Path Traversal Protection**: Validates and sanitizes all file paths
- **Size Limits**: Enforces maximum file sizes to prevent DoS attacks
- **Executable Detection**: Blocks processing of executable files

## 🧪 Testing

```bash
# Run Go tests
go test ./...

# Run frontend tests
cd frontend
npm test
```

## 📝 API Documentation

For backend API documentation, see:
- [Backend API Reference](file_server/advanced_file_operations/docs/backend-api-reference.md)

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

Copyright (c) 2025 FAZE3 DEVELOPMENT LLC. All rights reserved.

## 🐛 Known Issues

For known issues and feature requests, please check the GitHub Issues page.

## 📧 Support

For support, please contact: ole.abalo@faze3.dev
