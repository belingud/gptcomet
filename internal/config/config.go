package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gptcometerrors "github.com/belingud/gptcomet/internal/errors"
	"github.com/belingud/gptcomet/pkg/config/defaults"
	"github.com/belingud/gptcomet/pkg/types"

	"gopkg.in/yaml.v3"
)

type ManagerInterface interface {
	Get(key string) (interface{}, bool)
	GetWithDefault(key string, defaultValue interface{}) interface{}
	Set(key string, value interface{}) error
	ListWithoutPrompt() map[string]interface{}
	List() (string, error)
	Reset(promptOnly bool) error
	Remove(key string, value string) error
	Append(key string, value interface{}) error
	Save() error
	GetClientConfig(initProvider string) (*types.ClientConfig, error)
	GetSupportedKeys() []string
	UpdateProviderConfig(provider string, configs map[string]string) error
	GetPrompt(isRich bool) string
	GetReviewPrompt() string
	GetNestedValue(keys []string) (interface{}, bool)
	SetNestedValue(keys []string, value interface{})
	Load() error
	GetTranslationPrompt() string
	GetOutputTranslateTitle() bool
	GetFileIgnore() []string
}

// Manager handles configuration management
type Manager struct {
	config     map[string]interface{}
	configPath string
}

// New creates a new configuration manager
func New(configPath string) (*Manager, error) {
	if configPath == "" {
		var err error
		configPath, err = getConfigDir()
		if err != nil {
			return nil, gptcometerrors.NewConfigError("Configuration Error", "Failed to get config directory", err, nil)
		}
		configPath = configPath + "/gptcomet.yaml"
	} else {
		// Check if the config file exists
		if _, err := os.Stat(configPath); err != nil {
			return nil, gptcometerrors.ConfigFileNotFoundError(configPath)
		}
	}

	manager := &Manager{
		config:     make(map[string]interface{}),
		configPath: configPath,
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return nil, gptcometerrors.NewConfigError("Configuration Error", "Failed to create config directory", err, nil)
	}

	// Load existing config if it exists
	if _, err := os.Stat(configPath); err == nil {
		if err := manager.Load(); err != nil {
			return nil, gptcometerrors.NewConfigError("Configuration Error", "Failed to load config", err, nil)
		}
	} else {
		// Initialize with default configuration
		manager.config = defaults.DefaultConfig
		if err := manager.Save(); err != nil {
			return nil, gptcometerrors.NewConfigError("Configuration Error", "Failed to save default config", err, nil)
		}
	}

	return manager, nil
}

// GetClientConfig returns the client configuration for the currently selected provider.
//
// It parses the following configuration options from the provider configuration:
// - api_key: the API key for the provider
// - api_base: the base URL for the provider's API (defaults to the default API base)
// - model: the model to use with the provider (defaults to the default model)
// - proxy: the proxy URL to use with the provider (defaults to an empty string)
// - max_tokens: the maximum number of tokens to generate (defaults to 2048)
// - top_p: the top-p sampling parameter (defaults to 1.0)
// - temperature: the temperature parameter (defaults to 1.0)
// - frequency_penalty: the frequency penalty parameter (defaults to 0.5)
// - retries: the number of times to retry the request if it fails (defaults to 3)
// - answer_path: the JSON path to the answer field in the response (defaults to an empty string)
// - completion_path: the JSON path to the completion field in the response (defaults to an empty string)
//
// If any of the required configuration options are not set, an error is returned.
func (m *Manager) GetClientConfig(initProvider string) (*types.ClientConfig, error) {
	var provider string
	if initProvider == "" {
		var _ok bool
		provider, _ok = m.config["provider"].(string)
		if !_ok {
			return nil, gptcometerrors.NewConfigError("Configuration Error", "Provider not set", nil, []string{"Set a provider using 'gptcomet config set provider <name>'"})
		}
	} else {
		provider = initProvider
	}
	providerConfig, ok := m.config[provider].(map[string]interface{})
	if !ok {
		return nil, gptcometerrors.NewConfigError("Configuration Error", fmt.Sprintf("Provider config not found: %s", provider), nil, []string{fmt.Sprintf("Configure the provider: gptcomet config set %s.api_key <key>", provider)})
	}
	var apiKey string
	if provider == "ollama" {
		apiKey = ""
	} else {
		apiKeyValue, ok := providerConfig["api_key"].(string)
		if !ok || apiKeyValue == "" {
			return nil, gptcometerrors.APIKeyNotSetError(provider)
		}
		apiKey = apiKeyValue
	}

	apiBase := defaults.DefaultAPIBase
	if base, ok := providerConfig["api_base"].(string); ok {
		apiBase = base
	}

	model := defaults.DefaultModel
	if m, ok := providerConfig["model"].(string); ok {
		model = m
	}

	proxy := ""
	if p, ok := providerConfig["proxy"].(string); ok {
		proxy = p
	}

	maxTokens := defaults.DefaultMaxTokens
	maxTokens = getIntValue(providerConfig, "max_tokens", maxTokens)

	topP := defaults.DefaultTopP
	topP = getFloatValue(providerConfig, "top_p", topP)

	temperature := defaults.DefaultTemperature
	temperature = getFloatValue(providerConfig, "temperature", temperature)

	frequencyPenalty := defaults.DefaultFrequencyPenalty
	frequencyPenalty = getFloatValue(providerConfig, "frequency_penalty", frequencyPenalty)

	clientConfig := &types.ClientConfig{
		APIBase:          apiBase,
		APIKey:           apiKey,
		Model:            model,
		Provider:         provider,
		Retries:          defaults.DefaultRetries,
		Proxy:            proxy,
		MaxTokens:        maxTokens,
		TopP:             topP,
		Temperature:      temperature,
		FrequencyPenalty: frequencyPenalty,
	}
	if m, ok := providerConfig["retries"].(float64); ok {
		clientConfig.Retries = int(m)
	}

	if answerPath, ok := providerConfig["answer_path"].(string); ok {
		clientConfig.AnswerPath = answerPath
	}

	if completionPath, ok := providerConfig["completion_path"].(string); ok {
		clientConfig.CompletionPath = &completionPath
	}

	// Parse extra_headers (additional request headers)
	if extraHeadersStr, ok := providerConfig["extra_headers"].(string); ok && extraHeadersStr != "" && extraHeadersStr != "{}" {
		extraHeaders := make(map[string]string)
		if err := json.Unmarshal([]byte(extraHeadersStr), &extraHeaders); err != nil {
			return nil, gptcometerrors.NewConfigError("Configuration Error", "Failed to parse extra_headers", err, nil)
		}
		clientConfig.ExtraHeaders = extraHeaders
	}

	// Parse extra_body (additional request body)
	if extraBodyStr, ok := providerConfig["extra_body"].(string); ok && extraBodyStr != "" && extraBodyStr != "{}" {
		extraBody := make(map[string]interface{})
		if err := json.Unmarshal([]byte(extraBodyStr), &extraBody); err != nil {
			return nil, gptcometerrors.NewConfigError("Configuration Error", "Failed to parse extra_body", err, nil)
		}
		clientConfig.ExtraBody = extraBody
	}

	return clientConfig, nil
}

// SetProvider sets the provider configuration.
//
// It takes the provider name, API key, API base, and model name as arguments.
// If the API base or model name is not specified, it defaults to the default
// values.
//
// The method saves the configuration to the file and returns an error if
// the save fails.
func (m *Manager) SetProvider(provider, apiKey, apiBase, model string) error {
	if apiBase == "" {
		apiBase = defaults.DefaultAPIBase
	}
	if model == "" {
		model = defaults.DefaultModel
	}

	m.config[provider] = map[string]interface{}{
		"api_key":  apiKey,
		"api_base": apiBase,
		"model":    model,
	}
	m.config["provider"] = provider

	return m.Save()
}

// Set sets the value associated with the given key. The key is split
// on the '.' character and the value is set in the nested map.
//
// If the key is not found, the value is not set.
//
// If the key is "output.lang" or "output.review_lang", the value must be a valid language code.
// If the key is "output.translate_title", the value must be a boolean.
//
// The method saves the configuration to the file and returns an error if
// the save fails.
func (m *Manager) Set(key string, value interface{}) error {
	switch key {
	case "output.lang", "output.review_lang":
		if str, ok := value.(string); ok {
			if !IsValidLanguage(str) {
				return gptcometerrors.InvalidConfigValueError(key, str, "Invalid language code")
			}
		}
	case "output.translate_title":
		if _, ok := value.(bool); !ok {
			return gptcometerrors.InvalidConfigValueError(key, fmt.Sprintf("%v", value), "translate_title must be a boolean value")
		}
	}

	keys := strings.Split(key, ".")
	m.SetNestedValue(keys, value)
	return m.Save()
}

// ListWithoutPrompt returns a copy of the configuration that excludes the prompt section.
// This is useful when the user wants to list all configuration options without the prompt.
func (m *Manager) ListWithoutPrompt() map[string]interface{} {
	// Create a copy of the config without the prompt section
	result := make(map[string]interface{})
	for k, v := range m.config {
		if k != "prompt" {
			result[k] = v
		}
	}
	return result
}

// Reset resets the configuration to default values. If promptOnly is true, only the prompt
// configuration is reset. Otherwise, all configuration is reset.
func (m *Manager) Reset(promptOnly bool) error {
	if promptOnly {
		// Get default prompt config
		if promptConfig, ok := defaults.DefaultConfig["prompt"]; ok {
			// Handle type conversion
			switch pc := promptConfig.(type) {
			case map[string]interface{}:
				m.config["prompt"] = pc
			case map[string]string:
				// Convert map[string]string to map[string]interface{}
				interfaceMap := make(map[string]interface{})
				for k, v := range pc {
					interfaceMap[k] = v
				}
				m.config["prompt"] = interfaceMap
			}
		}
	} else {
		// Reset all config
		m.config = defaults.DefaultConfig
	}
	return m.Save()
}

// Remove removes a configuration value or a value from a list.
//
// If the value parameter is empty, the entire key is removed.
// If the value parameter is not empty, the value is removed from the list associated with the key.
//
// Returns an error if the key is not found or if the value is not a list.
func (m *Manager) Remove(key string, value string) error {
	keys := strings.Split(key, ".")
	if value == "" {
		// If no value is provided, remove the entire key
		lastKey := keys[len(keys)-1]
		parent, ok := m.GetNestedValue(keys[:len(keys)-1])
		if !ok {
			return nil
		}

		if parentMap, ok := parent.(map[string]interface{}); ok {
			delete(parentMap, lastKey)
		}
		return m.Save()
	}

	// If value is provided, try to remove it from a list
	current, ok := m.GetNestedValue(keys)
	if !ok {
		return nil
	}

	// Check if the current value is a list
	list, ok := current.([]interface{})
	if !ok {
		return gptcometerrors.InvalidConfigValueError(key, fmt.Sprintf("%v", current), "Value is not a list")
	}

	// Find and remove the value from the list
	newList := make([]interface{}, 0, len(list))
	for _, item := range list {
		if str, ok := item.(string); ok && str != value {
			newList = append(newList, item)
		} else if !ok {
			newList = append(newList, item)
		}
	}

	return m.Set(key, newList)
}

// GetPath returns the path to the configuration file.
func (m *Manager) GetPath() string {
	return m.configPath
}

// Append appends the given value to a list configuration.
//
// If the key doesn't exist, it creates a new list with the given value.
// If the key exists but is not a list, it returns an error.
// If the key exists and is a list, it appends the given value to the list.
//
// The method saves the configuration to the file and returns an error if
// the save fails.
func (m *Manager) Append(key string, value interface{}) error {
	keys := strings.Split(key, ".")
	current, ok := m.GetNestedValue(keys)
	if !ok {
		// If the key doesn't exist, create a new list
		return m.Set(key, []interface{}{value})
	}

	// Check if the current value is a list
	list, ok := current.([]interface{})
	if !ok {
		return gptcometerrors.InvalidConfigValueError(key, fmt.Sprintf("%v", current), "Value is not a list")
	}

	// Append the new value
	list = append(list, value)
	return m.Set(key, list)
}

// Load reads the configuration from the file specified by the configPath
// field and unmarshals it into the config field.
//
// If the file does not exist or an error occurs while reading or parsing the
// file, an error is returned.
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return gptcometerrors.WrapError(err, "Config Load Failed", "Failed to read config file")
	}

	if err := yaml.Unmarshal(data, &m.config); err != nil {
		return gptcometerrors.WrapError(err, "Config Parse Failed", "Failed to parse config file")
	}

	return nil
}

// Save writes the configuration in the config field to the file specified
// by the configPath field.
//
// If an error occurs while marshaling the configuration or writing the file,
// an error is returned.
func (m *Manager) Save() error {
	data, err := yaml.Marshal(m.config)
	if err != nil {
		return gptcometerrors.WrapError(err, "Config Marshal Failed", "Failed to marshal config")
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return gptcometerrors.WrapError(err, "Config Save Failed", "Failed to write config file")
	}

	return nil
}

// getConfigDir returns the path to the gptcomet configuration directory
// within the user's home directory.
//
// The function retrieves the home directory using os.UserHomeDir and
// constructs the configuration path by appending ".config/gptcomet".
// If any error occurs while retrieving the home directory, it returns
// an empty string and the error.
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".config", "gptcomet"), nil
}

// List returns the current configuration as a YAML-formatted string.
// This method masks sensitive information such as API keys before converting to YAML.
// It excludes the prompt section from the configuration.
// Returns:
//   - string: YAML representation of the configuration
//   - error: If there was an error converting the config to YAML
func (m *Manager) List() (string, error) {
	// Get config without prompt section
	configCopy := m.ListWithoutPrompt()

	// Mask API keys
	MaskConfigAPIKeys(configCopy)

	// Convert to YAML
	yamlBytes, err := yaml.Marshal(configCopy)
	if err != nil {
		return "", fmt.Errorf("failed to convert config to yaml: %w", err)
	}

	return string(yamlBytes), nil
}

// UpdateProviderConfig updates the configuration for a specific provider and saves it to the config file.
// It takes a provider name string and a map of configuration key-value pairs as input.
// The configuration values are converted from string to interface{} type before being stored.
// If there is an error updating or saving the configuration, it returns an error with context.
// Returns nil on successful update and save.
func (m *Manager) UpdateProviderConfig(provider string, configs map[string]string) error {
	// Convert string values to interface{}
	providerConfig := make(map[string]interface{})
	for k, v := range configs {
		providerConfig[k] = v
	}

	// Update the config
	err := m.Set(provider, providerConfig)
	if err != nil {
		return fmt.Errorf("failed to update provider config: %w", err)
	}

	// Save the config
	if err := m.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}
