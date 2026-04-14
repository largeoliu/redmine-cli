// Package projects provides commands for managing Redmine projects.
package projects

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/largeoliu/redmine-cli/internal/errors"
	"github.com/largeoliu/redmine-cli/internal/resources/helpers"
	"github.com/largeoliu/redmine-cli/internal/types"
)

// NewCommand creates a new projects command.
func NewCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Short:   "Manage Redmine projects",
		Aliases: []string{"projects"},
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
	var limit, offset int
	var include string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		RunE: func(cmd *cobra.Command, _ []string) error {
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
			if include != "" {
				params["include"] = include
			}
			result, err := NewClient(c).List(cmd.Context(), params)
			if err != nil {
				return err
			}
			return resolver.WriteOutput(cmd.OutOrStdout(), flags, result)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 25, "Limit number of results")
	cmd.Flags().IntVar(&offset, "offset", 0, "Offset for pagination")
	cmd.Flags().StringVar(&include, "include", "", "Include associated data (trackers,issue_categories,enabled_modules)")
	return cmd
}

func newGetCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	var include string
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get project details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			params := make(map[string]string)
			if include != "" {
				params["include"] = include
			}
			var result *Project
			id, parseErr := strconv.Atoi(args[0])
			if parseErr == nil {
				result, err = NewClient(c).Get(cmd.Context(), id, params)
			} else {
				result, err = NewClient(c).GetByIdentifier(cmd.Context(), args[0], params)
			}
			if err != nil {
				return err
			}
			return resolver.WriteOutput(cmd.OutOrStdout(), flags, result)
		},
	}
	cmd.Flags().StringVar(&include, "include", "", "Include associated data")
	return cmd
}

func newCreateCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	req := &ProjectCreateRequest{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if req.Name == "" {
				return errors.NewValidation("name is required")
			}
			if req.Identifier == "" {
				return errors.NewValidation("identifier is required")
			}
			if flags.DryRun {
				helpers.DryRunCreate("project", req)
				return nil
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			result, err := NewClient(c).Create(cmd.Context(), req)
			if err != nil {
				return err
			}
			fmt.Printf("Created project #%d: %s\n", result.ID, result.Name)
			return resolver.WriteOutput(cmd.OutOrStdout(), flags, result)
		},
	}
	cmd.Flags().StringVar(&req.Name, "name", "", "Project name (required)")
	cmd.Flags().StringVar(&req.Identifier, "identifier", "", "Project identifier (required)")
	cmd.Flags().StringVar(&req.Description, "description", "", "Project description")
	cmd.Flags().StringVar(&req.Homepage, "homepage", "", "Project homepage URL")
	cmd.Flags().BoolVar(&req.IsPublic, "public", true, "Is project public")
	cmd.Flags().IntVar(&req.ParentID, "parent-id", 0, "Parent project ID")
	return cmd
}

func newUpdateCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	req := &ProjectUpdateRequest{}
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "project")
			if err != nil {
				return err
			}
			if flags.DryRun {
				helpers.DryRunUpdate("project", id, req)
				return nil
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			if err := NewClient(c).Update(cmd.Context(), id, req); err != nil {
				return err
			}
			fmt.Printf("Updated project #%d\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.Name, "name", "", "Project name")
	cmd.Flags().StringVar(&req.Description, "description", "", "Project description")
	cmd.Flags().StringVar(&req.Homepage, "homepage", "", "Project homepage URL")
	cmd.Flags().IntVar(&req.Status, "status", 0, "Project status (1=active, 5=closed, 9=archived)")
	return cmd
}

func newDeleteCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "project")
			if err != nil {
				return err
			}
			if !helpers.ConfirmDelete("project", id, flags.Yes) {
				return nil
			}
			if flags.DryRun {
				helpers.DryRunDelete("project", id)
				return nil
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			if err := NewClient(c).Delete(cmd.Context(), id); err != nil {
				return err
			}
			fmt.Printf("Deleted project #%d\n", id)
			return nil
		},
	}
	return cmd
}
