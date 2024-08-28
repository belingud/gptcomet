GPTCOMET_PRE = "[GPTComet]"

DEFAULT_API_BASE = "https://api.openai.com/v1"
DEFAULT_MODEL = "text-davinci-003"
DEFAULT_RETRIES = 2

# GPTComet config keys
LANGUAGE_KEY = "output.lang"
PROVIDER_KEY = "provider"
FILE_IGNORE_KEY = "file_ignore"
CONSOLE_VERBOSE_KEY = "console.verbose"


SHORT_PREPARE_COMMIT_MSG = """\
#!/bin/sh
# GPTComet pre-commit hook

# Run GPTComet commit
gptcomet generate commit

# If GPTComet commit was successful, exit with 0
if [ $? -eq 0 ]; then
    exit 0
else
    echo "GPTComet commit failed. Please review your changes and try again."
    exit 1
fi
"""

RICH_PREPARE_COMMIT_MSG = """\
#!/bin/sh
# GPTComet pre-commit hook

# Run GPTComet commit
gptcomet generate commit --rich

# If GPTComet commit was successful, exit with 0
if [ $? -eq 0 ]; then
    exit 0
else
    echo "GPTComet commit failed. Please review your changes and try again."
    exit 1
fi
"""
