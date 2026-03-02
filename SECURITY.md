# Security Policy and Analysis

## Reporting Security Issues

If you discover a security vulnerability in Go File Renamer Pro, please email us at **ole.abalo@faze3.dev**. Please do not create public GitHub issues for security vulnerabilities.

We will respond to security reports within 48 hours and work to address confirmed vulnerabilities as quickly as possible.

## Security Features

Go File Renamer Pro implements multiple layers of security:

### Input Validation & Sanitization
- **XSS Prevention**: All user inputs are sanitized using `go-sanitize` library
- **Path Traversal Protection**: Filenames and paths are validated and sanitized
- **Injection Prevention**: Input validation prevents SQL injection and command injection
- **File Type Validation**: Magic number detection verifies actual file types
- **Executable Detection**: Prevents processing of potentially malicious executable files

### File Security
- **Content-Type Verification**: Validates MIME types against whitelists
- **Size Limits**: Enforces maximum file sizes to prevent DoS attacks
- **Extension Validation**: Whitelist/blacklist-based extension checking
- **Secure Permissions**: Configuration files use restricted permissions (0600)

### Network Security
- **Rate Limiting**: IP-based request throttling prevents abuse
- **CORS Configuration**: Configurable allowed origins
- **Security Headers**: HSTS, X-Content-Type-Options, X-Frame-Options, etc.
- **Request Size Limits**: Maximum payload sizes enforced
- **Timeout Configuration**: Request timeouts prevent resource exhaustion

### Application Security
- **Structured Error Handling**: Prevents information leakage
- **Panic Recovery**: Graceful handling of unexpected errors
- **Context-Based Auth**: Request-scoped security context
- **Logging**: Comprehensive security event logging

## Known Security Considerations

### Desktop Application Context
This is primarily a desktop application that runs locally. The HTTP server component is designed for:
- Local development and testing
- Integration with the desktop UI via Wails
- Optional cloud deployment (with proper security configuration)

### ExifTool Dependency
The application relies on ExifTool for metadata operations:
- **Version**: ExifTool 13.39 (bundled for Windows)
- **Recommendation**: Keep ExifTool updated to latest version
- **Cross-Platform**: Use system-installed ExifTool when possible

### File Processing Security
When processing user files:
- Files are validated before processing
- Temporary files are created with secure permissions
- Cleanup is performed after operations
- Original files are preserved (unless explicitly deleted)

## Security Best Practices for Users

### Running the Application
1. **Keep Updated**: Always use the latest version
2. **Trusted Sources**: Only download from official sources
3. **File Permissions**: Review file access permissions
4. **Network Exposure**: Avoid exposing the HTTP API publicly
5. **Backup Files**: Maintain backups before bulk operations

### Configuration
1. **Allowed Origins**: Configure CORS carefully if exposing API
2. **Rate Limits**: Adjust based on your usage patterns
3. **File Size Limits**: Set appropriate limits for your use case
4. **Watch Directories**: Be careful with recursive watching

### File Operations
1. **Dry Run Mode**: Test operations before executing
2. **Review Logs**: Monitor operation logs for issues
3. **Metadata Privacy**: Use metadata removal for privacy-sensitive files
4. **Trusted Files**: Only process files from trusted sources

## Security Audit History

### 2025-10-27 - Initial Security Review
- Comprehensive code review completed
- Input validation verified
- Rate limiting implementation reviewed
- File security mechanisms validated
- Identified areas for improvement (see GitHub Issues)

**Key Findings:**
- Overall strong security posture
- Comprehensive input validation
- Effective rate limiting
- Some areas identified for hardening (documented in issues)

## Security Roadmap

### Planned Improvements
1. **Enhanced Command Execution Safety**
   - Migrate all exec.Command calls to use centralized service
   - Add command argument validation
   - Implement command whitelisting

2. **Profile Management Hardening**
   - Add comprehensive profile name validation
   - Implement profile integrity checking
   - Add profile backup/restore functionality

3. **Dependency Management**
   - Regular security updates for dependencies
   - Automated vulnerability scanning in CI/CD
   - Dependency version pinning

4. **Enhanced Logging**
   - Implement log sanitization
   - Add security event audit trail
   - Structured logging with field redaction

5. **Testing**
   - Security-focused unit tests
   - Penetration testing
   - Fuzz testing for input validation

## Compliance & Standards

Go File Renamer Pro follows security best practices from:
- OWASP Top 10 Web Application Security Risks
- CWE (Common Weakness Enumeration)
- Go Secure Coding Practices

## Third-Party Dependencies

Key security-relevant dependencies:
- `github.com/mrz1836/go-sanitize` - Input sanitization
- `github.com/unrolled/secure` - Security headers
- `github.com/go-playground/validator/v10` - Struct validation
- `golang.org/x/time/rate` - Rate limiting
- `github.com/barasher/go-exiftool` - ExifTool wrapper

All dependencies are regularly reviewed and updated.

## Security Contact

For security-related questions or concerns:
- **Email**: ole.abalo@faze3.dev
- **Response Time**: Within 48 hours
- **Disclosure Policy**: Coordinated vulnerability disclosure

## Acknowledgments

We appreciate security researchers who responsibly disclose vulnerabilities. Contributors will be acknowledged (with permission) in release notes.

---

Copyright (c) 2025 FAZE3 DEVELOPMENT LLC. All rights reserved.
