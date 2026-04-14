// Package users provides commands for managing Redmine users.
package users

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/largeoliu/redmine-cli/internal/errors"
	"github.com/largeoliu/redmine-cli/internal/resources/helpers"
	"github.com/largeoliu/redmine-cli/internal/types"
)

// NewCommand creates a new users command.
func NewCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "user",
		Short:   "Manage Redmine users",
		Aliases: []string{"users"},
	}
	cmd.AddCommand(
		newListCommand(flags, resolver),
		newGetCommand(flags, resolver),
		newGetCurrentCommand(flags, resolver),
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
		Short: "List users",
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
	cmd.Flags().IntVar(&listFlags.Status, "status", 0, "Filter by status (1=active, 2=registered, 3=locked)")
	cmd.Flags().StringVar(&listFlags.Name, "name", "", "Filter by name (login, firstname, lastname)")
	cmd.Flags().IntVar(&listFlags.GroupID, "group-id", 0, "Filter by group ID")
	return cmd
}

func newGetCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	var include string
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get user details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "user")
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
	cmd.Flags().StringVar(&include, "include", "", "Include associated data (memberships,groups)")
	return cmd
}

func newGetCurrentCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-self",
		Short: "Get current user details",
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			result, err := NewClient(c).GetCurrent(cmd.Context())
			if err != nil {
				return err
			}
			return resolver.WriteOutput(cmd.OutOrStdout(), flags, result)
		},
	}
	return cmd
}

func newCreateCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	req := &UserCreateRequest{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new user (admin only)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if req.Login == "" {
				return errors.NewValidation("login is required")
			}
			if req.Firstname == "" {
				return errors.NewValidation("firstname is required")
			}
			if req.Lastname == "" {
				return errors.NewValidation("lastname is required")
			}
			if req.Mail == "" {
				return errors.NewValidation("mail is required")
			}
			if req.Password == "" {
				return errors.NewValidation("password is required")
			}
			if flags.DryRun {
				helpers.DryRunCreate("user", req)
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
			fmt.Printf("Created user #%d: %s\n", result.ID, result.Login)
			return resolver.WriteOutput(cmd.OutOrStdout(), flags, result)
		},
	}
	cmd.Flags().StringVar(&req.Login, "login", "", "Login name (required)")
	cmd.Flags().StringVar(&req.Firstname, "firstname", "", "First name (required)")
	cmd.Flags().StringVar(&req.Lastname, "lastname", "", "Last name (required)")
	cmd.Flags().StringVar(&req.Mail, "mail", "", "Email address (required)")
	cmd.Flags().StringVar(&req.Password, "password", "", "Password (required)")
	cmd.Flags().BoolVar(&req.Admin, "admin", false, "Set as administrator")
	cmd.Flags().IntVar(&req.Status, "status", 1, "User status (1=active, 2=registered, 3=locked)")
	cmd.Flags().IntVar(&req.AuthSourceID, "auth-source-id", 0, "Authentication source ID")
	cmd.Flags().BoolVar(&req.MustChangePassword, "must-change-password", false, "User must change password at next login")
	return cmd
}

func newUpdateCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	req := &UserUpdateRequest{}
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "user")
			if err != nil {
				return err
			}
			if flags.DryRun {
				helpers.DryRunUpdate("user", id, req)
				return nil
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			if err := NewClient(c).Update(cmd.Context(), id, req); err != nil {
				return err
			}
			fmt.Printf("Updated user #%d\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.Login, "login", "", "Login name")
	cmd.Flags().StringVar(&req.Firstname, "firstname", "", "First name")
	cmd.Flags().StringVar(&req.Lastname, "lastname", "", "Last name")
	cmd.Flags().StringVar(&req.Mail, "mail", "", "Email address")
	cmd.Flags().StringVar(&req.Password, "password", "", "Password")
	cmd.Flags().BoolVar(&req.Admin, "admin", false, "Set as administrator")
	cmd.Flags().IntVar(&req.Status, "status", 0, "User status (1=active, 2=registered, 3=locked)")
	cmd.Flags().IntVar(&req.AuthSourceID, "auth-source-id", 0, "Authentication source ID")
	cmd.Flags().BoolVar(&req.MustChangePassword, "must-change-password", false, "User must change password at next login")
	return cmd
}

func newDeleteCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a user (admin only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "user")
			if err != nil {
				return err
			}
			if !helpers.ConfirmDelete("user", id, flags.Yes) {
				return nil
			}
			if flags.DryRun {
				helpers.DryRunDelete("user", id)
				return nil
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			if err := NewClient(c).Delete(cmd.Context(), id); err != nil {
				return err
			}
			fmt.Printf("Deleted user #%d\n", id)
			return nil
		},
	}
	return cmd
}
