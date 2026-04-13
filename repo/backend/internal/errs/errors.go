package errs

import (
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus}
}

func WithMessage(base *AppError, message string) *AppError {
	return &AppError{Code: base.Code, Message: message, HTTPStatus: base.HTTPStatus}
}

func Is(err error, target *AppError) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == target.Code
	}
	return false
}

func HTTPStatusFromError(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

var (
	ErrNotFound            = New("NOT_FOUND", "Resource not found", http.StatusNotFound)
	ErrUnauthorized        = New("UNAUTHORIZED", "Authentication required", http.StatusUnauthorized)
	ErrForbidden           = New("FORBIDDEN", "Insufficient permissions", http.StatusForbidden)
	ErrValidation          = New("VALIDATION_ERROR", "Invalid input", http.StatusUnprocessableEntity)
	ErrConflict            = New("CONFLICT", "Resource already exists", http.StatusConflict)
	ErrCaptchaRequired     = New("CAPTCHA_REQUIRED", "CAPTCHA verification required", http.StatusPreconditionRequired)
	ErrCaptchaInvalid      = New("CAPTCHA_INVALID", "CAPTCHA verification failed", http.StatusBadRequest)
	ErrRateLimited         = New("RATE_LIMITED", "Too many requests", http.StatusTooManyRequests)
	ErrIdempotencyReplay   = New("IDEMPOTENCY_REPLAY", "Replayed response", http.StatusOK)
	ErrIPDenied            = New("IP_DENIED", "IP address is denied", http.StatusForbidden)
	ErrCSRFInvalid         = New("CSRF_INVALID", "Invalid CSRF token", http.StatusForbidden)
	ErrEventRateLimited    = New("EVENT_RATE_LIMITED", "Event rate limit exceeded", http.StatusTooManyRequests)
	ErrImageTypeInvalid    = New("IMAGE_TYPE_INVALID", "Unsupported image format", http.StatusBadRequest)
	ErrImageTooLarge       = New("IMAGE_TOO_LARGE", "Image exceeds maximum size", http.StatusBadRequest)
	ErrImageQuarantined    = New("IMAGE_QUARANTINED", "Image is under review", http.StatusForbidden)
	ErrDuplicateReview     = New("DUPLICATE_REVIEW", "You have already reviewed this item", http.StatusConflict)
	ErrAppealExists        = New("APPEAL_EXISTS", "An appeal already exists for this report", http.StatusConflict)
	ErrExperimentNotActive = New("EXPERIMENT_NOT_ACTIVE", "Experiment is not currently running", http.StatusBadRequest)
	ErrInternal            = New("INTERNAL_ERROR", "An internal error occurred", http.StatusInternalServerError)
)
