// Package app provides the CLI application commands and logic.
package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/config"
	"github.com/largeoliu/redmine-cli/internal/errors"
	"github.com/largeoliu/redmine-cli/internal/output"
	"github.com/largeoliu/redmine-cli/internal/resources/categories"
	"github.com/largeoliu/redmine-cli/internal/resources/issues"
	"github.com/largeoliu/redmine-cli/internal/resources/priorities"
	"github.com/largeoliu/redmine-cli/internal/resources/projects"
	"github.com/largeoliu/redmine-cli/internal/resources/sprints"
	"github.com/largeoliu/redmine-cli/internal/resources/statuses"
	"github.com/largeoliu/redmine-cli/internal/resources/time_entries"
	"github.com/largeoliu/redmine-cli/internal/resources/trackers"
	"github.com/largeoliu/redmine-cli/internal/resources/users"
	"github.com/largeoliu/redmine-cli/internal/resources/versions"
	"github.com/largeoliu/redmine-cli/internal/types"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// resolver implements types.Resolver interface
type resolver struct{}

func (r *resolver) ResolveClient(flags *types.GlobalFlags) (*client.Client, error) {
	return ResolveClient((*GlobalFlags)(flags))
}

func (r *resolver) WriteOutput(w io.Writer, flags *types.GlobalFlags, payload any) error {
	return WriteOutput(w, (*GlobalFlags)(flags), payload)
}

// defaultResolver is the default resolver instance
var defaultResolver types.Resolver = &resolver{}

// Execute runs the CLI application and returns the exit code.
func Execute() int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	root := NewRootCommand(ctx)
	cmd, err := root.ExecuteC()
	if err != nil {
		printError(cmd, err)
		return errors.ExitCode(err)
	}
	return 0
}

// NewRootCommand creates the root command for the CLI.
func NewRootCommand(ctx context.Context) *cobra.Command {
	flags := &GlobalFlags{}

	root := &cobra.Command{
		Use:               "redmine",
		Short:             "Redmine CLI - AI Agent friendly command line tool",
		Version:           version,
		SilenceErrors:     true,
		SilenceUsage:      true,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	bindGlobalFlags(root, flags)

	root.AddCommand(
		newLoginCommand(flags),
		newLogoutCommand(flags),
		newVersionCommand(),
		newUpgradeCommand(),
		newConfigCommand(flags),
		sprints.NewCommand(flags, defaultResolver),
		categories.NewCommand(flags, defaultResolver),
		issues.NewCommand(flags, defaultResolver),
		priorities.NewCommand(flags, defaultResolver),
		projects.NewCommand(flags, defaultResolver),
		statuses.NewCommand(flags, defaultResolver),
		time_entries.NewCommand(flags, defaultResolver),
		trackers.NewCommand(flags, defaultResolver),
		users.NewCommand(flags, defaultResolver),
		versions.NewCommand(flags, defaultResolver),
	)

	root.SetContext(ctx)
	return root
}

func bindGlobalFlags(cmd *cobra.Command, flags *GlobalFlags) {
	cmd.PersistentFlags().StringVarP(&flags.URL, "url", "u", "", "Redmine instance URL")
	cmd.PersistentFlags().StringVarP(&flags.Key, "key", "k", "", "API key")
	cmd.PersistentFlags().StringVarP(&flags.Format, "format", "f", "json", "Output format (json/table/raw)")
	cmd.PersistentFlags().StringVar(&flags.JQ, "jq", "", "jq filter expression")
	cmd.PersistentFlags().StringVar(&flags.Fields, "fields", "", "Fields to include in output")
	cmd.PersistentFlags().BoolVar(&flags.DryRun, "dry-run", false, "Preview without executing")
	cmd.PersistentFlags().BoolVarP(&flags.Yes, "yes", "y", false, "Skip confirmation prompts")
	cmd.PersistentFlags().StringVarP(&flags.Output, "output", "o", "", "Output file path")
	cmd.PersistentFlags().IntVarP(&flags.Limit, "limit", "l", 0, "Limit number of results")
	cmd.PersistentFlags().IntVar(&flags.Offset, "offset", 0, "Offset for pagination")
	cmd.PersistentFlags().DurationVar(&flags.Timeout, "timeout", 30*time.Second, "Request timeout (e.g. 30s, 1m)")
	cmd.PersistentFlags().IntVar(&flags.Retries, "retries", 3, "Number of retries for failed requests")
	cmd.PersistentFlags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Verbose output")
	cmd.PersistentFlags().BoolVar(&flags.Debug, "debug", false, "Debug mode")
	cmd.PersistentFlags().StringVar(&flags.Instance, "instance", "", "Instance name from config")
}

func printError(_ *cobra.Command, err error) {
	var appErr *errors.Error
	if errors.As(err, &appErr) {
		fmt.Fprintf(os.Stderr, "Error: %s\n", appErr.Message)
		if appErr.Hint != "" {
			fmt.Fprintf(os.Stderr, "Hint: %s\n", appErr.Hint)
		}
		if len(appErr.Actions) > 0 {
			fmt.Fprintf(os.Stderr, "Actions:\n")
			for _, action := range appErr.Actions {
				fmt.Fprintf(os.Stderr, "  - %s\n", action)
			}
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

// ResolveClient resolves the client based on the provided flags.
func ResolveClient(flags *GlobalFlags) (*client.Client, error) {
	url := flags.URL
	key := flags.Key

	if url == "" || key == "" {
		store := config.NewStore("")
		cfg, err := store.Load()
		if err != nil {
			return nil, errors.NewInternal("failed to load config", errors.WithCause(err))
		}

		instanceName := flags.Instance
		if instanceName == "" {
			instanceName = cfg.Default
		}

		if instanceName != "" {
			inst, ok := cfg.GetInstance(instanceName)
			if !ok {
				return nil, errors.NewValidation("instance not found: " + instanceName)
			}
			if url == "" {
				url = inst.URL
			}
			if key == "" {
				key = inst.APIKey
			}
		}
	}

	if url == "" {
		return nil, errors.NewValidation("Redmine URL is required",
			errors.WithHint("Use --url flag or run 'redmine login'"))
	}
	if key == "" {
		return nil, errors.NewValidation("API key is required",
			errors.WithHint("Use --key flag or run 'redmine login'"))
	}

	opts := []client.Option{}
	if flags.Timeout > 0 {
		opts = append(opts, client.WithTimeout(flags.Timeout))
	}
	if flags.Retries > 0 {
		opts = append(opts, client.WithRetry(flags.Retries, 500*time.Millisecond, 5*time.Second))
	}
	return client.NewClient(url, key, opts...), nil
}

// ResolveFormat resolves the output format based on the provided flags.
func ResolveFormat(flags *GlobalFlags) output.Format {
	switch flags.Format {
	case "table":
		return output.FormatTable
	case "raw":
		return output.FormatRaw
	default:
		return output.FormatJSON
	}
}

// WriteOutput writes the output based on the provided flags and payload.
func WriteOutput(w io.Writer, flags *GlobalFlags, payload any) error {
	format := ResolveFormat(flags)

	normalizedPayload, err := output.NormalizePayload(payload)
	if err != nil {
		return err
	}

	if flags.JQ != "" {
		query, err := output.ParseJQ(flags.JQ)
		if err != nil {
			return err
		}
		return output.ApplyJQNormalized(w, normalizedPayload, query)
	}

	if flags.Fields != "" {
		fields := parseFields(flags.Fields)
		filtered, err := output.SelectFieldsNormalized(normalizedPayload, fields)
		if err != nil {
			return err
		}
		return output.Write(w, format, filtered)
	}

	return output.Write(w, format, normalizedPayload)
}

func parseFields(s string) []string {
	var fields []string
	for _, f := range splitByComma(s) {
		if f != "" {
			fields = append(fields, f)
		}
	}
	return fields
}

func splitByComma(s string) []string {
	var result []string
	var current string
	for _, r := range s {
		if r == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
