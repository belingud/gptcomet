// Package prompt renders lightweight {{ key }} placeholders used by the
// provider prompt templates.
package prompt

import "regexp"

var placeholderPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_.]+)\s*\}\}`)

// Render replaces known {{ key }} placeholders with values from vars.
//
// Placeholders that are not present in vars are left untouched so custom
// templates remain readable and can be diagnosed by the caller.
func Render(template string, vars map[string]string) string {
	if template == "" || len(vars) == 0 {
		return template
	}

	return placeholderPattern.ReplaceAllStringFunc(template, func(match string) string {
		parts := placeholderPattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		if value, ok := vars[parts[1]]; ok {
			return value
		}
		return match
	})
}

// HasPlaceholder reports whether template contains at least one {{ key }}
// placeholder.
func HasPlaceholder(template string) bool {
	return placeholderPattern.MatchString(template)
}
