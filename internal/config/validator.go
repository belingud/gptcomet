package config

import (
	"strings"
)

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
