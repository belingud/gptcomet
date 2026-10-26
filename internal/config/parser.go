package config

import (
	"fmt"
	"strconv"
	"time"
)

// getIntValue converts a configuration value to an integer.
// If the value cannot be converted, it returns the default value and logs a warning.
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

// getFloatValue converts a configuration value to a float64.
// If the value cannot be converted, it returns the default value and logs a warning.
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

// getStringValue converts a configuration value to a string.
// If the value is missing or not a string, it returns the default value.
func getStringValue(config map[string]interface{}, key string, defaultValue string) string {
	if val, ok := config[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		case time.Time:
			// YAML decodes unquoted date-like values (e.g. api_version:
			// 2024-06-01) into time.Time. Preserve the user's date form.
			if v.Hour() == 0 && v.Minute() == 0 && v.Second() == 0 && v.Nanosecond() == 0 {
				return v.Format("2006-01-02")
			}
			return v.Format(time.RFC3339)
		}
	}
	return defaultValue
}
