import typing as t
from pathlib import Path

from git import Repo

from gptcomet.config_manager import ConfigManager
from gptcomet.const import FILE_IGNORE_KEY, GPTCOMET_PRE, LANGUAGE_KEY
from gptcomet.exceptions import GitNoStagedChanges
from gptcomet.llm_client import LLMClient
from gptcomet.log import logger
from gptcomet.styles import Colors, stylize
from gptcomet.utils import console, output_language_map, should_ignore


class MessageGenerator:
    """
    A class that generates concise git messages based on the staged changes.

    Args:
        config_manager (ConfigManager): The gptcomet ConfigManager instance.

    Attributes:
        config_manager (ConfigManager): The ConfigManager instance to use.
        llm_client (LLMClient): The LLMClient instance to use.
        repo (Repo): The Repo instance to use.

    Methods:
        generate_commit_message(rich: bool = True): Generate a commit message from the staged changes.
        make_ignored_options(ignored_files: list[str]): Make a list of ignored files.

    Raises:
        GitNoStagedChanges: If there are no staged changes.
        InvalidGitRepositoryError: If the current directory is not a Git repository.
        ConfigError: If the configuration is invalid.

    Examples:
        >>> from gptcomet.config_manager import ConfigManager
        >>> from gptcomet.message_generator import MessageGenerator
        >>> config_manager = ConfigManager()
        >>> message_generator = MessageGenerator(config_manager)
        >>> message_generator.generate_commit_message()
    """

    __slots__ = ("config_manager", "diff", "llm_client", "repo")

    def __init__(self, config_manager: ConfigManager, repo_path: t.Optional[str] = None):
        """
        Initialize MessageGenerator instance.

        Args:
            config_manager (ConfigManager): The ConfigManager instance to use.
            repo_path (str, optional): The path to the Git repository. Defaults to None.

        Raises:
            InvalidGitRepositoryError: If the path is not a Git repository.
            ConfigError: If the configuration is invalid.
        """
        self.config_manager = config_manager
        self.llm_client = LLMClient(config_manager)
        self.repo = Repo(repo_path or Path.cwd())
        self.diff = None

    @classmethod
    def from_config_manager(cls, config_manager: ConfigManager, repo_path: t.Optional[str] = None):
        """
        Creates an instance of the class from a ConfigManager.

        Args:
            config_manager (ConfigManager): The ConfigManager instance to create the class instance from.
            repo_path (Optional[str]): The path to the repository to use.
        Returns:
            An instance of the class.
        """
        return cls(config_manager, repo_path)

    def get_staged_diff(self, repo: t.Optional[Repo] = None) -> str:
        ignored_files: list = self.config_manager.get(FILE_IGNORE_KEY)
        diff_options = ["--staged", "-U2", *self.make_ignored_options(ignored_files)]
        logger.debug(f"{GPTCOMET_PRE} Diff options: {diff_options}")
        repo = repo or self.repo
        diff = repo.git.diff(diff_options)
        lines = diff.splitlines()
        return "\n".join(
            [
                line
                for line in lines
                if not line.startswith("index")
                and not line.startswith("---")
                and not line.startswith("+++")
            ]
        )

    def make_ignored_options(self, ignored_files: list[str]) -> list[str]:
        """
        Make a list of ignored files.

        Args:
            ignored_files (list[str]): The list of ignored files.

        Returns:
            list[str]: The list of ignored files.
        """
        staged_files = self.repo.git.diff(["--staged", "--name-only"]).splitlines()
        return [
            f":!{file}"
            for file in staged_files
            if should_ignore(file, ignored_files) and Path(file).exists()
        ]

    def generate_commit_message(self, rich: bool = False) -> str:
        """
        Generate a commit message from the staged changes.

        Args:
            rich (bool): Whether to use the rich commit message template. Defaults to False.

        Returns:
            str: The generated commit message.

        Raises:
            GitNoStagedChanges: If there are no staged changes.
            BadRequestError: If the completion API returns an error.
        """
        logger.debug(f"{GPTCOMET_PRE} Generating commit message, rich: {rich}")
        diff = self.get_staged_diff()
        logger.debug(f"{GPTCOMET_PRE} Diff length: {len(diff)}")
        self.llm_client.clear_history()
        if not diff:
            raise GitNoStagedChanges()
        return self._translate_msg(self._gen_msg_from_diff(diff, rich=rich))

    def _gen_msg_from_diff(self, diff: str, rich: bool = False) -> str:
        if not rich:
            return self._generate_brief_commit_message(diff)
        else:
            return self._generate_rich_commit_message(diff)

    def _translate_msg(self, msg: str) -> str:
        lang = self.config_manager.get(LANGUAGE_KEY, _type=str)
        title = ""
        if str(lang).lower() != "en" and lang is not None:
            full_lang = output_language_map.get(lang, None)
            if full_lang is None:
                console.print(
                    stylize(
                        f"{GPTCOMET_PRE} Language {lang} not supported, will use the original value {lang} for the attempt.",
                        Colors.MAGENTA,
                    )
                )
                full_lang = lang
            if self.config_manager.get("console.verbose"):
                console.print(f"{GPTCOMET_PRE} Original commit message: {msg}")
            if ":" in msg:
                title, msg = msg.split(":")
            # Default is English, but can be changed by the user
            console.print(f"{GPTCOMET_PRE} Translating commit message to {lang}")
            translation = str(self.config_manager.get("prompt.translation"))
            translation = translation.replace("{{ placeholder }}", msg).replace(
                "{{ output_language }}", full_lang
            )
            msg = self.llm_client.generate(translation)
        return f"{title}: {msg}" if title else msg

    def _generate_brief_commit_message(self, diff: str) -> str:
        """
        Generates a brief commit message based on the given diff.

        Args:
            diff (str): The diff string representing the changes made.

        Returns:
            str: The generated brief commit message.

        Raises:
            None

        Description:
            This function uses the "prompt.brief_commit_message" from the config manager to generate a prompt.
            The placeholder "{{ placeholder }}" in the prompt is replaced with the provided diff.

        Example:
            >>> diff = "Add new feature"
            >>> generator = MessageGenerator(ConfigManager())
            >>> commit_message = generator._generate_brief_commit_message(diff)
            >>> print(commit_message)
            Added new feature
        """
        prompt = str(self.config_manager.get("prompt.brief_commit_message"))
        prompt = prompt.replace("{{ placeholder }}", diff)
        return self.llm_client.generate(prompt)

    def _generate_rich_commit_message(self, diff: str) -> str:
        """
        Generates a rich commit message based on the given diff.

        Args:
            diff (str): The diff string representing the changes made.

        Returns:
            str: The generated rich commit message.

        Raises:
            None

        Description:
            This function uses the "prompt.rich_commit_message" and "output.rich_template" from the config manager
            to generate a prompt. The placeholders "{{ rich_template }}" and "{{ placeholder }}" in the prompt
            are replaced with the provided diff and the rich template, respectively.

        Example:
            >>> diff = "Add new feature"
            >>> generator = MessageGenerator(ConfigManager())
            >>> commit_message = generator._generate_rich_commit_message(diff)
            >>> print(commit_message)
            feat: Add new feature

            - Added new feature
        """
        rich_prompt = str(self.config_manager.get("prompt.rich_commit_message"))
        rich_template = str(self.config_manager.get("output.rich_template"))
        rich_prompt = rich_prompt.replace("{{ rich_template }}", rich_template)
        rich_prompt = rich_prompt.replace("{{ placeholder }}", diff)
        return self.llm_client.generate(rich_prompt)
