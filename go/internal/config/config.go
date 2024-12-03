package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gptcomet/pkg/types"

	"gopkg.in/yaml.v3"
)

// Manager handles configuration management
type Manager struct {
	config     map[string]interface{}
	configPath string
}

// New creates a new configuration manager
func New() (*Manager, error) {
	configDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(configDir, ".config", "gptcomet", "config.yaml")
	manager := &Manager{
		config:     make(map[string]interface{}),
		configPath: configPath,
	}

	// Create config directory if it doesn't exist
	configDirPath := filepath.Dir(configPath)
	if err := os.MkdirAll(configDirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing config if it exists
	if _, err := os.Stat(configPath); err == nil {
		if err := manager.load(); err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		// Initialize with default configuration
		defaultConfig := defaultConfig()
		manager.config = defaultConfig
		if err := manager.save(); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
	}

	return manager, nil
}

// GetClientConfig retrieves the client configuration
func (m *Manager) GetClientConfig() (*types.ClientConfig, error) {
	provider, ok := m.config["provider"].(string)
	if !ok {
		return nil, fmt.Errorf("provider not set")
	}

	providerConfig, ok := m.config[provider].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("provider config not found: %s", provider)
	}

	apiKey, ok := providerConfig["api_key"].(string)
	if !ok {
		return nil, fmt.Errorf("api_key not found for provider: %s", provider)
	}

	apiBase := types.DefaultAPIBase
	if base, ok := providerConfig["api_base"].(string); ok {
		apiBase = base
	}

	model := types.DefaultModel
	if m, ok := providerConfig["model"].(string); ok {
		model = m
	}

	completionPath := "/chat/completions"
	if path, ok := providerConfig["completion_path"].(string); ok {
		completionPath = path
	}

	return &types.ClientConfig{
		APIBase:        apiBase,
		APIKey:         apiKey,
		Model:          model,
		Provider:       provider,
		Retries:        types.DefaultRetries,
		CompletionPath: completionPath,
	}, nil
}

// SetProvider sets the provider configuration
func (m *Manager) SetProvider(provider, apiKey, apiBase, model string) error {
	if apiBase == "" {
		apiBase = types.DefaultAPIBase
	}
	if model == "" {
		model = types.DefaultModel
	}

	m.config[provider] = map[string]interface{}{
		"api_key":  apiKey,
		"api_base": apiBase,
		"model":    model,
	}
	m.config["provider"] = provider

	return m.save()
}

// Get retrieves a configuration value
func (m *Manager) Get(key string) (interface{}, bool) {
	return m.getNestedValue(strings.Split(key, "."))
}

// Set sets a configuration value
func (m *Manager) Set(key string, value interface{}) error {
	keys := strings.Split(key, ".")
	m.setNestedValue(keys, value)
	return m.save()
}

// ListWithoutPrompt returns all configuration as a map without the prompt section
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

// Reset resets the configuration to default values
// If promptOnly is true, only reset the prompt section
func (m *Manager) Reset(promptOnly bool) error {
	if promptOnly {
		// Get default prompt config
		defaultCfg := defaultConfig()
		if promptConfig, ok := defaultCfg["prompt"].(map[string]interface{}); ok {
			m.config["prompt"] = promptConfig
		}
	} else {
		// Reset all config
		m.config = defaultConfig()
	}
	return m.save()
}

// Remove removes a configuration value or a value from a list
func (m *Manager) Remove(key string, value string) error {
	keys := strings.Split(key, ".")
	if value == "" {
		// If no value is provided, remove the entire key
		lastKey := keys[len(keys)-1]
		parent, ok := m.getNestedValue(keys[:len(keys)-1])
		if !ok {
			return nil
		}

		if parentMap, ok := parent.(map[string]interface{}); ok {
			delete(parentMap, lastKey)
		}
		return m.save()
	}

	// If value is provided, try to remove it from a list
	current, ok := m.getNestedValue(keys)
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

// GetPath returns the configuration file path
func (m *Manager) GetPath() string {
	return m.configPath
}

// Append appends a value to a list configuration
func (m *Manager) Append(key string, value interface{}) error {
	keys := strings.Split(key, ".")
	current, ok := m.getNestedValue(keys)
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

// getNestedValue retrieves a nested configuration value
func (m *Manager) getNestedValue(keys []string) (interface{}, bool) {
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

// setNestedValue sets a nested configuration value
func (m *Manager) setNestedValue(keys []string, value interface{}) {
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

// load reads the configuration from file
func (m *Manager) load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &m.config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// save writes the configuration to file
func (m *Manager) save() error {
	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getConfigDir returns the configuration directory path
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".config", "gptcomet"), nil
}

// defaultConfig returns the default configuration
func defaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"provider": "openai",
		"file_ignore": []string{
			"bun.lockb",
			"Cargo.lock",
			"composer.lock",
			"Gemfile.lock",
			"package-lock.json",
			"pnpm-lock.yaml",
			"poetry.lock",
			"yarn.lock",
			"pdm.lock",
			"Pipfile.lock",
			"*.py[cod]",
			"go.mod",
			"go.sum",
			"uv.lock",
			"README.md",
			"README.MD",
			"*.md",
			"*.MD",
		},
		"output": map[string]interface{}{
			"lang":          "en",
			"rich_template": "<title>:<summary>\n\n<detail>",
		},
		"console": map[string]interface{}{
			"verbose": true,
		},
		"openai": map[string]interface{}{
			"api_base":        types.DefaultAPIBase,
			"api_key":         "",
			"model":           types.DefaultModel,
			"retries":         2,
			"proxy":           "",
			"max_tokens":      2048,
			"top_p":           0.7,
			"temperature":     0.7,
			"frequency_penalty": 0,
			"extra_headers":   "{}",
			"completion_path": "/chat/completions",
			"answer_path":     "choices.0.message.content",
		},
		"anthropic": map[string]interface{}{
			"api_base":        "https://api.anthropic.com",
			"api_key":         "",
			"model":           "claude-2",
			"retries":         2,
			"proxy":           "",
			"max_tokens":      2048,
			"top_p":           0.7,
			"temperature":     0.7,
			"frequency_penalty": 0,
			"extra_headers":   "{}",
			"completion_path": "/v1/messages",
			"answer_path":     "content.0.text",
		},
		"prompt": map[string]interface{}{
			"brief_commit_message": `you are an expert software engineer responsible for writing a clear and concise commit message.
Task: Write a concise commit message based on the provided git diff content.

Guidelines:
- start with a concise, informative title.
- follow with a high-level summary in bullet points (imperative tense).
- focus on the most significant changes.
- sometimes you need to judge the effect based on the type of files that have been modified.

use one of the following labels for the title:

- build: changes that affect the build system or external dependencies (example scopes: gulp, broccoli, npm)
- chore: updating libraries, copyrights or other setting, includes updating dependencies.
- ci: changes to our CI configuration files and scripts (example scopes: Travis, Circle, gitHub Actions)
- docs: non-code changes, such as fixing typos or adding new documentation
- feat: a commit of the type feat introduces a new feature to the codebase
- fix: a commit of the type fix patches a bug in your codebase
- perf: a code change that improves performance
- refactor: a code change that neither fixes a bug nor adds a feature
- style: changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)
- test: adding missing tests or correcting existing tests

the commit message templete is <title>: <summary>

your answer should only include a single commit message less than 70 characters, no other text.

Git diff like below example:
` + "```" + `
diff --git a/tests/test_stylize.py b/tests/test_stylize.py
@@ -7,5 +7,5 @@ def test_stylize_text():
    text = "Hello, world!"
    styles = ["bold", "italic"]
-    result = stylize(text, *styles)
+    result = stylize(text, *styles, "red")
` + "```" + `
No space before ` + "`diff`" + `, this example means function ` + "`test_stylize_text`" + ` in ` + "`test_stylize.py`" + ` is modified in this commit.
Then there is a specifier of the lines that were modified.
A line starting with ` + "`+`" + ` means it was added.
A line that starts with ` + "`-`" + ` means that line was deleted.
A line that starts with neither ` + "`+`" + ` nor ` + "`-`" + ` is code given for context and better understanding.
If there are some spaces before ` + "`+`" + `, ` + "`-`" + ` or ` + "`diff`" + ` at the beginning, it could be context. It is not part of the diff.
After the git diff of the first file, there will be an empty line, and then the git diff of the next file.

Examples:
test: update import of stylize test
fix: Fix password hashing vulnerability

Generate commit message by below git diff:
{{ placeholder }}

Commit Message:`,
			"rich_commit_message": `you are an expert software engineer responsible for writing a clear and concise commit message.
Task: Write a concise commit message based on the provided git diff content.

Guidelines:
- start with a concise, informative title.
- follow with a high-level summary in bullet points (imperative tense).
- focus on the most significant changes.
- sometimes you need to judge the effect based on the type of files that have been modified.

use one of the following labels for the title:

- build: changes that affect the build system or external dependencies (example scopes: gulp, broccoli, npm)
- chore: updating libraries, copyrights or other setting, includes updating dependencies.
- ci: changes to our CI configuration files and scripts (example scopes: Travis, Circle, gitHub Actions)
- docs: non-code changes, such as fixing typos or adding new documentation
- feat: a commit of the type feat introduces a new feature to the codebase
- fix: a commit of the type fix patches a bug in your codebase
- perf: a code change that improves performance
- refactor: a code change that neither fixes a bug nor adds a feature
- style: changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)
- test: adding missing tests or correcting existing tests

The commit message template is {{ output.rich_template }}. Your answer should only include commit message, no other text or ` + "`" + `.
If your answer includes details about the commit, please list each item on a new line.

Git diff like below example:
` + "```" + `
diff --git a/tests/test_stylize.py b/tests/test_stylize.py
@@ -7,5 +7,5 @@ def test_stylize_text():
    text = "Hello, world!"
    styles = ["bold", "italic"]
-    result = stylize(text, *styles)
+    result = stylize(text, *styles, "red")
` + "```" + `
No space before ` + "`diff`" + `, this example means function ` + "`test_stylize_text`" + ` in ` + "`test_stylize.py`" + ` is modified in this commit.
Then there is a specifier of the lines that were modified.
A line starting with ` + "`+`" + ` means it was added.
A line that starts with ` + "`-`" + ` means that line was deleted.
A line that starts with neither ` + "`+`" + ` nor ` + "`-`" + ` is code given for context and better understanding.
If there are some spaces before ` + "`+`" + `, ` + "`-`" + ` or ` + "`diff`" + ` at the beginning, it could be context. It is not part of the diff.
After the git diff of the first file, there will be an empty line, and then the git diff of the next file.

Example:
feat: support generating rich commit message

- implement rich commit message generate function
- delete unused functions in message generater

Generate commit message by below git diff:
{{ placeholder }}

Commit Message:`,
			"translation": `You are a professional polyglot programmer and translator. You are translating a git commit message.
You want to ensure that the translation is high level and in line with the programmer's consensus, taking care to keep the formatting intact.

Translate the following message into {{ output.lang }}.

GIT COMMIT MESSAGE:

{{ placeholder }}

Remember translate all given git commit message and give me only the translation.
THE TRANSLATION:`,
		},
	}
}

// GetSupportedKeys returns a list of supported configuration keys
func (m *Manager) GetSupportedKeys() []string {
	// Get current provider
	provider, _ := m.getNestedValue([]string{"provider"})
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

// GetPrompt retrieves the prompt configuration
func (m *Manager) GetPrompt(isRich bool) string {
	promptConfig, ok := m.config["prompt"].(map[string]interface{})
	if !ok {
		return ""
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
	return ""
}

// GetTranslationPrompt retrieves the translation prompt
func (m *Manager) GetTranslationPrompt() string {
	promptConfig, ok := m.config["prompt"].(map[string]interface{})
	if !ok {
		return ""
	}
	if translation, ok := promptConfig["translation"].(string); ok {
		return translation
	}
	return ""
}

// MaskAPIKey masks an API key by showing only the first few characters and replacing the rest with asterisks
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

// MaskConfigAPIKeys recursively masks all API keys in a map
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

// List returns the configuration as a string with masked API keys
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
