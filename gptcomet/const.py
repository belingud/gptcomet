DEFAULT_API_BASE = "https://api.openai.com/v1"
DEFAULT_MODEL = "gpt-3.5-turbo"
DEFAULT_RETRIES = 2

SUPPORTED_LANG = ["en", "zh"]

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
