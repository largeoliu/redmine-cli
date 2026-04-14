// Package config provides configuration management for the CLI.
package config

import (
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// Store handles configuration persistence.
type Store struct {
	configDir string
	keyring   Keyring
}

// NewStore creates a new config store.
func NewStore(configDir string) *Store {
	if configDir == "" {
		configDir = Dir()
	}
	return &Store{
		configDir: configDir,
		keyring:   NewKeyring(),
	}
}

// NewStoreWithKeyring creates a new config store with a specific keyring.
func NewStoreWithKeyring(configDir string, kr Keyring) *Store {
	if configDir == "" {
		configDir = Dir()
	}
	return &Store{
		configDir: configDir,
		keyring:   kr,
	}
}

// Load loads the configuration from disk.
// If a keyring is available, API keys are resolved from the keyring
// instead of the YAML file (which stores them only as a fallback).
func (s *Store) Load() (*Config, error) {
	path := filepath.Clean(filepath.Join(s.configDir, "config.yaml"))
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Instances == nil {
		cfg.Instances = make(map[string]Instance)
	}
	for name, inst := range cfg.Instances {
		if inst.APIKey == "" && s.keyring.IsAvailable() {
			if key, err := s.keyring.Get(name); err == nil {
				inst.APIKey = key
				cfg.Instances[name] = inst
			}
		}
	}
	return &cfg, nil
}

// Save saves the configuration to disk.
func (s *Store) Save(cfg *Config) error {
	if err := os.MkdirAll(s.configDir, 0750); err != nil {
		return err
	}
	path := filepath.Clean(filepath.Join(s.configDir, "config.yaml"))
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// SaveInstance saves an instance to the configuration.
// If the keyring is available, the API key is stored securely in the keyring
// and the YAML file only contains an empty api_key field.
func (s *Store) SaveInstance(name string, inst Instance) error {
	cfg, err := s.Load()
	if err != nil {
		return err
	}
	apiKey := inst.APIKey
	if s.keyring.IsAvailable() && apiKey != "" {
		if err := s.keyring.Set(name, apiKey); err != nil {
			return err
		}
		inst.APIKey = ""
	}
	cfg.Instances[name] = inst
	if cfg.Default == "" {
		cfg.Default = name
	}
	return s.Save(cfg)
}

// SetDefault sets the default instance.
func (s *Store) SetDefault(name string) error {
	cfg, err := s.Load()
	if err != nil {
		return err
	}
	if _, ok := cfg.Instances[name]; !ok {
		return ErrInstanceNotFound
	}
	cfg.Default = name
	return s.Save(cfg)
}

// DeleteInstance deletes an instance from the configuration and keyring.
func (s *Store) DeleteInstance(name string) error {
	cfg, err := s.Load()
	if err != nil {
		return err
	}
	delete(cfg.Instances, name)
	if s.keyring.IsAvailable() {
		if err := s.keyring.Delete(name); err != nil && err != ErrAPIKeyNotFound {
			return err
		}
	}
	if cfg.Default == name {
		cfg.Default = ""
		names := make([]string, 0, len(cfg.Instances))
		for n := range cfg.Instances {
			names = append(names, n)
		}
		sort.Strings(names)
		if len(names) > 0 {
			cfg.Default = names[0]
		}
	}
	return s.Save(cfg)
}
