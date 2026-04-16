// Package app provides the CLI application commands and logic.
package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/config"
	"github.com/largeoliu/redmine-cli/internal/errors"
)

var (
	green = color.New(color.FgGreen).SprintFunc()
	cyan  = color.New(color.FgCyan).SprintFunc()
)

func newLoginCommand(flags *GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Redmine instance",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runLogin(cmd.Context(), flags)
		},
	}
	cmd.Flags().String("name", "", "Instance name")
	cmd.Flags().Bool("set-default", true, "Set as default instance")
	return cmd
}

func runLogin(ctx context.Context, flags *GlobalFlags) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Redmine URL")
	url := promptInput(reader, "", flags.URL)
	if url == "" {
		return errors.NewValidation("URL 不能为空")
	}

	fmt.Print("正在检查连通性... ")
	c := client.NewClient(url, "")
	if err := c.Ping(ctx); err != nil {
		fmt.Println(color.RedString("✗"))
		return err
	}
	fmt.Println(green("✓"))

	fmt.Print("API Key: ")
	apiKey := promptSecret(reader, "")
	if apiKey == "" {
		return errors.NewValidation("API Key 不能为空")
	}

	fmt.Print("正在验证连接... ")
	c.SetAPIKey(apiKey)
	if err := c.TestAuth(ctx); err != nil {
		fmt.Println(color.RedString("✗"))
		return errors.NewAuth("API Key 无效", errors.WithActions(
			"1) API Key 正确",
			"2) 前往 Settings → API access 查看",
		), errors.WithCause(err))
	}
	fmt.Println(green("✓"))

	fmt.Print("实例名称 [default]: ")
	name := promptInput(reader, "", "")
	if name == "" {
		name = "default"
	}

	store := config.NewStore("")
	if err := store.SaveInstance(name, config.Instance{
		URL:    url,
		APIKey: apiKey,
	}); err != nil {
		return errors.NewInternal("配置保存失败", errors.WithCause(err))
	}

	setDefault := true
	cfg, err := store.Load()
	if err == nil && cfg != nil && cfg.Default != "" && cfg.Default != name {
		fmt.Print("设为默认实例? [Y/n]: ")
		setDefault = promptBool(reader, "", true)
	}

	if setDefault {
		if err := store.SetDefault(name); err != nil {
			return errors.NewInternal("设置默认实例失败", errors.WithCause(err))
		}
	}

	fmt.Println()
	fmt.Printf("%s 登录成功！配置已保存到 %s\n", green("✓"), config.Path())
	return nil
}

func promptInput(reader *bufio.Reader, _ string, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("[%s]: ", cyan(defaultValue))
	} else {
		fmt.Print(": ")
	}
	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

func promptSecret(reader *bufio.Reader, prompt string) string {
	if prompt != "" {
		fmt.Print(prompt + ": ")
	}

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		input, _ := reader.ReadString('\n')
		return strings.TrimSpace(input)
	}
	defer func() {
		_ = term.Restore(fd, oldState)
	}()

	input, err := term.ReadPassword(fd)
	fmt.Println()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(input))
}

func promptBool(reader *bufio.Reader, prompt string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "Y"
	}
	input := promptInput(reader, prompt, defaultStr)
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return defaultValue
	}
	return input == "y" || input == "yes"
}
