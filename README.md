# GPTComet: AI-Powered Git Commit Message Generator

## Table of content

<!-- TOC -->
- [GPTComet: AI-Powered Git Commit Message Generator](#gptcomet-ai-powered-git-commit-message-generator)
  - [Table of content](#table-of-content)
  - [Overview](#overview)
  - [Features](#features)
  - [Installation](#installation)
  - [Usage](#usage)
  - [Commands](#commands)
  - [Configuration](#configuration)
  - [Supported Keys](#supported-keys)
  - [Example](#example)
  - [Development](#development)
  - [License](#license)
  - [Contact](#contact)
<!-- TOC -->

## Overview

GPTComet is a Python library designed to automate the process of generating commit messages for Git repositories.
It leverages the power of AI to create meaningful commit messages based on the changes made in the codebase.

## Features

*   **Automatic Commit Message Generation**: GPTComet can generate commit messages based on the changes made in the code.
*   **Support for Multiple Languages**: GPTComet supports multiple languages, including English and Chinese.
*   **Customizable Configuration**: GPTComet allows users to customize the configuration to suit their needs.
*   **Support for Rich Commit Messages**: GPTComet supports rich commit messages, which include a title, summary, and detailed description.

## Installation

To use GPTComet, you need to have Python installed on your system. You can install the library using pip:

```shell
pip install gptcomet
```

Recommend install use `pipx` on Mac or Linux.

```shell
pipx install gptcomet
```


## Usage

To use gptcomet, follow these steps:

1.  **Install GPTComet**: Install GPTComet through pypi.
2.  **Configure GPTComet**: Configure GPTComet with your api_key The configuration file should contain the following keys:
    *   `provider`: The provider of the language model (e.g., `openai`).
    *   `api_base`: The base URL of the API (e.g., `https://api.openai.com/v1`).
    *   `api_key`: The API key for the provider.
    *   `model`: The model used for generating commit messages (e.g., `gpt-3.5-turbo`).
    *   `retries`: The number of retries for the API request (e.g., `2`).
3.  **Run GPTComet**: Run GPTComet using the following command: `gptcomet generate commit`.

## Commands

The following are the available commands for GPTComet:

* `gptcomet config`: Config manage commands group.
  * `set`: Set a configuration value.
  * `get`: Get a configuration value.
  * `list`: List all configuration values.
  * `reset`: Reset the configuration to its default values.
  * `keys`: List all supported keys.
* `gptcomet hook`: Hook manage commands group(Prototype phase.).
  * `install`: Install the GPTComet hook.
  * `uninstall`: Uninstall the GPTComet hook.
  * `status`: Check the status of the GPTComet hook.
* `gptcomet generate`: Generate messages by changes/diff.
  * `commit`: Generate a commit message based on the changes made in the code.
  * `pr`: Generate a pull request message based on the changes made in the code.


## Configuration

The configuration file for GPTComet is `gptcomet.toml`. The file should contain the following keys:

*   `provider`: The provider of the language model (e.g., `openai`).
*   `api_base`: The base URL of the API (e.g., `https://api.openai.com/v1`).
*   `api_key`: The API key for the provider.
*   `model`: The model used for generating commit messages (e.g., `gpt-3.5-turbo`).
*   `retries`: The number of retries for the API request (e.g., `2`).
*   `prompt.brief_commit_message`: The prompt for generating brief commit messages.
*   `prompt.translation`: The prompt for translating commit messages to a target language.
*   `output.lang`: The language of the commit message (e.g., `en`).


## Supported Keys

You can use `gptcomet config keys` to check supported keys.

## Example

Here is an example of how to use GPTComet:

1.  When you first set your OpenAI KEY by `gptcomet config set openai.api_key YOUR_API_KEY`, it will generate config file at `~/.local/gptcomet/gptcomet.toml`, includes:
    ```
    provider = "openai"
    api_base = "https://api.openai.com/v1"
    api_key = "YOUR_API_KEY"
    model = "gpt-3.5-turbo"
    retries = 2
    output.lang = "en"
    ```
2.  Run the following command to generate a commit message: `GPTComet generate commit`
3.  GPTComet will generate a commit message based on the changes made in the code and display it in the console.

Note: Replace `YOUR_API_KEY` with your actual API key for the provider.


## Development

If you'd like to contribute to GPTComet, feel free to fork this project and submit a pull request.

## License

GPTComet is licensed under the MIT License.

## Contact

If you have any questions or suggestions, feel free to contact.
