package config

import (
	"fmt"
	"strconv"
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
