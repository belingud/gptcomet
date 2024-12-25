GPTCOMET_PRE = "[GPTComet]"

DEFAULT_API_BASE = "https://api.openai.com/v1"
DEFAULT_MODEL = "gpt-4o"
DEFAULT_RETRIES = 2

# GPTComet config keys
LANGUAGE_KEY = "output.lang"
PROVIDER_KEY = "provider"
FILE_IGNORE_KEY = "file_ignore"
CONSOLE_VERBOSE_KEY = "console.verbose"

# git output
COMMIT_OUTPUT_TEMPLATE = """
Auther: {author} <{email}>
{branch}({commit_hash})

{commit_msg}

 {git_show_stat}
"""
