# Folder Watcher with UI

A file system watching tool that monitors a directory for changes and copies modified files to a destination directory. Features both command line and graphical user interfaces.

## Features

- Watch a source directory for file changes (create, modify, delete)
- Copy changed files to a destination directory
- Graphical user interface for easy configuration
- Command line interface for scripting and automation
- Configurable ignore patterns
- Recursive directory watching
- Adjustable buffer size for file copying
- Multiple log levels

## Installation

```bash
go get github.com/yourusername/folderwatcher
```

Or clone the repository and build:

```bash
git clone https://github.com/yourusername/folderwatcher.git
cd folderwatcher
go build
```

## Usage

### Graphical User Interface

Run the application with no arguments or with the `-ui` flag to start in UI mode:

```bash
./folderwatcher
# or
./folderwatcher -ui
```

The UI allows you to:
- Set source and destination directories
- Configure ignore patterns
- Toggle recursive watching
- Set copy buffer size
- Choose log level
- Start and stop watching
- View real-time logs

### Command Line Interface

Use the `-cli` flag to run in command line mode:

```bash
./folderwatcher -cli -src=/path/to/source -dst=/path/to/destination
```

Available command line options:

- `-src`: Source directory to watch (required in CLI mode)
- `-dst`: Destination directory to copy to (required in CLI mode)
- `-recursive`: Watch subdirectories recursively (default: true)
- `-ignore`: Comma-separated patterns to ignore (default: .git,.DS_Store,node_modules)
- `-buffer`: Copy buffer size in MB (default: 32)
- `-log-level`: Log level: debug, info, warning, error (default: info)

## Configuration

The application stores configuration in a settings file for persistence between runs. The UI allows saving and loading these configurations.

## Requirements

- Go 1.24 or later
- For UI: A windowing system (X11, Wayland, Windows, or macOS)

## License

MIT
