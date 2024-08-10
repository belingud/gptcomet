# aicommit: AI-Powered Git Commit Message Generator

## Table of Contents



## Overview

aicommit is a Python library designed to automate the process of generating commit messages for Git repositories.
It leverages the power of AI to create meaningful commit messages based on the changes made in the codebase.

## Features

*   **Automatic Commit Message Generation**: AICommit can generate commit messages based on the changes made in the code.
*   **Support for Multiple Languages**: AICommit supports multiple languages, including English and Chinese.
*   **Customizable Configuration**: AICommit allows users to customize the configuration to suit their needs.
*   **Support for Rich Commit Messages**: AICommit supports rich commit messages, which include a title, summary, and detailed description.

## Installation

To use aicommit, you need to have Python installed on your system. You can install the library using pip:

```shell
pip install aicommit
```

Recommend install use `pipx` on Mac or Linux.

```shell
pipx install aicommit
```


## Usage

To use AICommit, follow these steps:

1.  **Install AICommit**: Install AICommit through pypi.
2.  **Configure AICommit**: Configure AICommit with your api_key The configuration file should contain the following keys:
    *   `provider`: The provider of the language model (e.g., `openai`).
    *   `api_base`: The base URL of the API (e.g., `https://api.openai.com/v1`).
    *   `api_key`: The API key for the provider.
    *   `model`: The model used for generating commit messages (e.g., `gpt-3.5-turbo`).
    *   `retries`: The number of retries for the API request (e.g., `2`).
3.  **Run AICommit**: Run AICommit using the following command: `aicommit generate commit`.

## Commands

The following are the available commands for AICommit:

* `aicommit config`: Config manage commands group.
  * `set`: Set a configuration value.
  * `get`: Get a configuration value.
  * `list`: List all configuration values.
  * `reset`: Reset the configuration to its default values.
  * `keys`: List all supported keys.
* `aicommit hook`: Hook manage commands group(Prototype phase.).
  * `install`: Install the AICommit hook.
  * `uninstall`: Uninstall the AICommit hook.
  * `status`: Check the status of the AICommit hook.
* `aicommit generate`: Generate messages by changes/diff.
  * `aicommit generate commit`: Generate a commit message based on the changes made in the code.
  * `aicommit generate pr`: Generate a pull request message based on the changes made in the code.


## Configuration

The configuration file for AICommit is `aicommit.toml`. The file should contain the following keys:

*   `provider`: The provider of the language model (e.g., `openai`).
*   `api_base`: The base URL of the API (e.g., `https://api.openai.com/v1`).
*   `api_key`: The API key for the provider.
*   `model`: The model used for generating commit messages (e.g., `gpt-3.5-turbo`).
*   `retries`: The number of retries for the API request (e.g., `2`).
*   `prompt.brief_commit_message`: The prompt for generating brief commit messages.
*   `prompt.translation`: The prompt for translating commit messages to a target language.
*   `output.lang`: The language of the commit message (e.g., `en`).


## Supported Keys

You can use `aicommit config keys` to check supported keys.

## Example

Here is an example of how to use AICommit:

1.  When you first set your OpenAI KEY by `aicommit config set openai.api_key YOUR_API_KEY`, it will generate config file at `~/.local/aicommit/aicommit.toml`, includes:
    ```
    provider = "openai"
    api_base = "https://api.openai.com/v1"
    api_key = "YOUR_API_KEY"
    model = "gpt-3.5-turbo"
    retries = 2
    output.lang = "en"
    ```
2.  Run the following command to generate a commit message: `aicommit generate commit`
3.  AICommit will generate a commit message based on the changes made in the code and display it in the console.

Note: Replace `YOUR_API_KEY` with your actual API key for the provider.


## Development

If you'd like to contribute to AICommit, feel free to fork this project and submit a pull request.

## License

AICommit is licensed under the MIT License.

## Contact

If you have any questions or suggestions, feel free to contact.
