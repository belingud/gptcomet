from pathlib import Path
from unittest.mock import Mock, patch

import pytest
from git import Repo

from gptcomet.exceptions import GitNoStagedChanges
from gptcomet.message_generator import MessageGenerator

PROJECT_ROOT = str(Path(__file__).parent.parent.absolute())


@pytest.fixture
def mock_config_manager():
    """创建模拟的配置管理器"""
    config = Mock()
    config.get.side_effect = lambda key, default=None: {
        "file_ignore": ["*.log", "*.tmp"],
        "language": "en",
        "prompt.brief_commit_message": "Generate commit message for: {{ placeholder }}",
        "prompt.rich_commit_message": "Generate rich commit message for: {{ placeholder }} using {{ rich_template }}",
        "prompt.translation": "Translate: {{ placeholder }} to {{ output_language }}",
        "output.rich_template": "type: message\n\n- details",
    }.get(key, default)
    return config


@pytest.fixture
def mock_repo():
    """创建模拟的Git仓库"""
    repo = Mock(spec=Repo)
    repo.git.diff.return_value = ""
    # 设置工作目录为项目根目录
    repo.working_dir = PROJECT_ROOT
    return repo


@pytest.fixture
def message_generator(mock_config_manager, mock_repo):
    """创建MessageGenerator实例"""
    with patch("git.Repo") as mock_repo_class:
        # 确保Repo()调用返回正确工作目录的repo
        mock_repo_class.return_value = mock_repo
        # 显式传入项目根目录
        return MessageGenerator(mock_config_manager, repo_path=PROJECT_ROOT)


class TestMessageGenerator:
    def test_initialization(self, mock_config_manager):
        """测试初始化"""
        with patch("git.Repo") as mock_repo:
            mock_instance = Mock(spec=Repo)
            mock_instance.working_dir = PROJECT_ROOT
            mock_repo.return_value = mock_instance
            generator = MessageGenerator(mock_config_manager, repo_path=PROJECT_ROOT)
            assert generator.config_manager == mock_config_manager
            assert generator.diff is None

    def test_from_config_manager(self, mock_config_manager):
        """测试from_config_manager工厂方法"""
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
    def test_get_staged_diff_empty(self, message_generator, mock_repo):
        """测试没有暂存更改时的差异获取"""
        mock_repo.git.diff.return_value = ""
        diff = message_generator.get_staged_diff(mock_repo)
        assert diff == ""

    def test_get_staged_diff_with_changes(self, message_generator, mock_repo):
        """测试有暂存更改时的差异获取"""
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
        diff = message_generator.get_staged_diff(mock_repo)

        # 验证过滤掉了不需要的行
        assert "index 1234567..89abcdef 100644" not in diff
        assert "--- a/test.py" not in diff
        assert "+++ b/test.py" not in diff
        assert '+    print("hello")' in diff


class TestGenerateCommitMessage:
    def test_generate_commit_message_no_changes(self, message_generator):
        """测试没有暂存更改时生成提交信息"""
        with pytest.raises(GitNoStagedChanges):
            with patch("gptcomet.message_generator.MessageGenerator.get_staged_diff") as mock_get_staged_diff:
                mock_get_staged_diff.return_value = ""
                message_generator.generate_commit_message()

    def test_generate_brief_commit_message(self, message_generator, mock_repo):
        """测试生成简短提交信息"""
        mock_repo.git.diff.return_value = "test diff content"
        with patch("gptcomet.llm_client.LLMClient.generate") as mock_generate:
            def get_side_effect(key, _type=None):
                if key == "output.lang":
                    return "en"
                elif key == "prompt.translation":
                    return "Translate: {{ placeholder }} to {{ output_language }}"
                elif key == "file_ignore":
                    return ["*.log", "*.tmp"]
                return None
            message_generator.config_manager.get.side_effect = get_side_effect
            mock_generate.return_value = "Add new feature"
            msg = message_generator.generate_commit_message(rich=False)
            assert msg == "Add new feature"

    def test_generate_rich_commit_message(self, message_generator, mock_repo):
        """测试生成富文本提交信息"""
        mock_repo.git.diff.return_value = "test diff content"
        with patch("gptcomet.llm_client.LLMClient.generate") as mock_generate:
            def get_side_effect(key, _type=None):
                if key == "output.lang":
                    return "zh-cn"
                elif key == "prompt.translation":
                    return "Translate: {{ placeholder }} to {{ output_language }}"
                elif key == "file_ignore":
                    return ["*.log", "*.tmp"]
                return None
            message_generator.config_manager.get.side_effect = get_side_effect
            mock_generate.side_effect = ["Add new feature", "添加新功能"]

    def test_translate_commit_message(self, message_generator, mock_repo):
        """测试提交信息翻译"""
        mock_repo.git.diff.return_value = "test diff content"
        with patch("gptcomet.llm_client.LLMClient.generate") as mock_generate:
            def get_side_effect(key, _type=None):
                if key == "output.lang":
                    return "zh-cn"
                elif key == "prompt.translation":
                    return "Translate: {{ placeholder }} to {{ output_language }}"
                elif key == "file_ignore":
                    return ["*.log", "*.tmp"]
                return None
            mock_generate.side_effect = ["Add new feature", "添加新功能"]
            message_generator.config_manager.get.side_effect = get_side_effect

            msg = message_generator.generate_commit_message(rich=False)
            assert msg == "添加新功能"
            assert mock_generate.call_count == 2


class TestPrivateMethods:
    def test_generate_brief_commit_message(self, message_generator):
        """测试_generate_brief_commit_message方法"""
        with patch("gptcomet.llm_client.LLMClient.generate") as mock_generate:
            mock_generate.return_value = "Add new feature"
            msg = message_generator._generate_brief_commit_message("test diff")
            assert msg == "Add new feature"

    def test_generate_rich_commit_message(self, message_generator):
        """测试_generate_rich_commit_message方法"""
        with patch("gptcomet.llm_client.LLMClient.generate") as mock_generate:
            mock_generate.return_value = "feat: Add new feature"
            msg = message_generator._generate_rich_commit_message("test diff")
            assert msg == "feat: Add new feature"

    def test_translate_msg(self, message_generator):
        """测试_translate_msg方法"""
        with patch("gptcomet.llm_client.LLMClient.generate") as mock_generate:
            mock_generate.return_value = "添加新功能"

            def get_side_effect(key, _type=None):
                if key == "output.lang":
                    return "zh-cn"
                elif key == "prompt.translation":
                    return "Translate: {{ placeholder }} to {{ output_language }}"
                return None

            message_generator.config_manager.get.side_effect = get_side_effect

            msg = message_generator._translate_msg("Add new feature")
            assert msg == "添加新功能"
