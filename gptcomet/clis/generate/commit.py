import sys
from pathlib import Path
from typing import Annotated, Optional

import typer
from git import Commit, DiffIndex, HookExecutionError, Repo, safe_decode
from litellm.exceptions import BadRequestError
from rich.prompt import Prompt

from gptcomet.config_manager import ConfigManager, get_config_manager
from gptcomet.const import CONSOLE_VERBOSE_KEY, GPTCOMET_PRE
from gptcomet.exceptions import GitNoStagedChanges, KeyNotFound
from gptcomet.log import logger, set_debug
from gptcomet.message_generator import MessageGenerator
from gptcomet.styles import Colors, stylize
from gptcomet.utils import console, strtobool


def ask_for_retry() -> bool:
    """
    Ask the user whether to retry generating a commit message.

    This function is only used in interactive mode.
    Returns:
        bool: True if the user chooses to retry generating the commit message, False otherwise.
    """
    retry = True
    if sys.stdin.isatty():
        # Interactive mode will ask for confirmation
        char = Prompt.ask(
            "Do you want to use this commit message? y: yes, n: no, r: retry.",
            default="y",
            choices=["y", "n", "r"],
            case_sensitive=False,
        )
    else:
        console.print(
            "[yellow]Non-interactive mode detected, using the generated commit message directly.[/yellow]",
        )
        return False

    if char == "n":
        console.print("[yellow]Commit message discarded.[/yellow]")
        return False
    elif char == "y":
        return False
    return retry


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
    """
    Generates a commit message based on the staged changes. Will ignore files in `file_ignore`.
    """
    rich = False
    if config_path:
        cfg: ConfigManager = ConfigManager(config_path=config_path)
    else:
        cfg: ConfigManager = get_config_manager(local=local)
    if debug:
        set_debug()
        logger.debug(f"Using Config path: {cfg.current_config_path}")
    message_generator = MessageGenerator(cfg)
    retry = True
    commit_msg = ""
    console.print(
        stylize(
            "🤖 Hang tight! I'm having a chat with the AI to craft your commit message...",
            Colors.CYAN,
        ),
    )
    while retry:
        try:
            commit_msg = message_generator.generate_commit_message(rich)
        except KeyNotFound as e:
            console.print(f"Error: {e!s}, please check your configuration.")
            raise typer.Abort() from None
        except (GitNoStagedChanges, BadRequestError) as e:
            console.print(str(e))
            raise typer.Abort() from None

        if not commit_msg:
            console.print(stylize("No commit message generated.", Colors.MAGENTA))
            return

        console.print("Generated commit message:")
        console.print(stylize(commit_msg, Colors.GREEN))

        retry = ask_for_retry()
    if not commit_msg:
        console.print(stylize("No commit message generated.", Colors.MAGENTA))
        raise typer.Abort()
    try:
        commit: Commit = message_generator.repo.index.commit(message=commit_msg)
        if strtobool(cfg.get(CONSOLE_VERBOSE_KEY, True)):
            output: str = gen_output(message_generator.repo, commit)
            console.print(output)
    except (HookExecutionError, ValueError) as e:
        console.print(f"Commit Error: {e!s}")
        raise typer.Abort() from None
    console.print(stylize("Commit message saved.", Colors.CYAN))
