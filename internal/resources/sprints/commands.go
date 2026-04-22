package sprints

import (
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/largeoliu/redmine-cli/internal/errors"
	agilepkg "github.com/largeoliu/redmine-cli/internal/resources/agile"
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
	cmd.AddCommand(newGetCommand(flags, resolver))
	return cmd
}

func newListCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	var projectID int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project sprints",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if projectID == 0 {
				return errors.NewValidation("--project is required")
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}

			result, err := agilepkg.NewClient(c).ListSprints(cmd.Context(), projectID)
			if err != nil {
				return err
			}

			return resolver.WriteOutput(cmd.OutOrStdout(), flags, enrichSprintStatus(result.AgileSprints))
		},
	}
	cmd.Flags().IntVar(&projectID, "project", 0, "Project ID (required)")
	return cmd
}

func newGetCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	var projectID int
	cmd := &cobra.Command{
		Use:   "get <sprint_id>",
		Short: "Show sprint details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if projectID == 0 {
				return errors.NewValidation("--project-id is required")
			}
			sprintID, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.NewValidation("sprint_id must be an integer")
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			sprint, err := agilepkg.NewClient(c).GetSprint(cmd.Context(), projectID, sprintID)
			if err != nil {
				return err
			}
			return resolver.WriteOutput(cmd.OutOrStdout(), flags, sprint)
		},
	}
	cmd.Flags().IntVar(&projectID, "project-id", 0, "Project ID (required)")
	return cmd
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
