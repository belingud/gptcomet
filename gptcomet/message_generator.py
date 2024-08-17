import logging
from pathlib import Path

import tenacity  # noqa: F401
from git import Repo

from gptcomet.config_manager import ConfigManager
from gptcomet.exceptions import GitNoStagedChanges
from gptcomet.llm_client import LLMClient

logger = logging.getLogger(__name__)


class MessageGenerator:
    __slots__ = ("config_manager", "llm_client", "repo")

    def __init__(self, config_manager: ConfigManager):
        self.config_manager = config_manager
        self.llm_client = LLMClient(config_manager)
        logger.debug(type(self.llm_client.retries))
        self.repo = Repo(Path.cwd())

    @classmethod
    def from_config_manager(cls, config_manager: ConfigManager):
        """
        Creates an instance of the class from a ConfigManager.

        Args:
            config_manager (ConfigManager): The ConfigManager instance to create the class instance from.

        Returns:
            An instance of the class.
        """
        return cls(config_manager)

    def generate_commit_message(self, rich: bool = True) -> str:
        """
        Generate a commit message from the staged changes.

        Args:
            rich (bool): Whether to use the rich commit message template. Defaults to True.

        Returns:
            str: The generated commit message.

        Raises:
            GitNoStagedChanges: If there are no staged changes.
        """
        logger.debug(f"[GPTComet] Generating commit message, rich: {rich}")
        self.llm_client.clear_history()
        diff = self.repo.git.diff("--cached")
        if not diff:
            raise GitNoStagedChanges()
        if not rich:
            msg = self._generate_brief_commit_message(diff)
        else:
            msg = self._generate_rich_commit_message(diff)
        lang = self.config_manager.get("output.lang")
        if lang != "en":
            logger.debug(f"[GPTComet] Translating commit message to {lang}")
            translation = self.config_manager.get("prompt.translation")
            translation = translation.replace("{{ placeholder }}", msg)
            msg = self.llm_client.generate(translation)
        return msg

        # summary = self._generate_summary(diff)
        # prefix = self._generate_prefix(summary)
        # title = self._generate_title(summary)
        #
        # message = self._generate_detailed_message(summary)
        # return f"{prefix}: {title}\n\n{message}"

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
        prompt = self.config_manager.get("prompt.brief_commit_message")
        prompt = prompt.replace("{{ placeholder }}", diff)
        return self.llm_client.generate(prompt)

    def _generate_rich_commit_message(self, diff: str) -> str:
        """
        TODO: Support rich commit message
        1. Generate summary from diff
        2. Generate prefix from summary
        3. Generate title from summary
        4. Generate detailed message from summary
        """
        return self._generate_brief_commit_message(diff)

    def generate_pr_message(self) -> str:
        diff = self.repo.git.diff("origin/master...HEAD")
        if not diff:
            return "No changes to create a pull request."

        summary = self._generate_summary(diff)
        title = self._generate_pr_title(summary)
        description = self._generate_pr_description(summary)
        changes = self._generate_pr_changes(summary)
        testing = self._generate_pr_testing(summary)

        return f"Title: {title}\n\nDescription:\n{description}\n\nChanges:\n{changes}\n\nTesting:\n{testing}"

    def generate(self, prompt_key: str, content: str) -> str:
        prompt = self.config_manager.get(prompt_key)
        prompt.replace("{{ placeholder }}", content)
        return self.llm_client.generate(prompt)

    def _generate_summary(self, diff: str) -> str:
        prompt = self.config_manager.get("prompt.commit_summary")
        prompt = prompt.replace("{{ placeholder }}", diff)
        return self.llm_client.generate(prompt)

    def _generate_prefix(self, summary: str) -> str:
        prompt = self.config_manager.get("prompt.conventional_commit_prefix")
        prompt = prompt.replace("{{ placeholder }}", summary)
        return self.llm_client.generate(prompt)

    def _generate_title(self, summary: str) -> str:
        prompt = self.config_manager.get("prompt.commit_title")
        prompt = prompt.replace("{{ placeholder }}", summary)
        return self.llm_client.generate(prompt)

    def _generate_detailed_message(self, summary: str) -> str:
        prompt = self.config_manager.get("prompt.commit_message")
        prompt = prompt.replace("{{ placeholder }}", summary)
        return self.llm_client.generate(prompt)

    def _generate_pr_title(self, summary: str) -> str:
        prompt = self.config_manager.get("prompt.pr_title")
        prompt = prompt.replace("{{ placeholder }}", summary)
        return self.llm_client.generate(prompt)

    def _generate_pr_description(self, summary: str) -> str:
        prompt = self.config_manager.get("prompt.pr_description")
        prompt = prompt.replace("{{ placeholder }}", summary)
        return self.llm_client.generate(prompt)

    def _generate_pr_changes(self, summary: str) -> str:
        prompt = self.config_manager.get("prompt.pr_changes")
        prompt = prompt.replace("{{ placeholder }}", summary)
        return self.llm_client.generate(prompt)

    def _generate_pr_testing(self, summary: str) -> str:
        prompt = self.config_manager.get("prompt.pr_testing")
        prompt = prompt.replace("{{ placeholder }}", summary)
        return self.llm_client.generate(prompt)
