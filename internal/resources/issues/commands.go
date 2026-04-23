// Package issues provides commands for managing Redmine issues.
package issues

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/errors"
	"github.com/largeoliu/redmine-cli/internal/resources/helpers"
	"github.com/largeoliu/redmine-cli/internal/resources/trackers"
	"github.com/largeoliu/redmine-cli/internal/types"
)

type sprintListResponse struct {
	AgileSprints []sprintRef `json:"agile_sprints"`
}

type sprintRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func resolveSprintID(ctx context.Context, c *client.Client, projectID int, selector string) (int, error) {
	if selector == "" {
		return 0, nil
	}
	if id, err := strconv.Atoi(selector); err == nil {
		if id <= 0 {
			return 0, errors.NewValidation("--sprint must be a positive integer or sprint name")
		}
		return id, nil
	}

	// Avoid importing internal/resources/agile here to prevent an import cycle:
	// agile uses issues (board report), and issues needs sprint lookup.
	var resp sprintListResponse
	if err := c.Get(ctx, fmt.Sprintf("/projects/%d/agile_sprints.json", projectID), &resp); err != nil {
		return 0, err
	}

	var matches []sprintRef
	for _, s := range resp.AgileSprints {
		if s.Name == selector {
			matches = append(matches, s)
		}
	}

	switch len(matches) {
	case 0:
		return 0, errors.NewValidation("sprint not found: " + selector)
	case 1:
		return matches[0].ID, nil
	default:
		ids := make([]string, 0, len(matches))
		for _, s := range matches {
			ids = append(ids, strconv.Itoa(s.ID))
		}
		return 0, errors.NewValidation("multiple sprints match name: " + selector + " (ids: " + strings.Join(ids, ",") + ")")
	}
}

type customFieldFlags struct {
	Fields []string
}

func parseCustomFieldFlags(fields []string, tracker *trackers.Tracker) ([]CustomField, error) {
	if len(fields) == 0 {
		return nil, nil
	}

	cfMap := make(map[int]CustomField)

	for _, f := range fields {
		parts := strings.SplitN(f, ":", 3)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid custom field format: %s (expected name:value or id:X:value)", f)
		}

		cf := CustomField{}
		if parts[0] == "id" {
			if len(parts) < 3 {
				return nil, fmt.Errorf("invalid custom field format: %s (expected id:X:value)", f)
			}
			var id int
			if _, err := fmt.Sscanf(parts[1], "%d", &id); err != nil {
				return nil, fmt.Errorf("invalid custom field id: %s", parts[1])
			}
			cf.ID = id
			cf.Value = parts[2]
		} else {
			if tracker == nil {
				return nil, fmt.Errorf("tracker required to match custom field by name, use id:X:value format instead")
			}
			found := false
			for _, tf := range tracker.CustomFields {
				if tf.Name == parts[0] {
					cf.ID = tf.ID
					cf.Value = parts[1]
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("custom field not found: %s", parts[0])
			}
		}
		cfMap[cf.ID] = cf
	}

	result := make([]CustomField, 0, len(cfMap))
	for _, cf := range cfMap {
		result = append(result, cf)
	}
	return result, nil
}

// isTerminalFunc 检查是否为终端环境，可被测试覆盖
var isTerminalFunc = func() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// newStdinReader 创建标准输入读取器，可被测试覆盖
var newStdinReader = newBufioReader

// newBufioReader 创建一个从 os.Stdin 读取的 bufio.Reader
func newBufioReader() *bufio.Reader {
	return bufio.NewReader(os.Stdin)
}

//nolint:gocyclo
func promptCustomFields(tracker *trackers.Tracker, initialValues map[int]CustomField) ([]CustomField, error) {
	if len(tracker.CustomFields) == 0 {
		return nil, nil
	}

	if !isTerminalFunc() {
		return nil, nil
	}

	reader := newStdinReader()
	return promptCustomFieldsInteractive(tracker, initialValues, reader)
}

// promptCustomFieldsInteractive 是可测试的交互式自定义字段提示函数
//
//nolint:gocyclo
func promptCustomFieldsInteractive(tracker *trackers.Tracker, initialValues map[int]CustomField, inputReader *bufio.Reader) ([]CustomField, error) {
	values := make(map[int]CustomField)
	for k, v := range initialValues {
		values[k] = v
	}

	fmt.Println("\n📋 Custom fields detected, please fill in:")

	for _, cf := range tracker.CustomFields {
		current, hasCurrent := values[cf.ID]
		defaultVal := ""
		if hasCurrent {
			defaultVal = fmt.Sprintf("%v", current.Value)
		}

		fmt.Printf("[%d] %s\n", cf.ID, cf.Name)
		fmt.Printf("    Type: %s\n", cf.FieldFormat)

		var input string

		switch cf.FieldFormat {
		case "list":
			if len(cf.PossibleValues) > 0 {
				fmt.Print("    Options: ")
				for i, pv := range cf.PossibleValues {
					fmt.Printf("[%d. %s] ", i+1, pv.Label)
				}
				fmt.Println()
				fmt.Print("    Select: ")
				input, _ = inputReader.ReadString('\n') //nolint:errcheck,gosec
				input = strings.TrimSpace(input)
				if input != "" {
					idx, err := strconv.Atoi(input)
					if err == nil && idx >= 1 && idx <= len(cf.PossibleValues) {
						values[cf.ID] = CustomField{ID: cf.ID, Value: cf.PossibleValues[idx-1].Value}
					}
				} else if hasCurrent {
					values[cf.ID] = current
				}
			}
		case "bool":
			fmt.Print("    (y/n): ")
			input, _ = inputReader.ReadString('\n') //nolint:errcheck,gosec
			input = strings.TrimSpace(input)
			if input == "y" || input == "Y" {
				values[cf.ID] = CustomField{ID: cf.ID, Value: "1"}
			} else if input == "n" || input == "N" {
				values[cf.ID] = CustomField{ID: cf.ID, Value: "0"}
			} else if hasCurrent {
				values[cf.ID] = current
			}
		case "date":
			fmt.Printf("    Input (YYYY-MM-DD) [%s]: ", defaultVal)
			input, _ = inputReader.ReadString('\n') //nolint:errcheck,gosec
			input = strings.TrimSpace(input)
			if input == "" && defaultVal != "" {
				values[cf.ID] = CustomField{ID: cf.ID, Value: defaultVal}
			} else if input != "" {
				values[cf.ID] = CustomField{ID: cf.ID, Value: input}
			}
		default:
			fmt.Printf("    Input [%s]: ", defaultVal)
			input, _ = inputReader.ReadString('\n') //nolint:errcheck,gosec
			input = strings.TrimSpace(input)
			if input != "" {
				values[cf.ID] = CustomField{ID: cf.ID, Value: input}
			} else if hasCurrent {
				values[cf.ID] = current
			}
		}
		fmt.Println()
	}

	result := make([]CustomField, 0, len(values))
	for _, v := range values {
		result = append(result, v)
	}
	return result, nil
}

func mergeCustomFields(interactive, flags []CustomField) []CustomField {
	cfMap := make(map[int]CustomField)
	for _, cf := range interactive {
		cfMap[cf.ID] = cf
	}
	for _, cf := range flags {
		cfMap[cf.ID] = cf
	}
	result := make([]CustomField, 0, len(cfMap))
	for _, cf := range cfMap {
		result = append(result, cf)
	}
	return result
}

// NewCommand creates a new issue command with all subcommands.
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
	var trackerSelector string
	var sprintSelector string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if listFlags.ProjectID == 0 {
				return errors.NewValidation("--project-id is required")
			}
			statusChanged := cmd.Flags().Changed("status-id")
			// 使用全局标志的 limit 和 offset（如果设置了的话）
			if flags.Limit > 0 {
				listFlags.Limit = flags.Limit
			}
			if flags.Offset > 0 {
				listFlags.Offset = flags.Offset
			}

			c, resolveErr := resolver.ResolveClient(flags)
			if resolveErr != nil {
				return resolveErr
			}
			if cmd.Flags().Changed("sprint") {
				sprintID, sprintErr := resolveSprintID(cmd.Context(), c, listFlags.ProjectID, sprintSelector)
				if sprintErr != nil {
					return sprintErr
				}
				listFlags.SprintID = sprintID
			}
			if cmd.Flags().Changed("tracker") {
				switch {
				case trackerSelector == "", strings.EqualFold(trackerSelector, "全部"), strings.EqualFold(trackerSelector, "all"):
					listFlags.TrackerID = nil
				default:
					trackerDef, trackerErr := trackers.NewClient(c).FindByName(cmd.Context(), trackerSelector)
					if trackerErr != nil {
						return trackerErr
					}
					listFlags.TrackerID = []int{trackerDef.ID}
				}
			}
			params := BuildListParams(*listFlags)
			// Redmine defaults to "open" when status_id is omitted.
			// If user did not explicitly set --status-id, request all statuses.
			if !statusChanged {
				params["status_id"] = "*"
			}
			result, err := NewClient(c).List(cmd.Context(), params)
			if err != nil {
				return err
			}
			return resolver.WriteOutput(cmd.OutOrStdout(), flags, result)
		},
	}
	cmd.Flags().IntVar(&listFlags.ProjectID, "project-id", 0, "Filter by project ID")
	cmd.Flags().StringVar(&sprintSelector, "sprint", "", "Filter by sprint ID or exact sprint name")
	cmd.Flags().StringVar(&trackerSelector, "tracker", "", "Filter by tracker name (use 全部 to skip filtering)")
	cmd.Flags().IntSliceVar(&listFlags.TrackerID, "tracker-id", nil, "Filter by tracker ID (can be specified multiple times)")
	cmd.Flags().IntSliceVar(&listFlags.VersionID, "version-id", nil, "Filter by fixed version ID (can be specified multiple times)")
	cmd.Flags().IntSliceVar(&listFlags.StatusID, "status-id", nil, "Filter by status ID (can be specified multiple times)")
	cmd.Flags().IntSliceVar(&listFlags.AssignedToID, "assigned-to-id", nil, "Filter by assigned user ID (can be specified multiple times)")
	cmd.Flags().StringSliceVar(&listFlags.Query, "query", nil, "Filter by custom query ID (can be specified multiple times)")
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
	createFlags := &customFieldFlags{}
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
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			var trackerDef *trackers.Tracker
			if req.TrackerID > 0 {
				trackerDef, _ = trackers.NewClient(c).Get(cmd.Context(), req.TrackerID) //nolint:errcheck
			}
			var flagCFs []CustomField
			if len(createFlags.Fields) > 0 {
				flagCFs, err = parseCustomFieldFlags(createFlags.Fields, trackerDef)
				if err != nil {
					return err
				}
			}
			var interactiveCFs []CustomField
			if trackerDef != nil {
				interactiveCFs, _ = promptCustomFields(trackerDef, nil) //nolint:errcheck
			}
			mergedCFs := mergeCustomFields(interactiveCFs, flagCFs)
			if len(mergedCFs) > 0 {
				req.CustomFields = mergedCFs
			}
			if flags.DryRun {
				helpers.DryRunCreate("issue", req)
				return nil
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
	cmd.Flags().IntVar(&req.FixedVersionID, "version-id", 0, "Target version ID")
	cmd.Flags().StringVar(&req.StartDate, "start-date", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&req.DueDate, "due-date", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&req.DoneRatio, "done-ratio", 0, "Done ratio (0-100)")
	cmd.Flags().StringSliceVar(&createFlags.Fields, "custom-field", nil, "Custom field value (format: name:value or id:X:value, can be specified multiple times)")
	return cmd
}

func newUpdateCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	req := &IssueUpdateRequest{}
	updateFlags := &customFieldFlags{}
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an issue",
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
			var flagCFs []CustomField
			if len(updateFlags.Fields) > 0 {
				flagCFs, err = parseCustomFieldFlags(updateFlags.Fields, nil)
				if err != nil {
					return err
				}
			}
			if len(flagCFs) > 0 {
				req.CustomFields = flagCFs
			}
			if flags.DryRun {
				helpers.DryRunUpdate("issue", id, req)
				return nil
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
	cmd.Flags().IntVar(&req.FixedVersionID, "version-id", 0, "Target version ID")
	cmd.Flags().StringVar(&req.StartDate, "start-date", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&req.DueDate, "due-date", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&req.DoneRatio, "done-ratio", 0, "Done ratio (0-100)")
	cmd.Flags().StringVar(&req.Notes, "notes", "", "Notes to add")
	cmd.Flags().StringSliceVar(&updateFlags.Fields, "custom-field", nil, "Custom field value (format: name:value or id:X:value, can be specified multiple times)")
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
