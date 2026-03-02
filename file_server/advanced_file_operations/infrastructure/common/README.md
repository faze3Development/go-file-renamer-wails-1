# Common Package

This package defines shared data structures, constants, context utilities, and types used throughout the application. It serves as the foundation for cross-cutting concerns and enables consistent data handling across all layers.

## 📦 Key Components

### Data Structures

#### File Types
Core file processing types used throughout the application:

```go
// File represents a file with its content and metadata
type File struct {
    Filename    string
    Content     []byte
    ContentType string
    Size        int64
}

// FileOperation represents a file to be processed
type FileOperation struct {
    Filename    string
    Content     []byte
    ContentType string
    Size        int64
    Metadata    map[string]interface{}
}

// FileOperationResult represents the result of processing a file
type FileOperationResult struct {
    Filename    string
    NewName     string
    Success     bool
    Error       string
    Action      string
    ContentType string
    Content     []byte
}
```

#### User Types
User and profile data structures:

```go
// User represents a user in the system
type User struct {
    ID        string
    Email     string
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// UserProfile represents user-specific configuration
type UserProfile struct {
    UserID       string
    Preferences  map[string]interface{}
    JobHistory   []JobSummary
    LastActivity time.Time
}
```

### Constants

#### File Size Limits
Predefined file size constants:

```go
const (
    MaxFileSize10MB  = 10 * 1024 * 1024   // 10 MB - Standard uploads
    MaxFileSize50MB  = 50 * 1024 * 1024   // 50 MB - Large file limit
    MaxPayload32MB   = 32 * 1024 * 1024   // 32 MB - Request payload limit
    MaxFileSize100MB = 100 * 1024 * 1024  // 100 MB - Enterprise limit
)
```

Usage:
```go
if file.Size > common.MaxFileSize50MB {
    return errors.NewPayloadTooLargeError(common.MaxFileSize50MB, file.Size)
}
```

#### Processing Limits
Limits for batch operations:

```go
const (
    MaxConcurrentJobs     = 100  // Maximum concurrent processing jobs
    MaxFilesPerBatch      = 1000 // Maximum files per batch operation
    MaxFilesPerRequest    = 100  // Maximum files per single request
    JobCleanupInterval    = 3600 // Seconds until completed jobs are cleaned
)
```

### Context Keys

Context keys for dependency injection and request scoping:

```go
type ContextKey string

const (
    LoggerKey     ContextKey = "logger"
    UserIDKey     ContextKey = "userID"
    RequestIDKey  ContextKey = "requestID"
    UserKey       ContextKey = "user"
    DatabaseKey   ContextKey = "database"
)
```

### Helper Functions

#### Context Utilities
Safe context value extraction:

```go
// LoggerFromContext extracts the logger from context
func LoggerFromContext(ctx context.Context) *slog.Logger {
    if logger, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
        return logger
    }
    return slog.Default()
}

// UserIDFromContext extracts the user ID from context
func UserIDFromContext(ctx context.Context) string {
    if userID, ok := ctx.Value(UserIDKey).(string); ok {
        return userID
    }
    return ""
}

// RequestIDFromContext extracts the request ID from context
func RequestIDFromContext(ctx context.Context) string {
    if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
        return reqID
    }
    return ""
}

// UserFromContext extracts the user from context
func UserFromContext(ctx context.Context) (*User, bool) {
    user, ok := ctx.Value(UserKey).(*User)
    return user, ok
}
```

#### File Size Utilities

```go
// ParseMaxFileSizeFromString parses a file size string with units
func ParseMaxFileSizeFromString(sizeStr string, defaultSize int64) int64 {
    // Supports formats like "10MB", "50M", "1GB"
    // Returns defaultSize if parsing fails
}

// GetMaxFileSizeFromEnv gets file size limit from environment
func GetMaxFileSizeFromEnv(defaultSize int64) int64 {
    // Reads from environment variable with fallback
}

// FormatFileSize formats bytes as human-readable string
func FormatFileSize(bytes int64) string {
    // Returns "10 MB", "1.5 GB", etc.
}
```

## 📁 Files

### `common.go`
Defines core data structures, constants, and shared utilities.

**Contents:**
- File processing types (`File`, `FileOperation`, `FileOperationResult`)
- User types (`User`, `UserProfile`)
- File size constants
- Processing limit constants
- File size utility functions
- Common validation helpers

### `contextkeys.go`
Defines context keys and extraction helpers for dependency injection.

**Contents:**
- Context key constants
- Context value extraction functions
- Type-safe context helpers
- Default value handling

## 🔧 Usage Examples

### File Processing

```go
import "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"

// Create a file for processing
file := &common.File{
    Filename:    "photo.jpg",
    Content:     imageData,
    ContentType: "image/jpeg",
    Size:        int64(len(imageData)),
}

// Check file size
if file.Size > common.MaxFileSize50MB {
    return fmt.Errorf("file too large")
}

// Create operation result
result := &common.FileOperationResult{
    Filename:    file.Filename,
    NewName:     "IMG_20250127.jpg",
    Success:     true,
    Action:      "renamed",
    ContentType: file.ContentType,
}
```

### Context Management

```go
import (
    "context"
    "log/slog"
    "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
)

func ProcessFile(ctx context.Context, file *common.File) error {
    // Extract logger from context
    logger := common.LoggerFromContext(ctx)
    logger.Info("processing file", "filename", file.Filename)
    
    // Extract user ID if available
    userID := common.UserIDFromContext(ctx)
    if userID != "" {
        logger.Info("processing for user", "userID", userID)
    }
    
    // Extract request ID for tracing
    requestID := common.RequestIDFromContext(ctx)
    logger = logger.With("requestID", requestID)
    
    // Process file...
    return nil
}

// Setting context values
func SetupContext(ctx context.Context, logger *slog.Logger, userID string) context.Context {
    ctx = context.WithValue(ctx, common.LoggerKey, logger)
    ctx = context.WithValue(ctx, common.UserIDKey, userID)
    ctx = context.WithValue(ctx, common.RequestIDKey, generateRequestID())
    return ctx
}
```

### File Size Utilities

```go
import "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"

// Parse file size from configuration
maxSize := common.ParseMaxFileSizeFromString("50MB", common.MaxFileSize50MB)

// Get file size from environment with fallback
maxUploadSize := common.GetMaxFileSizeFromEnv(common.MaxFileSize10MB)

// Format file size for display
displaySize := common.FormatFileSize(file.Size)
fmt.Printf("File size: %s\n", displaySize)
// Output: "File size: 2.5 MB"
```

### User Management

```go
import (
    "time"
    "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
)

// Create user
user := &common.User{
    ID:        "user-123",
    Email:     "user@example.com",
    Name:      "John Doe",
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}

// Create user profile
profile := &common.UserProfile{
    UserID: user.ID,
    Preferences: map[string]interface{}{
        "theme":        "dark",
        "defaultPath":  "/home/user/files",
        "autoProcess":  true,
    },
    JobHistory:   []common.JobSummary{},
    LastActivity: time.Now(),
}
```

## 🏗️ Architecture Patterns

### Dependency Injection
Context keys enable clean dependency injection without global state:

```go
// In middleware or handler setup
ctx = context.WithValue(ctx, common.LoggerKey, logger)
ctx = context.WithValue(ctx, common.UserIDKey, userID)

// In business logic
func ProcessRequest(ctx context.Context) {
    logger := common.LoggerFromContext(ctx)
    userID := common.UserIDFromContext(ctx)
    // Use injected dependencies
}
```

### Type Safety
All context extraction functions are type-safe with default values:

```go
// Safe extraction - never panics
logger := common.LoggerFromContext(ctx) // Returns default logger if not in context
userID := common.UserIDFromContext(ctx) // Returns empty string if not in context
```

### Constants Over Magic Numbers
Use named constants instead of hardcoded values:

```go
// Bad
if file.Size > 52428800 {
    return errors.New("file too large")
}

// Good
if file.Size > common.MaxFileSize50MB {
    return errors.NewPayloadTooLargeError(common.MaxFileSize50MB, file.Size)
}
```

## 🔒 Best Practices

### Context Propagation
Always propagate context through call chains:

```go
func Handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    // Add values to context
    ctx = context.WithValue(ctx, common.LoggerKey, logger)
    
    // Pass context to business logic
    ProcessFile(ctx, file)
}

func ProcessFile(ctx context.Context, file *common.File) error {
    // Extract from context
    logger := common.LoggerFromContext(ctx)
    
    // Pass to lower layers
    return repository.Save(ctx, file)
}
```

### File Size Validation
Always validate file sizes early:

```go
func ValidateFile(file *common.File) error {
    if file.Size <= 0 {
        return errors.New("file is empty")
    }
    if file.Size > common.MaxFileSize50MB {
        return errors.NewPayloadTooLargeError(common.MaxFileSize50MB, file.Size)
    }
    return nil
}
```

### Struct Initialization
Use struct literals for clarity:

```go
// Good - explicit field names
file := &common.File{
    Filename:    filename,
    Content:     data,
    ContentType: contentType,
    Size:        int64(len(data)),
}

// Avoid - positional arguments are error-prone
// file := &common.File{filename, data, contentType, int64(len(data))}
```

## 🧪 Testing

```bash
# Test common package
go test ./file_server/advanced_file_operations/infrastructure/common/

# Test with coverage
go test -cover ./file_server/advanced_file_operations/infrastructure/common/

# Benchmark file size utilities
go test -bench=. ./file_server/advanced_file_operations/infrastructure/common/
```

## 📚 Related Documentation

- [Infrastructure Overview](../README.md)
- [Security Package](../security/README.md) - Uses common types for validation
- [Errors Package](../errors/README.md) - Uses common types in error responses

## 🤝 Contributing

When adding new common types or constants:

1. Consider if it's truly cross-cutting (used in multiple packages)
2. Document the type with clear examples
3. Add helper functions for common operations
4. Update this README with new additions
5. Maintain backwards compatibility for constants

## 📄 License

Copyright (c) 2025 FAZE3 DEVELOPMENT LLC. All rights reserved.
