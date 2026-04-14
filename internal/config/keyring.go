// Package config provides configuration management for the CLI.
package config

import (
	"errors"
	"fmt"
	"sync"

	"github.com/zalando/go-keyring"
)

// Keyring 定义密钥环接口
type Keyring interface {
	// Get 从密钥环获取 API Key
	Get(instanceName string) (string, error)
	// Set 将 API Key 存储到密钥环
	Set(instanceName, apiKey string) error
	// Delete 从密钥环删除 API Key
	Delete(instanceName string) error
	// IsAvailable 检查密钥环是否可用
	IsAvailable() bool
}

// keyringServiceName 用于标识密钥环服务
const keyringServiceName = "redmine-cli"

// realKeyring 使用系统密钥环实现
type realKeyring struct{}

// fallbackKeyring 降级实现，使用内存存储
type fallbackKeyring struct {
	mu        sync.RWMutex
	keys      map[string]string
	available bool
}

// NewKeyring 创建密钥环实例
// 如果系统密钥环不可用，则降级到内存存储
func NewKeyring() Keyring {
	// 尝试使用系统密钥环
	kr := &realKeyring{}
	if kr.IsAvailable() {
		return kr
	}

	// 降级到内存存储
	return &fallbackKeyring{
		keys:      make(map[string]string),
		available: false,
	}
}

// realKeyring 实现

func (k *realKeyring) Get(instanceName string) (string, error) {
	key := formatKeyringKey(instanceName)
	secret, err := keyring.Get(keyringServiceName, key)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", ErrAPIKeyNotFound
		}
		return "", fmt.Errorf("failed to get API key from keyring: %w", err)
	}
	return secret, nil
}

func (k *realKeyring) Set(instanceName, apiKey string) error {
	key := formatKeyringKey(instanceName)
	if err := keyring.Set(keyringServiceName, key, apiKey); err != nil {
		return fmt.Errorf("failed to set API key to keyring: %w", err)
	}
	return nil
}

func (k *realKeyring) Delete(instanceName string) error {
	key := formatKeyringKey(instanceName)
	if err := keyring.Delete(keyringServiceName, key); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return ErrAPIKeyNotFound
		}
		return fmt.Errorf("failed to delete API key from keyring: %w", err)
	}
	return nil
}

func (k *realKeyring) IsAvailable() bool {
	// 尝试执行一个简单的操作来检查密钥环是否可用
	testKey := "__test_keyring_availability__"
	testValue := "test"

	// 尝试写入
	if err := keyring.Set(keyringServiceName, testKey, testValue); err != nil {
		return false
	}

	// 尝试读取
	_, err := keyring.Get(keyringServiceName, testKey)
	if err != nil {
		return false
	}

	// 清理测试数据
	_ = keyring.Delete(keyringServiceName, testKey) //nolint:errcheck // 清理错误不影响可用性检查

	return true
}

// fallbackKeyring 实现

func (k *fallbackKeyring) Get(instanceName string) (string, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	key := formatKeyringKey(instanceName)
	apiKey, ok := k.keys[key]
	if !ok {
		return "", ErrAPIKeyNotFound
	}
	return apiKey, nil
}

func (k *fallbackKeyring) Set(instanceName, apiKey string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	key := formatKeyringKey(instanceName)
	k.keys[key] = apiKey
	return nil
}

func (k *fallbackKeyring) Delete(instanceName string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	key := formatKeyringKey(instanceName)
	if _, ok := k.keys[key]; !ok {
		return ErrAPIKeyNotFound
	}
	delete(k.keys, key)
	return nil
}

func (k *fallbackKeyring) IsAvailable() bool {
	return k.available
}

// formatKeyringKey 格式化密钥环键名
func formatKeyringKey(instanceName string) string {
	return fmt.Sprintf("instance-%s-api-key", instanceName)
}
