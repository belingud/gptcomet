# GPTComet: AI-Powered Git Commit Message Generator

[![PyPI version](https://img.shields.io/pypi/v/gptcomet?style=for-the-badge)](https://pypi.org/project/gptcomet/)
![GitHub Release](https://img.shields.io/github/v/release/belingud/gptcomet?style=for-the-badge)
[![License](https://img.shields.io/github/license/belingud/gptcomet.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/belingud/gptcomet?style=for-the-badge)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/belingud/gptcomet/release.yml?style=for-the-badge)
![PyPI - Downloads](https://img.shields.io/pypi/dm/gptcomet?logo=pypi&style=for-the-badge)
![Pepy Total Downloads](https://img.shields.io/pepy/dt/gptcomet?style=for-the-badge&logo=python)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/belingud/gptcomet/total?style=for-the-badge&label=Release%20Download)

![](artwork/commit.gif)

<!-- TOC -->

- [GPTComet: AI-Powered Git Commit Message Generator](#gptcomet-ai-powered-git-commit-message-generator)
  - [Overview](#overview)
  - [Features](#features)
  - [Installation](#installation)
  - [Usage](#usage)
  - [Setup](#setup)
    - [Configuration Methods](#configuration-methods)
    - [Provider Setup Guide](#provider-setup-guide)
      - [OpenAI](#openai)
      - [Gemini](#gemini)
      - [Claude/Anthropic](#claudeanthropic)
      - [Vertex](#vertex)
      - [Azure](#azure)
      - [Ollama](#ollama)
      - [Other Supported Providers](#other-supported-providers)
    - [Manual Provider Setup](#manual-provider-setup)
  - [Commands](#commands)
  - [Configuration](#configuration)
    - [Supported Configuration Keys](#supported-configuration-keys)
    - [file\_ignore](#file_ignore)
    - [provider](#provider)
    - [output](#output)
    - [Supported languages](#supported-languages)
    - [console](#console)
  - [Supported Keys](#supported-keys)
  - [Example](#example)
  - [known Issue](#known-issue)
  - [Development](#development)
  - [Contact](#contact)
  - [License](#license)

<!-- /TOC -->
## Overview

GPTComet is a go library designed to automate the process of generating commit messages for Git repositories.
It leverages the power of AI to create meaningful commit messages based on the changes made in the codebase.

## Features

-   **Automatic Commit Message Generation**: GPTComet can generate commit messages based on the changes made in the code.
-   **Support for Multiple Languages**: GPTComet supports multiple languages, including English, Chinese and so on.
-   **Customizable Configuration**: GPTComet allows users to customize the configuration to suit their needs, such llm model and prompt.
-   **Support for Rich Commit Messages**: GPTComet supports rich commit messages, which include a title, summary, and detailed description.
-   **Support for Multiple Providers**: GPTComet supports multiple providers, including OpenAI, Gemini, Claude/Anthropic, Vertex, Azure, Ollama, and others.
-   **Support SVN and Git**: GPTComet supports both SVN and Git repositories.

## Installation

To use GPTComet, you can download from [Github release](https://github.com/belingud/gptcomet/releases/latest), or by install scripts:

```bash
curl -sSL https://cdn.jsdelivr.net/gh/belingud/gptcomet@master/install.sh | bash
```

Windows:

```powershell
irm https://cdn.jsdelivr.net/gh/belingud/gptcomet@master/install.ps1 | iex
```

If you prefer to run in python, you can install by `pip` directly, it packaged the binary files corresponding to the platform already.

```shell
pip install gptcomet
```

Install use `pipx` or `uv` on Mac, Linux or Windows.

```shell
pipx install gptcomet
```

Install by `uv`

```shell
uv tool install gptcomet
Resolved 1 package in 1.33s
Installed 1 package in 8ms
 + gptcomet==0.1.6
Installed 2 executables: gmsg, gptcomet
```

## Usage

To use gptcomet, follow these steps:

1.  **Install GPTComet**: Install GPTComet through pypi.
2.  **Configure GPTComet**: See [Setup](#setup). Configure GPTComet with your api_key and other required keys like:

-   `provider`: The provider of the language model (default `openai`).
-   `api_base`: The base URL of the API (default `https://api.openai.com/v1`).
-   `api_key`: The API key for the provider.
-   `model`: The model used for generating commit messages (default `text-davinci-003`).

3.  **Run GPTComet**: Run GPTComet using the following command: `gmsg commit`.

If you are using `openai` provider, and finished set `api_key`, you can run `gmsg commit` directly.

## Setup

### Configuration Methods

1. **Direct Configuration**

    - Configure directly in `~/.config/gptcomet/gptcomet.yaml`.

2. **Interactive Setup**
    - Use the `gmsg newprovider` command for guided setup.

### Provider Setup Guide

![](artwork/newprovider.gif)

```bash
gmsg newprovider

    Select Provider

  > 1. azure
    2. chatglm
    3. claude
    4. cohere
    5. deepseek
    6. gemini
    7. kimi
    8. mistral
    9. ollama
    10. openai
    11. sambanova
    12. silicon
    13. tongyi
    14. vertex
    15. xai
    16. Input Manually

    ↑/k up • ↓/j down • ? more
```

#### OpenAI

OpenAI api key page: https://platform.openai.com/api-keys

```shell
gmsg newprovider

Selected provider: openai
Configure provider:

Previous inputs:
  Enter OpenAI API base: https://api.openai.com/v1
  Enter API key: sk-abc*********************************************
  Enter max tokens: 1024

Enter Enter model name (default: gpt-4o):
> gpt-4o


Provider openai configured successfully!
```

#### Gemini

Gemini api key page: https://aistudio.google.com/u/1/apikey

```shell
gmsg newprovider
Selected provider: gemini
Configure provider:

Previous inputs:
  Enter Gemini API base: https://generativelanguage.googleapis.com/v1beta/models
  Enter API key: AIz************************************
  Enter max tokens: 1024

Enter Enter model name (default: gemini-1.5-flash):
> gemini-2.0-flash-exp

Provider gemini already has a configuration. Do you want to overwrite it? (y/N): y

Provider gemini configured successfully!
```

#### Claude/Anthropic

I don't have an anthropic account yet, please see [Anthropic console](https://console.anthropic.com)

#### Vertex

Vertex console page: https://console.cloud.google.com

```shell
gmsg newprovider
Selected provider: vertex
Configure provider:

Previous inputs:
  Enter Vertex AI API Base URL: https://us-central1-aiplatform.googleapis.com/v1
  Enter API key: sk-awz*********************************************
  Enter location (e.g., us-central1): us-central1
  Enter max tokens: 1024
  Enter model name: gemini-1.5-pro

Enter Enter Google Cloud project ID:
> test-project


Provider vertex configured successfully!
```

#### Azure

```shell
gmsg newprovider

Selected provider: azure
Configure provider:

Previous inputs:
  Enter Azure OpenAI endpoint: https://gptcomet.openai.azure.com
  Enter API key: ********************************
  Enter API version: 2024-02-15-preview
  Enter Azure OpenAI deployment name: gpt4o
  Enter max tokens: 1024

Enter Enter deployment name (default: gpt-4o):
> gpt-4o


Provider azure configured successfully!
```

#### Ollama

```shell
gmsg newprovider
Selected provider: ollama
Configure provider:

Previous inputs:
  Enter Ollama API Base URL: http://localhost:11434/api
  Enter max tokens: 1024

Enter Enter model name (default: llama2):
> llama2


Provider ollama configured successfully!
```

#### Other Supported Providers

-   Groq
-   Mistral
-   Tongyi/Qwen
-   XAI
-   Sambanova
-   Silicon
-   Deepseek
-   ChatGLM

### Manual Provider Setup

Or you can enter the provider name manually, and setup config manually.

```shell
gmsg newprovider
You can either select one from the list or enter a custom provider name.
  ...
  vertex
> Input manually

Enter provider name: test
Enter OpenAI API Base URL [https://api.openai.com/v1]:
Enter model name [gpt-4o]:
Enter API key: ************************************
Enter max tokens [1024]:
[GPTComet] Provider test configured successfully.
```

## Commands

The following are the available commands for GPTComet:

-   `gmsg config`: Config manage commands group.
    -   `get <key>`: Get the value of a configuration key.
    -   `list`: List the entire configuration content.
    -   `reset`: Reset the configuration to default values (optionally reset only the prompt section with `--prompt`).
    -   `set <key> <value>`: Set a configuration value.
    -   `path`: Get the configuration file path.
    -   `remove <key> [value]`: Remove a configuration key or a value from a list. (List value only, like `fileignore`)
    -   `append <key> <value>`: Append a value to a list configuration.(List value only, like `fileignore`)
    -   `keys`: List all supported configuration keys.
-   `gmsg commit`: Generate commit message by changes/diff.
    -   `--svn`: Generate commit message for svn.
    -   `--dry-run`: Dry run the command without actually generating the commit message.
    -   `-y/--yes`: Skip the confirmation prompt.
-   `gmsg newprovider`: Add a new provider.

Global flags:

```shell
  -c, --config string   Config file path
  -d, --debug           Enable debug mode
```

## Configuration

The configuration file for GPTComet is `gptcomet.yaml`. The file should contain the following keys:

-   `file_ignore`: The file to ignore when generating a commit.
-   `provider`: The provider of the language model (default `openai`).
    -   `api_base`: The base URL of the API (default `https://api.openai.com/v1`).
    -   `api_key`: The API key for the provider.
    -   `model`: The model used for generating commit messages (default `text-davinci-003`).
    -   `retries`: The number of retries for the API request (default `2`).
    -   `proxy`: The proxy URL for the provider.
    -   `max_tokens`: The maximum number of tokens for the provider.
    -   `top_p`: The top_p parameter for the provider (default `0.7`).
    -   `temperature`: The temperature parameter for the provider (default `0.7`).
    -   `frequency_penalty`: The frequency_penalty parameter for the provider (default `0`).
    -   `extra_headers`: The extra headers for the provider, json string.
    -   `answer_path`: The json path for the answer. Default `choices[0].message.content`
    -   `completion_path`: The url path for the completion api. Default `/chat/completions`
-   `prompt`: The prompt for generating commit messages.
    -   `brief_commit_message`: The prompt for generating brief commit messages.
    -   `rich_commit_message`: The prompt for generating rich commit messages.
    -   `translation`: The prompt for translating commit messages to a target language.
-   `output`: The output configuration.
    -   `output.lang`: The language of the commit message (default `en`).
    -   `output.rich_template`: The template for generating rich commit messages.
    -   `output.translate_title`: Whether to translate the title of the commit message.
-   `console`: The console configuration.
    -   `console.verbose`: Whether to enable verbose output.

`output.translate_title` is used to determine whether to translate the title of the commit message.

For example in `output.lang: zh-cn`, the title of the commit message is `feat: Add new feature`

If `output.translate_title` is set to `true`, the commit message will be translated to `功能：新增功能`. otherwise, the commit message will be translated to `feat: 新增功能`:

### Supported Configuration Keys

Here's a summary of the main configuration keys:

| Key                            | Description                                                | Default Value                     |
| :----------------------------- | :--------------------------------------------------------- | :-------------------------------- |
| `provider`                     | The name of the LLM provider to use.                       | `openai`                          |
| `file_ignore`                  | A list of file patterns to ignore in the diff.             | (See [file_ignore](#file_ignore)) |
| `output.lang`                  | The language for commit message generation.                | `en`                              |
| `output.rich_template`         | The template to use for rich commit messages.              | `<title>:<summary>\n\n<detail>`   |
| `console.verbose`              | Enable verbose output.                                     | `true`                            |
| `output.translate_title`       | Translate the title of the commit message.                 | `false`                           |
| `<provider>.api_base`          | The API base URL for the provider.                         | (Provider-specific)               |
| `<provider>.api_key`           | The API key for the provider.                              |                                   |
| `<provider>.model`             | The model name to use.                                     | (Provider-specific)               |
| `<provider>.retries`           | The number of retry attempts for API requests.             | `2`                               |
| `<provider>.proxy`             | The proxy URL to use (if needed).                          |                                   |
| `<provider>.max_tokens`        | The maximum number of tokens to generate.                  | `2048`                            |
| `<provider>.top_p`             | The top-p value for nucleus sampling.                      | `0.7`                             |
| `<provider>.temperature`       | The temperature value for controlling randomness.          | `0.7`                             |
| `<provider>.frequency_penalty` | The frequency penalty value.                               | `0`                               |
| `<provider>.extra_headers`     | Extra headers to include in API requests (JSON string).    | `{}`                              |
| `<provider>.completion_path`   | The API path for completion requests.                      | (Provider-specific)               |
| `<provider>.answer_path`       | The JSON path to extract the answer from the API response. | (Provider-specific)               |
| `prompt.brief_commit_message`  | The prompt template for generating brief commit messages.  | (See `defaults/defaults.go`)      |
| `prompt.rich_commit_message`   | The prompt template for generating rich commit messages.   | (See `defaults/defaults.go`)      |
| `prompt.translation`           | The prompt template for translating commit messages.       | (See `defaults/defaults.go`)      |

**Note:** `<provider>` should be replaced with the actual provider name (e.g., `openai`, `gemini`, `claude`).

Some providers require specific keys, such as Vertex needing project ID, location, etc.

### file_ignore

The file to ignore when generating a commit. The default value is

```yaml
- bun.lockb
- Cargo.lock
- composer.lock
- Gemfile.lock
- package-lock.json
- pnpm-lock.yaml
- poetry.lock
- yarn.lock
- pdm.lock
- Pipfile.lock
- "*.py[cod]"
- go.sum
- uv.lock
```

You can add more file_ignore by using the `gmsg config append file_ignore <xxx>` command.
`<xxx>` is same syntax as `gitignore`, like `*.so` to ignore all `.so` suffix files.

### provider

The provider configuration of the language model.

The default provider is `openai`.

Provider config just like:

```yaml
provider: openai
openai:
    api_base: https://api.openai.com/v1
    api_key: YOUR_API_KEY
    model: gpt-4o
    retries: 2
    max_tokens: 1024
    temperature: 0.7
    top_p: 0.7
    frequency_penalty: 0
    extra_headers: {}
    answer_path: choices.0.message.content
    completion_path: /chat/completions
```

If you are using `openai`, just leave the `api_base` as default. Set your `api_key` in the `config` section.

If you are using an `openai` class provider, or a provider compatible interface, you can set the provider to `openai`.
And set your custom `api_base`, `api_key` and `model`.

For example:

`Openrouter` providers api interface compatible with openai,
you can set provider to `openai` and set `api_base` to `https://openrouter.ai/api/v1`,
`api_key` to your api key from [keys page](https://openrouter.ai/settings/keys)
and `model` to `meta-llama/llama-3.1-8b-instruct:free` or some other you prefer.

```shell
gmsg config set openai.api_base https://openrouter.ai/api/v1
gmsg config set openai.api_key YOUR_API_KEY
gmsg config set openai.model meta-llama/llama-3.1-8b-instruct:free
gmsg config set openai.max_tokens 1024
```

Silicon providers the similar interface with openrouter, so you can set provider to `openai`
and set `api_base` to `https://api.siliconflow.cn/v1`.

**Note that max tokens may vary, and will return an error if it is too large.**

### output

The output configuration of the commit message.

The default output is

```yaml
output:
    lang: en
    rich_template: "<title>:<summary>\n\n<detail>"
```

You can set `rich_template` to change the template of the rich commit message,
and set `lang` to change the language of the commit message.

### Supported languages

-   `en`: English
-   `zh-cn`: Simplified Chinese
-   `zh-tw`: Traditional Chinese
-   `fr`: French
-   `vi`: Vietnamese
-   `ja`: Japanese
-   `ko`: Korean
-   `ru`: Russian
-   `tr`: Turkish
-   `id`: Indonesian
-   `th`: Thai
-   `de`: German
-   `es`: Spanish
-   `pt`: Portuguese
-   `it`: Italian
-   `ar`: Arabic
-   `hi`: Hindi
-   `el`: Greek
-   `pl`: Polish
-   `nl`: Dutch
-   `sv`: Swedish
-   `fi`: Finnish
-   `hu`: Hungarian
-   `cs`: Czech
-   `ro`: Romanian
-   `bg`: Bulgarian
-   `uk`: Ukrainian
-   `he`: Hebrew
-   `lt`: Lithuanian
-   `la`: Latin
-   `ca`: Catalan
-   `sr`: Serbian
-   `sl`: Slovenian
-   `mk`: Macedonian
-   `lv`: Latvian

### console

The console output config.

The default console is

```yaml
console:
    verbose: true
```

When `verbose` is true, more information will be printed in the console.

## Supported Keys

You can use `gmsg config keys` to check supported keys.

## Example

Here is an example of how to use GPTComet:

1.  When you first set your OpenAI KEY by `gmsg config set openai.api_key YOUR_API_KEY`, it will generate config file at `~/.local/gptcomet/gptcomet.yaml`, includes:

```
provider: "openai"
api_base: "https://api.openai.com/v1"
api_key: "YOUR_API_KEY"
model: "gpt-4o"
retries: 2
output:
  lang: "en"
```

2.  Run the following command to generate a commit message: `gmsg commit`
3.  GPTComet will generate a commit message based on the changes made in the code and display it in the console.

Note: Replace `YOUR_API_KEY` with your actual API key for the provider.

## known Issue

- Groq with proxy not working as expected, cloudflare blocks the request. Not sure why yet.

## Development

If you'd like to contribute to GPTComet, feel free to fork this project and submit a pull request.

First, fork the project and clone your repo.

```shell
git clone https://github.com/<yourname>/gptcomet
```

Second, make sure you have `pdm`, you can install by `pip`, `brew` or other way in their [installation](https://github.com/pdm-project/pdm?tab=readme-ov-file#installation) docs

Use `just` command install dependence, `just` is a handy way to save and run project-specific commands, `just` docs [https://github.com/casey/just](https://github.com/casey/just)

```shell
just install
```

Or use `pdm` directly `pdm install`.

Then, you can submit a pull request.

## Contact

If you have any questions or suggestions, feel free to contact.

## License

GPTComet is licensed under the MIT License.

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fbelingud%2Fgptcomet.svg?type=large&issueType=license)](https://app.fossa.com/projects/git%2Bgithub.com%2Fbelingud%2Fgptcomet?ref=badge_large&issueType=license)
