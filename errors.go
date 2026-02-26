package nslsolver

import "fmt"

// APIError represents an error returned by the NSLSolver API.
type APIError struct {
	StatusCode int
	Message    string
	Retryable  bool
}

func (e *APIError) Error() string {
	return fmt.Sprintf("nslsolver: API error %d: %s", e.StatusCode, e.Message)
}

func newAPIError(statusCode int, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Retryable:  statusCode == 429 || statusCode == 503,
	}
}

// isAPIErrorWithStatus checks whether err is an *APIError with the given HTTP status code.
func isAPIErrorWithStatus(err error, status int) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.StatusCode == status
}

func IsAuthError(err error) bool       { return isAPIErrorWithStatus(err, 401) }
func IsBalanceError(err error) bool     { return isAPIErrorWithStatus(err, 402) }
func IsNotAllowedError(err error) bool  { return isAPIErrorWithStatus(err, 403) }
func IsBadRequestError(err error) bool  { return isAPIErrorWithStatus(err, 400) }
func IsRateLimitError(err error) bool   { return isAPIErrorWithStatus(err, 429) }
func IsBackendError(err error) bool     { return isAPIErrorWithStatus(err, 503) }

// IsRetryableError reports whether err is a retryable API error (429 or 503).
func IsRetryableError(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.Retryable
}
