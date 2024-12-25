# GPTComet: AI-Powered Git Commit Message Generator

[![PyPI version](https://img.shields.io/pypi/v/gptcomet?style=for-the-badge)](https://pypi.org/project/gptcomet/)
[![License](https://img.shields.io/github/license/belingud/gptcomet.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)
![Static Badge](https://img.shields.io/badge/language-Python-%233572A5?style=for-the-badge)
![PyPI - Downloads](https://img.shields.io/pypi/dm/gptcomet?logo=pypi&style=for-the-badge)
![Pepy Total Downloads](https://img.shields.io/pepy/dt/gptcomet?style=for-the-badge&logo=python)

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
    - [file\_ignore](#file_ignore)
    - [provider](#provider)
    - [output](#output)
    - [console](#console)
  - [Supported Keys](#supported-keys)
  - [Example](#example)
  - [Development](#development)
  - [Contact](#contact)
  - [License](#license)

<!-- /TOC -->

## Overview

GPTComet is a Python library designed to automate the process of generating commit messages for Git repositories.
It leverages the power of AI to create meaningful commit messages based on the changes made in the codebase.

## Features

* **Automatic Commit Message Generation**: GPTComet can generate commit messages based on the changes made in the code.
* **Support for Multiple Languages**: GPTComet supports multiple languages, including English, Chinese and so on.
* **Customizable Configuration**: GPTComet allows users to customize the configuration to suit their needs, such llm model and prompt.
* **Support for Rich Commit Messages**: GPTComet supports rich commit messages, which include a title, summary, and detailed description.

## Installation

To use GPTComet, you need to have Python installed on your system. You can install the library using pip:

```shell
pip install gptcomet
```

Install use `pipx` on Mac or Linux.

```shell
pipx install gptcomet
```
After installing GPTComet, you will have two commands: `gptcomet` and `gmsg`.

```shell
pipx install gptcomet
  installed package gptcomet 0.1.4, installed using Python 3.12.3
  These apps are now globally available
    - gmsg
    - gptcomet
done! ✨ 🌟 ✨
```

Install by `uv`

```shell
uv tool install gptcomet --index https://pypi.org/simple
Resolved 25 packages in 9ms
   Built ruamel-yaml-clib==0.2.12
Prepared 18 packages in 12.53s
Installed 25 packages in 28ms
 + attrs==24.3.0
 + boltons==24.1.0
 + certifi==2024.12.14
 + charset-normalizer==3.4.0
 + click==8.1.8
 + face==24.0.0
 + gitdb==4.0.11
 + gitpython==3.1.43
 + glom==24.11.0
 + gptcomet==0.1.4
 + idna==3.10
 + markdown-it-py==3.0.0
 + mdurl==0.1.2
 + prompt-toolkit==3.0.48
 + pygments==2.18.0
 + requests==2.32.3
 + rich==13.9.4
 + ruamel-yaml==0.18.6
 + ruamel-yaml-clib==0.2.12
 + shellingham==1.5.4
 + smmap==5.0.1
 + typer==0.15.1
 + typing-extensions==4.12.2
 + urllib3==2.3.0
 + wcwidth==0.2.13
Installed 2 executables: gmsg, gptcomet
```

## Usage

To use gptcomet, follow these steps:

1.  **Install GPTComet**: Install GPTComet through pypi.
2.  **Configure GPTComet**: See [Setup](#setup). Configure GPTComet with your api_key and other required keys like:
  * `provider`: The provider of the language model (default `openai`).
  * `api_base`: The base URL of the API (default `https://api.openai.com/v1`).
  * `api_key`: The API key for the provider.
  * `model`: The model used for generating commit messages (default `text-davinci-003`).
3.  **Run GPTComet**: Run GPTComet using the following command: `gmsg commit`.

If you are using `openai` provider, and finished set `api_key`, you can run `gmsg commit` directly.

## Setup

### Configuration Methods

1. **Direct Configuration**
   - Configure directly in `~/.config/gptcomet/gptcomet.yaml`.

2. **Interactive Setup**
   - Use the `gmsg newprovider` command for guided setup.

### Provider Setup Guide

#### OpenAI

OpenAI api key page: https://platform.openai.com/api-keys

```shell
gmsg newprovider
You can either select one from the list or enter a custom provider name.
  ...
  sambanova
> openai
  vertex
  Input manually

Enter OpenAI API Base URL [https://api.openai.com/v1]:
Enter model name [gpt-4o]:
Enter API key: ********************************************************************************************************************************************************************
Enter max tokens [1024]:
[GPTComet] Provider openai configured successfully.
```

#### Gemini

Gemini api key page: https://aistudio.google.com/u/1/apikey

```shell
gmsg newprovider
You can either select one from the list or enter a custom provider name.
> gemini
  ...
  Input manually

Enter Gemini API base [https://generativelanguage.googleapis.com/v1beta/models]:
Enter Gemini model [gemini-pro]: gemini-2.0-flash-exp
Enter Gemini API key: ***************************************
Enter max tokens [1024]: 500
[GPTComet] Provider gemini configured successfully.
```

#### Claude/Anthropic

I don't have an anthropic account yet, please see [Anthropic console](https://console.anthropic.com)

```shell
gmsg newprovider
You can either select one from the list or enter a custom provider name.
  gemini
  cohere
> claude
  ...
  Input manually

Enter Anthropic API Base URL [https://api.anthropic.com/v1]:
Enter model name [claude-3-5-sonnet]:
Enter API key: ***************************************
Enter max tokens [1024]:
Enter Anthropic API version [2023-06-01]:
[GPTComet] Provider claude configured successfully.
```

#### Vertex

Vertex console page: https://console.cloud.google.com

```shell
gmsg newprovider
You can either select one from the list or enter a custom provider name.
  ...
  openai
> vertex
  Input manually

Enter Vertex AI API Base URL [https://us-central1-aiplatform.googleapis.com/v1]:
Enter model name [gemini-pro]:
Enter API key: ********************************
Enter Google Cloud project ID []: test-project-id
Enter location (e.g., us-central1) [us-central1]:
Enter max tokens [1024]:
[GPTComet] Provider vertex configured successfully.
```

#### Azure

```shell
gmsg newprovider
You can either select one from the list or enter a custom provider name.
  ...
  mistral
> azure
  ...
  Input manually

Enter Azure OpenAI endpoint URL (e.g., https://YOUR_RESOURCE_NAME.openai.azure.com) []: https://gptcomet.openai.azure.com
Enter Azure OpenAI deployment name []: gpt4o
Enter API version [2024-02-15-preview]:
Enter model name [gpt-4o]:
Enter API key: ********************************
Enter max tokens [1024]:
[GPTComet] Provider azure configured successfully.
```

#### Ollama

```shell
gmsg newprovider
You can either select one from the list or enter a custom provider name.
  azure
> ollama
  ...
  Input manually

Enter Ollama API Base URL [http://localhost:11434/api]:
Enter model name [llama2]:
Enter max tokens [1024]:
[GPTComet] Provider ollama configured successfully.
```

#### Other Supported Providers

- Groq
- Mistral
- Tongyi/Qwen
- XAI
- Sambanova
- Silicon
- Deepseek
- ChatGLM

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

* `gmsg config`: Config manage commands group.
  * `set`: Set a configuration value.
  * `get`: Get a configuration value.
  * `list`: List all configuration values.
  * `reset`: Reset the configuration to its default values.
  * `keys`: List all supported keys.
  * `append`: Append a value to a configuration key. (List value only, like `fileignore`)
  * `remove`: Remove a value from a configuration key. (List value only, like `fileignore`)
* `gmsg commit`: Generate commit message by changes/diff.
* `gmsg newprovider`: Add a new provider.


## Configuration

The configuration file for GPTComet is `gptcomet.yaml`. The file should contain the following keys:

* `file_ignore`: The file to ignore when generating a commit.
* `provider`: The provider of the language model (default `openai`).
  * `api_base`: The base URL of the API (default `https://api.openai.com/v1`).
  * `api_key`: The API key for the provider.
  * `model`: The model used for generating commit messages (default `text-davinci-003`).
  * `retries`: The number of retries for the API request (default `2`).
  * `proxy`: The proxy URL for the provider.
  * `max_tokens`: The maximum number of tokens for the provider.
  * `top_p`: The top_p parameter for the provider (default `0.7`).
  * `temperature`: The temperature parameter for the provider (default `0.7`).
  * `frequency_penalty`: The frequency_penalty parameter for the provider (default `0`).
  * `extra_headers`: The extra headers for the provider, json string.
  * `answer_path`: The json path for the answer. Default `choices[0].message.content`
  * `completion_path`: The url path for the completion api. Default `/chat/completions`
* `prompt`: The prompt for generating commit messages.
  * `brief_commit_message`: The prompt for generating brief commit messages.
  * `rich_commit_message`: The prompt for generating rich commit messages.
  * `translation`: The prompt for translating commit messages to a target language.
* `output`: The output configuration.
  * `output.lang`: The language of the commit message (default `en`).
  * `output.rich_template`: The template for generating rich commit messages.

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
- '*.py[cod]'
- go.mod
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

* `en`: English
* `zh-cn`: Simplified Chinese
* `zh-tw`: Traditional Chinese
* `fr`: French
* `vi`: Vietnamese
* `ja`: Japanese
* `ko`: Korean
* `ru`: Russian
* `tr`: Turkish
* `id`: Indonesian
* `th`: Thai
* `de`: German
* `es`: Spanish
* `pt`: Portuguese
* `it`: Italian
* `ar`: Arabic
* `hi`: Hindi
* `el`: Greek
* `pl`: Polish
* `nl`: Dutch
* `sv`: Swedish
* `fi`: Finnish
* `hu`: Hungarian
* `cs`: Czech
* `ro`: Romanian
* `bg`: Bulgarian
* `uk`: Ukrainian
* `he`: Hebrew
* `lt`: Lithuanian
* `la`: Latin
* `ca`: Catalan
* `sr`: Serbian
* `sl`: Slovenian
* `mk`: Macedonian
* `lv`: Latvian

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
