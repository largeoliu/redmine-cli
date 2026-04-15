// internal/config/keyring_test.go
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

func TestFallbackKeyring(t *testing.T) {
	kr := &fallbackKeyring{
		keys:      make(map[string]string),
		available: false,
	}

	if kr.IsAvailable() {
		t.Error("fallback keyring should not be available")
	}

	instanceName := "test-instance"
	apiKey := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"

	err := kr.Set(instanceName, apiKey)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, err := kr.Get(instanceName)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got != apiKey {
		t.Errorf("expected API key %s, got %s", apiKey, got)
	}

	_, err = kr.Get("non-existent")
	if err != ErrAPIKeyNotFound {
		t.Errorf("expected ErrAPIKeyNotFound, got %v", err)
	}

	err = kr.Delete(instanceName)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = kr.Get(instanceName)
	if err != ErrAPIKeyNotFound {
		t.Errorf("expected ErrAPIKeyNotFound after delete, got %v", err)
	}

	err = kr.Delete("non-existent")
	if err != ErrAPIKeyNotFound {
		t.Errorf("expected ErrAPIKeyNotFound when deleting non-existent key, got %v", err)
	}
}

func TestStoreWithKeyring(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	instanceName := "test-instance"
	apiKey := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
	inst := Instance{
		URL:    "https://example.com",
		APIKey: apiKey,
	}

	err := store.SaveInstance(instanceName, inst)
	if err != nil {
		t.Fatalf("SaveInstance failed: %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	loadedInst, ok := cfg.Instances[instanceName]
	if !ok {
		t.Fatalf("instance %s not found", instanceName)
	}

	if loadedInst.APIKey != apiKey {
		t.Errorf("expected API key %s, got %s", apiKey, loadedInst.APIKey)
	}

	configPath := filepath.Join(dir, "config.yaml")
	//nolint:gosec // Test code reading from test directory
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config file failed: %v", err)
	}

	var fileCfg Config
	if err := yaml.Unmarshal(configData, &fileCfg); err != nil {
		t.Fatalf("unmarshal config failed: %v", err)
	}

	if fileCfg.Instances[instanceName].APIKey != apiKey {
		t.Errorf("API key should be saved in config file, got %s", fileCfg.Instances[instanceName].APIKey)
	}
}

func TestStoreDeleteInstance(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	instanceName := "test-instance"
	apiKey := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
	inst := Instance{
		URL:    "https://example.com",
		APIKey: apiKey,
	}

	err := store.SaveInstance(instanceName, inst)
	if err != nil {
		t.Fatalf("SaveInstance failed: %v", err)
	}

	err = store.DeleteInstance(instanceName)
	if err != nil {
		t.Fatalf("DeleteInstance failed: %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if _, ok := cfg.Instances[instanceName]; ok {
		t.Error("instance should be deleted")
	}

	kr := NewKeyring()
	_, err = kr.Get(instanceName)
	if err != ErrAPIKeyNotFound {
		t.Errorf("expected ErrAPIKeyNotFound, got %v", err)
	}
}

func TestStoreDeleteInstanceIgnoresMissingKeyringSecret(t *testing.T) {
	dir := t.TempDir()
	store := NewStoreWithKeyring(dir, &mockKeyring{
		deleteFunc: func(instanceName string) error {
			if instanceName != "test-instance" {
				t.Fatalf("expected instanceName test-instance, got %s", instanceName)
			}
			return ErrAPIKeyNotFound
		},
	})

	err := store.Save(&Config{
		Default: "test-instance",
		Instances: map[string]Instance{
			"test-instance": {URL: "https://example.com"},
		},
	})
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	err = store.DeleteInstance("test-instance")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if _, ok := cfg.Instances["test-instance"]; ok {
		t.Fatal("instance should be deleted even when keyring secret is missing")
	}
}

func TestFormatKeyringKey(t *testing.T) {
	tests := []struct {
		instanceName string
		expected     string
	}{
		{"default", "instance-default-api-key"},
		{"production", "instance-production-api-key"},
		{"test-123", "instance-test-123-api-key"},
	}

	for _, tt := range tests {
		got := formatKeyringKey(tt.instanceName)
		if got != tt.expected {
			t.Errorf("formatKeyringKey(%s) = %s, want %s", tt.instanceName, got, tt.expected)
		}
	}
}

func TestNewStore(t *testing.T) {
	// Test with custom config dir
	customDir := "/tmp/custom-config"
	store := NewStore(customDir)
	if store.configDir != customDir {
		t.Errorf("NewStore() configDir = %s, want %s", store.configDir, customDir)
	}

	// Test with empty config dir (should use default)
	store = NewStore("")
	expectedDir := Dir()
	if store.configDir != expectedDir {
		t.Errorf("NewStore() configDir = %s, want %s", store.configDir, expectedDir)
	}
}

func TestStoreLoadErrors(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	// Test loading non-existent config (should return default)
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	//nolint:staticcheck // SA5011: cfg is guaranteed non-nil after successful Load
	if cfg == nil {
		t.Error("Load() returned nil config")
	}
	//nolint:staticcheck // SA5011: cfg.Instances is initialized by Load()
	if cfg.Instances == nil {
		t.Error("Load() returned config with nil instances")
	}

	// Test loading invalid YAML
	configPath := filepath.Join(dir, "config.yaml")
	invalidYAML := []byte("invalid: yaml: content: [")
	if err := os.WriteFile(configPath, invalidYAML, 0600); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	_, err = store.Load()
	if err == nil {
		t.Error("Load() expected error for invalid YAML, got nil")
	}
}

func TestStoreSaveErrors(t *testing.T) {
	// Skip on Windows as path handling differs
	if os.PathSeparator == '\\' {
		t.Skip("skipping on Windows")
	}

	// Test saving to path should succeed
	// because Save() creates the directory
	store := NewStore(filepath.Join(t.TempDir(), "nonexistent", "path", "that", "can", "be", "created"))
	cfg := DefaultConfig()

	err := store.Save(cfg)
	if err != nil {
		t.Errorf("Save() should not return error for non-existent path, got: %v", err)
	}
}

func TestNewKeyring(t *testing.T) {
	// NewKeyring should return a keyring instance
	kr := NewKeyring()
	if kr == nil {
		t.Fatal("NewKeyring() returned nil")
	}

	// Test that it implements the Keyring interface
	_ = Keyring(kr)
}

func TestFallbackKeyringAvailable(t *testing.T) {
	kr := &fallbackKeyring{
		keys:      make(map[string]string),
		available: true,
	}

	if !kr.IsAvailable() {
		t.Error("fallback keyring should be available when set to true")
	}
}

// mockKeyring 用于测试 Keyring 接口
type mockKeyring struct {
	getFunc       func(instanceName string) (string, error)
	setFunc       func(instanceName, apiKey string) error
	deleteFunc    func(instanceName string) error
	isAvailableFn func() bool
}

func (m *mockKeyring) Get(instanceName string) (string, error) {
	if m.getFunc != nil {
		return m.getFunc(instanceName)
	}
	return "", nil
}

func (m *mockKeyring) Set(instanceName, apiKey string) error {
	if m.setFunc != nil {
		return m.setFunc(instanceName, apiKey)
	}
	return nil
}

func (m *mockKeyring) Delete(instanceName string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(instanceName)
	}
	return nil
}

func (m *mockKeyring) IsAvailable() bool {
	if m.isAvailableFn != nil {
		return m.isAvailableFn()
	}
	return true
}

//nolint:revive // Interface compliance tests - t unused
func TestMockKeyring(t *testing.T) {
	// Test that mockKeyring implements Keyring interface
	var _ Keyring = &mockKeyring{}
}

//nolint:revive // Interface compliance tests - t unused
func TestKeyringInterface(t *testing.T) {
	// Test that realKeyring implements Keyring interface
	var _ Keyring = &realKeyring{}

	// Test that fallbackKeyring implements Keyring interface
	var _ Keyring = &fallbackKeyring{}
}

//nolint:revive // Interface compliance tests - t unused
func TestRealKeyringMethods(t *testing.T) {
	// Test realKeyring methods when keyring is available
	// Note: These tests may not work in all environments
	kr := &realKeyring{}

	// Test IsAvailable - this will test the full flow
	// In CI environments without a keyring, this may return false
	_ = kr.IsAvailable()
}

//nolint:revive // Interface compliance tests - t unused
func TestFallbackKeyringConcurrent(t *testing.T) {
	kr := &fallbackKeyring{
		keys:      make(map[string]string),
		available: true,
	}

	// Test concurrent access
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			//nolint:gosec,errcheck // Test code - keyring operations in test goroutine
			kr.Set("instance", "api-key-value-1234567890")
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			//nolint:gosec,errcheck // Test code - keyring operations in test goroutine
			kr.Get("instance")
		}
		done <- true
	}()

	// Deleter goroutine
	go func() {
		for i := 0; i < 100; i++ {
			//nolint:gosec,errcheck // Test code - keyring operations in test goroutine
			kr.Delete("instance")
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done
}

func TestRealKeyringGet(t *testing.T) {
	kr := &realKeyring{}

	// Test Get with non-existent key
	_, err := kr.Get("non-existent-instance")
	// This should return ErrAPIKeyNotFound or an error from the keyring
	if err == nil {
		t.Error("expected error when getting non-existent key")
	}
}

//nolint:revive // Interface compliance tests - t unused
func TestRealKeyringSet(t *testing.T) {
	kr := &realKeyring{}

	// Test Set - this may fail if keyring is not available
	// We just verify the function can be called
	instanceName := "test-instance-for-set"
	apiKey := "test-api-key-1234567890"

	err := kr.Set(instanceName, apiKey)
	// We don't assert on the result because keyring may not be available
	// Just verify the function doesn't panic
	_ = err
}

func TestRealKeyringDelete(t *testing.T) {
	kr := &realKeyring{}

	// Test Delete with non-existent key
	err := kr.Delete("non-existent-instance")
	// This should return an error
	if err == nil {
		t.Error("expected error when deleting non-existent key")
	}
}

//nolint:revive // Interface compliance tests - t unused
func TestRealKeyringIsAvailable(t *testing.T) {
	kr := &realKeyring{}

	// Test IsAvailable - this will test the full flow
	// In CI environments without a keyring, this may return false
	available := kr.IsAvailable()
	// We just verify the function can be called and returns a boolean
	_ = available
}

func TestNewKeyringReturnsRealOrFallback(t *testing.T) {
	kr := NewKeyring()
	if kr == nil {
		t.Fatal("NewKeyring() returned nil")
	}

	// The keyring should either be realKeyring or fallbackKeyring
	// We can verify it implements the interface
	_ = Keyring(kr)
}

func TestFallbackKeyringGetNonExistent(t *testing.T) {
	kr := &fallbackKeyring{
		keys:      make(map[string]string),
		available: false,
	}

	_, err := kr.Get("non-existent")
	if err != ErrAPIKeyNotFound {
		t.Errorf("expected ErrAPIKeyNotFound, got %v", err)
	}
}

func TestFallbackKeyringDeleteNonExistent(t *testing.T) {
	kr := &fallbackKeyring{
		keys:      make(map[string]string),
		available: false,
	}

	err := kr.Delete("non-existent")
	if err != ErrAPIKeyNotFound {
		t.Errorf("expected ErrAPIKeyNotFound, got %v", err)
	}
}

func TestFallbackKeyringSetAndGet(t *testing.T) {
	kr := &fallbackKeyring{
		keys:      make(map[string]string),
		available: false,
	}

	instanceName := "test-instance"
	apiKey := "test-api-key-1234567890"

	// Set
	err := kr.Set(instanceName, apiKey)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get
	got, err := kr.Get(instanceName)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got != apiKey {
		t.Errorf("expected %s, got %s", apiKey, got)
	}
}

func TestFallbackKeyringDelete(t *testing.T) {
	kr := &fallbackKeyring{
		keys:      make(map[string]string),
		available: false,
	}

	instanceName := "test-instance"
	apiKey := "test-api-key-1234567890"

	// Set
	err := kr.Set(instanceName, apiKey)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Delete
	err = kr.Delete(instanceName)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	_, err = kr.Get(instanceName)
	if err != ErrAPIKeyNotFound {
		t.Errorf("expected ErrAPIKeyNotFound after delete, got %v", err)
	}
}

func TestNewStoreWithKeyring(t *testing.T) {
	dir := t.TempDir()
	kr := &mockKeyring{
		isAvailableFn: func() bool { return true },
	}
	store := NewStoreWithKeyring(dir, kr)
	if store == nil {
		t.Fatal("NewStoreWithKeyring() returned nil")
	}
	if store.configDir != dir {
		t.Errorf("configDir = %s, want %s", store.configDir, dir)
	}
	if store.keyring != kr {
		t.Error("keyring not set correctly")
	}

	emptyStore := NewStoreWithKeyring("", kr)
	if emptyStore.configDir != Dir() {
		t.Errorf("empty dir: configDir = %s, want %s", emptyStore.configDir, Dir())
	}
}

func TestStoreSaveInstanceWithKeyringAvailable(t *testing.T) {
	dir := t.TempDir()
	var capturedName, capturedKey string
	kr := &mockKeyring{
		isAvailableFn: func() bool { return true },
		setFunc: func(instanceName, apiKey string) error {
			capturedName = instanceName
			capturedKey = apiKey
			return nil
		},
	}
	store := NewStoreWithKeyring(dir, kr)

	err := store.SaveInstance("test-instance", Instance{
		URL:    "https://example.com",
		APIKey: "secret-key-123",
	})
	if err != nil {
		t.Fatalf("SaveInstance failed: %v", err)
	}

	if capturedName != "test-instance" {
		t.Errorf("keyring.Set name = %s, want test-instance", capturedName)
	}
	if capturedKey != "secret-key-123" {
		t.Errorf("keyring.Set key = %s, want secret-key-123", capturedKey)
	}

	configPath := filepath.Join(dir, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config file failed: %v", err)
	}
	var fileCfg Config
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if fileCfg.Instances["test-instance"].APIKey != "" {
		t.Errorf("API key should be cleared in YAML, got %s", fileCfg.Instances["test-instance"].APIKey)
	}
}

func TestStoreSaveInstanceKeyringSetError(t *testing.T) {
	dir := t.TempDir()
	setErr := errors.New("keyring set failed")
	kr := &mockKeyring{
		isAvailableFn: func() bool { return true },
		setFunc: func(instanceName, apiKey string) error {
			return setErr
		},
	}
	store := NewStoreWithKeyring(dir, kr)

	err := store.SaveInstance("test-instance", Instance{
		URL:    "https://example.com",
		APIKey: "secret-key-123",
	})
	if err != setErr {
		t.Errorf("expected setErr, got %v", err)
	}
}

func TestStoreDeleteInstanceKeyringError(t *testing.T) {
	dir := t.TempDir()
	delErr := errors.New("keyring delete failed")
	kr := &mockKeyring{
		isAvailableFn: func() bool { return true },
		deleteFunc: func(instanceName string) error {
			return delErr
		},
	}
	store := NewStoreWithKeyring(dir, kr)

	err := store.Save(&Config{
		Default: "test-instance",
		Instances: map[string]Instance{
			"test-instance": {URL: "https://example.com"},
		},
	})
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	err = store.DeleteInstance("test-instance")
	if err != delErr {
		t.Errorf("expected delErr, got %v", err)
	}
}

func TestStoreSaveMkdirAllError(t *testing.T) {
	dir := t.TempDir()
	fileAsDir := filepath.Join(dir, "blocked")
	if err := os.WriteFile(fileAsDir, []byte("x"), 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	store := NewStoreWithKeyring(filepath.Join(fileAsDir, "sub", "dir"), &mockKeyring{})
	cfg := DefaultConfig()
	err := store.Save(cfg)
	if err == nil {
		t.Error("expected error when MkdirAll fails")
	}
}

func TestStoreLoadReadFileError(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	store := NewStoreWithKeyring(dir, &mockKeyring{})
	_, err := store.Load()
	if err == nil {
		t.Error("expected error when config.yaml is a directory")
	}
}

func TestStoreSaveInstanceLoadError(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	store := NewStoreWithKeyring(dir, &mockKeyring{})
	err := store.SaveInstance("test", Instance{URL: "https://example.com", APIKey: "key"})
	if err == nil {
		t.Error("expected error when Load fails in SaveInstance")
	}
}

func TestStoreSetDefaultLoadError(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	store := NewStoreWithKeyring(dir, &mockKeyring{})
	err := store.SetDefault("test")
	if err == nil {
		t.Error("expected error when Load fails in SetDefault")
	}
}

func TestStoreDeleteInstanceLoadError(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	store := NewStoreWithKeyring(dir, &mockKeyring{})
	err := store.DeleteInstance("test")
	if err == nil {
		t.Error("expected error when Load fails in DeleteInstance")
	}
}

func TestStoreLoadWithKeyringAvailable(t *testing.T) {
	dir := t.TempDir()
	kr := &mockKeyring{
		isAvailableFn: func() bool { return true },
		getFunc: func(instanceName string) (string, error) {
			if instanceName == "test-instance" {
				return "resolved-from-keyring", nil
			}
			return "", ErrAPIKeyNotFound
		},
	}
	store := NewStoreWithKeyring(dir, kr)

	cfg := &Config{
		Default: "test-instance",
		Instances: map[string]Instance{
			"test-instance": {URL: "https://example.com", APIKey: ""},
		},
	}
	err := store.Save(cfg)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	inst, ok := loaded.Instances["test-instance"]
	if !ok {
		t.Fatal("test-instance not found")
	}
	if inst.APIKey != "resolved-from-keyring" {
		t.Errorf("API key = %s, want resolved-from-keyring", inst.APIKey)
	}
}

func TestRealKeyringIsAvailableWithMock(t *testing.T) {
	keyring.MockInit()
	defer keyring.MockInit()

	kr := &realKeyring{}
	if !kr.IsAvailable() {
		t.Error("realKeyring.IsAvailable() should return true with mock provider")
	}
}

func TestRealKeyringGetSetDeleteWithMock(t *testing.T) {
	keyring.MockInit()
	defer keyring.MockInit()

	kr := &realKeyring{}

	err := kr.Set("mock-instance", "mock-api-key")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, err := kr.Get("mock-instance")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != "mock-api-key" {
		t.Errorf("Get = %s, want mock-api-key", got)
	}

	err = kr.Delete("mock-instance")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = kr.Get("mock-instance")
	if err != ErrAPIKeyNotFound {
		t.Errorf("expected ErrAPIKeyNotFound after delete, got %v", err)
	}
}

func TestRealKeyringGetNonNotFoundError(t *testing.T) {
	mockErr := errors.New("keyring internal error")
	keyring.MockInitWithError(mockErr)
	defer keyring.MockInit()

	kr := &realKeyring{}
	_, err := kr.Get("any-instance")
	if err == nil {
		t.Error("expected error, got nil")
	}

	expected := fmt.Sprintf("failed to get API key from keyring: %v", mockErr)
	if err.Error() != expected {
		t.Errorf("error = %q, want %q", err.Error(), expected)
	}
}

func TestRealKeyringSetError(t *testing.T) {
	mockErr := errors.New("keyring set error")
	keyring.MockInitWithError(mockErr)
	defer keyring.MockInit()

	kr := &realKeyring{}
	err := kr.Set("any-instance", "any-key")
	if err == nil {
		t.Error("expected error, got nil")
	}

	expected := fmt.Sprintf("failed to set API key to keyring: %v", mockErr)
	if err.Error() != expected {
		t.Errorf("error = %q, want %q", err.Error(), expected)
	}
}

func TestRealKeyringDeleteNonNotFoundError(t *testing.T) {
	mockErr := errors.New("keyring delete error")
	keyring.MockInitWithError(mockErr)
	defer keyring.MockInit()

	kr := &realKeyring{}
	err := kr.Delete("any-instance")
	if err == nil {
		t.Error("expected error, got nil")
	}

	expected := fmt.Sprintf("failed to delete API key from keyring: %v", mockErr)
	if err.Error() != expected {
		t.Errorf("error = %q, want %q", err.Error(), expected)
	}
}

func TestNewKeyringFallbackWithMockError(t *testing.T) {
	keyring.MockInitWithError(errors.New("unavailable"))
	defer keyring.MockInit()

	kr := NewKeyring()
	if kr == nil {
		t.Fatal("NewKeyring() returned nil")
	}

	_, ok := kr.(*fallbackKeyring)
	if !ok {
		t.Error("expected fallbackKeyring when keyring is unavailable")
	}
}

func TestNewKeyringRealWithMock(t *testing.T) {
	keyring.MockInit()
	defer keyring.MockInit()

	kr := NewKeyring()
	if kr == nil {
		t.Fatal("NewKeyring() returned nil")
	}

	_, ok := kr.(*realKeyring)
	if !ok {
		t.Error("expected realKeyring when keyring is available")
	}
}

func TestRealKeyringDeleteNonExistentWithMock(t *testing.T) {
	keyring.MockInit()
	defer keyring.MockInit()

	kr := &realKeyring{}
	err := kr.Delete("non-existent-instance")
	if err != ErrAPIKeyNotFound {
		t.Errorf("expected ErrAPIKeyNotFound for non-existent key, got %v", err)
	}
}
