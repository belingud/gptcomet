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
  - [Setup](#setup)
  - [Usage](#usage)
  - [Commands](#commands)
  - [Configuration](#configuration)
    - [file_ignore](#file_ignore)
  - [Supported Keys](#supported-keys)
  - [Example](#example)
  - [Development](#development)
  - [License](#license)
  - [Contact](#contact)

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
$ pipx install gptcomet
  installed package gptcomet 0.0.3, installed using Python 3.12.3
  These apps are now globally available
  - gmsg
  - gptcomet
done! ✨ 🌟 ✨
```

Install by `uv`

```shell
uv tool install gptcomet
```

## Setup

Before using GPTComet, you need to configure it with your OpenAI API key and other settings.
You can do this by running the following command:

```shell
gmsg config set openai.api_key YOUR_API_KEY
gmsg config set openai.model gpt-4o
```

Replace `YOUR_API_KEY` with your actual API key for the OpenAI provider.
The configuration file will be created at `~/.config/gptcomet/gptcomet.yaml`

Or you can use `gmsg newprovider` to setup a custom provider.

```shell
$ gmsg newprovider
Enter provider name (lowercase) [openai]:
Enter API Base URL:  [https://api.openai.com/v1/]: https://api.siliconflow.cn/v1
Enter model name:  [text-davinci-003]: Qwen/Qwen2.5-7B-Instruct
Enter API key: ***************************************************
Enter max tokens [1024]:
[GPTComet] Provider silicon configured successfully.
```

## Usage

To use gptcomet, follow these steps:

1.  **Install GPTComet**: Install GPTComet through pypi.
2.  **Configure GPTComet**: Configure GPTComet with your api_key The configuration file should contain the following keys:
  * `provider`: The provider of the language model (default `openai`).
  * `api_base`: The base URL of the API (default `https://api.openai.com/v1`).
  * `api_key`: The API key for the provider.
  * `model`: The model used for generating commit messages (default `text-davinci-003`).
  * `retries`: The number of retries for the API request (default `2`).
3.  **Run GPTComet**: Run GPTComet using the following command: `gmsg commit`.

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
gmsg config set openai.max_tokens 1300
```

Silicon providers the similar interface with openrouter, so you can set provider to `openai`
and set `api_base` to `https://api.siliconflow.cn/v1`.

**Note that max tokens may vary, and will return an error if it is too large.**

## Supported Keys

You can use `gmsg config keys` to check supported keys.

## Example

Here is an example of how to use GPTComet:

1.  When you first set your OpenAI KEY by `gmsg config set openai.api_key YOUR_API_KEY`, it will generate config file at `~/.local/gptcomet/gptcomet.yaml`, includes:
  ```
  provider = "openai"
  api_base = "https://api.openai.com/v1"
  api_key = "YOUR_API_KEY"
  model = "gpt-3.5-turbo"
  retries = 2
  output.lang = "en"
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

## License

GPTComet is licensed under the MIT License.

## Contact

If you have any questions or suggestions, feel free to contact.
