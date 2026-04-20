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

			report, cards, err := buildBoardReport(cmd.Context(), c, project)
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
	return cmd
}

func resolveProject(ctx context.Context, c *client.Client, value string) (*projects.Project, error) {
	projectClient := projects.NewClient(c)

	if id, err := strconv.Atoi(value); err == nil {
		return projectClient.Get(ctx, id, nil)
	}

	project, err := projectClient.GetByIdentifier(ctx, value, nil)
	if err != nil {
		return nil, errors.NewValidation("project not found: "+value, errors.WithCause(err))
	}
	return project, nil
}
