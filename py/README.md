# GPTComet Python Package

This is the Python package for GPTComet, a tool that leverages AI to automatically generate Git commit messages.

## Installation

```bash
pip install gptcomet
```

## Usage

After installation, you can use the `gptcomet` command directly from your terminal:

```bash
# Generate a commit message for staged changes
gptcomet commit

# Translate a commit message
gptcomet translate "fix: update user interface" --lang zh

# Configure GPTComet
gptcomet config set api_key "your-api-key"
```

## Features

- Automatically generate meaningful commit messages using AI
- Support for multiple languages
- Easy configuration management
- Cross-platform support (Windows, macOS, Linux)

## Requirements

- Python 3.9 or higher
- Git

## License

MIT License
