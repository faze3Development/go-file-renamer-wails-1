// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package monitoring_and_statistics

import (
	"fmt"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"
)

// JobLookupService defines the interface for looking up jobs across different services
type JobLookupService interface {
	GetJob(jobID string) (any, error) // Returns either *common.Job or *bulk_file_processing.ProcessingJob
}

type Handler struct {
	service      ServiceAPI
	jobLookupSvc JobLookupService
	logger       *slog.Logger
}

// NewHandler creates a new monitoring and statistics handler.
func NewHandler(service ServiceAPI, jobLookupSvc JobLookupService, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		service:      service,
		jobLookupSvc: jobLookupSvc,
		logger:       logger,
	}
}

// RegisterRoutes registers processing monitoring routes.
func (h *Handler) RegisterRoutes(r *chi.Mux) {
	r.Route("/processing", func(r chi.Router) {
		r.Get("/stats", h.wrap(h.GetStats))
		r.Get("/jobs/active", h.wrap(h.GetActiveJobs))
		r.Get("/jobs/recent", h.wrap(h.GetRecentJobs))
		r.Get("/bulk/status/{jobID}", h.wrap(h.GetJobStatus))
	})
}

func (h *Handler) wrap(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		errors.WrapHandler(handler, h.logger, requestID)(w, r)
	}
}

// GetStats returns processing statistics for monitoring.
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Processing stats requested")
	stats := h.service.GetStats()
	return errors.SafeJSONResponse(w, stats)
}

// ActiveJobsResponse represents the response for the GetActiveJobs handler
type ActiveJobsResponse struct {
	ActiveJobs []Job `json:"active_jobs"`
}

// RecentJobsResponse represents the response for the GetRecentJobs handler
type RecentJobsResponse struct {
	RecentJobs []Job `json:"recent_jobs"`
}

// GetActiveJobs returns currently active processing jobs.
func (h *Handler) GetActiveJobs(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Active jobs requested")
	jobs := h.service.GetActiveJobs()
	// Dereference pointers to create a slice of values for JSON response
	jobsCopy := make([]Job, 0, len(jobs))
	for _, job := range jobs {
		if job != nil {
			jobsCopy = append(jobsCopy, *job)
		}
	}
	h.logger.Info("Returning active jobs", "count", len(jobsCopy))
	return errors.SafeJSONResponse(w, ActiveJobsResponse{ActiveJobs: jobsCopy})
}

// GetRecentJobs returns recently completed processing jobs.
func (h *Handler) GetRecentJobs(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Recent jobs requested")
	jobs := h.service.GetRecentJobs()
	h.logger.Info("Returning recent jobs", "count", len(jobs))
	return errors.SafeJSONResponse(w, RecentJobsResponse{RecentJobs: jobs})
}

// JobStatusResponse represents the response for the GetJobStatus handler
type JobStatusResponse struct {
	Status   string           `json:"status"`
	Progress int              `json:"progress,omitempty"` // Estimated progress percentage (0-100)
	Results  []map[string]any `json:"results,omitempty"`
	Error    string           `json:"error,omitempty"`
	JobID    string           `json:"jobId,omitempty"`
}

// GetJobStatus returns the status of a specific job.
func (h *Handler) GetJobStatus(w http.ResponseWriter, r *http.Request) error {
	jobID := chi.URLParam(r, "jobID")
	if jobID == "" {
		h.logger.Warn("Job status request missing job ID")
		return errors.SafeJSONResponse(w, JobStatusResponse{
			Status: "error",
			Error:  "Job ID is required",
		})
	}

	h.logger.Info("Job status requested", "jobID", jobID)

	// Check active jobs first
	activeJobs := h.service.GetActiveJobs()
	for _, job := range activeJobs {
		if job.ID == jobID {
			h.logger.Info("Found active job", "jobID", jobID, "status", job.Status)
			response := JobStatusResponse{
				Status:   string(job.Status),
				Progress: h.service.EstimateJobProgress(job),
				JobID:    job.ID,
			}

			// For completed jobs, include basic results
			// In a real implementation, you'd get detailed results from the bulk service
			if job.Status == "completed" {
				response.Results = []map[string]any{
					{
						"success": true,
						"action":  "completed",
						"message": fmt.Sprintf("Job completed with %d successes and %d failures", job.SuccessCount, job.FailureCount),
					},
				}
			}

			return errors.SafeJSONResponse(w, response)
		}
	}

	// Check recent jobs
	recentJobs := h.service.GetRecentJobs()
	for _, job := range recentJobs {
		if job.ID == jobID {
			h.logger.Info("Found completed job", "jobID", jobID, "status", job.Status)
			response := JobStatusResponse{
				Status:   string(job.Status),
				Progress: h.service.EstimateJobProgress(&job),
				JobID:    job.ID,
			}

			// For completed jobs, include results
			switch job.Status {
			case "completed":
				response.Results = []map[string]any{
					{
						"success": true,
						"action":  "completed",
						"message": fmt.Sprintf("Job completed with %d successes and %d failures", job.SuccessCount, job.FailureCount),
					},
				}
			case "failed":
				response.Error = job.Error
			}

			return errors.SafeJSONResponse(w, response)
		}
	}

	// Check other job services if available
	if h.jobLookupSvc != nil {
		if job, err := h.jobLookupSvc.GetJob(jobID); err == nil && job != nil {
			h.logger.Info("Found job in lookup service", "jobID", jobID)

			// Handle different job types
			switch j := job.(type) {
			case *Job:
				// Regular job from monitoring service
				response := JobStatusResponse{
					Status:   string(j.Status),
					Progress: h.service.EstimateJobProgress(j),
					JobID:    j.ID,
				}

				if j.Status == "completed" {
					response.Results = []map[string]any{
						{
							"success": true,
							"action":  "completed",
							"message": fmt.Sprintf("Job completed with %d successes and %d failures", j.SuccessCount, j.FailureCount),
						},
					}
				}

				return errors.SafeJSONResponse(w, response)

			default:
				// Check if this is a bulk processing job by examining its structure
				// Use reflection to safely access fields without direct import
				v := reflect.ValueOf(job)
				if v.Kind() == reflect.Ptr && !v.IsNil() {
					v = v.Elem()
				}

				// Check if it has the expected fields for a ProcessingJob
				if v.IsValid() && v.Kind() == reflect.Struct {
					// Try to get Status field
					statusField := v.FieldByName("Status")
					idField := v.FieldByName("ID")
					resultsField := v.FieldByName("Results")
					errorField := v.FieldByName("Error")

					if statusField.IsValid() && idField.IsValid() {
						status := statusField.String()
						jobID := idField.String()

						response := JobStatusResponse{
							Status: status,
							JobID:  jobID,
						}

						// Set progress based on status
						switch status {
						case "completed":
							response.Progress = 100
							// Include results if available
							if resultsField.IsValid() && resultsField.Len() > 0 {
								response.Results = make([]map[string]any, resultsField.Len())
								for i := 0; i < resultsField.Len(); i++ {
									result := resultsField.Index(i)
									if result.IsValid() {
										filename := result.FieldByName("Filename")
										success := result.FieldByName("Success")
										action := result.FieldByName("Action")
										newName := result.FieldByName("NewName")
										dataID := result.FieldByName("DataID")
										contentType := result.FieldByName("ContentType")

										resultMap := make(map[string]any)
										if filename.IsValid() {
											resultMap["filename"] = filename.String()
										}
										if success.IsValid() {
											resultMap["success"] = success.Bool()
										}
										if action.IsValid() {
											resultMap["action"] = action.String()
										}
										if newName.IsValid() {
											resultMap["newName"] = newName.String()
										}
										if dataID.IsValid() {
											resultMap["dataId"] = dataID.String()
										}
										if contentType.IsValid() {
											resultMap["contentType"] = contentType.String()
										}
										response.Results[i] = resultMap
									}
								}
							} else {
								response.Results = []map[string]any{
									{
										"success": true,
										"action":  "bulk_completed",
										"message": "Bulk job completed successfully",
									},
								}
							}
						case "processing":
							response.Progress = 50 // Estimate for processing
						case "failed":
							if errorField.IsValid() {
								response.Error = errorField.String()
							}
							response.Progress = 0
						case "queued":
							response.Progress = 0
						}

						return errors.SafeJSONResponse(w, response)
					}
				}

				// Unknown job type
				response := JobStatusResponse{
					Status: "unknown",
					JobID:  jobID,
					Error:  "Unknown job type",
				}
				return errors.SafeJSONResponse(w, response)
			}
		}
	}

	// Job not found
	h.logger.Warn("Job not found", "jobID", jobID)
	return errors.SafeJSONResponse(w, JobStatusResponse{
		Status: "error",
		Error:  "Job not found",
	})
}
