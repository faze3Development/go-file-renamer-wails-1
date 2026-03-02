# Infrastructure Package

This package contains the core infrastructure layer of the application, providing essential services that support all file processing operations. The infrastructure follows clean architecture principles with clear separation of concerns and dependency injection patterns.

## 📦 Sub-packages

### Common (`common/`)
Defines shared data structures, constants, and context utilities used throughout the application.

**Key Components:**
- `File`, `FileOperation`, `FileOperationResult`: Core file processing types
- Context keys for dependency injection
- File size constants and limits
- User and profile data structures

**Files:**
- `common.go`: Common data structures and constants
- `contextkeys.go`: Context keys and helper functions

### Database (`db/`)
Manages database connections and provides access to Firestore for persistent storage.

**Key Components:**
- Firestore client initialization and configuration
- Database service interface for dependency injection
- Connection pooling and lifecycle management

**Files:**
- `service.go`: Database service implementation
- `db.go`: Database utilities and helpers

**Note:** Currently configured for Firestore. Future versions may support additional databases.

### Errors (`errors/`)
Implements a centralized, structured error handling system with HTTP-aware error responses.

**Key Components:**
- Custom error types (`AppError`) with structured data
- HTTP status code mapping
- JSON error response formatting
- Error constructors for common scenarios

**Error Types:**
- `ValidationError`: Input validation failures
- `NotFoundError`: Resource not found
- `UnauthorizedError`: Authentication failures
- `PayloadTooLargeError`: File size violations
- `RateLimitError`: Rate limit exceeded
- `SystemError`: Internal system errors
- `FileProcessingError`: File operation failures

**Files:**
- `types.go`: Error type definitions and constructors
- `handlers.go`: HTTP error handlers and response formatting
- `validation.go`: Validation-specific error helpers

### ExifTool (`exiftool/`)
Provides enterprise-grade wrapper around ExifTool for comprehensive metadata operations.

**Key Features:**
- Thread-safe single-process architecture
- Metadata extraction from 100+ file formats
- Complete metadata removal for privacy
- Metadata writing capabilities
- Full ExifTool configuration support
- Proper lifecycle management

**Supported Formats:**
- Images: JPEG, PNG, GIF, WebP, TIFF, RAW formats (CR2, NEF, ARW, DNG)
- Videos: MP4, AVI, MOV, MKV, WebM
- Audio: MP3, FLAC, M4A, AAC
- Documents: PDF, Office files (DOC, DOCX, XLS, XLSX, PPT, PPTX)
- Archives: ZIP, RAR, 7Z

**Files:**
- `service.go`: Service initialization and interface
- `exiftool.go`: ExifTool wrapper implementation

**See:** [ExifTool README](exiftool/README.md) for detailed usage and configuration

### Security (`security/`)
Provides comprehensive security features including input validation, rate limiting, and file security checks.

**Key Features:**
- **Input Sanitization**: XSS and injection attack prevention
- **File Validation**: Content-type and magic number verification
- **Rate Limiting**: IP-based request throttling
- **Executable Detection**: Prevents processing malicious files
- **Size Validation**: Enforces file size limits
- **Path Sanitization**: Prevents directory traversal attacks

**Components:**
- `SanitizeInput()`: Sanitizes user text input
- `SanitizeFilename()`: Sanitizes filenames and paths
- `ValidateFileType()`: Validates file content types
- `CheckForInjection()`: Detects injection attempts
- `RequestValidator`: Comprehensive request validation
- `IPRateLimiter`: IP-based rate limiting
- `ValidateFile()`: Multi-layer file validation

**Files:**
- `validation.go`: Input and file validation functions
- `ratelimit.go`: Rate limiting middleware and implementation

**See:** [Security README](security/README.md) for detailed security features

### User (`user/`)
Manages user-related operations including profile management and user data persistence.

**Key Components:**
- User profile CRUD operations
- User configuration management
- Firestore-backed persistence
- User authentication helpers

**Files:**
- `service.go`: User business logic
- `repository.go`: Database operations
- `handler.go`: HTTP request handlers
- `interfaces.go`: Service and repository interfaces

**Note:** Follows repository pattern for clean separation of business logic and data access.

## 🏗️ Architecture Principles

### Dependency Injection
All services use constructor injection for dependencies:

```go
func NewService(logger *slog.Logger, db db.ServiceInterface) *Service {
    return &Service{
        logger: logger,
        db:     db,
    }
}
```

### Interface-Based Design
Services depend on interfaces, not concrete implementations:

```go
type ServiceAPI interface {
    GetUser(ctx context.Context, userID string) (*User, error)
    UpdateUser(ctx context.Context, user *User) error
}
```

### Context Propagation
Request context flows through all layers:

```go
func (s *Service) ProcessFile(ctx context.Context, file *common.File) error {
    logger := common.LoggerFromContext(ctx)
    userID := common.UserIDFromContext(ctx)
    // ... processing logic
}
```

### Error Handling
Consistent error types across all infrastructure:

```go
if err := validateInput(input); err != nil {
    return errors.NewValidationError("invalid input", err)
}
```

## 🔧 Configuration

### Environment Variables

Infrastructure components respect environment variables for configuration:

- `ENV`: Environment mode (`development`, `production`)
- `ALLOWED_ORIGINS`: CORS allowed origins (comma-separated)
- `FIRESTORE_PROJECT_ID`: Firebase project ID
- `UPLOAD_RATE_LIMIT_RPS`: Upload rate limit (requests per second)
- `UPLOAD_RATE_LIMIT_BURST`: Upload rate limit burst size
- `API_RATE_LIMIT_RPS`: API rate limit (requests per second)
- `API_RATE_LIMIT_BURST`: API rate limit burst size

### File Size Limits

Defined in `common/common.go`:

```go
const (
    MaxFileSize10MB  = 10 * 1024 * 1024   // 10MB
    MaxFileSize50MB  = 50 * 1024 * 1024   // 50MB
    MaxPayload32MB   = 32 * 1024 * 1024   // 32MB
)
```

## 🔒 Security Considerations

### Input Validation
All user inputs are sanitized before processing:
- XSS prevention via `go-sanitize` library
- SQL injection prevention (parameterized queries)
- Path traversal prevention (path sanitization)
- File type validation (magic number detection)

### Rate Limiting
Prevents abuse with configurable limits:
- Per-IP rate limiting
- Separate limits for uploads vs. API calls
- Automatic cleanup of old rate limiters

### File Security
Multiple layers of file validation:
- Content-type verification
- Magic number detection
- Executable file blocking
- File size enforcement
- Extension whitelisting/blacklisting

## 📝 Usage Examples

### Using Error Handling

```go
import "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"

func processFile(file *common.File) error {
    if file.Size > common.MaxFileSize50MB {
        return errors.NewPayloadTooLargeError(common.MaxFileSize50MB, file.Size)
    }
    
    if err := validateContent(file); err != nil {
        return errors.NewValidationError("invalid file content", err)
    }
    
    // Process file...
    return nil
}
```

### Using Security Validation

```go
import "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/security"

func handleUpload(r *http.Request) error {
    validator := security.NewRequestValidator(logger)
    
    files, config, err := validator.ValidateAndParseMultipartRequest(r, common.MaxPayload32MB)
    if err != nil {
        return err // Returns structured validation error
    }
    
    // Process validated files...
    return nil
}
```

### Using ExifTool Service

```go
import "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/exiftool"

// Initialize service
config := exiftool.DefaultConfig()
service, err := exiftool.NewService(config, logger)
if err != nil {
    log.Fatal(err)
}
defer service.Close()

// Extract metadata
metadata, err := service.ExtractMetadata(file)
if err != nil {
    return err
}

// Remove metadata
err = service.RemoveAllMetadata(file)
```

## 🧪 Testing

Each infrastructure package should have corresponding tests:

```bash
# Test all infrastructure packages
go test ./file_server/advanced_file_operations/infrastructure/...

# Test specific package
go test ./file_server/advanced_file_operations/infrastructure/security/

# Test with coverage
go test -cover ./file_server/advanced_file_operations/infrastructure/...
```

## 📚 Related Documentation

- [ExifTool Service](exiftool/README.md) - Detailed ExifTool configuration and usage
- [Security Package](security/README.md) - Security features and best practices
- [Backend API Reference](../docs/backend-api-reference.md) - HTTP API documentation

## 🤝 Contributing

When adding new infrastructure components:

1. Follow the dependency injection pattern
2. Use interfaces for external dependencies
3. Implement proper error handling with structured errors
4. Add comprehensive logging
5. Write unit tests
6. Update this README with new packages

## 📄 License

Copyright (c) 2025 FAZE3 DEVELOPMENT LLC. All rights reserved.
