package errors

import "fmt"

// ProcessingError represents an error that occurs during file processing.
type ProcessingError struct {
	Message string
	Err     error
}

func (e *ProcessingError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *ProcessingError) Unwrap() error {
	return e.Err
}

func NewProcessingError(message string, err error) error {
	return &ProcessingError{Message: message, Err: err}
}

// ValidationError represents an error in user-provided configuration.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func NewValidationError(message string) error {
	return &ValidationError{Message: message}
}
