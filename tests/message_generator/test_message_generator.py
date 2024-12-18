from pathlib import Path
from unittest.mock import Mock, patch

import glom
import pytest
from git import Repo

from gptcomet.exceptions import GitNoStagedChanges
from gptcomet.message_generator import MessageGenerator

PROJECT_ROOT = str(Path(__file__).parent.parent.absolute())


@pytest.fixture
def mock_config_manager():
    """Create a mock configuration manager"""
    config = Mock()
    config.get.side_effect = lambda key, default=None: glom.glom(
        {
            "provider": "openai",
            "openai": {
                "api_key": "test-key",
                "model": "gpt-3.5-turbo",
                "api_base": "https://api.openai.com/v1",
                "retries": "3",
                "completion_path": "/chat/completions",
                "answer_path": "choices.0.message.content",
                "proxy": "",
                "max_tokens": "100",
                "temperature": "0.7",
                "top_p": "1",
                "frequency_penalty": "0",
                "presence_penalty": "0",
            },
            "file_ignore": ["*.log", "*.tmp"],
            "language": "en",
            "prompt.brief_commit_message": "Generate commit message for: {{ placeholder }}",
            "prompt.rich_commit_message": "Generate rich commit message for: {{ placeholder }} using {{ rich_template }}",
            "prompt.translation": "Translate: {{ placeholder }} to {{ output_language }}",
            "output.rich_template": "type: message\n\n- details",
        },
        key,
        default=default,
    )
    config.is_api_key_set = True
    return config


@pytest.fixture
def mock_repo():
    """Create a mock Git repository"""
    repo = Mock(spec=Repo)
    repo.git.diff.return_value = ""
    # Set working directory as project root directory
    repo.working_dir = PROJECT_ROOT
    return repo


@pytest.fixture
def create_message_generator(mock_config_manager, mock_repo):
    """Create MessageGenerator instance"""
    mock_repo_class = patch("git.Repo").start()
    mock_repo_class.return_value = mock_repo
    # Explicitly pass project root directory
    msg_generator = MessageGenerator(mock_config_manager, repo_path=PROJECT_ROOT)
    msg_generator.repo = mock_repo
    return msg_generator


class TestMessageGenerator:
    def test_init(self, mock_config_manager):
        """Test initialization"""
        with patch("git.Repo") as mock_repo:
            mock_instance = Mock(spec=Repo)
            mock_instance.working_dir = PROJECT_ROOT
            mock_repo.return_value = mock_instance
            generator = MessageGenerator(mock_config_manager, repo_path=PROJECT_ROOT)
            assert generator.config_manager == mock_config_manager
            assert generator.diff is None

    def test_from_config_manager(self, mock_config_manager):
        """Test from_config_manager factory method"""
        with patch("git.Repo") as mock_repo:
            mock_instance = Mock(spec=Repo)
            mock_instance.working_dir = PROJECT_ROOT
            mock_repo.return_value = mock_instance
            generator = MessageGenerator.from_config_manager(
                mock_config_manager, repo_path=PROJECT_ROOT
            )
            assert isinstance(generator, MessageGenerator)
            assert generator.config_manager == mock_config_manager


class TestGetStagedDiff:
    def test_get_staged_diff_no_changes(self, create_message_generator, mock_repo):
        """Test getting diff when there are no staged changes"""
        mock_repo.git.diff.return_value = ""
        diff = create_message_generator.get_staged_diff(mock_repo)
        assert diff == ""

    def test_get_staged_diff_with_changes(self, create_message_generator, mock_repo):
        """Test getting diff when there are staged changes"""
        mock_diff = """
diff --git a/test.py b/test.py
index 1234567..89abcdef 100644
--- a/test.py
+++ b/test.py
@@ -1,2 +1,3 @@
 def test():
+    print("hello")
     pass
"""
        mock_repo.git.diff.return_value = mock_diff
        diff = create_message_generator.get_staged_diff(mock_repo)

        # Verify that unwanted lines are filtered out
        assert "index 1234567..89abcdef 100644" not in diff
        assert "--- a/test.py" not in diff
        assert "+++ b/test.py" not in diff
        assert '+    print("hello")' in diff


class TestGenerateCommitMessage:
    def test_generate_commit_message_no_changes(self, create_message_generator):
        """Test generating commit message when there are no staged changes"""
        with (
            pytest.raises(GitNoStagedChanges),
            patch(
                "gptcomet.message_generator.MessageGenerator.get_staged_diff"
            ) as mock_get_staged_diff,
        ):
            mock_get_staged_diff.return_value = ""
            create_message_generator.generate_commit_message()

    def test_generate_brief_commit_message(self, create_message_generator, mock_repo):
        """Test generating brief commit message"""
        mock_repo.git.diff.return_value = "test diff content"
        with patch("gptcomet.llm_client.LLMClient.generate") as mock_generate:

            def get_side_effect(key, _type=None):
                return {
                    "output.lang": "en",
                    "prompt.translation": "Translate: {{ placeholder }} to {{ output_language }}",
                    "file_ignore": ["*.log", "*.tmp"],
                }.get(key)

            create_message_generator.config_manager.get.side_effect = get_side_effect
            mock_generate.return_value = "Add new feature"
            msg = create_message_generator.generate_commit_message(rich=False)
            assert msg == "Add new feature"

    def test_generate_rich_commit_message(self, create_message_generator, mock_repo):
        """Test generating rich commit message"""
        mock_repo.git.diff.return_value = "test diff content"
        with patch("gptcomet.llm_client.LLMClient.generate") as mock_generate:

            def get_side_effect(key, _type=None):
                return {
                    "output.lang": "en",
                    "prompt.translation": "Translate: {{ placeholder }} to {{ output_language }}",
                    "file_ignore": ["*.log", "*.tmp"],
                }.get(key)

            create_message_generator.config_manager.get.side_effect = get_side_effect
            mock_generate.side_effect = ["Add new feature", "Add new feature"]

    def test_translate_message(self, create_message_generator, mock_repo):
        """Test message translation"""
        mock_repo.git.diff.return_value = "test diff content"
        with patch("gptcomet.llm_client.LLMClient.generate") as mock_generate:

            def get_side_effect(key, _type=None):
                return {
                    "output.lang": "zh-cn",
                    "prompt.translation": "Translate: {{ placeholder }} to {{ output_language }}",
                    "file_ignore": ["*.log", "*.tmp"],
                }.get(key)

            mock_generate.side_effect = ["Add new feature", "Add new feature"]
            create_message_generator.config_manager.get.side_effect = get_side_effect

            msg = create_message_generator.generate_commit_message(rich=False)
            assert msg == "Add new feature"
            assert mock_generate.call_count == 2


class TestPrivateMethods:
    def test__generate_brief_commit_message(self, create_message_generator):
        """Test _generate_brief_commit_message method"""
        with patch("gptcomet.llm_client.LLMClient.generate") as mock_generate:
            mock_generate.return_value = "Add new feature"
            msg = create_message_generator._generate_brief_commit_message("test diff")
            assert msg == "Add new feature"

    def test__generate_rich_commit_message(self, create_message_generator):
        """Test _generate_rich_commit_message method"""
        with patch("gptcomet.llm_client.LLMClient.generate") as mock_generate:
            mock_generate.return_value = "feat: Add new feature"
            msg = create_message_generator._generate_rich_commit_message("test diff")
            assert msg == "feat: Add new feature"

    def test__translate_msg(self, create_message_generator):
        """Test _translate_msg method"""
        with patch("gptcomet.llm_client.LLMClient.generate") as mock_generate:
            mock_generate.return_value = "Add new feature"

            def get_side_effect(key, _type=None):
                if key == "output.lang":
                    return "en"
                elif key == "prompt.translation":
                    return "Translate: {{ placeholder }} to {{ output_language }}"
                return None

            create_message_generator.config_manager.get.side_effect = get_side_effect

            msg = create_message_generator._translate_msg("Add new feature")
            assert msg == "Add new feature"
