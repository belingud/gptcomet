import sys
from pathlib import Path
from typing import Annotated, Literal, Optional, cast

import typer
from git import Commit, DiffIndex, HookExecutionError, Repo, safe_decode
from litellm.exceptions import BadRequestError
from rich.prompt import Prompt

from gptcomet.config_manager import ConfigManager, get_config_manager
from gptcomet.exceptions import GitNoStagedChanges, KeyNotFound
from gptcomet.log import set_debug
from gptcomet.message_generator import MessageGenerator
from gptcomet.styles import Colors, stylize
from gptcomet.utils import console


def ask_for_retry() -> Literal["y", "n", "r"]:
    """
    Ask the user whether to retry generating a commit message.

    This function is only used in interactive mode.
    Returns:
        Literal["y", "n", "r"]: The user's choice.
    """
    char: Literal["y", "n", "r"] = "y"
    if sys.stdin.isatty():
        # Interactive mode will ask for confirmation
        char = cast(
            Literal["y", "n", "r"],
            Prompt.ask(
                "Do you want to use this commit message? y: yes, n: no, r: retry.",
                default="y",
                choices=["y", "n", "r"],
                case_sensitive=False,
            ),
        )
    else:
        console.print(
            "[yellow]Non-interactive mode detected, using the generated commit message directly.[/yellow]",
        )
        char = "y"

    return char


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
    commit_hash: str = commit.hexsha[:7]
    branch: str = repo.active_branch.name
    commit_msg: str = safe_decode(commit.message)
    author: Optional[str] = safe_decode(getattr(commit.author, commit.author.conf_name, None))
    email: Optional[str] = safe_decode(getattr(commit.author, commit.author.conf_email, None))

    # Get the diff details from the previous commit
    diffs: DiffIndex = commit.diff(commit.parents[0] if commit.parents else None)
    file_count: int = 0
    insertions: int = 0
    deletions: int = 0
    file_changes: list[str] = []

    for change in diffs:
        file_count += 1
        insertions += change.diff.count("\n+")  # Count added lines
        deletions += change.diff.count("\n-")  # Count deleted lines
        # Track file mode changes (e.g., new files)
        if change.new_file:
            file_changes.append(f" create mode {change.b_mode} {change.b_path}")
        elif change.deleted_file:
            file_changes.append(f" delete mode {change.b_mode} {change.b_path}")
        elif change.renamed_file:
            file_changes.append(f" rename {change.a_path} {change.b_path}")
        else:
            file_changes.append(f" modify {change.a_path}")

    # Prepare the output format
    if rich is True:
        author_info = (
            f":construction_worker: [green]{author}[/green]  :email: [blue]{email}[/blue]\n"
        )
    else:
        author_info = f"{author}  {email}\n"
    output = author_info + f"[{branch} {commit_hash}] {commit_msg}\n"

    output += f" {file_count} files changed, {insertions} insertions(+), {deletions} deletions(-)\n"

    # Append file mode changes if any
    if file_changes:
        output += "\n".join(file_changes) + "\n"
    return output


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
        except (KeyNotFound, GitNoStagedChanges, BadRequestError) as error:
            console.print(str(error))
            raise typer.Abort() from None

        if not commit_message:
            console.print(stylize("No commit message generated.", Colors.MAGENTA))
            return

        console.print("Generated commit message:")
        console.print(stylize(commit_message, Colors.GREEN, panel=True))

        user_input = ask_for_retry()
        if user_input == "y":
            break
        elif user_input == "n":
            console.print(stylize("Commit message discarded.", Colors.YELLOW))
            return

    try:
        commit = message_generator.repo.index.commit(message=commit_message)
        output = gen_output(message_generator.repo, commit)
        console.print(output)
    except (HookExecutionError, ValueError) as error:
        console.print(f"Commit Error: {error!s}")
        raise typer.Abort() from None

    console.print(stylize("Commit message saved.", Colors.CYAN))
