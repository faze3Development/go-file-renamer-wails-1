# Security Package

This package provides comprehensive security features to protect the application from common attacks and abuse. It implements defense-in-depth with multiple layers of protection including input validation, rate limiting, file security checks, and request sanitization.

## 🛡️ Security Features

### Input Sanitization
- **XSS Prevention**: Removes cross-site scripting attack vectors using `go-sanitize`
- **HTML Stripping**: Removes HTML tags from user input
- **SQL Injection Prevention**: Sanitizes inputs to prevent SQL injection
- **Path Traversal Protection**: Validates and sanitizes file paths
- **Control Character Removal**: Strips dangerous control characters
- **Length Limits**: Enforces maximum input lengths

### File Validation
- **Content-Type Verification**: Validates MIME types against whitelists
- **Magic Number Detection**: Verifies actual file type from content
- **Executable Detection**: Prevents processing of executable files
- **Extension Validation**: Whitelist and blacklist-based extension checking
- **Size Limits**: Enforces individual and total file size restrictions
- **Filename Sanitization**: Cleans dangerous characters from filenames

### Rate Limiting
- **IP-Based Limiting**: Per-IP request throttling
- **Configurable Rates**: Separate limits for uploads vs. API calls
- **Burst Support**: Allows temporary bursts while maintaining limits
- **Automatic Cleanup**: Removes inactive rate limiters from memory
- **Context-Aware**: Different limits for different endpoints

### Request Validation
- **Multipart Form Parsing**: Secure parsing of file uploads
- **File Count Limits**: Prevents processing too many files
- **Payload Size Limits**: Enforces maximum request sizes
- **Configuration Validation**: Validates processing options
- **Profile Name Validation**: Sanitizes and validates profile names

## 📁 Files

### `validation.go`
Provides comprehensive input and file validation functions.

**Key Functions:**
- `SanitizeInput(string) string`: Sanitizes user text input
- `SanitizeFilename(string) string`: Sanitizes filenames and paths
- `ValidateFileType(string) bool`: Validates content types
- `CheckForInjection(string) bool`: Detects injection attempts
- `ValidateFileCount(int, int) error`: Validates file counts
- `ValidateIndividualFileSize(int64, int64) error`: Validates file sizes
- `ValidateFile(*http.Request, *FileValidationConfig) (*multipart.FileHeader, error)`: Complete file validation
- `ValidateStruct(interface{}) error`: Struct validation using `go-playground/validator`
- `ValidateProfileName(string) (string, error)`: Profile name validation

**Request Validator:**
```go
type RequestValidator struct {
    logger *slog.Logger
}

// Key Methods:
- ValidateAndParseMultipartRequest(r *http.Request, maxPayloadSize int64) ([]common.File, map[string]interface{}, error)
- ValidateJobRequest(fileCount int) error
```

**File Validation Config:**
```go
type FileValidationConfig struct {
    MaxSizeBytes int64       // Maximum file size
    AllowedTypes []string    // Whitelisted MIME types
    BlockedTypes []string    // Blacklisted MIME types
    AllowedExts  []string    // Whitelisted extensions
    BlockedExts  []string    // Blacklisted extensions
}
```

### `ratelimit.go`
Implements IP-based rate limiting middleware.

**Key Components:**
- `IPRateLimiter`: Thread-safe rate limiter with automatic cleanup
- `RateLimitingMiddleware(limiter, logger)`: HTTP middleware
- `CreateFileUploadRateLimiter()`: Specialized for file uploads
- `CreateAPIRateLimiter()`: Specialized for API endpoints

**IPRateLimiter Methods:**
```go
type IPRateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
    cleanup  time.Duration
}

// Key Methods:
- Allow(ip string) bool: Check if request is allowed
- getLimiter(ip string) *rate.Limiter: Get/create limiter for IP
```

## 🔧 Usage Examples

### Input Sanitization

```go
import "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/security"

// Sanitize user input
userInput := "<script>alert('xss')</script>Hello"
clean := security.SanitizeInput(userInput)
// Result: "Hello"

// Sanitize filename
filename := "../../etc/passwd"
safe := security.SanitizeFilename(filename)
// Result: "etc_passwd" (path separators removed)

// Check for injection
malicious := "'; DROP TABLE users; --"
if security.CheckForInjection(malicious) {
    // Handle potential injection attack
}
```

### File Validation

```go
import "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/security"

// Create validation config
config := security.DefaultFileConfig()
// Or custom config:
config := &security.FileValidationConfig{
    MaxSizeBytes: 50 * 1024 * 1024, // 50MB
    AllowedTypes: []string{"image/jpeg", "image/png"},
    AllowedExts:  []string{".jpg", ".jpeg", ".png"},
    BlockedExts:  []string{".exe", ".bat", ".sh"},
}

// Validate file from HTTP request
fileHeader, err := security.ValidateFile(r, config)
if err != nil {
    // Handle validation error
    return err
}

// File is validated and safe to process
file, _ := fileHeader.Open()
defer file.Close()
```

### Request Validation

```go
import "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/security"

func handleUpload(w http.ResponseWriter, r *http.Request) {
    // Create validator
    validator := security.NewRequestValidator(logger)
    
    // Validate and parse request
    files, config, err := validator.ValidateAndParseMultipartRequest(
        r,
        common.MaxPayload32MB,
    )
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Validate file count
    if err := validator.ValidateJobRequest(len(files)); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Process validated files...
}
```

### Rate Limiting

```go
import "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/security"

// Create rate limiters
uploadLimiter := security.CreateFileUploadRateLimiter()
apiLimiter := security.CreateAPIRateLimiter()

// Apply as middleware
router.Use(security.RateLimitingMiddleware(apiLimiter, logger))

// Or check manually
ip := getClientIP(r)
if !uploadLimiter.Allow(ip) {
    http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
    return
}
```

### Custom Rate Limiter

```go
import (
    "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/security"
    "golang.org/x/time/rate"
)

// Create custom rate limiter
// Allow 10 requests per second with burst of 20
limiter := security.NewIPRateLimiter(
    rate.Limit(10),  // 10 rps
    20,              // burst size
)

// Use in middleware
router.Use(security.RateLimitingMiddleware(limiter, logger))
```

## 🔒 Security Best Practices

### Input Handling
1. **Always sanitize**: Sanitize all user inputs before processing
2. **Validate types**: Use structured validation for complex inputs
3. **Check injection**: Verify inputs don't contain injection patterns
4. **Limit length**: Enforce maximum lengths on all text inputs

### File Handling
1. **Validate content**: Don't trust file extensions or MIME types
2. **Check magic numbers**: Verify actual file content
3. **Enforce limits**: Restrict file sizes and counts
4. **Block executables**: Never process executable files
5. **Sanitize names**: Clean all filenames before storage

### Request Processing
1. **Rate limit**: Apply rate limiting to all endpoints
2. **Size limits**: Enforce maximum request sizes
3. **Validate early**: Fail fast on invalid requests
4. **Log suspicious**: Log potential security issues
5. **Context propagation**: Pass security context through request chain

## ⚙️ Configuration

### Environment Variables

```bash
# Rate limiting configuration
UPLOAD_RATE_LIMIT_RPS=0.083      # ~5 requests per minute
UPLOAD_RATE_LIMIT_BURST=5        # Allow burst of 5

API_RATE_LIMIT_RPS=1             # 60 requests per minute
API_RATE_LIMIT_BURST=60          # Allow burst of 60
```

### Default Limits

```go
// From common/common.go
MaxFileSize10MB  = 10 * 1024 * 1024   // Individual file limit
MaxFileSize50MB  = 50 * 1024 * 1024   // Large file limit
MaxPayload32MB   = 32 * 1024 * 1024   // Request payload limit
```

### Allowed File Types

Default configuration from `DefaultFileConfig()`:

**Allowed MIME Types:**
- `image/jpeg`
- `image/png`
- `image/gif`
- `image/webp`
- `application/pdf`

**Blocked MIME Types:**
- `application/x-executable`
- `application/x-msdownload`
- `application/x-script`

**Blocked Extensions:**
- `.exe`, `.bat`, `.cmd`, `.com`
- `.pif`, `.scr`, `.vbs`, `.js`
- `.jar`

## 🔍 Threat Model

### Threats Mitigated

1. **Cross-Site Scripting (XSS)**
   - Mitigation: Input sanitization, HTML stripping
   - Status: ✅ Protected

2. **SQL Injection**
   - Mitigation: Input sanitization, parameterized queries
   - Status: ✅ Protected

3. **Path Traversal**
   - Mitigation: Filename sanitization, path validation
   - Status: ✅ Protected

4. **Malicious File Uploads**
   - Mitigation: Magic number detection, executable blocking
   - Status: ✅ Protected

5. **Denial of Service (DoS)**
   - Mitigation: Rate limiting, size limits, file count limits
   - Status: ✅ Protected

6. **Content-Type Spoofing**
   - Mitigation: Magic number verification
   - Status: ✅ Protected

7. **Resource Exhaustion**
   - Mitigation: Automatic cleanup, size limits
   - Status: ✅ Protected

## 🧪 Testing

```bash
# Test security package
go test ./file_server/advanced_file_operations/infrastructure/security/

# Test with race detector
go test -race ./file_server/advanced_file_operations/infrastructure/security/

# Test with coverage
go test -cover ./file_server/advanced_file_operations/infrastructure/security/
```

## 📚 Dependencies

- `github.com/go-playground/validator/v10`: Struct validation
- `github.com/mrz1836/go-sanitize`: Input sanitization
- `golang.org/x/time/rate`: Rate limiting implementation

## 🐛 Known Limitations

1. **Rate Limiter Memory**: Inactive limiters are cleaned up every 10 minutes, but very high IP diversity could increase memory usage
2. **Magic Number Detection**: Limited to common executable signatures
3. **Content Validation**: Deep content inspection (e.g., steganography) is not performed

## 🤝 Contributing

When enhancing security:

1. Add tests for new validation rules
2. Document threat model changes
3. Update configuration examples
4. Consider performance impact
5. Review with security-focused mindset

## 📄 License

Copyright (c) 2025 FAZE3 DEVELOPMENT LLC. All rights reserved.
