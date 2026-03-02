// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package monitoring_and_statistics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Job represents a processing job that can contain multiple file operations
type Job struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"userId"`
	Type         string                 `json:"type"`
	Status       string                 `json:"status"` // "queued", "processing", "completed", "failed"
	Config       map[string]interface{} `json:"config"`
	CreatedAt    time.Time              `json:"createdAt"`
	StartedAt    *time.Time             `json:"startedAt,omitempty"`
	CompletedAt  *time.Time             `json:"completedAt,omitempty"`
	Duration     *int64                 `json:"durationMs,omitempty"`
	SuccessCount int                    `json:"successCount"`
	FailureCount int                    `json:"failureCount"`
	Error        string                 `json:"error,omitempty"`
}

// JobStats represents job processing statistics
type JobStats struct {
	QueueLength       int       `json:"queueLength"`
	Processing        int       `json:"processing"`
	Completed         int       `json:"completed"`
	Failed            int       `json:"failed"`
	AvgProcessingTime float64   `json:"avgProcessingTimeMs"`
	SuccessRate       float64   `json:"successRate"`
	ActiveJobs        []Job     `json:"activeJobs"`
	RecentJobs        []Job     `json:"recentJobs"`
	LastUpdated       time.Time `json:"lastUpdated"`
}

// ServiceAPI defines the interface for the monitoring and statistics service.
type ServiceAPI interface {
	QueueJob(userID, jobType string, fileCount int, config map[string]interface{}) *Job
	StartJob(id string) (*Job, bool)
	CompleteJob(id string, successCount, failureCount int)
	FailJob(id string, errorMsg string)
	GetStats() JobStats
	GetActiveJobs() []*Job
	GetRecentJobs() []Job
	ProcessJob(ctx context.Context, id string)
	EstimateJobProgress(job *Job) int
}

// Service manages processing jobs and statistics
type Service struct {
	mu            sync.RWMutex
	jobs          map[string]*Job
	jobFileCounts map[string]int // Store file counts for jobs
	activeJobs    []*Job
	recentJobs    []*Job
	maxRecentJobs int
}

// NewService creates a new processing service
func NewService() *Service {
	return &Service{
		jobs:          make(map[string]*Job),
		jobFileCounts: make(map[string]int),
		activeJobs:    make([]*Job, 0),
		recentJobs:    make([]*Job, 0),
		maxRecentJobs: 10,
	}
}

// QueueJob creates a new job and puts it in the queue.
func (s *Service) QueueJob(userID, jobType string, fileCount int, config map[string]interface{}) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := &Job{
		ID:        generateJobID(),
		UserID:    userID,
		Type:      jobType,
		Status:    "queued", // Start as Queued
		Config:    config,
		CreatedAt: time.Now(),
	}
	s.jobFileCounts[job.ID] = fileCount

	s.jobs[job.ID] = job
	s.activeJobs = append(s.activeJobs, job) // A "queued" job is still considered active

	return job
}

// StartJob marks a queued job as processing. This would be called by a worker.
func (s *Service) StartJob(id string) (*Job, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists || job.Status != "queued" {
		return nil, false
	}

	now := time.Now()
	job.Status = "processing"
	job.StartedAt = &now
	return job, true
}

// CompleteJob completes a processing job
func (s *Service) CompleteJob(id string, successCount, failureCount int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return
	}

	now := time.Now()
	var duration int64
	if job.StartedAt != nil {
		duration = now.Sub(*job.StartedAt).Milliseconds()
	}

	job.Status = "completed"
	job.CompletedAt = &now
	job.Duration = &duration
	job.SuccessCount = successCount
	job.FailureCount = failureCount

	// Move from active to recent
	s.moveJobToRecent(job)
}

// FailJob marks a job as failed
func (s *Service) FailJob(id string, errorMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return
	}

	now := time.Now()
	var duration int64
	if job.StartedAt != nil {
		duration = now.Sub(*job.StartedAt).Milliseconds()
	}

	job.Status = "failed"
	job.CompletedAt = &now
	job.Duration = &duration
	job.Error = errorMsg
	if fileCount, ok := s.jobFileCounts[job.ID]; ok {
		job.FailureCount = fileCount // Assume all operations failed if job fails
	} else {
		job.FailureCount = 0
	}

	// Move from active to recent
	s.moveJobToRecent(job)
}

// moveJobToRecent moves a job from active to recent jobs list
func (s *Service) moveJobToRecent(job *Job) {
	// Remove from active jobs
	for i, activeJob := range s.activeJobs {
		if activeJob.ID == job.ID {
			s.activeJobs = append(s.activeJobs[:i], s.activeJobs[i+1:]...)
			break
		}
	}

	// Add to recent jobs
	s.recentJobs = append([]*Job{job}, s.recentJobs...)

	// Keep only the most recent jobs
	if len(s.recentJobs) > s.maxRecentJobs {
		s.recentJobs = s.recentJobs[:s.maxRecentJobs]
	}
}

// EstimateJobProgress calculates rough progress percentage for a job
func (s *Service) EstimateJobProgress(job *Job) int {
	if job == nil {
		return 0
	}

	// If job is completed or failed, return 100%
	if job.Status == "completed" || job.Status == "failed" {
		return 100
	}

	// If job hasn't started yet, return 5% (queued)
	if job.Status == "queued" || job.StartedAt == nil {
		return 5
	}

	// Calculate time-based progress for processing jobs
	timeElapsed := time.Since(*job.StartedAt)

	// Estimate: most jobs complete within 30 seconds to 2 minutes
	// Start at 10% and progress to 95% over time
	var estimatedTotalDuration time.Duration
	if job.Type == "bulk" {
		// Bulk jobs: estimate 1-3 minutes depending on file count
		fileCount := 1 // default
		if count, ok := s.jobFileCounts[job.ID]; ok && count > 0 {
			fileCount = count
		}
		// Estimate 2-5 seconds per file, minimum 30 seconds
		estimatedSeconds := fileCount * 3
		if estimatedSeconds < 30 {
			estimatedSeconds = 30
		}
		estimatedTotalDuration = time.Duration(estimatedSeconds) * time.Second
	} else {
		// Other jobs: 30 seconds default
		estimatedTotalDuration = 30 * time.Second
	}

	// Calculate progress: 10% + (85% * time_elapsed / estimated_total)
	progress := 10 + int(85*timeElapsed.Seconds()/estimatedTotalDuration.Seconds())

	// Clamp to 95% until actually completed
	if progress > 95 {
		progress = 95
	}
	if progress < 10 {
		progress = 10
	}

	return progress
}

// GetStats returns current processing statistics
func (s *Service) GetStats() JobStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	queueLength := 0
	for _, job := range s.activeJobs {
		if count, ok := s.jobFileCounts[job.ID]; ok {
			queueLength += count
		}
	}

	// To avoid race conditions, we create copies of the jobs for the stats object.
	// We also pre-allocate slices to the correct capacity.
	activeJobsCopy := make([]Job, 0, len(s.activeJobs))
	for _, job := range s.activeJobs {
		activeJobsCopy = append(activeJobsCopy, *job)
	}

	recentJobsCopy := make([]Job, 0, len(s.recentJobs))
	stats := JobStats{
		QueueLength: queueLength,
		Processing:  len(s.activeJobs),
		ActiveJobs:  activeJobsCopy,
		LastUpdated: time.Now(),
	}

	// Convert recent jobs and calculate totals
	totalJobs := 0
	totalDuration := int64(0)
	// Note: These stats are calculated based on the `maxRecentJobs` window.

	for _, job := range s.recentJobs {
		recentJobsCopy = append(recentJobsCopy, *job)

		switch job.Status {
		case "completed":
			stats.Completed++
			if job.Duration != nil {
				totalDuration += *job.Duration
			}
		case "failed":
			stats.Failed++
		}
		totalJobs++
	}

	stats.RecentJobs = recentJobsCopy

	// Calculate average processing time and success rate
	if stats.Completed > 0 {
		stats.AvgProcessingTime = float64(totalDuration) / float64(stats.Completed)
	}

	if totalJobs > 0 {
		stats.SuccessRate = float64(stats.Completed) / float64(totalJobs) * 100
	}

	return stats
}

// GetActiveJobs returns currently active jobs
func (s *Service) GetActiveJobs() []*Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return append(make([]*Job, 0, len(s.activeJobs)), s.activeJobs...)
}

// GetRecentJobs returns recently completed jobs
func (s *Service) GetRecentJobs() []Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]Job, 0, len(s.recentJobs))
	for _, job := range s.recentJobs {
		jobs = append(jobs, *job)
	}
	return jobs
}

// generateJobID generates a unique job ID
func generateJobID() string {
	return uuid.NewString()
}

// ProcessJob is the entry point for processing a job
func (s *Service) ProcessJob(ctx context.Context, id string) {
	job, ok := s.StartJob(id)
	if !ok {
		// Job either doesn't exist or is not in a startable state (e.g., already processing).
		// This can be logged, but we'll just return for now.
		return
	}

	var successCount int
	var processingErr error

	// Use defer to ensure job status is updated even if a panic occurs.
	defer func() {
		if r := recover(); r != nil {
			processingErr = fmt.Errorf("panic recovered in ProcessJob: %v", r)
		}
		if processingErr != nil {
			s.FailJob(job.ID, processingErr.Error())
		} else {
			s.CompleteJob(job.ID, successCount, 0)
		}
	}()

	// This service is a lightweight monitor. The actual file processing is handled
	// by the bulk_file_processing service. This function simulates a successful
	// job completion for statistical purposes.

	select {
	case <-ctx.Done():
		processingErr = fmt.Errorf("job processing cancelled: %w", ctx.Err())
		return
	default:
		if fileCount, ok := s.jobFileCounts[job.ID]; ok {
			successCount = fileCount
		} else {
			processingErr = fmt.Errorf("inconsistent state: file count for job %s not found", job.ID)
		}
	}
}
