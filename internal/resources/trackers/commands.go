// Package trackers provides commands for managing Redmine trackers.
package trackers

import (
	"github.com/spf13/cobra"

	"github.com/largeoliu/redmine-cli/internal/resources/helpers"
	"github.com/largeoliu/redmine-cli/internal/types"
)

func NewCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tracker",
		Short:   "Manage Redmine trackers",
		Aliases: []string{"trackers"},
	}
	cmd.AddCommand(newListCommand(flags, resolver))
	cmd.AddCommand(newGetCommand(flags, resolver))
	return cmd
}

func newListCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all trackers",
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

func newGetCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get tracker details including custom fields",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "tracker")
			if err != nil {
				return err
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			result, err := NewClient(c).Get(cmd.Context(), id)
			if err != nil {
				return err
			}
			return resolver.WriteOutput(cmd.OutOrStdout(), flags, result)
		},
	}
	return cmd
}
