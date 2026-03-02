# Bulk File Processing Package

This package handles asynchronous processing of large file batches with comprehensive job management, progress tracking, and concurrent execution. It's designed for high-throughput file operations with robust error handling and result aggregation.

## 🎯 Features

- **Asynchronous Processing**: Non-blocking job submission and execution
- **Job Management**: Complete lifecycle tracking from submission to completion
- **Progress Tracking**: Real-time progress updates and status monitoring
- **Concurrent Execution**: Worker pool pattern for parallel file processing
- **Result Aggregation**: Combine results from multiple concurrent operations
- **Error Recovery**: Graceful handling of individual file failures
- **Job Storage**: Maintain job history and results
- **Resource Management**: Configurable limits and cleanup policies

## 📦 Key Components

### Service Interface

```go
type ServiceAPI interface {
    // ProcessBulkFiles submits a bulk processing job
    ProcessBulkFiles(ctx context.Context, userID string, files []common.File, options ProcessingOptions) (*ProcessingJob, error)
    
    // GetJob retrieves job status and results
    GetJob(jobID string) (*ProcessingJob, error)
    
    // GetProcessedData retrieves processed file content
    GetProcessedData(jobID, filename string) ([]byte, bool)
    
    // CancelJob cancels a running job
    CancelJob(jobID string) error
    
    // CleanupJob removes completed job data
    CleanupJob(jobID string) error
}
```

### Processing Options

```go
type ProcessingOptions struct {
    // Operation flags
    RenameFiles    bool   `json:"renameFiles"`
    RemoveMetadata bool   `json:"removeMetadata"`
    OptimizeFiles  bool   `json:"optimizeFiles"`
    CompressFiles  bool   `json:"compressFiles"`
    
    // Renaming configuration
    Pattern string                 `json:"pattern,omitempty"`
    Namer   string                 `json:"namer,omitempty"`
    Rename  RenameOperationOptions `json:"renameOptions"`
    
    // Validation
    AllowedTypes []string `json:"allowedTypes,omitempty"`
    MaxFileSize  int64    `json:"maxFileSize,omitempty"`
}

type RenameOperationOptions struct {
    Prefix      string `json:"prefix,omitempty"`
    Suffix      string `json:"suffix,omitempty"`
    StartNumber int    `json:"startNumber,omitempty"`
    Padding     int    `json:"padding,omitempty"`
    DateFormat  string `json:"dateFormat,omitempty"`
}

type SequentialNamingOptions struct {
    BaseName    string `json:"baseName"`
    StartNumber int    `json:"startNumber"`
    Padding     int    `json:"padding"`
    Extension   string `json:"extension,omitempty"`
}
```

### Processing Job

```go
type ProcessingJob struct {
    JobID          string                   `json:"jobId"`
    UserID         string                   `json:"userId"`
    Status         string                   `json:"status"` // queued, processing, completed, failed
    TotalFiles     int                      `json:"totalFiles"`
    ProcessedFiles int                      `json:"processedFiles"`
    SuccessCount   int                      `json:"successCount"`
    FailureCount   int                      `json:"failureCount"`
    Duration       int64                    `json:"durationMs"`
    Results        []ProcessingJobFileResult `json:"results"`
    CreatedAt      time.Time                `json:"createdAt"`
    CompletedAt    time.Time                `json:"completedAt,omitempty"`
    Error          string                   `json:"error,omitempty"`
}

type ProcessingJobFileResult struct {
    Filename    string `json:"filename"`
    NewName     string `json:"newName,omitempty"`
    Success     bool   `json:"success"`
    Error       string `json:"error,omitempty"`
    Action      string `json:"action,omitempty"`
    ContentType string `json:"contentType,omitempty"`
}
```

## 📁 Files

### `service.go`
Core service implementation with job processing logic.

**Key Functions:**
- `ProcessBulkFiles()`: Submit and execute bulk processing job
- `GetJob()`: Retrieve job status and results
- `GetProcessedData()`: Get processed file content
- `processFiles()`: Internal processing orchestration
- `processFileWithRetry()`: Retry logic for failed operations

### `handler.go`
HTTP request handlers for bulk processing endpoints.

**Endpoints:**
- `POST /api/bulk/process`: Submit bulk processing job
- `GET /api/bulk/jobs/{jobID}`: Get job status
- `GET /api/bulk/jobs/{jobID}/download`: Download processed files
- `DELETE /api/bulk/jobs/{jobID}`: Cancel or cleanup job

### `types.go`
Type definitions for processing options and results.

**Contents:**
- Processing option structures
- Job status enums
- Result aggregation types

## 🔧 Usage Examples

### Submit Bulk Processing Job

```go
import (
    "go-file-renamer-wails/file_server/advanced_file_operations/operations/file_operations/bulk_file_processing"
    "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
)

// Prepare files
files := []common.File{
    {Filename: "photo1.jpg", Content: data1, ContentType: "image/jpeg", Size: int64(len(data1))},
    {Filename: "photo2.jpg", Content: data2, ContentType: "image/jpeg", Size: int64(len(data2))},
    {Filename: "photo3.jpg", Content: data3, ContentType: "image/jpeg", Size: int64(len(data3))},
}

// Configure processing
options := bulk_file_processing.ProcessingOptions{
    RenameFiles:    true,
    RemoveMetadata: true,
    OptimizeFiles:  true,
    Pattern:        "IMG_%Y%m%d",
    Namer:          "sequential",
    Rename: bulk_file_processing.RenameOperationOptions{
        Prefix:      "vacation_",
        StartNumber: 1,
        Padding:     3,
    },
    AllowedTypes: []string{"image/jpeg", "image/png"},
    MaxFileSize:  common.MaxFileSize50MB,
}

// Submit job
job, err := bulkService.ProcessBulkFiles(ctx, userID, files, options)
if err != nil {
    return err
}

fmt.Printf("Job submitted: %s\n", job.JobID)
```

### Monitor Job Progress

```go
// Poll for job status
ticker := time.NewTicker(1 * time.Second)
defer ticker.Stop()

for {
    select {
    case <-ticker.C:
        job, err := bulkService.GetJob(jobID)
        if err != nil {
            return err
        }
        
        fmt.Printf("Status: %s, Progress: %d/%d\n", 
            job.Status, 
            job.ProcessedFiles, 
            job.TotalFiles,
        )
        
        if job.Status == "completed" || job.Status == "failed" {
            return nil
        }
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### Retrieve Processed Files

```go
// Get job results
job, err := bulkService.GetJob(jobID)
if err != nil {
    return err
}

// Download processed files
for _, result := range job.Results {
    if result.Success {
        data, ok := bulkService.GetProcessedData(jobID, result.Filename)
        if ok {
            // Save or transmit processed file
            err := os.WriteFile(result.NewName, data, 0644)
            if err != nil {
                log.Printf("Failed to save %s: %v\n", result.NewName, err)
            }
        }
    } else {
        log.Printf("File %s failed: %s\n", result.Filename, result.Error)
    }
}
```

### Sequential Naming

```go
options := bulk_file_processing.ProcessingOptions{
    RenameFiles: true,
    Namer:       "sequential",
    Rename: bulk_file_processing.RenameOperationOptions{
        Prefix:      "IMG_",
        StartNumber: 1,
        Padding:     4,
        Suffix:      "_final",
    },
}

// Result: IMG_0001_final.jpg, IMG_0002_final.jpg, IMG_0003_final.jpg, ...
```

### Date-Based Naming

```go
options := bulk_file_processing.ProcessingOptions{
    RenameFiles: true,
    Namer:       "date",
    Rename: bulk_file_processing.RenameOperationOptions{
        DateFormat: "2006-01-02_15-04-05",
        Prefix:     "photo_",
    },
}

// Result: photo_2025-01-27_10-30-45.jpg
```

## 🌐 HTTP API

### Submit Job

**Request:**
```http
POST /api/bulk/process
Content-Type: application/json

{
  "userId": "user-123",
  "files": [
    {
      "filename": "photo1.jpg",
      "contentBase64": "...",
      "contentType": "image/jpeg",
      "size": 204800
    }
  ],
  "options": {
    "renameFiles": true,
    "removeMetadata": true,
    "optimizeFiles": true,
    "pattern": "IMG_%Y%m%d",
    "namer": "sequential",
    "renameOptions": {
      "prefix": "vacation_",
      "startNumber": 1,
      "padding": 3
    }
  }
}
```

**Response:**
```json
{
  "jobId": "job-abc123",
  "status": "processing",
  "totalFiles": 3,
  "processedFiles": 0,
  "successCount": 0,
  "failureCount": 0,
  "createdAt": "2025-01-27T10:30:00Z"
}
```

### Get Job Status

**Request:**
```http
GET /api/bulk/jobs/job-abc123
```

**Response:**
```json
{
  "jobId": "job-abc123",
  "userId": "user-123",
  "status": "completed",
  "totalFiles": 3,
  "processedFiles": 3,
  "successCount": 3,
  "failureCount": 0,
  "durationMs": 2345,
  "results": [
    {
      "filename": "photo1.jpg",
      "newName": "vacation_001.jpg",
      "success": true,
      "action": "renamed, metadata removed, optimized",
      "contentType": "image/jpeg"
    }
  ],
  "createdAt": "2025-01-27T10:30:00Z",
  "completedAt": "2025-01-27T10:30:02Z"
}
```

## 🏗️ Architecture

### Worker Pool Pattern

```
┌─────────────────┐
│  Job Submission │
└────────┬────────┘
         │
         ▼
┌─────────────────────┐
│   Job Queue         │
└────────┬────────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌────────┐ ┌────────┐
│Worker 1│ │Worker 2│  ... Worker N
└────┬───┘ └────┬───┘
     │          │
     └─────┬────┘
           │
           ▼
    ┌──────────────┐
    │ Result Store │
    └──────────────┘
```

### Processing Flow

```
1. Job Submission
   ├─ Validate files and options
   ├─ Generate job ID
   ├─ Create job record
   └─ Queue for processing

2. Job Execution
   ├─ Initialize worker pool
   ├─ Distribute files to workers
   ├─ Process files concurrently
   │  ├─ Apply operations in order
   │  ├─ Handle errors gracefully
   │  └─ Store results
   └─ Aggregate results

3. Job Completion
   ├─ Update job status
   ├─ Calculate statistics
   ├─ Store final results
   └─ Cleanup resources
```

## 🔒 Security & Validation

### File Validation
```go
// Validate file count
if len(files) > MaxFilesPerBatch {
    return errors.NewValidationError("too many files", len(files))
}

// Validate individual files
for _, file := range files {
    if file.Size > options.MaxFileSize {
        return errors.NewPayloadTooLargeError(options.MaxFileSize, file.Size)
    }
    
    if !isAllowedType(file.ContentType, options.AllowedTypes) {
        return errors.NewValidationError("invalid file type", file.ContentType)
    }
}
```

### Rate Limiting
Applied at handler level to prevent abuse:
```go
router.Use(security.RateLimitingMiddleware(uploadLimiter, logger))
```

### Resource Limits
```go
const (
    MaxFilesPerBatch      = 1000
    MaxConcurrentWorkers  = 10
    MaxJobRetention       = 24 * time.Hour
    MaxStoredJobs         = 10000
)
```

## 📊 Performance

### Concurrency Configuration
```go
// Adjust worker pool size based on load
workerCount := runtime.NumCPU() * 2

// Process files concurrently
var wg sync.WaitGroup
workers := make(chan struct{}, workerCount)

for _, file := range files {
    wg.Add(1)
    workers <- struct{}{}
    
    go func(f common.File) {
        defer wg.Done()
        defer func() { <-workers }()
        
        result := processFile(f, options)
        results <- result
    }(file)
}

wg.Wait()
```

### Memory Management
- Stream large files to avoid loading entirely in memory
- Clean up processed data after download
- Implement job result TTL
- Use background goroutines for cleanup

## 🧪 Testing

```bash
# Unit tests
go test ./file_server/advanced_file_operations/operations/file_operations/bulk_file_processing/

# Integration tests
go test -tags=integration ./file_server/advanced_file_operations/operations/file_operations/bulk_file_processing/

# Load tests
go test -bench=. ./file_server/advanced_file_operations/operations/file_operations/bulk_file_processing/
```

## 📚 Related Documentation

- [Operations Overview](../README.md)
- [Limited File Processing](../limited_file_processing/README.md)
- [Optimizer Package](../optimizer/README.md)
- [Renamers Package](../renamers/README.md)

## 🤝 Contributing

When enhancing bulk processing:

1. Maintain backwards compatibility for job structures
2. Add proper error handling for all operations
3. Implement retry logic for transient failures
4. Write comprehensive tests including edge cases
5. Document new processing options
6. Consider performance impact of changes

## 📄 License

Copyright (c) 2025 FAZE3 DEVELOPMENT LLC. All rights reserved.
