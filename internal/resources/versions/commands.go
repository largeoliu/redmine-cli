// Package versions provides commands for managing Redmine versions.
package versions

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/largeoliu/redmine-cli/internal/errors"
	"github.com/largeoliu/redmine-cli/internal/resources/helpers"
	"github.com/largeoliu/redmine-cli/internal/types"
)

// NewCommand creates a new versions command.
func NewCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Manage Redmine versions",
		Aliases: []string{"versions"},
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
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List versions for a project",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if projectID == 0 {
				return errors.NewValidation("project-id is required")
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			params := make(map[string]string)
			if limit > 0 {
				params["limit"] = strconv.Itoa(limit)
			}
			if offset > 0 {
				params["offset"] = strconv.Itoa(offset)
			}
			result, err := NewClient(c).List(cmd.Context(), projectID, params)
			if err != nil {
				return err
			}
			return resolver.WriteOutput(cmd.OutOrStdout(), flags, result)
		},
	}
	cmd.Flags().IntVar(&projectID, "project-id", 0, "Project ID (required)")
	cmd.Flags().IntVar(&limit, "limit", 25, "Limit number of results")
	cmd.Flags().IntVar(&offset, "offset", 0, "Offset for pagination")
	//nolint:errcheck // MarkFlagRequired returns error only if flag doesn't exist, which is safe to ignore
	_ = cmd.MarkFlagRequired("project-id")
	return cmd
}

func newGetCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get version details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "version")
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
	req := &VersionCreateRequest{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if projectID == 0 {
				return errors.NewValidation("project-id is required")
			}
			if req.Name == "" {
				return errors.NewValidation("name is required")
			}
			if flags.DryRun {
				helpers.DryRunCreate("version", req)
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
			fmt.Printf("Created version #%d: %s\n", result.ID, result.Name)
			return resolver.WriteOutput(cmd.OutOrStdout(), flags, result)
		},
	}
	cmd.Flags().IntVar(&projectID, "project-id", 0, "Project ID (required)")
	cmd.Flags().StringVar(&req.Name, "name", "", "Version name (required)")
	cmd.Flags().StringVar(&req.Description, "description", "", "Version description")
	cmd.Flags().StringVar(&req.Status, "status", "open", "Version status (open, locked, closed)")
	cmd.Flags().StringVar(&req.DueDate, "due-date", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&req.Sharing, "sharing", "none", "Sharing mode (none, descendants, hierarchy, tree, system)")
	//nolint:errcheck // MarkFlagRequired returns error only if flag doesn't exist, which is safe to ignore
	_ = cmd.MarkFlagRequired("project-id")
	return cmd
}

func newUpdateCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	req := &VersionUpdateRequest{}
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "version")
			if err != nil {
				return err
			}
			if flags.DryRun {
				helpers.DryRunUpdate("version", id, req)
				return nil
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			if err := NewClient(c).Update(cmd.Context(), id, req); err != nil {
				return err
			}
			fmt.Printf("Updated version #%d\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.Name, "name", "", "Version name")
	cmd.Flags().StringVar(&req.Description, "description", "", "Version description")
	cmd.Flags().StringVar(&req.Status, "status", "", "Version status (open, locked, closed)")
	cmd.Flags().StringVar(&req.DueDate, "due-date", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&req.Sharing, "sharing", "", "Sharing mode (none, descendants, hierarchy, tree, system)")
	return cmd
}

func newDeleteCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "version")
			if err != nil {
				return err
			}
			if !helpers.ConfirmDelete("version", id, flags.Yes) {
				return nil
			}
			if flags.DryRun {
				helpers.DryRunDelete("version", id)
				return nil
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			if err := NewClient(c).Delete(cmd.Context(), id); err != nil {
				return err
			}
			fmt.Printf("Deleted version #%d\n", id)
			return nil
		},
	}
	return cmd
}
