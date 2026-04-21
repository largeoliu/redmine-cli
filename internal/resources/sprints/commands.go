package sprints

import (
	"context"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/errors"
	agilepkg "github.com/largeoliu/redmine-cli/internal/resources/agile"
	projectspkg "github.com/largeoliu/redmine-cli/internal/resources/projects"
	"github.com/largeoliu/redmine-cli/internal/types"
)

// NewCommand creates a new sprint command.
func NewCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sprint",
		Short:   "View Redmine sprints",
		Aliases: []string{"sprints"},
	}
	cmd.AddCommand(newListCommand(flags, resolver))
	return cmd
}

func newListCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project>",
		Short: "List project sprints",
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

			result, err := agilepkg.NewClient(c).ListSprints(cmd.Context(), project.ID)
			if err != nil {
				return err
			}

			return resolver.WriteOutput(cmd.OutOrStdout(), flags, result.AgileSprints)
		},
	}
	return cmd
}

func resolveProject(ctx context.Context, c *client.Client, value string) (*projectspkg.Project, error) {
	projectClient := projectspkg.NewClient(c)

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
