SUPPORT_KEYS: str = """\
provider                     # llm provider
file_ignore                  # file to ignore when generating a commit
{provider}.api_base          # openai base url
{provider}.api_key
{provider}.model
{provider}.retries
{provider}.proxy
{provider}.max_tokens
{provider}.top_p
{provider}.temperature
{provider}.frequency_penalty
{provider}.extra_headers     # json string
prompt.brief_commit_message  # prmpt for brief commit msg
prompt.translation           # prompt for translation commit message to target language
output.lang                  # commit message language
"""
