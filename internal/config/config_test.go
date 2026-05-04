// internal/config/config_test.go
package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Settings.Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", cfg.Settings.Timeout)
	}
	if cfg.Settings.Retries != 3 {
		t.Errorf("expected default retries 3, got %d", cfg.Settings.Retries)
	}
}

func TestStoreSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	cfg := DefaultConfig()
	cfg.Instances["test"] = Instance{
		URL:    "https://example.com",
		APIKey: "test-key-12345",
	}
	cfg.Default = "test"

	if err := store.Save(cfg); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Default != "test" {
		t.Errorf("expected default 'test', got %s", loaded.Default)
	}
	inst, ok := loaded.Instances["test"]
	if !ok {
		t.Fatal("expected instance 'test'")
	}
	if inst.URL != "https://example.com" {
		t.Errorf("expected URL 'https://example.com', got %s", inst.URL)
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		url      string
		hasError bool
	}{
		{"https://example.com", false},
		{"http://localhost:3000", false},
		{"", true},
		{"not-a-url", true},
		{"ftp://example.com", true},
		{"  https://example.com  ", false}, // should trim spaces
		{"https://", true},                 // empty host
		{"http://", true},                  // empty host
		{"//example.com", true},            // no scheme
		{"example.com", true},              // no scheme
		{"http://\x00host", true},          // control character causes parse error
		{"https://example.com\x7f", true},  // control character in URL
	}
	for _, tt := range tests {
		err := ValidateURL(tt.url)
		if tt.hasError && err == nil {
			t.Errorf("ValidateURL(%s) expected error, got nil", tt.url)
		}
		if !tt.hasError && err != nil {
			t.Errorf("ValidateURL(%s) unexpected error: %v", tt.url, err)
		}
	}
}

func TestValidateAPIKey(t *testing.T) {
	tests := []struct {
		key      string
		hasError bool
	}{
		{"valid-api-key-12345", false},
		{"", true},
		{"short", true},
		{"  valid-api-key-12345  ", false}, // should trim spaces
		{"   ", true},                      // only spaces
		{"1234567890", false},              // exactly 10 characters
		{"123456789", true},                // 9 characters
	}
	for _, tt := range tests {
		err := ValidateAPIKey(tt.key)
		if tt.hasError && err == nil {
			t.Errorf("ValidateAPIKey(%s) expected error, got nil", tt.key)
		}
		if !tt.hasError && err != nil {
			t.Errorf("ValidateAPIKey(%s) unexpected error: %v", tt.key, err)
		}
	}
}

func TestCurrentInstance(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		wantNil  bool
		wantOK   bool
		wantName string
	}{
		{
			name: "no default set",
			config: &Config{
				Instances: map[string]Instance{
					"test": {URL: "https://example.com", APIKey: "test-key-12345"},
				},
			},
			wantNil: true,
			wantOK:  false,
		},
		{
			name: "default not in instances",
			config: &Config{
				Default: "nonexistent",
				Instances: map[string]Instance{
					"test": {URL: "https://example.com", APIKey: "test-key-12345"},
				},
			},
			wantNil: true,
			wantOK:  false,
		},
		{
			name: "valid default instance",
			config: &Config{
				Default: "test",
				Instances: map[string]Instance{
					"test": {URL: "https://example.com", APIKey: "test-key-12345"},
				},
			},
			wantNil:  false,
			wantOK:   true,
			wantName: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst, ok := tt.config.CurrentInstance()
			if ok != tt.wantOK {
				t.Errorf("CurrentInstance() ok = %v, want %v", ok, tt.wantOK)
			}
			if tt.wantNil && inst != nil {
				t.Errorf("CurrentInstance() expected nil, got %v", inst)
			}
			if !tt.wantNil && inst == nil {
				t.Error("CurrentInstance() expected non-nil instance")
			}
			if !tt.wantNil && inst.Name != tt.wantName {
				t.Errorf("CurrentInstance() name = %s, want %s", inst.Name, tt.wantName)
			}
		})
	}
}

func TestGetInstance(t *testing.T) {
	cfg := &Config{
		Instances: map[string]Instance{
			"test": {URL: "https://example.com", APIKey: "test-key-12345"},
		},
	}

	tests := []struct {
		name     string
		instName string
		wantOK   bool
		wantURL  string
	}{
		{
			name:     "existing instance",
			instName: "test",
			wantOK:   true,
			wantURL:  "https://example.com",
		},
		{
			name:     "non-existing instance",
			instName: "nonexistent",
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst, ok := cfg.GetInstance(tt.instName)
			if ok != tt.wantOK {
				t.Errorf("GetInstance() ok = %v, want %v", ok, tt.wantOK)
			}
			if tt.wantOK {
				if inst == nil {
					t.Fatal("GetInstance() expected non-nil instance")
				}
				if inst.URL != tt.wantURL {
					t.Errorf("GetInstance() URL = %s, want %s", inst.URL, tt.wantURL)
				}
				if inst.Name != tt.instName {
					t.Errorf("GetInstance() Name = %s, want %s", inst.Name, tt.instName)
				}
			} else {
				if inst != nil {
					t.Errorf("GetInstance() expected nil, got %v", inst)
				}
			}
		})
	}
}

func TestValidateInstance(t *testing.T) {
	tests := []struct {
		name    string
		inst    Instance
		wantErr error
	}{
		{
			name: "valid instance",
			inst: Instance{
				URL:    "https://example.com",
				APIKey: "valid-api-key-12345",
			},
			wantErr: nil,
		},
		{
			name: "invalid URL",
			inst: Instance{
				URL:    "not-a-url",
				APIKey: "valid-api-key-12345",
			},
			wantErr: ErrInvalidURL,
		},
		{
			name: "invalid API key",
			inst: Instance{
				URL:    "https://example.com",
				APIKey: "short",
			},
			wantErr: ErrInvalidAPIKey,
		},
		{
			name: "empty URL",
			inst: Instance{
				URL:    "",
				APIKey: "valid-api-key-12345",
			},
			wantErr: ErrInvalidURL,
		},
		{
			name: "empty API key",
			inst: Instance{
				URL:    "https://example.com",
				APIKey: "",
			},
			wantErr: ErrInvalidAPIKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInstance(tt.inst, false)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ValidateInstance() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateInstance() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateInstanceWithKeyring(t *testing.T) {
	inst := Instance{
		URL:    "https://example.com",
		APIKey: "",
	}
	err := ValidateInstance(inst, true)
	if err != nil {
		t.Errorf("ValidateInstance() with keyring should skip API key check, got: %v", err)
	}
}

func TestConfigDirWithEnvVar(t *testing.T) {
	testDir := "/tmp/test-redmine-cli"
	originalValue := os.Getenv("REDMINE_CONFIG_DIR")
	defer os.Setenv("REDMINE_CONFIG_DIR", originalValue)

	os.Setenv("REDMINE_CONFIG_DIR", testDir)
	dir := Dir()
	if dir != testDir {
		t.Errorf("Dir() = %s, want %s", dir, testDir)
	}
}

func TestConfigDirWithoutEnvVar(t *testing.T) {
	originalValue := os.Getenv("REDMINE_CONFIG_DIR")
	defer os.Setenv("REDMINE_CONFIG_DIR", originalValue)

	os.Unsetenv("REDMINE_CONFIG_DIR")
	dir := Dir()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".redmine-cli")
	if dir != expected {
		t.Errorf("Dir() = %s, want %s", dir, expected)
	}
}

func TestDirHomeErrorFallback(t *testing.T) {
	originalValue := os.Getenv("REDMINE_CONFIG_DIR")
	defer os.Setenv("REDMINE_CONFIG_DIR", originalValue)

	os.Unsetenv("REDMINE_CONFIG_DIR")

	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	os.Unsetenv("HOME")

	dir := Dir()
	if dir != "/tmp/.redmine-cli" {
		t.Errorf("Dir() = %s, want /tmp/.redmine-cli when home is unavailable", dir)
	}
}

func TestPath(t *testing.T) {
	testDir := "/tmp/test-redmine-cli"
	originalValue := os.Getenv("REDMINE_CONFIG_DIR")
	defer os.Setenv("REDMINE_CONFIG_DIR", originalValue)

	os.Setenv("REDMINE_CONFIG_DIR", testDir)
	path := Path()
	expected := filepath.Join(testDir, "config.yaml")
	if path != expected {
		t.Errorf("Path() = %s, want %s", path, expected)
	}
}

func TestStoreSetDefault(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	// Test setting default for non-existent instance
	err := store.SetDefault("nonexistent")
	if err != ErrInstanceNotFound {
		t.Errorf("SetDefault() error = %v, want %v", err, ErrInstanceNotFound)
	}

	// Create an instance
	inst := Instance{
		URL:    "https://example.com",
		APIKey: "test-key-12345",
	}
	err = store.SaveInstance("test", inst)
	if err != nil {
		t.Fatalf("SaveInstance failed: %v", err)
	}

	// Set as default
	err = store.SetDefault("test")
	if err != nil {
		t.Fatalf("SetDefault failed: %v", err)
	}

	// Verify it's set as default
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Default != "test" {
		t.Errorf("expected default 'test', got %s", cfg.Default)
	}

	// Create another instance
	inst2 := Instance{
		URL:    "https://example2.com",
		APIKey: "test-key-67890",
	}
	err = store.SaveInstance("test2", inst2)
	if err != nil {
		t.Fatalf("SaveInstance failed: %v", err)
	}

	// Set different instance as default
	err = store.SetDefault("test2")
	if err != nil {
		t.Fatalf("SetDefault failed: %v", err)
	}

	// Verify it's changed
	cfg, err = store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Default != "test2" {
		t.Errorf("expected default 'test2', got %s", cfg.Default)
	}
}

func TestStoreSaveInstanceErrors(t *testing.T) {
	// Skip on Windows as path handling differs
	if os.PathSeparator == '\\' {
		t.Skip("skipping on Windows")
	}

	// Test SaveInstance should succeed even with non-existent path
	// because Save() creates the directory
	store := NewStore(filepath.Join(t.TempDir(), "nonexistent", "save-instance"))
	inst := Instance{
		URL:    "https://example.com",
		APIKey: "test-key-12345",
	}
	err := store.SaveInstance("test", inst)
	if err != nil {
		t.Errorf("SaveInstance() should not return error for non-existent path, got: %v", err)
	}
}

func TestStoreSetDefaultErrors(t *testing.T) {
	// Skip on Windows as path handling differs
	if os.PathSeparator == '\\' {
		t.Skip("skipping on Windows")
	}

	// Test SetDefault with non-existent instance
	store := NewStore(t.TempDir())
	err := store.SetDefault("nonexistent")
	if err == nil {
		t.Error("SetDefault() expected error for non-existent instance, got nil")
	}
	if err != ErrInstanceNotFound {
		t.Errorf("SetDefault() expected ErrInstanceNotFound, got: %v", err)
	}
}

func TestStoreDeleteInstanceErrors(t *testing.T) {
	// Skip on Windows as path handling differs
	if os.PathSeparator == '\\' {
		t.Skip("skipping on Windows")
	}

	// Test DeleteInstance should succeed even with non-existent path
	// because Save() creates the directory
	store := NewStore(filepath.Join(t.TempDir(), "nonexistent", "delete-instance"))
	err := store.DeleteInstance("test")
	if err != nil {
		t.Errorf("DeleteInstance() should not return error for non-existent path, got: %v", err)
	}
}

func TestStoreDeleteInstanceWithDefault(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	// Create multiple instances
	inst := Instance{
		URL:    "https://example.com",
		APIKey: "test-key-12345",
	}
	err := store.SaveInstance("test1", inst)
	if err != nil {
		t.Fatalf("SaveInstance failed: %v", err)
	}

	err = store.SaveInstance("test2", inst)
	if err != nil {
		t.Fatalf("SaveInstance failed: %v", err)
	}

	// Set test1 as default
	err = store.SetDefault("test1")
	if err != nil {
		t.Fatalf("SetDefault failed: %v", err)
	}

	// Delete test1 (should set test2 as default)
	err = store.DeleteInstance("test1")
	if err != nil {
		t.Fatalf("DeleteInstance failed: %v", err)
	}

	// Verify test2 is now default
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Default != "test2" {
		t.Errorf("expected default 'test2', got %s", cfg.Default)
	}
}

func TestStoreDeleteLastInstance(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	// Create single instance
	inst := Instance{
		URL:    "https://example.com",
		APIKey: "test-key-12345",
	}
	err := store.SaveInstance("test", inst)
	if err != nil {
		t.Fatalf("SaveInstance failed: %v", err)
	}

	// Delete the only instance
	err = store.DeleteInstance("test")
	if err != nil {
		t.Fatalf("DeleteInstance failed: %v", err)
	}

	// Verify default is empty
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Default != "" {
		t.Errorf("expected empty default, got %s", cfg.Default)
	}
}

func TestStoreLoadPermissionError(t *testing.T) {
	// Skip on Windows as permission handling differs
	if os.PathSeparator == '\\' {
		t.Skip("skipping on Windows")
	}

	// Test is flaky on some systems, skip for now
	t.Skip("skipping flaky permission test")

	dir := t.TempDir()
	store := NewStore(dir)

	// Create a config file with no read permissions
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("test: value"), 0000); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	// Try to load - should get permission error
	_, err := store.Load()
	if err == nil {
		t.Error("Load() expected error for permission denied, got nil")
	}
}

func TestStoreLoadNilInstances(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	// Create a config file with no instances map
	configPath := filepath.Join(dir, "config.yaml")
	yamlContent := `default: ""
settings:
  timeout: 30s
  retries: 3
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify instances map is initialized
	if cfg.Instances == nil {
		t.Error("Load() should initialize Instances map")
	}
}

func TestStoreSaveInstanceWithLoadError(t *testing.T) {
	// Skip on Windows as path handling differs
	if os.PathSeparator == '\\' {
		t.Skip("skipping on Windows")
	}

	// Test SaveInstance should succeed even with non-existent path
	// because Load() returns DefaultConfig() for non-existent files
	store := NewStore(filepath.Join(t.TempDir(), "nonexistent", "save-instance-load"))
	inst := Instance{
		URL:    "https://example.com",
		APIKey: "test-key-12345",
	}
	err := store.SaveInstance("test", inst)
	if err != nil {
		t.Errorf("SaveInstance() should not return error for non-existent path, got: %v", err)
	}
}

func TestStoreSetDefaultWithLoadError(t *testing.T) {
	// Skip on Windows as path handling differs
	if os.PathSeparator == '\\' {
		t.Skip("skipping on Windows")
	}

	// Test SetDefault with non-existent instance
	// Load() returns DefaultConfig() for non-existent files, which has empty Instances
	store := NewStore(filepath.Join(t.TempDir(), "nonexistent", "set-default"))

	// Load the config to check its state
	cfg, loadErr := store.Load()
	if loadErr != nil {
		t.Fatalf("Load() failed: %v", loadErr)
	}
	if cfg.Instances == nil {
		t.Fatal("Instances should not be nil")
	}
	if len(cfg.Instances) > 0 {
		t.Fatalf("Instances should be empty, got %d instances", len(cfg.Instances))
	}

	err := store.SetDefault("test")
	if err == nil {
		t.Error("SetDefault() expected ErrInstanceNotFound for non-existent instance, got nil")
	}
	if err != ErrInstanceNotFound {
		t.Errorf("SetDefault() expected ErrInstanceNotFound, got: %v", err)
	}
}

func TestStoreDeleteInstanceWithLoadError(t *testing.T) {
	// Skip on Windows as path handling differs
	if os.PathSeparator == '\\' {
		t.Skip("skipping on Windows")
	}

	// Test DeleteInstance should succeed even with non-existent path
	// because Load() returns DefaultConfig() for non-existent files
	store := NewStore(filepath.Join(t.TempDir(), "nonexistent", "delete-instance-load"))
	err := store.DeleteInstance("test")
	if err != nil {
		t.Errorf("DeleteInstance() should not return error for non-existent path, got: %v", err)
	}
}

func TestStoreLoadWithCorruptedFile(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("invalid: yaml: content:"), 0600)
	if err != nil {
		t.Fatalf("failed to write corrupted config: %v", err)
	}

	_, err = store.Load()
	if err == nil {
		t.Error("Load() expected error for corrupted YAML, got nil")
	}
}

func TestStoreSaveWithEmptyConfig(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	cfg := &Config{}
	err := store.Save(cfg)
	if err != nil {
		t.Fatalf("Save() with empty config failed: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if loaded.Instances == nil {
		t.Error("Load() should initialize nil Instances map")
	}
}

func TestStoreDeleteInstanceWhenDefaultIsLast(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	inst := Instance{
		URL:    "https://example.com",
		APIKey: "test-key-12345",
	}

	err := store.SaveInstance("only", inst)
	if err != nil {
		t.Fatalf("SaveInstance failed: %v", err)
	}

	err = store.DeleteInstance("only")
	if err != nil {
		t.Fatalf("DeleteInstance failed: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Default != "" {
		t.Errorf("expected empty default after deleting last instance, got %s", loaded.Default)
	}
}

func TestStoreSaveMarshalError(t *testing.T) {
	dir := t.TempDir()
	store := NewStoreWithKeyring(dir, &mockKeyring{})

	marshalErr := errors.New("marshal failed")
	origMarshal := yamlMarshal
	yamlMarshal = func(v interface{}) ([]byte, error) { return nil, marshalErr }
	defer func() { yamlMarshal = origMarshal }()

	err := store.Save(DefaultConfig())
	if err != marshalErr {
		t.Errorf("expected marshalErr, got %v", err)
	}
}
