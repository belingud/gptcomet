package defaults

import (
	"github.com/charmbracelet/glamour/styles"
)

const (
	DefaultAPIBase          = "https://api.openai.com/v1"
	DefaultModel            = "gpt-4o"
	DefaultRetries          = 3
	DefaultMaxTokens        = 1024
	DefaultTemperature      = 0.3
	DefaultTopP             = 1.0
	DefaultFrequencyPenalty = 0.0
)

// defaultConfig returns a default configuration map for gptcomet.
//
// The configuration map contains the default values for the provider, file
// ignore, output, console, openai, and claude configuration options.
//
// The default values are as follows:
//
//   - provider: "openai"
//   - file_ignore: the default list of file patterns to ignore when generating
//     commit messages
//   - output:
//   - lang: "en"
//   - review_lang: "en"
//   - rich_template: "<title>:<summary>\n\n<detail>"
//   - translate_title: false
//   - markdown_theme: the default markdown theme for the output
//   - console:
//   - verbose: true
//   - openai:
//   - api_base: the default API base for the OpenAI provider
//   - api_key: an empty string (must be set by the user)
//   - model: the default model for the OpenAI provider
//   - retries: 2
//   - proxy: an empty string (must be set by the user)
//   - max_tokens: 1024
//   - top_p: 0.7
//   - temperature: 0.7
//   - frequency_penalty: 0
//   - extra_headers: an empty string (must be set by the user)
//   - completion_path: "/chat/completions"
//   - answer_path: "choices.0.message.content"
//   - claude:
//   - api_base: "https://api.anthropic.com"
//   - api_key: an empty string (must be set by the user)
//   - model: "claude-3.5-sonnet"
//   - retries: 2
//   - proxy: an empty string (must be set by the user)
//   - max_tokens: 1024
//   - top_p: 0.7
//   - temperature: 0.7
//   - frequency_penalty: 0
//   - extra_headers: an empty string (must be set by the user)
//   - completion_path: "/v1/messages"
//   - answer_path: "content.0.text"
//   - prompt: the default prompt templates
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
			"go.sum",
			"uv.lock",
		},
		"output": map[string]interface{}{
			"lang":            "en",
			"rich_template":   "<title>:<summary>\n\n<detail>",
			"translate_title": false,
			"review_lang":     "en",
			"markdown_theme":  styles.AutoStyle,
		},
		"console": map[string]interface{}{
			"verbose": true,
		},
		"openai": map[string]interface{}{
			"api_base":          DefaultAPIBase,
			"api_key":           "",
			"model":             DefaultModel,
			"retries":           2,
			"proxy":             "",
			"max_tokens":        1024,
			"top_p":             0.7,
			"temperature":       0.7,
			"frequency_penalty": 0,
			"extra_headers":     "{}",
			"extra_body":        "{}",
			"completion_path":   "/chat/completions",
			"answer_path":       "choices.0.message.content",
		},
		"claude": map[string]interface{}{
			"api_base":          "https://api.anthropic.com",
			"api_key":           "",
			"model":             "claude-3.5-sonnet",
			"retries":           2,
			"proxy":             "",
			"max_tokens":        1024,
			"top_p":             0.7,
			"temperature":       0.7,
			"frequency_penalty": 0,
			"extra_headers":     "{}",
			"extra_body":        "{}",
			"completion_path":   "/v1/messages",
			"answer_path":       "content.0.text",
		},
		"prompt": PromptDefaults,
	}
}

// PromptDefaults contains default prompt configurations
var PromptDefaults = map[string]string{
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

The commit message template is <title>: <summary>. Your answer should only include a single commit message less than 70 characters, no other text or ` + "`" + `.
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
	"review": `Please review the following code patch and provide feedback in {{ output.review_lang }}.:  
Requirements:  
1. Identify and list necessary improvements (e.g., bug risks, security vulnerabilities).  
2. Suggest optional improvements (e.g., code readability, maintainability).  
Clearly separate necessary improvements from optional suggestions. Only include points relevant to the provided diff.  

THE CODE PATCH TO BE REVIEWED:
{{ placeholder }}`,
}

var DefaultConfig = defaultConfig()
