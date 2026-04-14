// Package types provides global flags and resolver interfaces for the application.
package types

import (
	"io"

	"github.com/largeoliu/redmine-cli/internal/client"
)

// ClientResolver resolves a Redmine API client from global flags.
type ClientResolver interface {
	ResolveClient(flags *GlobalFlags) (*client.Client, error)
}

// OutputWriter writes formatted output based on global flags.
type OutputWriter interface {
	WriteOutput(w io.Writer, flags *GlobalFlags, payload any) error
}

// Resolver combines client resolution and output writing.
// Commands can depend on only the subset they need.
type Resolver interface {
	ClientResolver
	OutputWriter
}
