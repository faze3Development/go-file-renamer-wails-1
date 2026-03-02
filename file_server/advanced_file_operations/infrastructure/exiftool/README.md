# ExifTool Service

The ExifTool service provides a robust, enterprise-grade wrapper around the ExifTool utility for metadata extraction and removal operations. It follows the dependency injection pattern used throughout the application and provides full configuration support for all ExifTool options.

## Architecture

The service uses a single ExifTool process with the `stay_open` feature for optimal performance. It's designed to be thread-safe and provides proper lifecycle management with explicit initialization and cleanup.

### Key Features

- **Thread-Safe**: The underlying go-exiftool library uses mutex locking for concurrent access
- **Single Instance**: One ExifTool process per application for optimal resource usage
- **Full Configuration**: Support for all ExifTool command-line options
- **Proper Lifecycle**: Explicit initialization and cleanup with graceful shutdown
- **Error Handling**: Consistent error types using the infrastructure/errors package
- **Dependency Injection**: Follows the same pattern as other services (auth, db)

## Configuration

The service supports all ExifTool configuration options through the `Config` struct:

```go
type Config struct {
    // Buffer settings for reading ExifTool output
    Buffer     []byte
    BufferSize int

    // Character encoding settings
    Charset string

    // API options (can be multiple)
    ApiValues []string

    // Processing options
    NoPrintConversion         bool
    ExtractEmbedded          bool
    ExtractAllBinaryMetadata bool

    // Format options
    DateFormat string
    CoordFormat string

    // Output options
    PrintGroupNames string

    // File handling options
    BackupOriginal           bool
    ClearFieldsBeforeWriting bool

    // Binary path
    ExiftoolBinaryPath string
}
```

### Default Configuration

The `DefaultConfig()` function provides sensible defaults for production use:

```go
config := exiftool.DefaultConfig()
// Returns:
// - 128KB buffer with 64KB max size
// - UTF-8 charset for filenames
// - Standard date format: %Y:%m:%d %H:%M:%S
// - Coordinate format: %+f
// - Overwrite original files (no backup)
// - Use system PATH for exiftool binary
```

### Custom Configuration Examples

#### Production Configuration
```go
config := &exiftool.Config{
    Buffer:     make([]byte, 256*1024), // 256KB buffer
    BufferSize: 128 * 1024,             // 128KB max size
    Charset:    "filename=utf8",
    DateFormat: "%Y-%m-%d %H:%M:%S",
    CoordFormat: "%+f",
    // Use system PATH for exiftool binary
}
```

#### Development Configuration
```go
config := &exiftool.Config{
    Buffer:     make([]byte, 64*1024),  // Smaller buffer for dev
    BufferSize: 32 * 1024,              // 32KB max size
    Charset:    "filename=utf8",
    DateFormat: "%Y:%m:%d %H:%M:%S",
    CoordFormat: "%+f",
    PrintGroupNames: "0", // Show group names for debugging
    ExiftoolBinaryPath: "/usr/local/bin/exiftool", // Explicit path
}
```

#### High-Performance Configuration
```go
config := &exiftool.Config{
    Buffer:     make([]byte, 512*1024), // Large buffer
    BufferSize: 256 * 1024,             // 256KB max size
    Charset:    "filename=utf8",
    NoPrintConversion: true,            // Faster processing
    ExtractEmbedded: true,              // Extract embedded metadata
    DateFormat: "%s",                   // Unix timestamp format
    CoordFormat: "%+f",
}
```

## Usage

### Service Initialization

The service is initialized in `main.go` with proper lifecycle management:

```go
// Initialize ExifTool service
exiftoolConfig := exiftool.DefaultConfig()
exiftoolService, err := exiftool.NewService(exiftoolConfig, logger)
if err != nil {
    logger.Error("Failed to initialize ExifTool service", "error", err)
    os.Exit(1)
}
defer exiftoolService.Close()
```

### Dependency Injection

The service is passed through the dependency injection chain:

```go
// In operations.NewOrchestrator
func NewOrchestrator(userService user.ServiceAPI, dbService db.ServiceInterface, exiftoolService exiftool.Service, logger *slog.Logger) *Orchestrator {
    // ... other services
    return &Orchestrator{
        ExifToolService: exiftoolService,
        // ... other fields
    }
}
```

### Service Interface

The service provides two main operations:

```go
type Service interface {
    // RemoveAllMetadata removes all EXIF and other metadata from a file's content.
    RemoveAllMetadata(file *common.File) error

    // ExtractMetadata reads all metadata from a file's content and returns it as a map.
    ExtractMetadata(file *common.File) (map[string]interface{}, error)

    // Close gracefully shuts down the ExifTool service and releases resources.
    Close() error
}
```

### Usage Examples

#### Metadata Extraction
```go
// Create a file object
file := &common.File{
    Filename: "image.jpg",
    Content:  imageData,
    Size:     int64(len(imageData)),
}

// Extract metadata
metadata, err := exiftoolService.ExtractMetadata(file)
if err != nil {
    log.Printf("Failed to extract metadata: %v", err)
    return
}

// Access specific metadata fields
if cameraMake, ok := metadata["Make"]; ok {
    fmt.Printf("Camera Make: %v\n", cameraMake)
}

if exposureTime, ok := metadata["ExposureTime"]; ok {
    fmt.Printf("Exposure Time: %v\n", exposureTime)
}
```

#### Metadata Removal
```go
// Create a file object
file := &common.File{
    Filename: "image.jpg",
    Content:  imageData,
    Size:     int64(len(imageData)),
}

// Remove all metadata
err := exiftoolService.RemoveAllMetadata(file)
if err != nil {
    log.Printf("Failed to remove metadata: %v", err)
    return
}

// The file.Content now contains the cleaned image data
cleanedImageData := file.Content
```

#### Metadata Writing
```go
// Create a file object
file := &common.File{
    Filename: "image.jpg",
    Content:  imageData,
    Size:     int64(len(imageData)),
}

// Define metadata to write
metadata := map[string]interface{}{
    "Author":    "John Doe",
    "Copyright": "2024",
    "Keywords":  "travel,sunset,beach",
    "Title":     "Beautiful Sunset",
    "Rating":    5,
}

// Write metadata to file
err := exiftoolService.WriteMetadata(file, metadata)
if err != nil {
    log.Printf("Failed to write metadata: %v", err)
    return
}

// The file.Content now contains the image with new metadata
modifiedImageData := file.Content
```

## Thread Safety

The service is thread-safe by design:

- The underlying go-exiftool library uses mutex locking for concurrent access
- Multiple goroutines can safely call `ExtractMetadata` and `RemoveAllMetadata` simultaneously
- The single ExifTool process handles all requests efficiently

## Error Handling

The service uses the infrastructure/errors package for consistent error types:

- `NewSystemError`: For initialization and system-level errors
- `NewFileProcessingError`: For file-specific processing errors
- All errors include context and can be safely logged

## Lifecycle Management

The service follows a strict lifecycle:

1. **Initialization**: Created in `main.go` with configuration
2. **Usage**: Passed through dependency injection to processors and handlers
3. **Cleanup**: Explicitly closed with `defer exiftoolService.Close()` in `main.go`

## Requirements

- ExifTool must be installed on the system
- By default, the service looks for `exiftool` in the system PATH
- Can be configured to use a specific binary path via `ExiftoolBinaryPath`

## Performance Considerations

- Uses ExifTool's `stay_open` feature for optimal performance
- Single process handles all requests (no process spawning overhead)
- Configurable buffer sizes for different use cases
- Thread-safe concurrent access

## Troubleshooting

### Common Issues

1. **ExifTool not found**: Ensure ExifTool is installed and in PATH, or set `ExiftoolBinaryPath`
2. **Buffer too small**: Increase `Buffer` and `BufferSize` for large files
3. **Permission errors**: Ensure the service has read/write access to temporary directories

### Debugging

Enable debug logging by setting the log level to debug in your logger configuration. The service logs:
- Initialization success/failure
- File processing operations
- Cleanup operations
- Error details with context
