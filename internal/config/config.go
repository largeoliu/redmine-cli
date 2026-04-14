// Package config provides configuration management for the CLI.
package config

import (
	"os"
	"path/filepath"
	"time"
)

// Config holds the CLI configuration.
type Config struct {
	Default   string              `yaml:"default"`
	Instances map[string]Instance `yaml:"instances"`
	Settings  Settings            `yaml:"settings"`
	Git       GitConfig           `yaml:"git"`
	Report    ReportConfig        `yaml:"report"`
}

// Instance represents a Redmine instance configuration.
type Instance struct {
	Name   string `yaml:"-"`
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

// Settings holds general CLI settings.
type Settings struct {
	Timeout      time.Duration `yaml:"timeout"`
	Retries      int           `yaml:"retries"`
	OutputFormat string        `yaml:"output_format"`
	PageSize     int           `yaml:"page_size"`
}

// GitConfig holds Git integration settings.
type GitConfig struct {
	AutoLink       bool   `yaml:"auto_link"`
	CommitPattern  string `yaml:"commit_pattern"`
	DefaultProject int    `yaml:"default_project"`
}

// ReportConfig holds report generation settings.
type ReportConfig struct {
	TemplatesDir  string `yaml:"templates_dir"`
	DefaultFormat string `yaml:"default_format"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Instances: make(map[string]Instance),
		Settings: Settings{
			Timeout:      30 * time.Second,
			Retries:      3,
			OutputFormat: "json",
			PageSize:     100,
		},
		Git: GitConfig{
			AutoLink:      true,
			CommitPattern: `#(\d+)`,
		},
		Report: ReportConfig{
			TemplatesDir:  "",
			DefaultFormat: "table",
		},
	}
}

// CurrentInstance returns the current default instance.
func (c *Config) CurrentInstance() (*Instance, bool) {
	if c.Default == "" {
		return nil, false
	}
	inst, ok := c.Instances[c.Default]
	if !ok {
		return nil, false
	}
	inst.Name = c.Default
	return &inst, true
}

// GetInstance returns an instance by name.
func (c *Config) GetInstance(name string) (*Instance, bool) {
	inst, ok := c.Instances[name]
	if !ok {
		return nil, false
	}
	inst.Name = name
	return &inst, true
}

// Dir returns the configuration directory.
func Dir() string {
	if dir := os.Getenv("REDMINE_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/.redmine-cli"
	}
	return filepath.Join(home, ".redmine-cli")
}

// Path returns the configuration file path.
func Path() string {
	return filepath.Join(Dir(), "config.yaml")
}
