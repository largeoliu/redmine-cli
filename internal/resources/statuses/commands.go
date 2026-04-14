// Package statuses provides commands for managing Redmine issue statuses.
package statuses

import (
	"github.com/spf13/cobra"

	"github.com/largeoliu/redmine-cli/internal/types"
)

// NewCommand creates a new statuses command.
func NewCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Short:   "Manage Redmine issue statuses",
		Aliases: []string{"statuses"},
	}
	cmd.AddCommand(newListCommand(flags, resolver))
	return cmd
}

func newListCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issue statuses",
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			result, err := NewClient(c).List(cmd.Context())
			if err != nil {
				return err
			}
			return resolver.WriteOutput(cmd.OutOrStdout(), flags, result)
		},
	}
	return cmd
}
