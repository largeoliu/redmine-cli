package agile

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/errors"
	"github.com/largeoliu/redmine-cli/internal/resources/projects"
	"github.com/largeoliu/redmine-cli/internal/types"
)

// NewCommand creates a new agile command with content-oriented subcommands.
func NewCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agile",
		Short: "View Redmine agile content",
	}
	cmd.AddCommand(newBoardCommand(flags, resolver))
	return cmd
}

func newBoardCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	sprintSelector := "current"
	trackerSelector := "全部"
	cmd := &cobra.Command{
		Use:   "board <project>",
		Short: "Show sprint content for a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}

			project, err := resolveProject(cmd.Context(), c, args[0])
			if err != nil {
				return err
			}

			report, cards, err := buildBoardReportWithOptions(cmd.Context(), c, project, boardOptions{
				Sprint:  sprintSelector,
				Tracker: trackerSelector,
			})
			if err != nil {
				return err
			}

			switch strings.ToLower(flags.Format) {
			case "", "json":
				return resolver.WriteOutput(cmd.OutOrStdout(), flags, report)
			case "table":
				return resolver.WriteOutput(cmd.OutOrStdout(), flags, cards)
			default:
				_, err := fmt.Fprint(cmd.OutOrStdout(), renderBoardReport(report))
				return err
			}
		},
	}
	cmd.Flags().StringVar(&sprintSelector, "sprint", "current", "Sprint selector: current or a sprint ID")
	cmd.Flags().StringVar(&trackerSelector, "tracker", "全部", "Filter by tracker name (use 全部 to show all trackers)")
	return cmd
}

func resolveProject(ctx context.Context, c *client.Client, value string) (*projects.Project, error) {
	projectClient := projects.NewClient(c)

	if id, err := strconv.Atoi(value); err == nil {
		project, err := projectClient.Get(ctx, id, nil)
		if err != nil {
			return nil, wrapProjectLookupError(value, err)
		}
		return project, nil
	}

	project, err := projectClient.GetByIdentifier(ctx, value, nil)
	if err != nil {
		return nil, wrapProjectLookupError(value, err)
	}
	return project, nil
}

func wrapProjectLookupError(value string, err error) error {
	if !isNotFoundError(err) {
		return err
	}

	return errors.NewValidation(
		"project not found: "+value,
		errors.WithHint("Use a project ID or identifier. Run 'redmine project list --fields id,identifier,name' to find the correct value."),
		errors.WithActions("redmine project list --fields id,identifier,name"),
		errors.WithCause(err),
	)
}

func isNotFoundError(err error) bool {
	var appErr *errors.Error
	if !errors.As(err, &appErr) {
		return false
	}
	return appErr.Category == errors.CategoryAPI && strings.EqualFold(appErr.Message, "resource not found")
}
