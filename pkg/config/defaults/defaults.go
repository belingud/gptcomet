package defaults

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
}
