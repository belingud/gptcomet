package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	gptcometerrors "github.com/belingud/gptcomet/internal/errors"
)

var validMarkdownThemes = map[string]struct{}{
	"auto":        {},
	"ascii":       {},
	"dark":        {},
	"dracula":     {},
	"light":       {},
	"tokyo-night": {},
	"notty":       {},
	"pink":        {},
}

// IsValidMarkdownTheme reports whether theme is one of the themes supported by
// the markdown renderer.
func IsValidMarkdownTheme(theme string) bool {
	_, ok := validMarkdownThemes[theme]
	return ok
}

// Validate checks the current configuration and returns human-readable issues.
// An empty slice means the configuration passed structural validation.
func (m *Manager) Validate() []string {
	if m == nil {
		return []string{"configuration manager is nil"}
	}

	var issues []string

	provider, _ := m.config["provider"].(string)
	provider = strings.TrimSpace(provider)
	if provider == "" {
		issues = append(issues, "provider is not set")
	}

	providerConfig, providerOK := m.config[provider].(map[string]interface{})
	if provider != "" && !providerOK {
		issues = append(issues, fmt.Sprintf("provider %q has no configuration section", provider))
	}
	if providerOK {
		issues = append(issues, validateAPIKey(provider, providerConfig)...)
		issues = append(issues, validateProviderSpecific(provider, providerConfig)...)
		issues = append(issues, validateProviderNumbers(providerConfig)...)
	}

	issues = append(issues, validateOutputSection(m.config["output"])...)

	if raw, exists := m.config["file_ignore"]; exists {
		if _, ok := toStringSlice(raw); !ok {
			issues = append(issues, "file_ignore must be a list of strings")
		}
	}

	issues = append(issues, validatePromptSection(m.config["prompt"])...)

	return issues
}

func validateAPIKey(provider string, providerConfig map[string]interface{}) []string {
	if provider == "ollama" {
		return nil
	}

	raw, exists := providerConfig["api_key"]
	if exists {
		value, ok := raw.(string)
		if ok && strings.TrimSpace(value) != "" {
			return nil
		}
		if raw != nil && !ok {
			return []string{fmt.Sprintf("provider %q api_key must be a string", provider)}
		}
	}

	envVar := gptcometerrors.GetAPIKeyEnvVar(provider)
	if strings.TrimSpace(os.Getenv(envVar)) != "" {
		return nil
	}

	return []string{fmt.Sprintf("provider %q api_key is not set (set it in the config file or export %s)", provider, envVar)}
}

func validateProviderSpecific(provider string, providerConfig map[string]interface{}) []string {
	switch provider {
	case "azure":
		if strings.TrimSpace(getStringValue(providerConfig, "deployment_name", "")) == "" {
			return []string{"provider \"azure\" requires deployment_name"}
		}
	case "vertex":
		var issues []string
		if strings.TrimSpace(getStringValue(providerConfig, "project_id", "")) == "" {
			issues = append(issues, "provider \"vertex\" requires project_id")
		}
		if strings.TrimSpace(getStringValue(providerConfig, "location", "")) == "" {
			issues = append(issues, "provider \"vertex\" requires location")
		}
		return issues
	}
	return nil
}

func validateProviderNumbers(providerConfig map[string]interface{}) []string {
	var issues []string
	issues = append(issues, validateIntField(providerConfig, "max_tokens", 1)...)
	issues = append(issues, validateIntField(providerConfig, "retries", 1)...)
	issues = append(issues, validateFloatField(providerConfig, "top_p", 0, 1)...)
	issues = append(issues, validateFloatField(providerConfig, "temperature", 0, 2)...)
	issues = append(issues, validateFloatField(providerConfig, "frequency_penalty", -2, 2)...)
	return issues
}

func validateIntField(config map[string]interface{}, key string, min int) []string {
	raw, exists := config[key]
	if !exists {
		return nil
	}

	var value int
	switch v := raw.(type) {
	case int:
		value = v
	case float64:
		if v != float64(int(v)) {
			return []string{fmt.Sprintf("%s must be an integer", key)}
		}
		value = int(v)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return []string{fmt.Sprintf("%s must be an integer", key)}
		}
		value = parsed
	default:
		return []string{fmt.Sprintf("%s must be an integer", key)}
	}

	if value < min {
		return []string{fmt.Sprintf("%s must be >= %d (got %d)", key, min, value)}
	}
	return nil
}

func validateFloatField(config map[string]interface{}, key string, min, max float64) []string {
	raw, exists := config[key]
	if !exists {
		return nil
	}

	var value float64
	switch v := raw.(type) {
	case int:
		value = float64(v)
	case float64:
		value = v
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return []string{fmt.Sprintf("%s must be a number", key)}
		}
		value = parsed
	default:
		return []string{fmt.Sprintf("%s must be a number", key)}
	}

	if value < min || value > max {
		return []string{fmt.Sprintf("%s must be between %g and %g (got %g)", key, min, max, value)}
	}
	return nil
}

func validateOutputSection(raw interface{}) []string {
	if raw == nil {
		return nil
	}

	output, ok := raw.(map[string]interface{})
	if !ok {
		return []string{"output must be a mapping"}
	}

	var issues []string
	for _, key := range []string{"lang", "review_lang"} {
		value, exists := output[key]
		if !exists {
			continue
		}
		lang, ok := value.(string)
		if !ok || !IsValidLanguage(lang) {
			issues = append(issues, fmt.Sprintf("output.%s must be a supported language code", key))
		}
	}

	if value, exists := output["markdown_theme"]; exists {
		theme, ok := value.(string)
		if !ok || !IsValidMarkdownTheme(theme) {
			issues = append(issues, "output.markdown_theme is not a supported theme")
		}
	}

	if value, exists := output["translate_title"]; exists {
		if _, ok := value.(bool); !ok {
			issues = append(issues, "output.translate_title must be a boolean")
		}
	}

	return issues
}

func validatePromptSection(raw interface{}) []string {
	if raw == nil {
		return nil
	}

	prompt, ok := toStringMap(raw)
	if !ok {
		return []string{"prompt must be a mapping"}
	}

	var issues []string
	for _, key := range []string{"brief_commit_message", "rich_commit_message", "translation", "review"} {
		value, exists := prompt[key]
		if !exists {
			continue
		}
		if text, ok := value.(string); !ok || strings.TrimSpace(text) == "" {
			issues = append(issues, fmt.Sprintf("prompt.%s must be a non-empty string", key))
		}
	}
	return issues
}

func toStringMap(raw interface{}) (map[string]interface{}, bool) {
	switch value := raw.(type) {
	case map[string]interface{}:
		return value, true
	case map[string]string:
		result := make(map[string]interface{}, len(value))
		for key, item := range value {
			result[key] = item
		}
		return result, true
	default:
		return nil, false
	}
}

func toStringSlice(raw interface{}) ([]string, bool) {
	switch value := raw.(type) {
	case []string:
		return value, true
	case []interface{}:
		result := make([]string, 0, len(value))
		for _, item := range value {
			text, ok := item.(string)
			if !ok {
				return nil, false
			}
			result = append(result, text)
		}
		return result, true
	default:
		return nil, false
	}
}
