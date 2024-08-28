SUPPORT_KEYS: str = """\
provider                      # LLM provider
file_ignore                   # File to ignore when generating a commit
{provider}.api_base           # GPT base URL, default openai api
{provider}.api_key
{provider}.model
{provider}.retries
{provider}.proxy
{provider}.max_tokens
{provider}.top_p
{provider}.temperature
{provider}.frequency_penalty
{provider}.extra_headers      # JSON string
prompt.brief_commit_message   # Prompt for brief commit message
prompt.translation            # Prompt for translation commit message to target language
output.lang                   # Commit message language
"""
