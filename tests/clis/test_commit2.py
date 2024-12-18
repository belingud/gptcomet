from unittest.mock import MagicMock, Mock, patch

import pytest
import typer
from typer.testing import CliRunner

from gptcomet.clis.commit import entry
from gptcomet.config_manager import ConfigManager
from gptcomet.exceptions import KeyNotFound
from gptcomet.message_generator import MessageGenerator

app = typer.Typer()
app.command()(entry)
runner = CliRunner()


@pytest.fixture
def mock_message_generator():
    with patch("gptcomet.clis.commit.MessageGenerator") as mock:
        generator = Mock(spec=MessageGenerator)
        mock.return_value = generator
        generator.generate_commit_message.return_value = "test commit message"
        yield generator


class TestCommitEntry:
    def test_basic_commit(self, config_manager, mock_message_generator):
        """Test basic commit without any options"""
        mock_message_generator.generate_commit_message.return_value = "test commit message"

        with patch("gptcomet.clis.commit.commit") as mock_commit:
            result = runner.invoke(app, [])

        assert result.exit_code == 0
        mock_message_generator.generate_commit_message.assert_called_once_with(rich=False)
        mock_commit.assert_called_once()

    def test_rich_commit(self, config_manager, mock_message_generator):
        """Test commit with rich option"""
        mock_message_generator.generate_commit_message.return_value = "test commit message"

        with patch("gptcomet.clis.commit.commit"):
            result = runner.invoke(app, ["--rich"])

        assert result.exit_code == 0
        mock_message_generator.generate_commit_message.assert_called_once_with(rich=True)

    def test_provider_override(self, config_manager, mock_message_generator):
        """Test commit with provider override"""
        original_set_cli_overrides = ConfigManager.set_cli_overrides
        ConfigManager.set_cli_overrides = MagicMock()
        with (
            patch("gptcomet.clis.commit.commit"),
            patch("gptcomet.message_generator.MessageGenerator"),
        ):
            result = runner.invoke(app, ["--provider", "anthropic"])

            assert result.exit_code == 0
            config_manager.set_cli_overrides.assert_called_once_with(
                provider="anthropic", api_config={}
            )
        ConfigManager.set_cli_overrides = original_set_cli_overrides

    def test_api_config_override(self, config_manager, mock_message_generator):
        """Test commit with API configuration overrides"""
        original_set_cli_overrides = ConfigManager.set_cli_overrides
        ConfigManager.set_cli_overrides = MagicMock()

        with (
            patch("gptcomet.clis.commit.commit"),
            patch("gptcomet.message_generator.MessageGenerator"),
        ):
            result = runner.invoke(
                app, ["--api-key", "test-key", "--model", "gpt-4", "--temperature", "0.8"]
            )

        assert result.exit_code == 0
        config_manager.set_cli_overrides.assert_called_once_with(
            provider=None, api_config={"api_key": "test-key", "model": "gpt-4", "temperature": 0.8}
        )
        ConfigManager.set_cli_overrides = original_set_cli_overrides

    def test_invalid_temperature(self, config_manager):
        """Test commit with invalid temperature value"""
        result = runner.invoke(app, ["--temperature", "2.0"])

        assert result.exit_code != 0
        assert "Invalid value for '--temperature'" in result.stdout

    def test_invalid_max_tokens(self, config_manager):
        """Test commit with invalid max_tokens value"""
        result = runner.invoke(app, ["--max-tokens", "0"])

        assert result.exit_code != 0
        assert "Invalid value for '--max-tokens'" in result.stdout

    def test_local_config(self, config_manager, mock_message_generator):
        """Test commit using local configuration"""
        with (
            patch("gptcomet.clis.commit.commit"),
            patch("gptcomet.message_generator.MessageGenerator"),
        ):
            result = runner.invoke(app, ["--local"])

        assert result.exit_code == 0
        # from gptcomet.clis.commit import get_config_manager
        # get_config_manager.assert_called_once_with(local=True)

    def test_debug_mode(self, config_manager, mock_message_generator):
        """Test commit in debug mode"""
        with (
            patch("gptcomet.clis.commit.set_debug") as mock_set_debug,
            patch("gptcomet.clis.commit.commit"),
        ):
            result = runner.invoke(app, ["--debug"])

        assert result.exit_code == 0
        mock_set_debug.assert_called_once()

    def test_all_options_combined(self, config_manager, mock_message_generator):
        """Test commit with all options combined"""
        original_set_cli_overrides = ConfigManager.set_cli_overrides
        ConfigManager.set_cli_overrides = MagicMock()
        with (
            patch("gptcomet.clis.commit.commit"),
            patch("gptcomet.message_generator.MessageGenerator"),
        ):
            result = runner.invoke(
                app,
                [
                    "--rich",
                    "--debug",
                    "--provider",
                    "anthropic",
                    "--api-key",
                    "test-key",
                    "--model",
                    "claude-2",
                    "--temperature",
                    "0.7",
                    "--max-tokens",
                    "2000",
                    "--top-p",
                    "0.9",
                    "--frequency-penalty",
                    "0.5",
                ],
            )

        assert result.exit_code == 0
        config_manager.set_cli_overrides.assert_called_once_with(
            provider="anthropic",
            api_config={
                "api_key": "test-key",
                "model": "claude-2",
                "temperature": 0.7,
                "max_tokens": 2000,
                "top_p": 0.9,
                "frequency_penalty": 0.5,
            },
        )
        ConfigManager.set_cli_overrides = original_set_cli_overrides

    def test_error_handling(self, config_manager, mock_message_generator):
        """Test error handling in the entry function."""
        mock_message_generator.generate_commit_message.side_effect = KeyNotFound("Test error")

        result = runner.invoke(app, [])
        assert "Key 'Test error' not found in the configuration." in result.output
        assert result.exit_code == 1

    @patch("sys.stdin.isatty", return_value=False)
    def test_non_interactive_mode(self, mock_isatty, config_manager, mock_message_generator):
        """Test commit in non-interactive mode"""
        mock_message_generator.generate_commit_message.return_value = "test commit message"

        with patch("gptcomet.clis.commit.commit"):
            result = runner.invoke(app, [])

        assert result.exit_code == 0
        assert "Non-interactive mode detected" in result.stdout
