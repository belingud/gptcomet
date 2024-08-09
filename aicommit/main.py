import asyncio
from pathlib import Path

import click
import toml
from git import GitCommandError, Repo

CONFIG_PATH = Path("~/.config/aicommit/aicommit.toml").expanduser()


class AICommitError(Exception):
    pass


class NotSupportedError(AICommitError):
    pass


class UnknownProviderError(Exception):
    def __init__(self, provider_name: str):
        self.provider_name = provider_name

    def __str__(self):
        return f"Unknown provider: {self.provider_name}"


class GitError(AICommitError):
    pass


class ConfigError(AICommitError):
    API_KEY_MISSING_ERROR = "API key must be set in the configuration file."
    CONFIG_FILE_MISSING_ERROR = (
        "Configuration file not found. You have to set API_KEY by aicommit config set API_KEY=YOUR_API_KEY"
    )


class APIError(AICommitError):
    pass


class AIProvider:
    """Base class for AI providers."""

    async def generate_response(self, prompt: str) -> str:
        raise NotImplementedError("This method should be implemented by subclasses.")


class OpenAIProvider(AIProvider):
    """Implementation for OpenAI API."""

    def __init__(self, cfg):
        import openai

        self.api_key = cfg["openai"]["api_key"]
        self.api_base = cfg["openai"]["api_base"]
        self.model = cfg["openai"]["model"]
        self.temperature = cfg["openai"]["temperature"]
        self.max_tokens = cfg["openai"]["max_tokens"]
        self.retries = cfg["openai"]["retries"]

        openai.api_key = self.api_key
        openai.api_base = self.api_base

    async def generate_response(self, prompt: str) -> str:
        import openai

        try:
            response = await openai.ChatCompletion.acreate(
                model=self.model,
                messages=[{"role": "user", "content": prompt}],
                max_tokens=self.max_tokens,
                temperature=self.temperature,
            )
            return response.choices[0].message["content"].strip()
        except openai.OpenAIError as e:
            raise APIError from e


# You can define similar classes for other providers, e.g., ClaudeProvider


class AICommit:
    def __init__(self):
        self.load_config()
        self.repo = Repo(Path().absolute())
        self.provider = self.initialize_provider()

    def load_config(self):
        if not CONFIG_PATH.exists():
            raise ConfigError(ConfigError.CONFIG_FILE_MISSING_ERROR)
        if CONFIG_PATH.exists():
            with open(CONFIG_PATH) as f:
                self.config = toml.load(f)
        else:
            self.config = {
                "openai": {
                    "api_key": "",
                    "api_base": "https://api.siliconflow.cn/v1/chat/completions",
                    "model": "Qwen/Qwen2-7B-Instruct",
                    "temperature": 0.5,
                    "max_tokens": 60,
                    "retries": 2,
                    "proxy": "",
                },
                "model_provider": "openai",
                "allow_amend": True,
                "file_ignore": [
                    "bun.lockb",
                    "Cargo.lock",
                    "composer.lock",
                    "Gemfile.lock",
                    "package-lock.json",
                    "pnpm-lock.yaml",
                    "poetry.lock",
                    "yarn.lock",
                ],
                "prompt": {
                    "conventional_commit_prefix": "You are an expert programmer summarizing a code change...",
                    "commit_summary": "You are an expert programmer writing a commit message...",
                    "commit_title": "You are an expert programmer writing a commit message title...",
                    "file_diff": "You are an expert programmer summarizing a git diff...",
                    "translation": "You are a professional polyglot programmer and translator...",
                },
                "output": {
                    "conventional_commit": True,
                    "conventional_commit_prefix_format": "{{ prefix }}: ",
                    "lang": "en",
                    "show_per_file_summary": False,
                },
            }

        if not self.config["openai"]["api_key"]:
            raise ConfigError(ConfigError.API_KEY_MISSING_ERROR)

    def save_config(self):
        CONFIG_PATH.parent.mkdir(parents=True, exist_ok=True)
        with open(CONFIG_PATH, "w") as f:
            toml.dump(self.config, f)

    def initialize_provider(self) -> AIProvider:
        provider_name = self.config.get("model_provider", "openai")
        if provider_name == "openai":
            return OpenAIProvider(self.config)
        # elif provider_name == "claude":
        #     return ClaudeProvider(self.config)
        else:
            raise UnknownProviderError(provider_name)

    async def get_git_diff(self) -> str:
        try:
            diff = self.repo.git.diff("HEAD", "--cached")
        except GitCommandError as e:
            raise GitError(str(e)) from None
        else:
            return diff

    async def generate_commit_message(self, diff: str) -> str:
        prompts = self.config["prompt"]

        # Generate prefix
        prefix_prompt = prompts["conventional_commit_prefix"].replace("{{ file_diff }}", diff)
        prefix = await self.provider.generate_response(prefix_prompt)

        # Generate title
        title_prompt = prompts["commit_title"].replace("{{ file_diff }}", diff)
        title = await self.provider.generate_response(title_prompt)

        # Generate summary
        summary_prompt = prompts["commit_summary"].replace("{{ file_diff }}", diff)
        summary = await self.provider.generate_response(summary_prompt)

        # Format commit message
        commit_message = f"{prefix}\n\n{title}\n\n{summary}"
        return commit_message

    async def commit(self, auto_commit: bool):
        diff = await self.get_git_diff()
        commit_message = await self.generate_commit_message(diff)
        click.echo(f"Generated commit message:\n{commit_message}")
        if auto_commit:
            self.repo.index.commit(commit_message)
            click.echo("Changes committed.")


@click.group()
def cli():
    pass


@cli.command()
@click.option("--api-key", prompt=True, hide_input=True, help="OpenAI API key.")
@click.option("--api-base", default="https://api.siliconflow.cn/v1/chat/completions", help="OpenAI API base URL.")
@click.option("--model", default="Qwen/Qwen2-7B-Instruct", help="Model to use.")
@click.option("--temperature", default=0.5, help="Temperature for the model.")
@click.option("--max-tokens", default=60, help="Maximum tokens for the response.")
@click.option("--language", default="en", type=click.Choice(["en", "zh"]), help="Language for commit messages.")
@click.option("--auto-commit/--no-auto-commit", default=True, help="Automatically commit changes.")
def config(api_key, api_base, model, temperature, max_tokens, language, auto_commit):
    """Set configuration for aicommit."""
    ai_commit = AICommit()
    ai_commit.config["openai"]["api_key"] = api_key
    ai_commit.config["openai"]["api_base"] = api_base
    ai_commit.config["openai"]["model"] = model
    ai_commit.config["openai"]["temperature"] = temperature
    ai_commit.config["openai"]["max_tokens"] = max_tokens
    ai_commit.config["output"]["lang"] = language
    ai_commit.config["output"]["auto_commit"] = auto_commit
    ai_commit.save_config()
    click.echo("Configuration saved.")


@cli.command()
@click.option("--auto-commit/--no-auto-commit", default=True, help="Automatically commit changes.")
def commit(auto_commit):
    """Generate a commit message and optionally commit."""
    ai_commit = AICommit()
    asyncio.run(ai_commit.commit(auto_commit))


@cli.command()
@click.option("--install", is_flag=True, help="Install pre-commit hook.")
@click.option("--uninstall", is_flag=True, help="Uninstall pre-commit hook.")
def hook(install, uninstall):
    """Manage git hooks for aicommit."""
    hook_script = Path(".git/hooks/prepare-commit-msg")
    if install:
        hook_script.parent.mkdir(parents=True, exist_ok=True)
        with open(hook_script, "w") as f:
            f.write("#!/bin/sh\nexec aicommit commit\n")
        hook_script.chmod(0o755)
        click.echo("Pre-commit hook installed.")
    elif uninstall:
        if hook_script.exists():
            hook_script.unlink()
            click.echo("Pre-commit hook uninstalled.")
        else:
            click.echo("No pre-commit hook found.")


if __name__ == "__main__":
    cli()
