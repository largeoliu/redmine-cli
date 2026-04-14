// Package issues provides commands for managing Redmine issues.
package issues

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/largeoliu/redmine-cli/internal/errors"
	"github.com/largeoliu/redmine-cli/internal/resources/helpers"
	"github.com/largeoliu/redmine-cli/internal/types"
)

// NewCommand creates a new issues command.
func NewCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "issue",
		Short:   "Manage Redmine issues",
		Aliases: []string{"issues"},
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
	listFlags := &ListFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// 使用全局标志的 limit 和 offset（如果设置了的话）
			if flags.Limit > 0 {
				listFlags.Limit = flags.Limit
			}
			if flags.Offset > 0 {
				listFlags.Offset = flags.Offset
			}

			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			params := BuildListParams(*listFlags)
			result, err := NewClient(c).List(cmd.Context(), params)
			if err != nil {
				return err
			}
			return resolver.WriteOutput(cmd.OutOrStdout(), flags, result)
		},
	}
	cmd.Flags().IntVar(&listFlags.ProjectID, "project-id", 0, "Filter by project ID")
	cmd.Flags().IntVar(&listFlags.TrackerID, "tracker-id", 0, "Filter by tracker ID")
	cmd.Flags().IntVar(&listFlags.StatusID, "status-id", 0, "Filter by status ID")
	cmd.Flags().IntVar(&listFlags.AssignedToID, "assigned-to-id", 0, "Filter by assigned user ID")
	cmd.Flags().StringVar(&listFlags.Query, "query", "", "Filter by custom query ID")
	cmd.Flags().StringVar(&listFlags.Sort, "sort", "", "Sort field (e.g., created_on:desc)")
	return cmd
}

func newGetCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	var include string
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get issue details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "issue")
			if err != nil {
				return err
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			params := make(map[string]string)
			if include != "" {
				params["include"] = include
			}
			result, err := NewClient(c).Get(cmd.Context(), id, params)
			if err != nil {
				return err
			}
			return resolver.WriteOutput(cmd.OutOrStdout(), flags, result)
		},
	}
	cmd.Flags().StringVar(&include, "include", "", "Include associated data (children,attachments,relations,changesets,journals,watchers)")
	return cmd
}

func newCreateCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	req := &IssueCreateRequest{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issue",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if req.ProjectID == 0 {
				return errors.NewValidation("project-id is required")
			}
			if req.Subject == "" {
				return errors.NewValidation("subject is required")
			}
			if flags.DryRun {
				helpers.DryRunCreate("issue", req)
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
			fmt.Printf("Created issue #%d: %s\n", result.ID, result.Subject)
			return resolver.WriteOutput(cmd.OutOrStdout(), flags, result)
		},
	}
	cmd.Flags().IntVar(&req.ProjectID, "project-id", 0, "Project ID (required)")
	cmd.Flags().StringVar(&req.Subject, "subject", "", "Issue subject (required)")
	cmd.Flags().StringVar(&req.Description, "description", "", "Issue description")
	cmd.Flags().IntVar(&req.TrackerID, "tracker-id", 0, "Tracker ID")
	cmd.Flags().IntVar(&req.StatusID, "status-id", 0, "Status ID")
	cmd.Flags().IntVar(&req.PriorityID, "priority-id", 0, "Priority ID")
	cmd.Flags().IntVar(&req.AssignedToID, "assigned-to-id", 0, "Assigned user ID")
	cmd.Flags().IntVar(&req.CategoryID, "category-id", 0, "Category ID")
	cmd.Flags().IntVar(&req.FixedVersionID, "fixed-version-id", 0, "Target version ID")
	cmd.Flags().StringVar(&req.StartDate, "start-date", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&req.DueDate, "due-date", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&req.DoneRatio, "done-ratio", 0, "Done ratio (0-100)")
	return cmd
}

func newUpdateCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	req := &IssueUpdateRequest{}
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "issue")
			if err != nil {
				return err
			}
			if flags.DryRun {
				helpers.DryRunUpdate("issue", id, req)
				return nil
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			if err := NewClient(c).Update(cmd.Context(), id, req); err != nil {
				return err
			}
			fmt.Printf("Updated issue #%d\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.Subject, "subject", "", "Issue subject")
	cmd.Flags().StringVar(&req.Description, "description", "", "Issue description")
	cmd.Flags().IntVar(&req.StatusID, "status-id", 0, "Status ID")
	cmd.Flags().IntVar(&req.PriorityID, "priority-id", 0, "Priority ID")
	cmd.Flags().IntVar(&req.AssignedToID, "assigned-to-id", 0, "Assigned user ID")
	cmd.Flags().IntVar(&req.CategoryID, "category-id", 0, "Category ID")
	cmd.Flags().IntVar(&req.FixedVersionID, "fixed-version-id", 0, "Target version ID")
	cmd.Flags().StringVar(&req.StartDate, "start-date", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&req.DueDate, "due-date", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&req.DoneRatio, "done-ratio", 0, "Done ratio (0-100)")
	cmd.Flags().StringVar(&req.Notes, "notes", "", "Notes to add")
	return cmd
}

func newDeleteCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "issue")
			if err != nil {
				return err
			}
			if !helpers.ConfirmDelete("issue", id, flags.Yes) {
				return nil
			}
			if flags.DryRun {
				helpers.DryRunDelete("issue", id)
				return nil
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			if err := NewClient(c).Delete(cmd.Context(), id); err != nil {
				return err
			}
			fmt.Printf("Deleted issue #%d\n", id)
			return nil
		},
	}
	return cmd
}
