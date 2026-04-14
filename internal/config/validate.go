// Package config provides configuration management for the CLI.
package config

import (
	"net/url"
	"strings"

	apperrors "github.com/largeoliu/redmine-cli/internal/errors"
)

var (
	// ErrInvalidURL is returned when the URL is invalid.
	ErrInvalidURL = apperrors.NewValidation("invalid URL",
		apperrors.WithHint("URL must start with http:// or https://"))
	// ErrInvalidAPIKey is returned when the API key is invalid.
	ErrInvalidAPIKey = apperrors.NewValidation("invalid API key",
		apperrors.WithHint("API key must be at least 10 characters"))
	// ErrInstanceNotFound is returned when the instance is not found.
	ErrInstanceNotFound = apperrors.NewValidation("instance not found",
		apperrors.WithHint("check available instances with 'redmine config list'"))
	// ErrAPIKeyNotFound is returned when the API key is not found in keyring.
	ErrAPIKeyNotFound = apperrors.NewValidation("API key not found in keyring",
		apperrors.WithHint("run 'redmine login' to store your API key"))
)

// ValidateURL validates the given URL.
func ValidateURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ErrInvalidURL
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ErrInvalidURL
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrInvalidURL
	}
	if u.Host == "" {
		return ErrInvalidURL
	}
	return nil
}

// ValidateAPIKey validates the given API key.
func ValidateAPIKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrInvalidAPIKey
	}
	if len(key) < 10 {
		return ErrInvalidAPIKey
	}
	return nil
}

// ValidateInstance validates the given instance configuration.
// If the API key is stored in a keyring (apiKeyFromKeyring=true), the API key
// check is skipped since the key is not in the Instance struct.
func ValidateInstance(inst Instance, apiKeyFromKeyring bool) error {
	if err := ValidateURL(inst.URL); err != nil {
		return err
	}
	if !apiKeyFromKeyring {
		return ValidateAPIKey(inst.APIKey)
	}
	return nil
}
