package config

import (
	"sort"
	"strings"

	"github.com/belingud/gptcomet/pkg/config/defaults"
)

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

// GetReviewPrompt returns the review prompt from the configuration.
// If the prompt is not set, it returns the default review prompt.
func (m *Manager) GetReviewPrompt() string {
	promptConfig, ok := m.config["prompt"].(map[string]interface{})
	if !ok {
		// return default prompt if not set in config
		return defaults.PromptDefaults["review"]
	}
	if review, ok := promptConfig["review"].(string); ok {
		return review
	}
	// return default prompt if not set in config
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

// GetOutputTranslateTitle returns whether the title should be translated in the output.
// If the configuration value is not found, it returns false by default.
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
