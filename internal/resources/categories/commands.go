// Package categories provides commands for managing Redmine issue categories.
package categories

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/largeoliu/redmine-cli/internal/errors"
	"github.com/largeoliu/redmine-cli/internal/resources/helpers"
	"github.com/largeoliu/redmine-cli/internal/types"
)

// NewCommand creates the category command.
func NewCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "category",
		Short:   "Manage Redmine issue categories",
		Aliases: []string{"categories"},
	}
	cmd.AddCommand(
		newListCommand(flags, resolver),
		newGetCommand(flags, resolver),
		newCreateCommand(flags, resolver),
		newUpdateCommand(flags, resolver),
		newDeleteCommand(flags, resolver),
	)
	return cmd
}

func newListCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	var projectID int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issue categories for a project",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if projectID == 0 {
				return errors.NewValidation("project-id is required")
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			result, err := NewClient(c).List(cmd.Context(), projectID)
			if err != nil {
				return err
			}
			return resolver.WriteOutput(cmd.OutOrStdout(), flags, result)
		},
	}
	cmd.Flags().IntVar(&projectID, "project-id", 0, "Project ID (required)")
	return cmd
}

func newGetCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get issue category details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "category")
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

func newCreateCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	var projectID int
	req := &CategoryCreateRequest{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issue category",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if projectID == 0 {
				return errors.NewValidation("project-id is required")
			}
			if req.Name == "" {
				return errors.NewValidation("name is required")
			}
			if flags.DryRun {
				helpers.DryRunCreate("issue category", req)
				return nil
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			result, err := NewClient(c).Create(cmd.Context(), projectID, req)
			if err != nil {
				return err
			}
			fmt.Printf("Created issue category #%d: %s\n", result.ID, result.Name)
			return resolver.WriteOutput(cmd.OutOrStdout(), flags, result)
		},
	}
	cmd.Flags().IntVar(&projectID, "project-id", 0, "Project ID (required)")
	cmd.Flags().StringVar(&req.Name, "name", "", "Category name (required)")
	cmd.Flags().IntVar(&req.AssignedToID, "assigned-to-id", 0, "Default assigned user ID")
	return cmd
}

func newUpdateCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	req := &CategoryUpdateRequest{}
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an issue category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "category")
			if err != nil {
				return err
			}
			if flags.DryRun {
				helpers.DryRunUpdate("issue category", id, req)
				return nil
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			if err := NewClient(c).Update(cmd.Context(), id, req); err != nil {
				return err
			}
			fmt.Printf("Updated issue category #%d\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.Name, "name", "", "Category name")
	cmd.Flags().IntVar(&req.AssignedToID, "assigned-to-id", 0, "Default assigned user ID")
	return cmd
}

func newDeleteCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an issue category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "category")
			if err != nil {
				return err
			}
			if !helpers.ConfirmDelete("issue category", id, flags.Yes) {
				return nil
			}
			if flags.DryRun {
				helpers.DryRunDelete("issue category", id)
				return nil
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			if err := NewClient(c).Delete(cmd.Context(), id); err != nil {
				return err
			}
			fmt.Printf("Deleted issue category #%d\n", id)
			return nil
		},
	}
	return cmd
}
