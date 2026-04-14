// Package types provides global flags and resolver interface for the application.
package types

import "time"

// GlobalFlags contains global command line flags
type GlobalFlags struct {
	URL      string
	Key      string
	Format   string
	JQ       string
	Fields   string
	DryRun   bool
	Yes      bool
	Output   string
	Limit    int
	Offset   int
	Timeout  time.Duration
	Retries  int
	Verbose  bool
	Debug    bool
	Instance string
}
