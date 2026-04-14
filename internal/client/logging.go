// Package client provides the HTTP client for interacting with Redmine API.
package client

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// requestIDCounter is used to generate unique request IDs
var requestIDCounter uint64

// generateRequestID generates a unique request ID
func generateRequestID() string {
	id := atomic.AddUint64(&requestIDCounter, 1)
	return fmt.Sprintf("req-%d-%d", time.Now().UnixNano(), id)
}

// safeFprintf writes to the given writer and ignores any errors.
// This is used for logging where write failures should not interrupt the request.
func safeFprintf(w io.Writer, format string, args ...any) {
	if _, err := fmt.Fprintf(w, format, args...); err != nil {
		return
	}
}

// LoggingTransport is an http.RoundTripper that logs requests and responses
type LoggingTransport struct {
	transport http.RoundTripper
	logger    io.Writer
	verbose   bool
}

// NewLoggingTransport creates a new LoggingTransport
func NewLoggingTransport(transport http.RoundTripper, logger io.Writer, verbose bool) *LoggingTransport {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &LoggingTransport{
		transport: transport,
		logger:    logger,
		verbose:   verbose,
	}
}

// RoundTrip implements http.RoundTripper
func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	requestID := generateRequestID()

	// Log request
	if t.verbose {
		safeFprintf(t.logger, "[%s] --> %s %s\n", requestID, req.Method, req.URL.String())
		for key, values := range req.Header {
			if strings.ToLower(key) == "x-redmine-api-key" {
				safeFprintf(t.logger, "[%s]     %s: ***\n", requestID, key)
			} else {
				for _, value := range values {
					safeFprintf(t.logger, "[%s]     %s: %s\n", requestID, key, value)
				}
			}
		}
	} else {
		safeFprintf(t.logger, "[%s] --> %s %s\n", requestID, req.Method, req.URL.Path)
	}

	start := time.Now()
	resp, err := t.transport.RoundTrip(req)
	duration := time.Since(start)

	if err != nil {
		safeFprintf(t.logger, "[%s] <-- ERROR: %v (%s)\n", requestID, err, duration)
		return nil, err
	}

	// Log response
	if t.verbose {
		safeFprintf(t.logger, "[%s] <-- %d %s (%s)\n", requestID, resp.StatusCode, http.StatusText(resp.StatusCode), duration)
		for key, values := range resp.Header {
			for _, value := range values {
				safeFprintf(t.logger, "[%s]     %s: %s\n", requestID, key, value)
			}
		}
	} else {
		safeFprintf(t.logger, "[%s] <-- %d (%s)\n", requestID, resp.StatusCode, duration)
	}

	return resp, nil
}

// SetVerbose sets the verbose mode
func (t *LoggingTransport) SetVerbose(verbose bool) {
	t.verbose = verbose
}
