package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidLanguage(t *testing.T) {
	tests := []struct {
		name string
		lang string
		want bool
	}{
		// Valid language codes
		{
			name: "Valid - English",
			lang: "en",
			want: true,
		},
		{
			name: "Valid - Simplified Chinese",
			lang: "zh-cn",
			want: true,
		},
		{
			name: "Valid - Traditional Chinese",
			lang: "zh-tw",
			want: true,
		},
		{
			name: "Valid - French",
			lang: "fr",
			want: true,
		},
		{
			name: "Valid - Vietnamese",
			lang: "vi",
			want: true,
		},
		{
			name: "Valid - Japanese",
			lang: "ja",
			want: true,
		},
		{
			name: "Valid - Korean",
			lang: "ko",
			want: true,
		},
		{
			name: "Valid - German",
			lang: "de",
			want: true,
		},
		{
			name: "Valid - Spanish",
			lang: "es",
			want: true,
		},
		{
			name: "Valid - Arabic",
			lang: "ar",
			want: true,
		},
		// Invalid language codes
		{
			name: "Invalid - empty string",
			lang: "",
			want: false,
		},
		{
			name: "Invalid - lowercase 'xx'",
			lang: "xx",
			want: false,
		},
		{
			name: "Invalid - uppercase",
			lang: "EN",
			want: false,
		},
		{
			name: "Invalid - mixed case",
			lang: "En",
			want: false,
		},
		{
			name: "Invalid - number",
			lang: "123",
			want: false,
		},
		{
			name: "Invalid - special characters",
			lang: "en-us!",
			want: false,
		},
		{
			name: "Invalid - partial code",
			lang: "zh-",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidLanguage(tt.lang)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOutputLanguageMap(t *testing.T) {
	// Test that the map is not empty
	assert.NotEmpty(t, OutputLanguageMap, "OutputLanguageMap should not be empty")

	// Test that common language codes exist
	commonLanguages := []string{
		"en", "zh-cn", "zh-tw", "fr", "de", "es", "ja", "ko", "ru", "ar",
	}
	for _, lang := range commonLanguages {
		_, ok := OutputLanguageMap[lang]
		assert.True(t, ok, "Language code %s should exist in OutputLanguageMap", lang)
	}

	// Test that values are non-empty strings
	for lang, name := range OutputLanguageMap {
		assert.NotEmpty(t, name, "Language name for code %s should not be empty", lang)
		assert.NotEqual(t, lang, name, "Language name should not be the same as code for %s", lang)
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		showFirst int
		want      string
	}{
		{
			name:      "Empty string",
			apiKey:    "",
			showFirst: 3,
			want:      "",
		},
		{
			name:      "Short key - no prefix",
			apiKey:    "abc",
			showFirst: 3,
			want:      "abc",
		},
		{
			name:      "Short key - shorter than showFirst",
			apiKey:    "ab",
			showFirst: 3,
			want:      "ab",
		},
		{
			name:      "OpenAI sk- prefix",
			apiKey:    "sk-abc123def456",
			showFirst: 3,
			want:      "sk-abc*********",
		},
		{
			name:      "OpenAI sk-or-v1- prefix",
			apiKey:    "sk-or-v1-abc123def456",
			showFirst: 3,
			want:      "sk-or-v1-abc*********",
		},
		{
			name:      "Grok gsk_ prefix",
			apiKey:    "gsk_abc123def456",
			showFirst: 3,
			want:      "gsk_abc*********",
		},
		{
			name:      "xAI xai- prefix",
			apiKey:    "xai-abc123def456",
			showFirst: 3,
			want:      "xai-abc*********",
		},
		{
			name:      "No known prefix",
			apiKey:    "mykey123456",
			showFirst: 4,
			want:      "myke*******",
		},
		{
			name:      "ShowFirst = 0",
			apiKey:    "sk-abc123",
			showFirst: 0,
			want:      "sk-******",
		},
		{
			name:      "ShowFirst = 1",
			apiKey:    "sk-abc123",
			showFirst: 1,
			want:      "sk-a*****",
		},
		{
			name:      "ShowFirst = 10",
			apiKey:    "sk-abc123def456",
			showFirst: 10,
			want:      "sk-abc123def4**",
		},
		{
			name:      "Very long key",
			apiKey:    "sk-" + "a" + "b" + "c" + "1234567890abcdef",
			showFirst: 3,
			want:      "sk-abc****************",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskAPIKey(tt.apiKey, tt.showFirst)
			assert.Equal(t, tt.want, got)

			// Verify that masked keys have asterisks at the end (unless empty or shorter than showFirst)
			if tt.apiKey != "" && len(tt.apiKey) > tt.showFirst {
				if len(got) > 0 {
					assert.Contains(t, got, "*", "Masked key should contain asterisks")
					// Count asterisks
					asteriskCount := 0
					for _, c := range got {
						if c == '*' {
							asteriskCount++
						}
					}
					assert.Greater(t, asteriskCount, 0, "Should have at least one asterisk")
				}
			}

			// Verify that the length is preserved
			assert.Equal(t, len(tt.apiKey), len(got), "Masked key should have same length as original")
		})
	}
}

func TestMaskConfigAPIKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		verifyFn func(t *testing.T, result map[string]interface{})
	}{
		{
			name:  "Empty map",
			input: map[string]interface{}{},
			verifyFn: func(t *testing.T, result map[string]interface{}) {
				assert.Empty(t, result)
			},
		},
		{
			name: "Single API key at root",
			input: map[string]interface{}{
				"api_key": "sk-abc123def456",
			},
			verifyFn: func(t *testing.T, result map[string]interface{}) {
				value, ok := result["api_key"].(string)
				require.True(t, ok)
				assert.Equal(t, "sk-abc*********", value)
				assert.Contains(t, value, "*")
			},
		},
		{
			name: "Multiple keys including api_key",
			input: map[string]interface{}{
				"api_key":    "sk-abc123",
				"model":      "gpt-4",
				"max_tokens": 2048,
			},
			verifyFn: func(t *testing.T, result map[string]interface{}) {
				// api_key should be masked
				apiKey, ok := result["api_key"].(string)
				require.True(t, ok)
				assert.Contains(t, apiKey, "*")

				// Other keys should remain unchanged
				assert.Equal(t, "gpt-4", result["model"])
				assert.Equal(t, 2048, result["max_tokens"])
			},
		},
		{
			name: "Nested structure with api_key",
			input: map[string]interface{}{
				"provider": "openai",
				"openai": map[string]interface{}{
					"api_key": "sk-abc123",
					"model":   "gpt-4",
				},
			},
			verifyFn: func(t *testing.T, result map[string]interface{}) {
				// Root provider should not be masked
				assert.Equal(t, "openai", result["provider"])

				// Nested api_key should be masked
				openai, ok := result["openai"].(map[string]interface{})
				require.True(t, ok)
				apiKey, ok := openai["api_key"].(string)
				require.True(t, ok)
				assert.Contains(t, apiKey, "*")

				// Nested model should not be masked
				assert.Equal(t, "gpt-4", openai["model"])
			},
		},
		{
			name: "Deeply nested api_key",
			input: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"api_key": "sk-abc123",
					},
				},
			},
			verifyFn: func(t *testing.T, result map[string]interface{}) {
				level1, ok := result["level1"].(map[string]interface{})
				require.True(t, ok)
				level2, ok := level1["level2"].(map[string]interface{})
				require.True(t, ok)
				apiKey, ok := level2["api_key"].(string)
				require.True(t, ok)
				assert.Contains(t, apiKey, "*")
			},
		},
		{
			name: "Multiple api_keys at different levels",
			input: map[string]interface{}{
				"api_key": "root-key",
				"openai": map[string]interface{}{
					"api_key": "sk-abc123",
				},
				"anthropic": map[string]interface{}{
					"api_key": "sk-ant-123",
				},
			},
			verifyFn: func(t *testing.T, result map[string]interface{}) {
				// All api_keys should be masked
				rootKey, ok := result["api_key"].(string)
				require.True(t, ok)
				assert.Contains(t, rootKey, "*")

				openai, ok := result["openai"].(map[string]interface{})
				require.True(t, ok)
				openaiKey, ok := openai["api_key"].(string)
				require.True(t, ok)
				assert.Contains(t, openaiKey, "*")

				anthropic, ok := result["anthropic"].(map[string]interface{})
				require.True(t, ok)
				anthropicKey, ok := anthropic["api_key"].(string)
				require.True(t, ok)
				assert.Contains(t, anthropicKey, "*")
			},
		},
		{
			name: "Key named api_key but not a string type",
			input: map[string]interface{}{
				"api_key": 12345,
			},
			verifyFn: func(t *testing.T, result map[string]interface{}) {
				// Should not mask non-string values
				value, ok := result["api_key"]
				require.True(t, ok)
				assert.Equal(t, 12345, value)
			},
		},
		{
			name: "String value in key not named api_key",
			input: map[string]interface{}{
				"other_key": "sk-abc123",
			},
			verifyFn: func(t *testing.T, result map[string]interface{}) {
				// Should not mask keys not named api_key
				value, ok := result["other_key"].(string)
				require.True(t, ok)
				assert.Equal(t, "sk-abc123", value)
				assert.NotContains(t, value, "*")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of input to avoid modifying it
			inputCopy := make(map[string]interface{})
			for k, v := range tt.input {
				inputCopy[k] = v
			}

			MaskConfigAPIKeys(inputCopy)

			if tt.verifyFn != nil {
				tt.verifyFn(t, inputCopy)
			}
		})
	}
}

func TestMaskAPIKey_PrefixOrder(t *testing.T) {
	// Test that prefixes are checked in the correct order
	// sk-or-v1- should be checked before sk- to avoid incorrect masking
	tests := []struct {
		name      string
		apiKey    string
		showFirst int
		want      string
	}{
		{
			name:      "sk-or-v1- prefix should match",
			apiKey:    "sk-or-v1-abc123",
			showFirst: 3,
			want:      "sk-or-v1-abc***",
		},
		{
			name:      "sk- prefix (not sk-or-v1-)",
			apiKey:    "sk-abc123",
			showFirst: 3,
			want:      "sk-abc***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskAPIKey(tt.apiKey, tt.showFirst)
			assert.Equal(t, tt.want, got)
		})
	}
}
