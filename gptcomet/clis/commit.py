import sys
from pathlib import Path
from typing import Annotated, Literal, Optional, cast

import typer
from git import Commit, HookExecutionError, Repo, safe_decode
from httpx import HTTPStatusError
from prompt_toolkit import prompt
from prompt_toolkit.cursor_shapes import CursorShape
from prompt_toolkit.styles import Style
from rich.panel import Panel
from rich.prompt import Prompt

from gptcomet.config_manager import ConfigManager, get_config_manager
from gptcomet.const import COMMIT_OUTPUT_TEMPLATE
from gptcomet.exceptions import GitNoStagedChanges, KeyNotFound
from gptcomet.log import set_debug
from gptcomet.message_generator import MessageGenerator
from gptcomet.styles import Colors, stylize
from gptcomet.utils import console


def ask_for_retry() -> Literal["y", "n", "r", "e"]:
    """
    Ask the user whether to retry generating a commit message.

    This function is only used in interactive mode.
    Returns:
        Literal["y", "n", "r"]: The user's choice.
    """
    char: Literal["y", "n", "r", "e"] = "y"
    if sys.stdin.isatty():
        # Interactive mode will ask for confirmation
        char = cast(
            Literal["y", "n", "r", "e"],
            Prompt.ask(
                "Do you want to use this commit message? y: yes, n: no, r: retry, e: edit.",
                default="y",
                choices=["y", "n", "r", "e"],
                case_sensitive=False,
            ),
        )
    else:
        console.print(
            "[yellow]Non-interactive mode detected, using the generated commit message directly.[/yellow]",
        )
        char = "y"

    return char


def edit_text_in_place(initial_message: str) -> str:
    """
    Edit a given text in place with a prompt.

    Args:
        initial_message (str): The initial message to be edited.

    Returns:
        str: The edited message.
    """
    bottom_bar = "Support multiple lines. Press `Ctrl+D` to continue."

    def bottom_toolbar():
        return [('class:bottom-toolbar', f' {bottom_bar} ')]

    style = Style.from_dict({
        'bottom-toolbar': 'fg:#aaaaaa bg:#FDF5E6',
    })

    edited_message = prompt(
        "Edit the message\n",
        default=initial_message,
        multiline=True,
        enable_open_in_editor=True,
        mouse_support=True,
        cursor=CursorShape.BEAM,
        vi_mode=True,
        bottom_toolbar=bottom_toolbar,
        style=style
    )
    console.print(Panel(stylize(edited_message, Colors.GREEN), title="Updated Msg"))
    return edited_message


def gen_output(repo: Repo, commit: Commit, rich=True) -> str:
    """
    Generate a formatted output string for a commit message.

    Args:
        repo (Repo): The git repository object.
        commit (Commit): The git commit object.
        rich (bool, optional): If True, the output will be formatted with rich text markup.
            Defaults to True.

    Returns:
        str: A formatted output string containing information about the commit.
    """
    commit_hash: str = commit.hexsha
    branch: str = repo.active_branch.name
    commit_msg: Optional[str] = safe_decode(commit.message)
    if commit_msg is not None:
        commit_msg = commit_msg.strip()
    no_author = "No Author"
    no_email = "No Email"
    author: Optional[str] = (
        safe_decode(getattr(commit.author, commit.author.conf_name, no_author)) or no_author
    )
    email: Optional[str] = (
        safe_decode(getattr(commit.author, commit.author.conf_email, no_email)) or no_email
    )

    git_show_stat: str = repo.git.show("--pretty=format:%b", "--stat", commit_hash)

    # Prepare the output format
    if rich is True:
        git_show_stat = (
            git_show_stat.strip().replace("+", "[green]+[/green]").replace("-", "[red]-[/red]")
        )
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


def entry(
    # rich: Annotated[bool, typer.Option(False, "--rich", "-r", help="Use rich output.")] = False,
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
):
    """Generates a commit message based on the staged changes. Will ignore files in `file_ignore`."""
    config_manager = (
        ConfigManager(config_path=config_path) if config_path else get_config_manager(local=local)
    )
    if debug:
        set_debug()

    message_generator = MessageGenerator(config_manager)
    while True:
        try:
            commit_message = message_generator.generate_commit_message()
        except (KeyNotFound, GitNoStagedChanges, HTTPStatusError) as error:
            console.print(stylize(str(error), Colors.YELLOW))
            raise typer.Exit(0) from None

        if not commit_message:
            console.print(stylize("No commit message generated.", Colors.MAGENTA))
            return

        console.print("Generated commit message:")
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

    try:
        commit = message_generator.repo.index.commit(message=commit_message)
        output = gen_output(message_generator.repo, commit)
        console.print(output)
    except (HookExecutionError, ValueError) as error:
        console.print(f"Commit Error: {error!s}")
        raise typer.Abort() from None

    console.print(stylize("Commit message saved!", Colors.GREEN))
