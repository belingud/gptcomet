package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIntValue(t *testing.T) {
	tests := []struct {
		name         string
		config       map[string]interface{}
		key          string
		defaultValue int
		want         int
	}{
		{
			name:         "Key not present - return default",
			config:       map[string]interface{}{},
			key:          "missing",
			defaultValue: 100,
			want:         100,
		},
		{
			name: "Int value - return as is",
			config: map[string]interface{}{
				"value": 42,
			},
			key:          "value",
			defaultValue: 100,
			want:         42,
		},
		{
			name: "Float64 value that is an integer - convert to int",
			config: map[string]interface{}{
				"value": 42.0,
			},
			key:          "value",
			defaultValue: 100,
			want:         42,
		},
		{
			name: "Float64 value that is not an integer - return default with warning",
			config: map[string]interface{}{
				"value": 42.5,
			},
			key:          "value",
			defaultValue: 100,
			want:         100,
		},
		{
			name: "String value that is a valid integer - convert",
			config: map[string]interface{}{
				"value": "42",
			},
			key:          "value",
			defaultValue: 100,
			want:         42,
		},
		{
			name: "String value that is not a valid integer - return default with warning",
			config: map[string]interface{}{
				"value": "not-a-number",
			},
			key:          "value",
			defaultValue: 100,
			want:         100,
		},
		{
			name: "Unexpected type - return default with warning",
			config: map[string]interface{}{
				"value": []string{"not", "an", "int"},
			},
			key:          "value",
			defaultValue: 100,
			want:         100,
		},
		{
			name: "Zero value",
			config: map[string]interface{}{
				"value": 0,
			},
			key:          "value",
			defaultValue: 100,
			want:         0,
		},
		{
			name: "Negative value",
			config: map[string]interface{}{
				"value": -10,
			},
			key:          "value",
			defaultValue: 100,
			want:         -10,
		},
		{
			name: "Large value",
			config: map[string]interface{}{
				"value": 999999,
			},
			key:          "value",
			defaultValue: 100,
			want:         999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getIntValue(tt.config, tt.key, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetFloatValue(t *testing.T) {
	tests := []struct {
		name         string
		config       map[string]interface{}
		key          string
		defaultValue float64
		want         float64
	}{
		{
			name:         "Key not present - return default",
			config:       map[string]interface{}{},
			key:          "missing",
			defaultValue: 1.0,
			want:         1.0,
		},
		{
			name: "Float64 value - return as is",
			config: map[string]interface{}{
				"value": 3.14,
			},
			key:          "value",
			defaultValue: 1.0,
			want:         3.14,
		},
		{
			name: "Int value - convert to float64",
			config: map[string]interface{}{
				"value": 42,
			},
			key:          "value",
			defaultValue: 1.0,
			want:         42.0,
		},
		{
			name: "String value that is a valid float - convert",
			config: map[string]interface{}{
				"value": "3.14",
			},
			key:          "value",
			defaultValue: 1.0,
			want:         3.14,
		},
		{
			name: "String value that is not a valid float - return default with warning",
			config: map[string]interface{}{
				"value": "not-a-number",
			},
			key:          "value",
			defaultValue: 1.0,
			want:         1.0,
		},
		{
			name: "Unexpected type - return default with warning",
			config: map[string]interface{}{
				"value": []string{"not", "a", "float"},
			},
			key:          "value",
			defaultValue: 1.0,
			want:         1.0,
		},
		{
			name: "Zero value",
			config: map[string]interface{}{
				"value": 0.0,
			},
			key:          "value",
			defaultValue: 1.0,
			want:         0.0,
		},
		{
			name: "Negative value",
			config: map[string]interface{}{
				"value": -0.5,
			},
			key:          "value",
			defaultValue: 1.0,
			want:         -0.5,
		},
		{
			name: "Very small value",
			config: map[string]interface{}{
				"value": 0.0001,
			},
			key:          "value",
			defaultValue: 1.0,
			want:         0.0001,
		},
		{
			name: "Very large value",
			config: map[string]interface{}{
				"value": 999999.999,
			},
			key:          "value",
			defaultValue: 1.0,
			want:         999999.999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getFloatValue(tt.config, tt.key, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetIntValue_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		config       map[string]interface{}
		key          string
		defaultValue int
		want         int
	}{
		{
			name: "String with leading zeros",
			config: map[string]interface{}{
				"value": "042",
			},
			key:          "value",
			defaultValue: 100,
			want:         42,
		},
		{
			name: "String with whitespace",
			config: map[string]interface{}{
				"value": " 42 ",
			},
			key:          "value",
			defaultValue: 100,
			want:         100, // strconv.Atoi doesn't trim whitespace
		},
		{
			name: "Float64 with .0",
			config: map[string]interface{}{
				"value": 42.0,
			},
			key:          "value",
			defaultValue: 100,
			want:         42,
		},
		{
			name: "Float64 with .999 (should floor)",
			config: map[string]interface{}{
				"value": 42.999,
			},
			key:          "value",
			defaultValue: 100,
			want:         100, // Not exactly equal to int(42.999)
		},
		{
			name: "Scientific notation in string",
			config: map[string]interface{}{
				"value": "1e2",
			},
			key:          "value",
			defaultValue: 100,
			want:         100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getIntValue(tt.config, tt.key, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetFloatValue_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		config       map[string]interface{}
		key          string
		defaultValue float64
		want         float64
	}{
		{
			name: "String with leading zeros",
			config: map[string]interface{}{
				"value": "007.5",
			},
			key:          "value",
			defaultValue: 1.0,
			want:         7.5,
		},
		{
			name: "String with scientific notation",
			config: map[string]interface{}{
				"value": "1.5e2",
			},
			key:          "value",
			defaultValue: 1.0,
			want:         150.0,
		},
		{
			name: "Negative float string",
			config: map[string]interface{}{
				"value": "-3.14",
			},
			key:          "value",
			defaultValue: 1.0,
			want:         -3.14,
		},
		{
			name: "String with whitespace",
			config: map[string]interface{}{
				"value": " 3.14 ",
			},
			key:          "value",
			defaultValue: 1.0,
			want:         1.0, // strconv.ParseFloat doesn't trim whitespace
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getFloatValue(tt.config, tt.key, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}
