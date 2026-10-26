package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]string
		want     string
	}{
		{
			name:     "known placeholder",
			template: "hello {{ name }}",
			vars:     map[string]string{"name": "world"},
			want:     "hello world",
		},
		{
			name:     "multiple placeholders",
			template: "{{ greeting }}, {{ name }}!",
			vars:     map[string]string{"greeting": "hi", "name": "gptcomet"},
			want:     "hi, gptcomet!",
		},
		{
			name:     "no whitespace",
			template: "{{name}}",
			vars:     map[string]string{"name": "value"},
			want:     "value",
		},
		{
			name:     "unknown placeholder is kept",
			template: "hello {{ unknown }}",
			vars:     map[string]string{"name": "world"},
			want:     "hello {{ unknown }}",
		},
		{
			name:     "empty template",
			template: "",
			vars:     map[string]string{"name": "world"},
			want:     "",
		},
		{
			name:     "nil vars",
			template: "hello {{ name }}",
			vars:     nil,
			want:     "hello {{ name }}",
		},
		{
			name:     "value containing braces",
			template: "{{ code }}",
			vars:     map[string]string{"code": "{{ not-a-key }}"},
			want:     "{{ not-a-key }}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Render(tt.template, tt.vars))
		})
	}
}

func TestHasPlaceholder(t *testing.T) {
	assert.True(t, HasPlaceholder("hello {{ name }}"))
	assert.True(t, HasPlaceholder("{{name}}"))
	assert.False(t, HasPlaceholder("hello world"))
	assert.False(t, HasPlaceholder(""))
	assert.False(t, HasPlaceholder("{{ invalid-key }}"))
}
