// Package app provides the CLI application commands and logic.
package app

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/largeoliu/redmine-cli/internal/config"
	"github.com/largeoliu/redmine-cli/internal/errors"
)

func newConfigCommand(flags *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
	}
	cmd.AddCommand(
		newConfigGetCommand(flags),
		newConfigSetCommand(flags),
		newConfigListCommand(flags),
	)
	return cmd
}

func newConfigGetCommand(_ *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Show current configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
			store := config.NewStore("")
			cfg, err := store.Load()
			if err != nil {
				return errors.NewInternal("failed to load config", errors.WithCause(err))
			}

			fmt.Printf("Config file: %s\n\n", config.Path())
			fmt.Printf("Default instance: %s\n", cfg.Default)
			if len(cfg.Instances) == 0 {
				fmt.Println("\nNo instances configured.")
				fmt.Println("Run 'redmine login' to add an instance.")
				return nil
			}

			fmt.Println("\nInstances:")
			for name, inst := range cfg.Instances {
				marker := " "
				if name == cfg.Default {
					marker = "*"
				}
				fmt.Printf("  %s %s: %s\n", marker, name, inst.URL)
			}
			return nil
		},
	}
}

func newConfigSetCommand(_ *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "set <instance-name>",
		Short: "Set default instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			store := config.NewStore("")
			if err := store.SetDefault(name); err != nil {
				return err
			}
			fmt.Printf("Set default instance to: %s\n", name)
			return nil
		},
	}
}

func newConfigListCommand(flags *GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configured instances",
		RunE: func(cmd *cobra.Command, _ []string) error {
			store := config.NewStore("")
			cfg, err := store.Load()
			if err != nil {
				return errors.NewInternal("failed to load config", errors.WithCause(err))
			}

			result := make([]map[string]any, 0, len(cfg.Instances))
			for name, inst := range cfg.Instances {
				result = append(result, map[string]any{
					"name":    name,
					"url":     inst.URL,
					"default": name == cfg.Default,
				})
			}
			return WriteOutput(cmd.OutOrStdout(), flags, map[string]any{
				"instances": result,
			})
		},
	}
}
