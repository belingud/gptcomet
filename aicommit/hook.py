import logging
from pathlib import Path

import click
from git import Repo

from aicommit.const import RICH_PREPARE_COMMIT_MSG, SHORT_PREPARE_COMMIT_MSG

logger = logging.getLogger(__name__)


class AICommitHook:
    __slots__ = ("repo", "hook_path")

    def __init__(self):
        """
        Initialize git repository and hook path.

        Raises:
            git.InvalidGitRepositoryError: If the current directory is not a git repository.
            git.NoSuchPathError:
        """
        self.repo = Repo(Path.cwd())
        self.hook_path = Path(self.repo.git_dir) / "hooks" / "prepare-commit-msg"

    def install_hook(self, use_rich: bool = False):
        """
        Install the AICommit prepare-commit-msg hook.

        Args:
            use_rich (bool): Whether to use the rich commit message template. Defaults to False.
        """
        if use_rich:
            self.hook_path.write_text(RICH_PREPARE_COMMIT_MSG)
        else:
            self.hook_path.write_text(SHORT_PREPARE_COMMIT_MSG)
        self.hook_path.chmod(0o755)

    def is_hook_installed(self):
        """
        Check if the AICommit hook is installed.
        Returns True if the hook path exists and 'aicommit' is in the hook path content, False otherwise.
        """
        return self.hook_path.exists() and "aicommit" in self.hook_path.open().read()

    def uninstall_hook(self):
        """
        Uninstalls the AICommit prepare-commit-msg hook if it is installed.
        """
        if self.is_hook_installed():
            logger.debug("Uninstalling prepare-commit-msg hook...")
            self.hook_path.unlink()
            click.echo(
                f"[{click.style('AICommit', fg='green')}] prepare-commit-msg hook has been uninstalled successfully."
            )
        else:
            click.echo(
                f"[{click.style('AICommit', fg='green')}] prepare-commit-msg hook is not installed."
            )
