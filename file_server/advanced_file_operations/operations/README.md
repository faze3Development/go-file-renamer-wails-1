# Operations Package

This package implements the core business logic and orchestration layer for all file processing operations. It provides a unified interface for file manipulation, metadata operations, optimization, and bulk processing through a centralized orchestrator pattern.

## 🎯 Features

### File Operations
- **Bulk File Processing**: Asynchronous processing of large file batches with job tracking
- **Limited File Processing**: Synchronous processing for small, immediate operations
- **File Optimization**: Image compression and optimization with concurrent processing
- **File Renaming**: Multiple naming strategies and pattern-based renaming

### Metadata Operations
- **Metadata Extraction**: Extract EXIF, XMP, and other metadata from 100+ file formats
- **Metadata Removal**: Strip all metadata for privacy and security
- **Metadata Writing**: Add custom metadata tags to files
- **Metadata-Based Renaming**: Rename files using extracted metadata

### Monitoring & Statistics
- **Job Status Tracking**: Real-time job progress and status monitoring
- **Statistics Collection**: Processing metrics and performance data
- **Job History**: Complete audit trail of file operations

## 📦 Architecture

### Orchestrator Pattern

The `Orchestrator` serves as the central coordination point for all operations:

```go
type Orchestrator struct {
    // Core services
    JobStatusService          monitoring_and_statistics.ServiceAPI
    BulkFileProcessingService bulk_file_processing.ServiceAPI
    ExifToolService           exiftool.Service
    
    // Feature services
    MetadataExtractionService *metadata_extraction.Service
    MetadataRemovalService    *metadata_removal.Service
    MetadataRenamerService    *metadata_renamer.Service
    OptimizerService          *optimizer.Service
    
    // Feature handlers (HTTP endpoints)
    MetadataRemovalHandler *metadata_removal.Handler
    MetadataRenamerHandler *metadata_renamer.Handler
    MetadataWriterHandler  *metadata_writer.Handler
    OptimizerHandler       *optimizer.Handler
    RenamerHandler         *renamers.Handler
}
```

### Service Initialization

```go
func NewOrchestrator(exiftoolService exiftool.Service, logger *slog.Logger) *Orchestrator {
    // Initialize monitoring services
    jobStatusService := monitoring_and_statistics.NewService()
    
    // Initialize processing services
    bulkProcessingService := bulk_file_processing.NewService(
        jobStatusService, 
        exiftoolService, 
        logger,
    )
    
    // Initialize feature services
    metadataExtractionService := metadata_extraction.NewService(exiftoolService, logger)
    metadataRemovalService := metadata_removal.NewService(exiftoolService, logger)
    metadataRenamerService := metadata_renamer.NewService(metadataExtractionService, logger)
    optimizerService := optimizer.NewService(logger)
    
    // Initialize HTTP handlers
    metadataRemovalHandler := metadata_removal.NewHandler(metadataRemovalService, logger)
    // ... other handlers
    
    return &Orchestrator{
        JobStatusService:          jobStatusService,
        BulkFileProcessingService: bulkProcessingService,
        // ... other services and handlers
    }
}
```

## 📁 Sub-Packages

### File Operations (`file_operations/`)

#### Bulk File Processing
Handles asynchronous processing of large file batches:
- Job queuing and management
- Progress tracking
- Concurrent processing
- Result aggregation
- Error handling and recovery

**See:** [Bulk File Processing README](file_operations/bulk_file_processing/README.md)

#### Limited File Processing
Handles synchronous processing of small file sets:
- Immediate processing
- No job tracking overhead
- Direct response
- Suitable for real-time operations

**See:** [Limited File Processing README](file_operations/limited_file_processing/README.md)

#### Optimizer
File compression and optimization:
- Image optimization (JPEG, PNG, WebP)
- Concurrent processing with goroutines
- Quality vs. size trade-offs
- Format conversion support

**See:** [Optimizer README](file_operations/optimizer/README.md)

#### Renamers
Various file naming strategies:
- Sequential numbering
- Timestamp-based naming
- Random string generation
- Pattern-based templates
- Metadata extraction

**See:** [Renamers README](file_operations/renamers/README.md)

### Metadata Operations (`metadata_Operations/`)

#### Metadata Extraction
Extract metadata from files:
- EXIF data from images
- Document properties from PDFs
- Video codec information
- Audio tags
- Archive metadata

**See:** [Metadata Extraction README](metadata_Operations/metadata_extraction/README.md)

#### Metadata Removal
Strip metadata for privacy:
- Remove EXIF from images
- Clean PDF metadata
- Strip video metadata
- Remove audio tags
- Sanitize archives

**See:** [Metadata Removal README](metadata_Operations/metadata_removal/README.md)

#### Metadata Writer
Add custom metadata to files:
- Write EXIF tags
- Set PDF properties
- Add custom fields
- Batch metadata updates

**See:** [Metadata Writer README](metadata_Operations/metadata_writer/README.md)

#### Metadata Renamer
Rename files based on extracted metadata:
- Date-based naming from EXIF
- Camera model in filename
- GPS location naming
- Custom metadata templates

**See:** [Metadata Renamer README](metadata_Operations/metadata_renamer/README.md)

### Monitoring & Statistics (`monitoring_and_statistics/`)

Job tracking and performance metrics:
- Real-time job status
- Progress tracking
- Error tracking
- Performance statistics
- Job history

## 🔧 Usage Examples

### Using the Orchestrator

```go
import (
    "go-file-renamer-wails/file_server/advanced_file_operations/operations"
    "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/exiftool"
)

// Initialize ExifTool service
exiftoolConfig := exiftool.DefaultConfig()
exiftoolService, err := exiftool.NewService(exiftoolConfig, logger)
if err != nil {
    log.Fatal(err)
}
defer exiftoolService.Close()

// Create orchestrator
orchestrator := operations.NewOrchestrator(exiftoolService, logger)

// Access services through orchestrator
bulkService := orchestrator.BulkFileProcessingService
metadataService := orchestrator.MetadataExtractionService
```

### Bulk File Processing

```go
// Prepare files
files := []common.File{
    {Filename: "photo1.jpg", Content: data1, Size: len(data1)},
    {Filename: "photo2.jpg", Content: data2, Size: len(data2)},
}

// Configure processing options
options := bulk_file_processing.ProcessingOptions{
    RenameFiles:    true,
    RemoveMetadata: true,
    OptimizeFiles:  true,
    Pattern:        "IMG_%Y%m%d",
    Namer:          "sequential",
}

// Submit job
result, err := orchestrator.BulkFileProcessingService.ProcessBulkFiles(
    ctx,
    "user-123",
    files,
    options,
)
if err != nil {
    return err
}

// Check results
fmt.Printf("Job ID: %s\n", result.JobID)
fmt.Printf("Processed: %d/%d\n", result.SuccessCount, result.TotalFiles)
```

### Metadata Operations

```go
// Extract metadata
file := &common.File{
    Filename:    "photo.jpg",
    Content:     imageData,
    ContentType: "image/jpeg",
    Size:        int64(len(imageData)),
}

metadata, err := orchestrator.MetadataExtractionService.ExtractMetadata(ctx, file)
if err != nil {
    return err
}

// Access metadata
if cameraMake, ok := metadata["Make"]; ok {
    fmt.Printf("Camera: %v\n", cameraMake)
}

// Remove metadata
err = orchestrator.MetadataRemovalService.RemoveAllMetadata(ctx, file)
if err != nil {
    return err
}
```

### File Optimization

```go
// Optimize images
files := []common.File{
    {Filename: "large.jpg", Content: largeImageData},
}

results, err := orchestrator.OptimizerService.OptimizeFiles(ctx, files, options)
if err != nil {
    return err
}

for _, result := range results {
    if result.Success {
        fmt.Printf("Reduced %s from %d to %d bytes\n", 
            result.Filename, 
            result.OriginalSize, 
            result.OptimizedSize,
        )
    }
}
```

### Job Tracking

```go
// Submit a job
jobID := result.JobID

// Check job status
job, err := orchestrator.BulkFileProcessingService.GetJob(jobID)
if err != nil {
    return err
}

fmt.Printf("Status: %s\n", job.Status)
fmt.Printf("Progress: %d/%d files\n", job.ProcessedFiles, job.TotalFiles)
fmt.Printf("Success: %d, Failed: %d\n", job.SuccessCount, job.FailureCount)

// Get job statistics
stats := orchestrator.JobStatusService.GetStatistics()
fmt.Printf("Total jobs: %d\n", stats.TotalJobs)
fmt.Printf("Active jobs: %d\n", stats.ActiveJobs)
```

## 🌐 HTTP Routes

Routes are registered in `routes.go`:

```go
func RegisterRoutes(router *chi.Mux, orchestrator *Orchestrator, logger *slog.Logger) {
    // Bulk processing
    router.Post("/api/bulk/process", bulkHandler.ProcessFiles)
    router.Get("/api/bulk/jobs/{jobID}", bulkHandler.GetJob)
    
    // Metadata operations
    router.Post("/api/metadata/extract", metadataHandler.Extract)
    router.Post("/api/metadata/remove", metadataHandler.Remove)
    router.Post("/api/metadata/write", metadataHandler.Write)
    
    // Optimization
    router.Post("/api/optimize", optimizerHandler.Optimize)
    
    // Renaming
    router.Get("/api/renamers/patterns", renamerHandler.GetPatterns)
    router.Post("/api/rename", renamerHandler.Rename)
    
    // Job monitoring
    router.Get("/api/jobs/{jobID}", jobHandler.GetJob)
    router.Get("/api/jobs/{jobID}/status", jobHandler.GetStatus)
    router.Get("/api/statistics", jobHandler.GetStatistics)
}
```

## 🏗️ Processing Pipeline

### Standard Pipeline

```
┌──────────────┐
│  File Input  │
└──────┬───────┘
       │
       ▼
┌──────────────────┐
│   Validation     │ ◄── Security checks, size limits
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│  Metadata Ops    │ ◄── Extract, Remove, or Write
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│   Rename/Move    │ ◄── Apply naming strategy
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│   Optimization   │ ◄── Compress, optimize
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│  Result Output   │
└──────────────────┘
```

### Bulk Processing Pipeline

```
┌──────────────┐
│  Submit Job  │
└──────┬───────┘
       │
       ▼
┌──────────────────┐
│   Create Job     │ ◄── Generate job ID, initialize
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│  Queue Files     │ ◄── Partition into batches
└──────┬───────────┘
       │
       ▼
┌──────────────────────────┐
│  Process Concurrently    │ ◄── Worker pool
│  (goroutines)            │
└──────┬───────────────────┘
       │
       ▼
┌──────────────────┐
│  Aggregate       │ ◄── Combine results
│  Results         │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│  Update Status   │ ◄── Mark complete, update stats
└──────────────────┘
```

## 🔒 Security Considerations

### Input Validation
All operations validate inputs through security package:
- File type verification
- Size limit enforcement
- Filename sanitization
- Content validation

### Rate Limiting
Applied at the handler level:
- Per-IP rate limiting
- Different limits for different operations
- Burst support for legitimate use

### Error Handling
Consistent error handling across all operations:
- Structured error types
- Detailed error context
- Safe error messages to clients
- Full error logging

## 📊 Performance

### Concurrency
- Bulk processing uses worker pools
- Configurable concurrency limits
- Efficient resource utilization
- Graceful degradation under load

### Optimization
- Reuse ExifTool process (stay_open)
- Connection pooling for database
- Caching of frequently accessed data
- Efficient memory management

### Monitoring
- Real-time performance metrics
- Job duration tracking
- Resource usage monitoring
- Error rate tracking

## 🧪 Testing

```bash
# Test all operations
go test ./file_server/advanced_file_operations/operations/...

# Test specific package
go test ./file_server/advanced_file_operations/operations/file_operations/bulk_file_processing/

# Test with coverage
go test -cover ./file_server/advanced_file_operations/operations/...

# Integration tests
go test -tags=integration ./file_server/advanced_file_operations/operations/...
```

## 📚 Related Documentation

- [Backend API Reference](../docs/backend-api-reference.md)
- [Infrastructure Package](../infrastructure/README.md)
- [ExifTool Service](../infrastructure/exiftool/README.md)
- [Security Package](../infrastructure/security/README.md)

## 🤝 Contributing

When adding new operations:

1. Follow the existing service pattern
2. Use the orchestrator for coordination
3. Implement proper error handling
4. Add comprehensive logging
5. Write unit and integration tests
6. Update route registration in `routes.go`
7. Document the new operation
8. Update this README

## 📄 License

Copyright (c) 2025 FAZE3 DEVELOPMENT LLC. All rights reserved.
