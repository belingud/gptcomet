SUPPORT_KEYS: str = """\
provider                      # LLM provider
file_ignore                   # File to ignore when generating a commit
output.lang                   # Commit message language
output.rich_template
console.verbose
{provider}.api_base           # GPT base URL, default openai api
{provider}.api_key
{provider}.model
{provider}.retries
{provider}.proxy
{provider}.max_tokens
{provider}.top_p
{provider}.temperature
{provider}.frequency_penalty
{provider}.extra_headers      # JSON string, default `{}`
{provider}.answer_path        # completion response path, default `choices[0].message.content`
{provider}.completion_path    # completion api path, default `/chat/completions`
prompt.brief_commit_message   # Prompt for brief commit message
prompt.rich_commit_message    # Prompt for rich commit message
prompt.translation            # Prompt for translation commit message to target language
"""
