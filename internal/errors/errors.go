// Package errors provides custom error types for the CLI.
package errors

import (
	stderrors "errors"
)

// Category represents the error category.
type Category string

const (
	// CategoryValidation represents validation errors.
	CategoryValidation Category = "validation"
	// CategoryAuth represents authentication errors.
	CategoryAuth Category = "auth"
	// CategoryAPI represents API errors.
	CategoryAPI Category = "api"
	// CategoryNetwork represents network errors.
	CategoryNetwork Category = "network"
	// CategoryInternal represents internal errors.
	CategoryInternal Category = "internal"
	// CategoryTimeout represents timeout errors.
	CategoryTimeout Category = "timeout"
	// CategoryRateLimit represents rate limit errors.
	CategoryRateLimit Category = "rate_limit"
)

// Error is a custom error type with additional context.
type Error struct {
	Category  Category `json:"category"`
	Message   string   `json:"message"`
	Hint      string   `json:"hint,omitempty"`
	Actions   []string `json:"actions,omitempty"`
	Retryable bool     `json:"retryable"`
	RequestID string   `json:"request_id,omitempty"`
	Cause     error    `json:"-"`
}

// Error returns the error message.
func (e *Error) Error() string {
	return e.Message
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	return e.Cause
}

// NewValidation creates a new validation error.
func NewValidation(message string, opts ...Option) *Error {
	return newError(CategoryValidation, message, opts...)
}

// NewAuth creates a new authentication error.
func NewAuth(message string, opts ...Option) *Error {
	return newError(CategoryAuth, message, opts...)
}

// NewAPI creates a new API error.
func NewAPI(message string, opts ...Option) *Error {
	return newError(CategoryAPI, message, opts...)
}

// NewNetwork creates a new network error.
func NewNetwork(message string, opts ...Option) *Error {
	e := newError(CategoryNetwork, message, opts...)
	e.Retryable = true
	return e
}

// NewInternal creates a new internal error.
func NewInternal(message string, opts ...Option) *Error {
	return newError(CategoryInternal, message, opts...)
}

// NewTimeout creates a new timeout error.
func NewTimeout(message string, opts ...Option) *Error {
	return newError(CategoryTimeout, message, opts...)
}

// NewRateLimit creates a new rate limit error.
func NewRateLimit(message string, opts ...Option) *Error {
	e := newError(CategoryRateLimit, message, opts...)
	e.Retryable = true
	return e
}

func newError(category Category, message string, opts ...Option) *Error {
	e := &Error{
		Category: category,
		Message:  message,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Option is a functional option for configuring errors.
type Option func(*Error)

// WithHint adds a hint to the error.
func WithHint(hint string) Option {
	return func(e *Error) {
		e.Hint = hint
	}
}

// WithActions adds suggested actions to the error.
func WithActions(actions ...string) Option {
	return func(e *Error) {
		e.Actions = actions
	}
}

// WithRetryable sets whether the error is retryable.
func WithRetryable(retryable bool) Option {
	return func(e *Error) {
		e.Retryable = retryable
	}
}

// WithRequestID adds a request ID to the error.
func WithRequestID(id string) Option {
	return func(e *Error) {
		e.RequestID = id
	}
}

// WithCause sets the underlying cause of the error.
func WithCause(cause error) Option {
	return func(e *Error) {
		e.Cause = cause
	}
}

// ExitCode returns the exit code for the given error.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var appErr *Error
	if As(err, &appErr) {
		switch appErr.Category {
		case CategoryValidation:
			return 1
		case CategoryAuth:
			return 2
		case CategoryAPI:
			return 3
		case CategoryNetwork:
			return 4
		case CategoryInternal:
			return 5
		case CategoryTimeout:
			return 6
		case CategoryRateLimit:
			return 7
		}
	}
	return 1
}

// As wraps stderrors.As.
func As(err error, target any) bool {
	return stderrors.As(err, target)
}
