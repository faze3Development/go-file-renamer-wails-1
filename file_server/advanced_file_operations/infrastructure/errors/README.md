# Errors Package

This package implements a comprehensive, centralized error handling system with structured error types, HTTP-aware responses, and consistent error formatting. It provides type-safe error constructors and automatic HTTP status code mapping for web services.

## 🎯 Features

- **Structured Errors**: Rich error objects with context and metadata
- **HTTP Integration**: Automatic HTTP status code mapping
- **JSON Responses**: Consistent JSON error response formatting
- **Type Safety**: Specific error types for different scenarios
- **Context Preservation**: Maintains error context and stack information
- **Logging Integration**: Structured logging of errors
- **Recovery Handlers**: Panic recovery with graceful degradation

## 📦 Error Types

### AppError
Core error structure with full context:

```go
type AppError struct {
    Type       string                 `json:"type"`       // Error category
    Message    string                 `json:"message"`    // Human-readable message
    StatusCode int                    `json:"statusCode"` // HTTP status code
    Details    map[string]interface{} `json:"details"`    // Additional context
    Err        error                  `json:"-"`          // Original error (not exposed)
}
```

### Error Type Constants

```go
const (
    TypeValidationError      = "ValidationError"
    TypeNotFoundError        = "NotFoundError"
    TypeUnauthorizedError    = "UnauthorizedError"
    TypeForbiddenError       = "ForbiddenError"
    TypePayloadTooLargeError = "PayloadTooLargeError"
    TypeRateLimitError       = "RateLimitError"
    TypeSystemError          = "SystemError"
    TypeFileProcessingError  = "FileProcessingError"
    TypeDatabaseError        = "DatabaseError"
    TypeNetworkError         = "NetworkError"
)
```

## 📁 Files

### `types.go`
Defines error structures and constructor functions.

**Key Components:**
- `AppError` struct definition
- Error type constants
- Constructor functions for each error type
- Error interface implementation

### `handlers.go`
Provides HTTP error handling and response formatting.

**Key Components:**
- `HTTPErrorHandler()`: Main HTTP error handler
- `SafeJSONResponse()`: Safe JSON response writer
- `WrapHandler()`: Handler wrapper for automatic error handling
- `PanicRecovery()`: Panic recovery middleware

### `validation.go`
Contains validation-specific error constructors.

**Key Components:**
- File validation error constructors
- Request validation error helpers
- Field-specific validation errors

## 🔧 Error Constructors

### Validation Errors
```go
// NewValidationError creates a validation error
func NewValidationError(message string, details interface{}) *AppError {
    return &AppError{
        Type:       TypeValidationError,
        Message:    message,
        StatusCode: http.StatusBadRequest,
        Details:    map[string]interface{}{"validation": details},
    }
}

// Usage
if len(username) < 3 {
    return errors.NewValidationError(
        "username too short",
        map[string]interface{}{"minLength": 3, "actual": len(username)},
    )
}
```

### Not Found Errors
```go
// NewNotFoundError creates a not found error
func NewNotFoundError(resource, id string) *AppError {
    return &AppError{
        Type:       TypeNotFoundError,
        Message:    fmt.Sprintf("%s not found", resource),
        StatusCode: http.StatusNotFound,
        Details:    map[string]interface{}{"resource": resource, "id": id},
    }
}

// Usage
job, err := service.GetJob(jobID)
if err != nil {
    return errors.NewNotFoundError("job", jobID)
}
```

### Authorization Errors
```go
// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
    return &AppError{
        Type:       TypeUnauthorizedError,
        Message:    message,
        StatusCode: http.StatusUnauthorized,
    }
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
    return &AppError{
        Type:       TypeForbiddenError,
        Message:    message,
        StatusCode: http.StatusForbidden,
    }
}

// Usage
if !user.HasPermission(resource) {
    return errors.NewForbiddenError("insufficient permissions")
}
```

### Rate Limit Errors
```go
// NewRateLimitError creates a rate limit error
func NewRateLimitError(message string) *AppError {
    return &AppError{
        Type:       TypeRateLimitError,
        Message:    message,
        StatusCode: http.StatusTooManyRequests,
        Details:    map[string]interface{}{"retryAfter": 60},
    }
}

// Usage
if !rateLimiter.Allow(ip) {
    return errors.NewRateLimitError("too many requests")
}
```

### Payload Size Errors
```go
// NewPayloadTooLargeError creates a payload size error
func NewPayloadTooLargeError(maxSize, actualSize int64) *AppError {
    return &AppError{
        Type:       TypePayloadTooLargeError,
        Message:    "request payload too large",
        StatusCode: http.StatusRequestEntityTooLarge,
        Details: map[string]interface{}{
            "maxSizeBytes":    maxSize,
            "actualSizeBytes": actualSize,
        },
    }
}

// Usage
if fileSize > common.MaxFileSize50MB {
    return errors.NewPayloadTooLargeError(common.MaxFileSize50MB, fileSize)
}
```

### System Errors
```go
// NewSystemError creates a system/internal error
func NewSystemError(message string, err error) *AppError {
    return &AppError{
        Type:       TypeSystemError,
        Message:    message,
        StatusCode: http.StatusInternalServerError,
        Err:        err,
    }
}

// Usage
data, err := os.ReadFile(path)
if err != nil {
    return errors.NewSystemError("failed to read file", err)
}
```

### File Processing Errors
```go
// NewFileProcessingError creates a file processing error
func NewFileProcessingError(filename, message string, err error) *AppError {
    return &AppError{
        Type:       TypeFileProcessingError,
        Message:    message,
        StatusCode: http.StatusUnprocessableEntity,
        Details:    map[string]interface{}{"filename": filename},
        Err:        err,
    }
}

// Usage
if err := processImage(file); err != nil {
    return errors.NewFileProcessingError(
        file.Filename,
        "image processing failed",
        err,
    )
}
```

## 🌐 HTTP Integration

### HTTP Error Handler

```go
// HTTPErrorHandler handles errors and writes appropriate HTTP responses
func HTTPErrorHandler(w http.ResponseWriter, err error, logger *slog.Logger, requestID string) {
    // Converts any error to structured JSON response
    // Logs errors with context
    // Sets appropriate status codes
}

// Usage in handler
func Handler(w http.ResponseWriter, r *http.Request) {
    result, err := processRequest(r)
    if err != nil {
        errors.HTTPErrorHandler(w, err, logger, requestID)
        return
    }
    
    json.NewEncoder(w).Encode(result)
}
```

### Handler Wrapper

```go
// WrapHandler wraps an http.HandlerFunc to automatically handle errors
func WrapHandler(handler func(w http.ResponseWriter, r *http.Request) error, logger *slog.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := handler(w, r); err != nil {
            requestID := r.Context().Value("requestID").(string)
            HTTPErrorHandler(w, err, logger, requestID)
        }
    }
}

// Usage
router.HandleFunc("/api/files", errors.WrapHandler(uploadHandler, logger))

func uploadHandler(w http.ResponseWriter, r *http.Request) error {
    // Return error directly, wrapper handles it
    if err := validateRequest(r); err != nil {
        return err // Wrapper converts to HTTP response
    }
    return nil
}
```

### Panic Recovery

```go
// PanicRecovery middleware recovers from panics and returns 500 error
func PanicRecovery(logger *slog.Logger, requestIDKey string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    logger.Error("panic recovered", "error", err, "path", r.URL.Path)
                    
                    appErr := NewSystemError("internal server error", fmt.Errorf("%v", err))
                    HTTPErrorHandler(w, appErr, logger, "")
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}

// Usage in middleware chain
router.Use(errors.PanicRecovery(logger, "requestID"))
```

### Safe JSON Response

```go
// SafeJSONResponse writes JSON response with error handling
func SafeJSONResponse(w http.ResponseWriter, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(data); err != nil {
        // Logs error but prevents secondary error response
    }
}

// Usage
errors.SafeJSONResponse(w, map[string]string{"status": "ok"})
```

## 📋 JSON Response Format

### Success Response
```json
{
  "status": "success",
  "data": {
    "id": "123",
    "name": "file.jpg"
  }
}
```

### Error Response
```json
{
  "error": {
    "type": "ValidationError",
    "message": "invalid file size",
    "statusCode": 400,
    "details": {
      "maxSizeBytes": 52428800,
      "actualSizeBytes": 104857600
    }
  }
}
```

### Multiple Errors
```json
{
  "error": {
    "type": "ValidationError",
    "message": "multiple validation errors",
    "statusCode": 400,
    "details": {
      "errors": [
        {"field": "filename", "message": "required"},
        {"field": "size", "message": "too large"}
      ]
    }
  }
}
```

## 🔍 Usage Examples

### Basic Error Handling

```go
func ProcessFile(file *common.File) error {
    // Validation error
    if file.Filename == "" {
        return errors.NewValidationError(
            "filename is required",
            map[string]interface{}{"field": "filename"},
        )
    }
    
    // Size check error
    if file.Size > common.MaxFileSize50MB {
        return errors.NewPayloadTooLargeError(
            common.MaxFileSize50MB,
            file.Size,
        )
    }
    
    // System error
    data, err := processData(file.Content)
    if err != nil {
        return errors.NewSystemError("processing failed", err)
    }
    
    return nil
}
```

### HTTP Handler with Error Handling

```go
func UploadHandler(w http.ResponseWriter, r *http.Request) {
    logger := common.LoggerFromContext(r.Context())
    requestID := common.RequestIDFromContext(r.Context())
    
    // Parse request
    file, err := parseFileFromRequest(r)
    if err != nil {
        errors.HTTPErrorHandler(w, err, logger, requestID)
        return
    }
    
    // Process file
    result, err := processFile(file)
    if err != nil {
        errors.HTTPErrorHandler(w, err, logger, requestID)
        return
    }
    
    // Success response
    w.WriteHeader(http.StatusOK)
    errors.SafeJSONResponse(w, map[string]interface{}{
        "status": "success",
        "result": result,
    })
}
```

### Wrapped Handler

```go
func UploadHandler(w http.ResponseWriter, r *http.Request) error {
    file, err := parseFileFromRequest(r)
    if err != nil {
        return err // Wrapper handles conversion to HTTP response
    }
    
    result, err := processFile(file)
    if err != nil {
        return err // Wrapper handles conversion to HTTP response
    }
    
    w.WriteHeader(http.StatusOK)
    errors.SafeJSONResponse(w, map[string]interface{}{
        "status": "success",
        "result": result,
    })
    return nil
}

// Register with wrapper
router.HandleFunc("/upload", errors.WrapHandler(UploadHandler, logger))
```

## 🎯 Best Practices

### Error Context
Always include relevant context in errors:

```go
// Bad
return errors.New("processing failed")

// Good
return errors.NewFileProcessingError(
    filename,
    "EXIF extraction failed",
    err,
)
```

### Error Types
Use specific error types for different scenarios:

```go
// Validation errors for user input
if len(input) < 3 {
    return errors.NewValidationError("input too short", details)
}

// System errors for internal failures
if err := db.Save(data); err != nil {
    return errors.NewSystemError("database save failed", err)
}

// Not found errors for missing resources
if job == nil {
    return errors.NewNotFoundError("job", jobID)
}
```

### Error Logging
Errors are automatically logged by `HTTPErrorHandler`:

```go
// This logs the error with full context
errors.HTTPErrorHandler(w, err, logger, requestID)
```

### Original Error Preservation
AppError preserves the original error:

```go
appErr := errors.NewSystemError("failed to read", err)
// appErr.Err contains the original error for debugging
```

## 🧪 Testing

```bash
# Test errors package
go test ./file_server/advanced_file_operations/infrastructure/errors/

# Test with coverage
go test -cover ./file_server/advanced_file_operations/infrastructure/errors/

# Test error serialization
go test -run TestAppErrorJSON ./file_server/advanced_file_operations/infrastructure/errors/
```

## 📚 Related Documentation

- [Infrastructure Overview](../README.md)
- [Security Package](../security/README.md) - Uses error types for validation
- [Common Package](../common/README.md) - Error types reference common constants

## 🤝 Contributing

When adding new error types:

1. Add error type constant to `types.go`
2. Create constructor function following naming convention
3. Include appropriate HTTP status code
4. Add details map for context
5. Document with usage examples
6. Add tests for the new error type

## 📄 License

Copyright (c) 2025 FAZE3 DEVELOPMENT LLC. All rights reserved.
