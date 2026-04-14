// Package time_entries provides commands for managing Redmine time entries.
//
//nolint:revive // package name has underscore due to Redmine API naming convention
package time_entries

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/largeoliu/redmine-cli/internal/errors"
	"github.com/largeoliu/redmine-cli/internal/resources/helpers"
	"github.com/largeoliu/redmine-cli/internal/types"
)

func NewCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "time-entry",
		Short:   "Manage Redmine time entries",
		Aliases: []string{"time-entries"},
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
		Short: "List time entries",
		RunE: func(cmd *cobra.Command, _ []string) error {
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
	cmd.Flags().IntVar(&listFlags.IssueID, "issue-id", 0, "Filter by issue ID")
	cmd.Flags().IntVar(&listFlags.UserID, "user-id", 0, "Filter by user ID")
	cmd.Flags().IntVar(&listFlags.ActivityID, "activity-id", 0, "Filter by activity ID")
	cmd.Flags().StringVar(&listFlags.From, "from", "", "Filter by start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&listFlags.To, "to", "", "Filter by end date (YYYY-MM-DD)")
	return cmd
}

func newGetCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get time entry details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "time entry")
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
	req := &TimeEntryCreateRequest{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new time entry",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if req.IssueID == 0 && req.ProjectID == 0 {
				return errors.NewValidation("issue-id or project-id is required")
			}
			if req.Hours == 0 {
				return errors.NewValidation("hours is required")
			}
			if req.SpentOn == "" {
				return errors.NewValidation("spent-on is required")
			}
			if flags.DryRun {
				helpers.DryRunCreate("time entry", req)
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
			fmt.Printf("Created time entry #%d: %.2f hours\n", result.ID, result.Hours)
			return resolver.WriteOutput(cmd.OutOrStdout(), flags, result)
		},
	}
	cmd.Flags().IntVar(&req.IssueID, "issue-id", 0, "Issue ID")
	cmd.Flags().IntVar(&req.ProjectID, "project-id", 0, "Project ID")
	cmd.Flags().StringVar(&req.SpentOn, "spent-on", "", "Spent on date (YYYY-MM-DD) (required)")
	cmd.Flags().Float64Var(&req.Hours, "hours", 0, "Hours spent (required)")
	cmd.Flags().IntVar(&req.ActivityID, "activity-id", 0, "Activity ID")
	cmd.Flags().StringVar(&req.Comments, "comments", "", "Comments")
	return cmd
}

func newUpdateCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	req := &TimeEntryUpdateRequest{}
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a time entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "time entry")
			if err != nil {
				return err
			}
			if flags.DryRun {
				helpers.DryRunUpdate("time entry", id, req)
				return nil
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			if err := NewClient(c).Update(cmd.Context(), id, req); err != nil {
				return err
			}
			fmt.Printf("Updated time entry #%d\n", id)
			return nil
		},
	}
	cmd.Flags().IntVar(&req.IssueID, "issue-id", 0, "Issue ID")
	cmd.Flags().IntVar(&req.ProjectID, "project-id", 0, "Project ID")
	cmd.Flags().StringVar(&req.SpentOn, "spent-on", "", "Spent on date (YYYY-MM-DD)")
	cmd.Flags().Float64Var(&req.Hours, "hours", 0, "Hours spent")
	cmd.Flags().IntVar(&req.ActivityID, "activity-id", 0, "Activity ID")
	cmd.Flags().StringVar(&req.Comments, "comments", "", "Comments")
	return cmd
}

func newDeleteCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a time entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "time entry")
			if err != nil {
				return err
			}
			if !helpers.ConfirmDelete("time entry", id, flags.Yes) {
				return nil
			}
			if flags.DryRun {
				helpers.DryRunDelete("time entry", id)
				return nil
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			if err := NewClient(c).Delete(cmd.Context(), id); err != nil {
				return err
			}
			fmt.Printf("Deleted time entry #%d\n", id)
			return nil
		},
	}
	return cmd
}
