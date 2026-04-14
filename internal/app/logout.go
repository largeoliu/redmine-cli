// Package app provides the CLI application commands and logic.
package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/largeoliu/redmine-cli/internal/config"
	"github.com/largeoliu/redmine-cli/internal/errors"
)

func newLogoutCommand(flags *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from a Redmine instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceName := ""
			if len(args) > 0 {
				instanceName = args[0]
			}
			return runLogout(cmd.Context(), flags, instanceName)
		},
	}
	return cmd
}

func runLogout(_ context.Context, flags *GlobalFlags, instanceName string) error {
	store := config.NewStore("")

	if instanceName == "" {
		cfg, err := store.Load()
		if err != nil {
			return errors.NewInternal("failed to load config", errors.WithCause(err))
		}
		if cfg.Default == "" {
			return errors.NewValidation("No default instance configured")
		}
		instanceName = cfg.Default
	}

	cfg, err := store.Load()
	if err != nil {
		return errors.NewInternal("failed to load config", errors.WithCause(err))
	}
	inst, ok := cfg.Instances[instanceName]
	if !ok {
		return errors.NewValidation(fmt.Sprintf("Instance %q not found", instanceName))
	}

	if !flags.Yes {
		fmt.Printf("实例: %s\n", instanceName)
		fmt.Printf("URL: %s\n\n", inst.URL)
		fmt.Print("确定要删除此实例吗？请输入 \"yes\" 确认: ")

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil && input == "" {
			fmt.Println("取消删除")
			return nil
		}
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "yes" {
			fmt.Println("取消删除")
			return nil
		}
	}

	if err := store.DeleteInstance(instanceName); err != nil {
		return errors.NewInternal("failed to delete instance", errors.WithCause(err))
	}

	fmt.Printf("已删除实例: %s\n", instanceName)
	return nil
}
