import sys
from pathlib import Path
from typing import Annotated, Any, Literal, Optional, cast

import click
import requests
import typer
from git import (
    Commit,
    HookExecutionError,
    InvalidGitRepositoryError,
    NoSuchPathError,
    Repo,
    safe_decode,
)
from prompt_toolkit import prompt
from prompt_toolkit.key_binding import KeyBindings
from prompt_toolkit.keys import Keys
from prompt_toolkit.styles import Style
from rich.panel import Panel
from rich.prompt import Prompt

from gptcomet.config_manager import ConfigManager, get_config_manager
from gptcomet.const import COMMIT_OUTPUT_TEMPLATE
from gptcomet.exceptions import ConfigError, GitNoStagedChanges, KeyNotFound
from gptcomet.log import set_debug
from gptcomet.message_generator import MessageGenerator
from gptcomet.styles import Colors, stylize
from gptcomet.utils import console

RETRY_INPUT = Literal["y", "n", "r", "e"]
RETRY_CHOICES = {"y": "yes", "n": "no", "r": "retry", "e": "edit"}


def ask_for_retry() -> RETRY_INPUT:
    """
    Ask the user whether to retry generating a commit message.

    Returns:
        RETRY_INPUT: The user's choice (y/n/r/e).
    """
    if not sys.stdin.isatty():
        console.print(
            "[yellow]Non-interactive mode detected, using the generated commit message directly.[/yellow]",
        )
        return "y"

    prompt_text = "Do you want to use this commit message?\n" + ", ".join(
        [f"{k}: {v}" for k, v in RETRY_CHOICES.items()]
    )

    return cast(
        RETRY_INPUT,
        Prompt.ask(
            prompt_text,
            default="y",
            choices=list(RETRY_CHOICES.keys()),
            case_sensitive=False,
        ).lower(),
    )


def edit_text_in_place(initial_message: str) -> Optional[str]:
    """
    Edit a given text in place with a prompt.

    Args:
        initial_message (str): The initial message to be edited.

    Returns:
        Optional[str]: The edited message or None if cancelled.
    """
    if not initial_message:
        return None
    bottom_bar = (
        "[VIM mode] Support multiple lines. Press ESC then Enter to save, Ctrl+C to cancel."
    )

    def bottom_toolbar():
        return [("class:bottom-toolbar", f" {bottom_bar} ")]

    style = Style.from_dict(
        {
            "bottom-toolbar": "fg:#aaaaaa bg:#FDF5E6",
            "prompt": "bold",
        }
    )

    kb = KeyBindings()

    @kb.add(Keys.ControlC)
    def _(event):
        event.app.exit(result=None)

    try:
        edited_message = prompt(
            "Edit the message:\n",
            default=initial_message,
            multiline=True,
            style=style,
            bottom_toolbar=bottom_toolbar,
            key_bindings=kb,
        )
        if edited_message is None:  # Ctrl+C was pressed
            console.print("\n[yellow]Commit cancelled.[/yellow]")
            return None
        if edited_message:
            console.print(Panel(stylize(edited_message, Colors.GREEN), title="Updated Message"))
            return edited_message
        else:
            return initial_message
    except KeyboardInterrupt:
        console.print("\n[yellow]Commit cancelled.[/yellow]")
        return None


def gen_output(repo: Repo, commit: Commit, rich: bool = True) -> str:
    """
    Generate a formatted output string for a commit message.

    Args:
        repo (Repo): The git repository object.
        commit (Commit): The git commit object.
        rich (bool): If True, the output will be formatted with rich text markup.

    Returns:
        str: A formatted output string containing information about the commit.
    """
    NO_AUTHOR = "No Author"
    NO_EMAIL = "No Email"

    commit_hash: str = commit.hexsha
    branch: str = repo.active_branch.name
    commit_msg: str = safe_decode(commit.message).strip() if commit.message else ""

    author: str = (
        safe_decode(getattr(commit.author, commit.author.conf_name, NO_AUTHOR)) or NO_AUTHOR
    )
    email: str = safe_decode(getattr(commit.author, commit.author.conf_email, NO_EMAIL)) or NO_EMAIL

    git_show_stat: str = repo.git.show("--pretty=format:", "--stat", commit_hash).strip()

    if rich:
        git_show_stat = git_show_stat.replace("+", "[green]+[/green]").replace("-", "[red]-[/red]")
        author = f":construction_worker: [green]{author}[/]"
        email = f"[blue]{email}[/blue]"

    return COMMIT_OUTPUT_TEMPLATE.format(
        author=author,
        email=email,
        branch=branch,
        commit_hash=commit_hash,
        commit_msg=commit_msg,
        git_show_stat=git_show_stat,
    )


def commit(
    message_generator: MessageGenerator, commit_message: Optional[str] = None, rich: bool = True
) -> None:
    """
    Commit the generated commit message.

    Args:
        message_generator (MessageGenerator): The message generator object.
        commit_message (Optional[str]): The commit message to be committed.
        rich (bool): If True, the commit message will be formatted with rich text markup.
    """
    if not commit_message:
        return
    try:
        commit = message_generator.repo.index.commit(message=commit_message)
        output = gen_output(message_generator.repo, commit, rich)
        console.print(output)
        console.print(stylize("Commit message saved successfully!", Colors.GREEN))
    except (HookExecutionError, ValueError) as error:
        console.print(f"[red]Commit Error: {error}[/red]")
        raise typer.Exit(1) from None


def _setup_config(
    config_path: Optional[Path],
    local: bool,
    provider: Optional[str],
    cli_args: dict[str, Any],
) -> ConfigManager:
    """Set up configuration manager with CLI overrides."""
    config_manager = (
        ConfigManager(config_path=config_path) if config_path else get_config_manager(local=local)
    )
    if provider or cli_args:
        config_manager.set_cli_overrides(provider=provider, api_config=cli_args)
    return config_manager


def _get_cli_args(**kwargs: Any) -> dict[str, Any]:
    """Extract CLI arguments for API configuration."""
    return {k: v for k, v in kwargs.items() if v is not None}


def _generate_message(
    message_generator: MessageGenerator,
    rich: bool,
    dry_run: bool,
) -> Optional[str]:
    """Generate and handle commit message with user interaction."""
    commit_message: Optional[str] = None
    while True:
        try:
            if not commit_message:
                console.print(
                    stylize("🤖 Hang tight, I'm cooking up something good!", Colors.GREEN)
                )
                commit_message = message_generator.generate_commit_message(rich=rich)

            console.print("\nCommit message:")
            console.print(Panel(stylize(commit_message, Colors.GREEN)))
            if dry_run:
                console.print(stylize("Commit message not saved.", Colors.YELLOW))
                return None

            user_input = ask_for_retry()
            if user_input == "y":
                return commit_message
            if user_input == "n":
                console.print(stylize("Commit message discarded.", Colors.YELLOW))
                return None
            if user_input == "e":
                commit_message = edit_text_in_place(commit_message)
            elif user_input == "r":
                commit_message = None
        except KeyboardInterrupt:
            console.print("\n[yellow]Operation cancelled by user.[/yellow]")
            return None
    return None


def entry(
    rich: Annotated[
        bool, typer.Option("--rich", "-r", help="Generate rich commit message.")
    ] = False,
    debug: Annotated[bool, typer.Option("--debug", "-d", help="Print debug information.")] = False,
    local: Annotated[bool, typer.Option("--local", help="Use local configuration file.")] = False,
    config_path: Annotated[
        Optional[Path],
        typer.Option(
            "--config",
            "-c",
            help="Path to config file.",
            exists=True,
            file_okay=True,
            dir_okay=False,
            writable=False,
            readable=True,
            resolve_path=True,
        ),
    ] = None,
    api_key: Annotated[
        Optional[str],
        typer.Option("--api-key", help="API key for the provider.", click_type=click.STRING),
    ] = None,
    api_base: Annotated[
        Optional[str],
        typer.Option("--api-base", help="Base URL for the API.", click_type=click.STRING),
    ] = None,
    model: Annotated[
        Optional[str],
        typer.Option("--model", help="Model to use for generation.", click_type=click.STRING),
    ] = None,
    provider: Annotated[
        Optional[str],
        typer.Option(
            "--provider",
            help="Provider to use (e.g., openai, anthropic).",
            click_type=click.STRING,
        ),
    ] = None,
    proxy: Annotated[
        Optional[str],
        typer.Option("--proxy", help="Proxy URL to use for API calls.", click_type=click.STRING),
    ] = None,
    max_tokens: Annotated[
        Optional[int],
        typer.Option(
            "--max-tokens",
            help="Maximum tokens for generation.",
            click_type=click.IntRange(min=1),
        ),
    ] = None,
    top_p: Annotated[
        Optional[float],
        typer.Option(
            "--top-p",
            help="Top-p sampling parameter.",
            click_type=click.FloatRange(min=0, max=1),
        ),
    ] = None,
    temperature: Annotated[
        Optional[float],
        typer.Option(
            "--temperature",
            help="Temperature for generation.",
            click_type=click.FloatRange(min=0, max=1),
        ),
    ] = None,
    extra_headers: Annotated[
        Optional[str],
        typer.Option(
            "--extra-headers",
            help="Extra headers for API calls (JSON string).",
            click_type=click.STRING,
        ),
    ] = None,
    frequency_penalty: Annotated[
        Optional[float],
        typer.Option(
            "--frequency-penalty",
            help="Frequency penalty for generation.",
            click_type=click.FloatRange(min=0, max=1),
        ),
    ] = None,
    dry_run: Annotated[
        bool,
        typer.Option(
            "--dry-run",
            help="Print the commit message without committing.",
            is_eager=True,
        ),
    ] = False,
) -> None:
    """
    Generate and commit a message based on staged changes.

    The commit message will be generated automatically based on the staged changes,
    ignoring files specified in `file_ignore`.
    """
    debug and set_debug()

    cli_args = _get_cli_args(
        api_key=api_key,
        api_base=api_base,
        model=model,
        proxy=proxy,
        max_tokens=max_tokens,
        top_p=top_p,
        temperature=temperature,
        extra_headers=extra_headers,
        frequency_penalty=frequency_penalty,
    )

    try:
        config_manager = _setup_config(config_path, local, provider, cli_args)
        message_generator = MessageGenerator(config_manager)
    except (ConfigError, InvalidGitRepositoryError, NoSuchPathError) as error:
        console.print(stylize(str(error), Colors.YELLOW))
        raise typer.Exit(1) from None

    try:
        commit_message = _generate_message(message_generator, rich, dry_run)
        if commit_message:
            commit(message_generator, commit_message, rich)
    except (
        KeyNotFound,
        GitNoStagedChanges,
        requests.RequestException,
        requests.Timeout,
    ) as error:
        console.print(stylize(str(error), Colors.YELLOW))
        raise typer.Exit(1) from None
