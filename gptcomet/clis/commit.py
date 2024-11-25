import sys
from pathlib import Path
from typing import Annotated, Literal, Optional, cast

import typer
from git import (
    Commit,
    HookExecutionError,
    InvalidGitRepositoryError,
    NoSuchPathError,
    Repo,
    safe_decode,
)
from httpx import HTTPStatusError
from prompt_toolkit import prompt
from prompt_toolkit.cursor_shapes import CursorShape
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


def edit_text_in_place(initial_message: str) -> str:
    """
    Edit a given text in place with a prompt.

    Args:
        initial_message (str): The initial message to be edited.

    Returns:
        str: The edited message.
    """
    bottom_bar = "Support multiple lines. Press ESC then Enter to save, Ctrl+C to cancel."

    def bottom_toolbar():
        return [("class:bottom-toolbar", f" {bottom_bar} ")]

    style = Style.from_dict(
        {
            "bottom-toolbar": "fg:#aaaaaa bg:#FDF5E6",
            "prompt": "bold",
        }
    )

    try:
        edited_message = prompt(
            "Edit the message:\n",
            default=initial_message,
            multiline=True,
            enable_open_in_editor=True,
            mouse_support=True,
            cursor=CursorShape.BEAM,
            vi_mode=True,
            bottom_toolbar=bottom_toolbar,
            style=style,
        ).strip()

        if edited_message:
            console.print(Panel(stylize(edited_message, Colors.GREEN), title="Updated Message"))
            return edited_message
        else:
            return initial_message
    except KeyboardInterrupt:
        console.print("\n[yellow]Edit cancelled, keeping original message.[/yellow]")
        return initial_message


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


def commit(message_generator: MessageGenerator, commit_message: str, rich: bool = True) -> None:
    """
    Commit the generated commit message.

    Args:
        message_generator (MessageGenerator): The message generator object.
        commit_message (str): The commit message to be committed.
        rich (bool): If True, the commit message will be formatted with rich text markup.
    """
    try:
        commit = message_generator.repo.index.commit(message=commit_message)
        output = gen_output(message_generator.repo, commit, rich)
        console.print(output)
        console.print(stylize("Commit message saved successfully!", Colors.GREEN))
    except (HookExecutionError, ValueError) as error:
        console.print(f"[red]Commit Error: {error}[/red]")
        raise typer.Exit(1) from None


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
) -> None:
    """
    Generate and commit a message based on staged changes.

    The commit message will be generated automatically based on the staged changes,
    ignoring files specified in `file_ignore`.
    """
    if debug:
        set_debug()

    config_manager = (
        ConfigManager(config_path=config_path) if config_path else get_config_manager(local=local)
    )
    try:
        message_generator = MessageGenerator(config_manager)
    except (ConfigError, InvalidGitRepositoryError, NoSuchPathError) as error:
        console.print(stylize(str(error), Colors.YELLOW))
        raise typer.Exit(1) from None

    while True:
        try:
            commit_message = message_generator.generate_commit_message(rich=rich)
            if not commit_message:
                console.print(stylize("No commit message generated.", Colors.MAGENTA))
                raise typer.Exit(0) from None

            console.print("\nGenerated commit message:")
            console.print(Panel(stylize(commit_message, Colors.GREEN)))

            user_input = ask_for_retry()
            if user_input == "y":
                break
            elif user_input == "n":
                console.print(stylize("Commit message discarded.", Colors.YELLOW))
                return
            elif user_input == "e":
                commit_message = edit_text_in_place(commit_message)
                break

        except (KeyNotFound, GitNoStagedChanges, HTTPStatusError) as error:
            console.print(stylize(str(error), Colors.YELLOW))
            raise typer.Exit(1) from None
        except KeyboardInterrupt:
            console.print("\n[yellow]Operation cancelled by user.[/yellow]")
            return

    # Commit and show the output
    commit(message_generator, commit_message, rich)
