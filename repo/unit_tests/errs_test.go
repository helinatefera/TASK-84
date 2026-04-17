package unit_tests_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/localinsights/portal/internal/errs"
)

func TestAppErrorError(t *testing.T) {
	e := errs.New("TEST_CODE", "a test message", http.StatusBadRequest)
	want := "TEST_CODE: a test message"
	if e.Error() != want {
		t.Fatalf("Error() = %q, want %q", e.Error(), want)
	}
}

func TestNewPreservesFields(t *testing.T) {
	e := errs.New("X", "y", http.StatusTeapot)
	if e.Code != "X" {
		t.Errorf("Code = %q, want %q", e.Code, "X")
	}
	if e.Message != "y" {
		t.Errorf("Message = %q, want %q", e.Message, "y")
	}
	if e.HTTPStatus != http.StatusTeapot {
		t.Errorf("HTTPStatus = %d, want %d", e.HTTPStatus, http.StatusTeapot)
	}
}

func TestWithMessageOverridesMessageButKeepsCodeAndStatus(t *testing.T) {
	base := errs.ErrValidation
	custom := errs.WithMessage(base, "email looks wrong")

	if custom.Code != base.Code {
		t.Errorf("Code = %q, want %q", custom.Code, base.Code)
	}
	if custom.HTTPStatus != base.HTTPStatus {
		t.Errorf("HTTPStatus = %d, want %d", custom.HTTPStatus, base.HTTPStatus)
	}
	if custom.Message != "email looks wrong" {
		t.Errorf("Message = %q, want %q", custom.Message, "email looks wrong")
	}
	// Base must not be mutated (immutability check).
	if base.Message == "email looks wrong" {
		t.Errorf("base error was mutated — WithMessage must return a copy")
	}
}

func TestIsMatchesByCode(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		target *errs.AppError
		want   bool
	}{
		{"same pointer", errs.ErrNotFound, errs.ErrNotFound, true},
		{"same code different message", errs.WithMessage(errs.ErrNotFound, "custom"), errs.ErrNotFound, true},
		{"different code", errs.ErrNotFound, errs.ErrValidation, false},
		{"plain error", errors.New("oops"), errs.ErrNotFound, false},
		{"wrapped AppError", fmt.Errorf("wrap: %w", errs.ErrConflict), errs.ErrConflict, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := errs.Is(tc.err, tc.target); got != tc.want {
				t.Errorf("Is() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHTTPStatusFromError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"AppError returns its status", errs.ErrNotFound, http.StatusNotFound},
		{"Custom AppError", errs.New("X", "y", http.StatusTeapot), http.StatusTeapot},
		{"Wrapped AppError", fmt.Errorf("wrap: %w", errs.ErrForbidden), http.StatusForbidden},
		{"Plain error defaults to 500", errors.New("boom"), http.StatusInternalServerError},
		{"Nil error also defaults to 500", nil, http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := errs.HTTPStatusFromError(tc.err); got != tc.want {
				t.Errorf("HTTPStatusFromError = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestSentinelErrorsHaveExpectedStatusCodes(t *testing.T) {
	// Spot-check that the frequently-used sentinels map to correct HTTP statuses.
	// This lock documents the public contract so refactors cannot silently change it.
	checks := []struct {
		err  *errs.AppError
		code string
		http int
	}{
		{errs.ErrNotFound, "NOT_FOUND", http.StatusNotFound},
		{errs.ErrUnauthorized, "UNAUTHORIZED", http.StatusUnauthorized},
		{errs.ErrForbidden, "FORBIDDEN", http.StatusForbidden},
		{errs.ErrValidation, "VALIDATION_ERROR", http.StatusUnprocessableEntity},
		{errs.ErrConflict, "CONFLICT", http.StatusConflict},
		{errs.ErrRateLimited, "RATE_LIMITED", http.StatusTooManyRequests},
		{errs.ErrCSRFInvalid, "CSRF_INVALID", http.StatusForbidden},
		{errs.ErrCaptchaRequired, "CAPTCHA_REQUIRED", http.StatusPreconditionRequired},
		{errs.ErrCaptchaInvalid, "CAPTCHA_INVALID", http.StatusBadRequest},
		{errs.ErrInternal, "INTERNAL_ERROR", http.StatusInternalServerError},
	}
	for _, c := range checks {
		if c.err.Code != c.code {
			t.Errorf("%s: Code = %q, want %q", c.code, c.err.Code, c.code)
		}
		if c.err.HTTPStatus != c.http {
			t.Errorf("%s: HTTPStatus = %d, want %d", c.code, c.err.HTTPStatus, c.http)
		}
	}
}
