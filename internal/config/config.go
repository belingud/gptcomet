package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/belingud/gptcomet/internal/debug"
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
			return nil, fmt.Errorf("failed to get config directory: %w", err)
		}
		configPath = configPath + "/gptcomet.yaml"
	} else {
		// Check if the config file exists
		if _, err := os.Stat(configPath); err != nil {
			return nil, fmt.Errorf("config file does not exist: %w", err)
		}
	}

	manager := &Manager{
		config:     make(map[string]interface{}),
		configPath: configPath,
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing config if it exists
	if _, err := os.Stat(configPath); err == nil {
		if err := manager.Load(); err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		// Initialize with default configuration
		manager.config = defaults.DefaultConfig
		if err := manager.Save(); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
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
			return nil, fmt.Errorf("provider not set")
		}
	} else {
		provider = initProvider
	}
	providerConfig, ok := m.config[provider].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("provider config not found: %s", provider)
	}
	var apiKey string
	if provider == "ollama" {
		apiKey = ""
	} else {
		apiKeyValue, ok := providerConfig["api_key"].(string)
		if !ok || apiKeyValue == "" {
			return nil, fmt.Errorf("api_key not found for provider: %s", provider)
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
			return nil, fmt.Errorf("failed to parse extra_headers: %w", err)
		}
		clientConfig.ExtraHeaders = extraHeaders
	}

	// Parse extra_body (additional request body)
	if extraBodyStr, ok := providerConfig["extra_body"].(string); ok && extraBodyStr != "" && extraBodyStr != "{}" {
		extraBody := make(map[string]interface{})
		if err := json.Unmarshal([]byte(extraBodyStr), &extraBody); err != nil {
			return nil, fmt.Errorf("failed to parse extra_body: %w", err)
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

// Get returns the value associated with the given key. The key is split
// on the '.' character and the value is retrieved from the nested map.
//
// If the key is not found, the second return value is false.
func (m *Manager) Get(key string) (interface{}, bool) {
	return m.GetNestedValue(strings.Split(key, "."))
}

// GetWithDefault retrieves the value associated with the given key from the configuration.
// If the key is not found, it returns the provided defaultValue.
//
// The key is split on the '.' character to navigate the nested map structure.
//
// Parameters:
//
//	key - The configuration key to look up.
//	defaultValue - The value to return if the key is not found.
//
// Returns:
//
//	The configuration value associated with the key, or defaultValue if the key does not exist.
func (m *Manager) GetWithDefault(key string, defaultValue interface{}) interface{} {
	value, ok := m.GetNestedValue(strings.Split(key, "."))
	if !ok {
		return defaultValue
	}
	return value
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
				return fmt.Errorf("invalid language code: %s", str)
			}
		}
	case "output.translate_title":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("translate_title must be a boolean value")
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
		return fmt.Errorf("value at key '%s' is not a list", key)
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
		return fmt.Errorf("value at key '%s' is not a list", key)
	}

	// Append the new value
	list = append(list, value)
	return m.Set(key, list)
}

// GetNestedValue retrieves the value associated with the given key path from the
// configuration.
//
// The key path is a slice of strings where each string is a key in a nested map.
// For example, the key path ["a", "b", "c"] would retrieve the value associated with
// the key "c" from the map "b" which is a value in the map "a".
//
// If any of the keys in the path do not exist, the method returns (nil, false).
// If the key path is valid, the method returns the value associated with the last
// key in the path and true.
func (m *Manager) GetNestedValue(keys []string) (interface{}, bool) {
	current := interface{}(m.config)
	for _, key := range keys {
		currentMap, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current, ok = currentMap[key]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

// SetNestedValue sets the value associated with the given key path in the
// configuration.
//
// The key path is a slice of strings where each string is a key in a nested map.
// For example, the key path ["a", "b", "c"] would set the value associated with
// the key "c" in the map "b" which is a value in the map "a".
//
// If any of the keys in the path do not exist, the method creates them as needed.
// The method returns the value associated with the last key in the path.
func (m *Manager) SetNestedValue(keys []string, value interface{}) {
	current := m.config
	for _, key := range keys[:len(keys)-1] {
		next, ok := current[key]
		if !ok {
			next = make(map[string]interface{})
			current[key] = next
		}
		if nextMap, ok := next.(map[string]interface{}); ok {
			current = nextMap
		} else {
			newMap := make(map[string]interface{})
			current[key] = newMap
			current = newMap
		}
	}
	current[keys[len(keys)-1]] = value
}

// Load reads the configuration from the file specified by the configPath
// field and unmarshals it into the config field.
//
// If the file does not exist or an error occurs while reading or parsing the
// file, an error is returned.
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &m.config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
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
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
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

// OutputLanguageMap maps language codes to their names
var OutputLanguageMap = map[string]string{
	"en":    "English",
	"zh-cn": "Simplified Chinese",
	"zh-tw": "Traditional Chinese",
	"fr":    "French",
	"vi":    "Vietnamese",
	"ja":    "Japanese",
	"ko":    "Korean",
	"ru":    "Russian",
	"tr":    "Turkish",
	"id":    "Indonesian",
	"th":    "Thai",
	"de":    "German",
	"es":    "Spanish",
	"pt":    "Portuguese",
	"it":    "Italian",
	"ar":    "Arabic",
	"hi":    "Hindi",
	"el":    "Greek",
	"pl":    "Polish",
	"nl":    "Dutch",
	"sv":    "Swedish",
	"fi":    "Finnish",
	"hu":    "Hungarian",
	"cs":    "Czech",
	"ro":    "Romanian",
	"bg":    "Bulgarian",
	"uk":    "Ukrainian",
	"he":    "Hebrew",
	"lt":    "Lithuanian",
	"la":    "Latin",
	"ca":    "Catalan",
	"sr":    "Serbian",
	"sl":    "Slovenian",
	"mk":    "Macedonian",
	"lv":    "Latvian",
	"bn":    "Bengali",
	"ta":    "Tamil",
	"te":    "Telugu",
	"ml":    "Malayalam",
	"si":    "Sinhala",
	"fa":    "Persian",
	"ur":    "Urdu",
	"pa":    "Punjabi",
	"mr":    "Marathi",
}

// IsValidLanguage returns true if the given language code is valid, false otherwise.
// It uses the OutputLanguageMap to check if the language code is valid.
func IsValidLanguage(lang string) bool {
	_, ok := OutputLanguageMap[lang]
	return ok
}

// GetSupportedKeys returns a sorted list of all supported configuration keys.
//
// The returned list will include the following keys:
//   - provider
//   - file_ignore
//   - output.lang
//   - output.rich_template
//   - output.translate_title
//   - console.verbose
//   - <provider>.api_base
//   - <provider>.api_key
//   - <provider>.model
//   - <provider>.retries
//   - <provider>.proxy
//   - <provider>.max_tokens
//   - <provider>.top_p
//   - <provider>.temperature
//   - <provider>.frequency_penalty
//   - <provider>.extra_headers
//   - <provider>.completion_path
//   - <provider>.answer_path
//   - prompt.brief_commit_message
//   - prompt.rich_commit_message
//   - prompt.translation
//
// The <provider> placeholder in the returned list will be replaced with the name of the current provider.
func (m *Manager) GetSupportedKeys() []string {
	// Get current provider
	provider, _ := m.GetNestedValue([]string{"provider"})
	providerStr, ok := provider.(string)
	if !ok || providerStr == "" {
		providerStr = "openai"
	}

	// Generate keys from code logic
	keys := make(map[string]bool)

	// Root level keys
	keys["provider"] = true
	keys["file_ignore"] = true

	// Output keys
	outputKeys := []string{
		"lang",
		"rich_template",
		"translate_title",
		"review_lang",
	}
	for _, key := range outputKeys {
		keys["output."+key] = true
	}

	// Console keys
	consoleKeys := []string{
		"verbose",
	}
	for _, key := range consoleKeys {
		keys["console."+key] = true
	}

	// Provider keys
	providerKeys := []string{
		"api_base",
		"api_key",
		"model",
		"retries",
		"proxy",
		"max_tokens",
		"top_p",
		"temperature",
		"frequency_penalty",
		"extra_headers",
		"extra_body",
		"completion_path",
		"answer_path",
	}
	for _, key := range providerKeys {
		keys["<provider>."+key] = true
	}

	// Prompt keys
	promptKeys := []string{
		"brief_commit_message",
		"rich_commit_message",
		"translation",
	}
	for _, key := range promptKeys {
		keys["prompt."+key] = true
	}

	// Convert map to sorted slice
	result := make([]string, 0, len(keys))
	for key := range keys {
		// Replace provider name with <provider> placeholder
		if strings.HasPrefix(key, providerStr+".") {
			key = strings.Replace(key, providerStr+".", "<provider>.", 1)
		}
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

// GetPrompt returns the prompt template for the given rich/non-rich option.
// If the prompt is not set in the configuration, it returns the default prompt.
func (m *Manager) GetPrompt(isRich bool) string {
	promptConfig, ok := m.config["prompt"].(map[string]interface{})
	if !ok {
		// return default prompt if not set in config
		if isRich {
			return defaults.PromptDefaults["rich_commit_message"]
		} else {
			return defaults.PromptDefaults["brief_commit_message"]
		}
	}

	if isRich {
		if rich, ok := promptConfig["rich_commit_message"].(string); ok {
			return rich
		}
	} else {
		if brief, ok := promptConfig["brief_commit_message"].(string); ok {
			return brief
		}
	}

	// return default prompt if not set in config
	if isRich {
		return defaults.PromptDefaults["rich_commit_message"]
	} else {
		return defaults.PromptDefaults["brief_commit_message"]
	}
}

func (m *Manager) GetReviewPrompt() string {
	promptConfig, ok := m.config["prompt"].(map[string]interface{})
	if !ok {
		// return default prompt if not set in config
		debug.Printf("Prompt not found in config, using default")
		return defaults.PromptDefaults["review"]
	}
	if review, ok := promptConfig["review"].(string); ok {
		return review
	}
	// return default prompt if not set in config
	debug.Printf("Using default review prompt in the end")
	return defaults.PromptDefaults["review"]
}

// GetTranslationPrompt retrieves the translation prompt from the configuration.
// If the prompt configuration is not set or if the translation prompt is not found,
// it returns the default translation prompt from defaults package.
//
// Returns:
//   - string: The translation prompt to be used
func (m *Manager) GetTranslationPrompt() string {
	promptConfig, ok := m.config["prompt"].(map[string]interface{})
	if !ok {
		// return default prompt if not set in config
		return defaults.PromptDefaults["translation"]
	}
	if translation, ok := promptConfig["translation"].(string); ok {
		return translation
	}
	// return default prompt if not set in config
	return defaults.PromptDefaults["translation"]
}

// MaskAPIKey masks an API key by showing only the first few characters and replacing the rest with asterisks.
// It preserves common API key prefixes (e.g., "sk-", "gsk_") in the visible part.
//
// Parameters:
//   - apiKey: The API key string to be masked
//   - showFirst: The number of characters to show after the prefix (or from start if no prefix)
//
// Returns:
//   - A string with the masked API key, preserving the prefix (if any) and showing the specified
//     number of characters, with the remainder replaced by asterisks.
//   - Returns empty string if input is empty
//
// Example:
//
//	MaskAPIKey("sk-abc123def456", 3) returns "sk-abc***********"
//	MaskAPIKey("mykey123456", 4) returns "mkey******"
func MaskAPIKey(apiKey string, showFirst int) string {
	if apiKey == "" {
		return apiKey
	}

	// Check common API key prefixes
	prefixes := []string{"sk-or-v1-", "sk-", "gsk_", "xai-"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(apiKey, prefix) {
			visiblePart := apiKey[:len(prefix)+showFirst]
			return visiblePart + strings.Repeat("*", len(apiKey)-len(visiblePart))
		}
	}

	// No prefix found, mask all but the first few characters
	if len(apiKey) <= showFirst {
		return apiKey
	}
	return apiKey[:showFirst] + strings.Repeat("*", len(apiKey)-showFirst)
}

// MaskConfigAPIKeys recursively traverses a configuration map and masks any API keys found within it.
// It specifically looks for keys named "api_key" and masks their string values using MaskAPIKey function,
// preserving only the first 3 characters visible.
//
// Parameters:
//   - data: A map[string]interface{} containing configuration data that may include API keys
//
// The function modifies the input map in place, replacing sensitive API key values with masked versions.
// It handles nested maps by recursively processing them with the same masking logic.
func MaskConfigAPIKeys(data map[string]interface{}) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			if key == "api_key" {
				data[key] = MaskAPIKey(v, 3)
			}
		case map[string]interface{}:
			MaskConfigAPIKeys(v)
		}
	}
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

// GetFileIgnore retrieves the list of file patterns to ignore from the configuration.
// It returns:
//   - A slice of strings containing the ignore patterns if configured
//   - nil if no patterns are configured or if the configuration is invalid
//
// The patterns should be in a format compatible with filepath.Match
func (m *Manager) GetFileIgnore() []string {
	value, ok := m.Get("file_ignore")
	if !ok {
		return nil
	}

	if patterns, ok := value.([]interface{}); ok {
		result := make([]string, 0, len(patterns))
		for _, pattern := range patterns {
			if str, ok := pattern.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}

	return nil
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

// GetOutputTranslateTitle returns whether the title should be translated in the output.
// If the configuration value is not found, it returns true by default.
// If the configuration value exists but cannot be converted to boolean, it returns true.
func (m *Manager) GetOutputTranslateTitle() bool {
	value, ok := m.Get("output.translate_title")
	if !ok {
		return false
	}

	if b, ok := value.(bool); ok {
		return b
	}

	return false
}

func getIntValue(config map[string]interface{}, key string, defaultValue int) int {
	if val, ok := config[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			// check if it is an integer
			if v == float64(int(v)) {
				return int(v)
			}
			fmt.Printf("Warning: %s value '%v' is not an integer, using default: %d\n", key, v, defaultValue)
		case string:
			if intVal, err := strconv.Atoi(v); err == nil {
				return intVal
			}
			fmt.Printf("Warning: %s value '%s' is not a valid integer, using default: %d\n", key, v, defaultValue)
		default:
			fmt.Printf("Warning: %s has unexpected type %T, using default: %d\n", key, v, defaultValue)
		}
	}
	return defaultValue
}

func getFloatValue(config map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := config[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
				return floatVal
			}
			fmt.Printf("Warning: %s value '%s' is not a valid float, using default: %f\n", key, v, defaultValue)
		default:
			fmt.Printf("Warning: %s has unexpected type %T, using default: %f\n", key, v, defaultValue)
		}
	}
	return defaultValue
}
