package sprints

import (
	"context"
	"strconv"
	"strings"
	"time"

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
	details := false

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

			payload := result.AgileSprints
			if details {
				// 列表端点已包含完整数据，直接使用
				payload = enrichSprintStatus(result.AgileSprints)
			}

			return resolver.WriteOutput(cmd.OutOrStdout(), flags, payload)
		},
	}
	cmd.Flags().BoolVar(&details, "details", false, "Expand each sprint with full details")
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

// enrichSprintStatus 根据日期为没有 status 字段的 sprint 推断状态
func enrichSprintStatus(sprints []agilepkg.Sprint) []agilepkg.Sprint {
	now := time.Now().UTC()
	result := make([]agilepkg.Sprint, len(sprints))

	for i, sprint := range sprints {
		result[i] = sprint

		// 如果 API 已经返回了 status，跳过
		if sprint.Status != "" {
			continue
		}

		// 根据日期推断状态
		if sprint.IsClosed {
			result[i].Status = "closed"
			continue
		}

		if sprint.IsArchived {
			result[i].Status = "archived"
			continue
		}

		// 检查是否是当前 sprint
		if sprint.StartDate != "" && sprint.EndDate != "" {
			start, startErr := time.Parse("2006-01-02", sprint.StartDate)
			end, endErr := time.Parse("2006-01-02", sprint.EndDate)

			if startErr == nil && endErr == nil {
				if !start.After(now) && !end.Before(now) {
					result[i].Status = "active"
					continue
				}
				if end.Before(now) {
					result[i].Status = "closed"
					continue
				}
				if start.After(now) {
					result[i].Status = "open"
					continue
				}
			}
		}

		// 默认为 open
		if result[i].Status == "" {
			result[i].Status = "open"
		}
	}

	return result
}
