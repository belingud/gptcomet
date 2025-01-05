package defaults

import (
	"strings"
	"testing"
)

func TestPromptDefaults(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		contains []string // key phrases that should be in the prompt
	}{
		{
			name: "brief commit message prompt",
			key:  "brief_commit_message",
			contains: []string{
				"software engineer",
				"commit message",
				"Guidelines",
				"build:",
				"feat:",
				"fix:",
			},
		},
		{
			name: "rich commit message prompt",
			key:  "rich_commit_message",
			contains: []string{
				"software engineer",
				"commit message",
				"Guidelines",
				"build:",
				"feat:",
				"fix:",
				"{{ output.rich_template }}",
			},
		},
		{
			name: "translation prompt",
			key:  "translation",
			contains: []string{
				"polyglot programmer",
				"translator",
				"git commit message",
				"{{ output.lang }}",
				"{{ placeholder }}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, ok := PromptDefaults[tt.key]
			if !ok {
				t.Errorf("prompt %s not found", tt.key)
				return
			}

			for _, phrase := range tt.contains {
				if !strings.Contains(prompt, phrase) {
					t.Errorf("prompt %s does not contain expected phrase: %s", tt.key, phrase)
				}
			}
		})
	}
}
