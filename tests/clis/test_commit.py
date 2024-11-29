from unittest.mock import Mock, patch

import pytest
import typer
from git import Commit, Repo

from gptcomet.clis.commit import (
    RETRY_CHOICES,
    ask_for_retry,
    commit,
    edit_text_in_place,
    entry,
    gen_output,
)
from gptcomet.exceptions import ConfigError, ConfigErrorEnum, GitNoStagedChanges
from gptcomet.message_generator import MessageGenerator
from gptcomet.styles import Colors, stylize


@pytest.fixture
def mock_console():
    with patch("gptcomet.clis.commit.console") as mock:
        yield mock


@pytest.fixture
def mock_repo():
    repo = Mock(spec=Repo)
    repo.active_branch.name = "main"
    return repo


@pytest.fixture
def mock_commit():
    commit = Mock(spec=Commit)
    commit.hexsha = "abcd1234"
    commit.message = b"test commit message"
    commit.author.conf_name = "name"
    commit.author.conf_email = "email"
    commit.author.name = "Test Author"
    commit.author.email = "test@example.com"
    return commit


@pytest.fixture
def mock_message_generator(mock_repo):
    generator = Mock(spec=MessageGenerator)
    generator.repo = mock_repo
    generator.generate_commit_message.return_value = "Generated commit message"
    return generator


def test_gen_output_normal(mock_repo, mock_commit):
    mock_repo.git.show.return_value = "1 file changed, 2 insertions(+), 1 deletion(-)"
    output = gen_output(mock_repo, mock_commit, rich=False)
    assert "test commit message" in output
    assert "Test Author" in output
    assert "test@example.com" in output
    assert "main" in output
    assert "abcd1234" in output
    assert "1 file changed" in output


def test_gen_output_rich(mock_repo, mock_commit):
    mock_repo.git.show.return_value = "1 file changed, 2 insertions(+), 1 deletion(-)"
    output = gen_output(mock_repo, mock_commit, rich=True)
    assert "[green]Test Author[/]" in output
    assert "[blue]test@example.com[/blue]" in output
    assert "[green]+[/green]" in output
    assert "[red]-[/red]" in output


def test_gen_output_no_author_email(mock_repo, mock_commit):
    mock_commit.author.name = None
    mock_commit.author.email = None
    output = gen_output(mock_repo, mock_commit, rich=False)
    assert "No Author" in output
    assert "No Email" in output


@patch("gptcomet.clis.commit.sys.stdin")
def test_ask_for_retry_non_interactive(mock_stdin, mock_console):
    mock_stdin.isatty.return_value = False
    result = ask_for_retry()
    assert result == "y"
    mock_console.print.assert_called_with(
        "[yellow]Non-interactive mode detected, using the generated commit message directly.[/yellow]",
    )


@patch("gptcomet.clis.commit.sys.stdin")
@patch("gptcomet.clis.commit.Prompt.ask")
def test_ask_for_retry_interactive(mock_prompt, mock_stdin):
    mock_stdin.isatty.return_value = True
    mock_prompt.return_value = "y"
    result = ask_for_retry()
    assert result == "y"
    prompt_text = "Do you want to use this commit message?\n" + ", ".join(
        [f"{k}: {v}" for k, v in RETRY_CHOICES.items()]
    )
    mock_prompt.assert_called_with(
        prompt_text, default="y", choices=list(RETRY_CHOICES.keys()), case_sensitive=False
    )


@patch("gptcomet.clis.commit.prompt")
def test_edit_text_in_place_normal(mock_prompt, mock_console):
    mock_prompt.return_value = "edited message"
    result = edit_text_in_place("initial message")
    assert result == "edited message"
    mock_prompt.assert_called_once()


@patch("gptcomet.clis.commit.prompt")
def test_edit_text_in_place_empty(mock_prompt, mock_console):
    mock_prompt.return_value = ""
    result = edit_text_in_place("initial message")
    assert result == "initial message"


@patch("gptcomet.clis.commit.prompt")
def test_edit_text_in_place_keyboard_interrupt(mock_prompt, mock_console):
    mock_prompt.side_effect = KeyboardInterrupt()
    result = edit_text_in_place("initial message")
    assert result is None
    mock_console.print.assert_called_with("\n[yellow]Commit cancelled.[/yellow]")


def test_commit_success(mock_message_generator, mock_console):
    commit_obj = Mock()
    commit_obj.message = b"test message"
    commit_obj.hexsha = "abcd1234"
    commit_obj.author.conf_name = "name"
    commit_obj.author.conf_email = "email"
    commit_obj.author.name = "Test Author"
    commit_obj.author.email = "test@example.com"
    mock_message_generator.repo.index.commit.return_value = commit_obj
    mock_message_generator.repo.git.show.return_value = "1 file changed"

    commit(mock_message_generator, "test message")
    mock_message_generator.repo.index.commit.assert_called_with(message="test message")
    assert mock_console.print.call_count >= 2


def test_commit_error(mock_message_generator, mock_console):
    mock_message_generator.repo.index.commit.side_effect = ValueError("test error")
    with pytest.raises(typer.Exit):
        commit(mock_message_generator, "test message")
    mock_console.print.assert_called_with("[red]Commit Error: test error[/red]")


@patch("gptcomet.clis.commit.MessageGenerator")
def test_entry_config_error(mock_message_generator_class, mock_console):
    error = ConfigError(ConfigErrorEnum.API_KEY_MISSING, provider="test")
    mock_message_generator_class.side_effect = error
    with pytest.raises(typer.Exit):
        entry()
    mock_console.print.assert_called_with(stylize(str(error), Colors.YELLOW))


@patch("gptcomet.clis.commit.MessageGenerator")
def test_entry_no_message(mock_message_generator_class, mock_console):
    mock_generator = Mock()
    mock_generator.generate_commit_message.return_value = ""
    mock_message_generator_class.return_value = mock_generator
    with pytest.raises(typer.Exit):
        entry()
    mock_console.print.assert_called_with(stylize("No commit message generated.", Colors.MAGENTA))


@patch("gptcomet.clis.commit.MessageGenerator")
@patch("gptcomet.clis.commit.ask_for_retry")
def test_entry_user_reject(mock_ask_retry, mock_message_generator_class, mock_console):
    mock_generator = Mock()
    mock_generator.generate_commit_message.return_value = "test message"
    mock_message_generator_class.return_value = mock_generator
    mock_ask_retry.return_value = "n"

    entry()
    mock_console.print.assert_any_call(stylize("Commit message discarded.", Colors.YELLOW))


@patch("gptcomet.clis.commit.MessageGenerator")
def test_entry_git_error(mock_message_generator_class, mock_console):
    error = GitNoStagedChanges()
    mock_message_generator_class.return_value = Mock(spec=MessageGenerator)
    mock_message_generator_class.return_value.generate_commit_message.side_effect = error
    with pytest.raises(typer.Exit):
        entry()
    mock_console.print.assert_called_with(stylize(str(error), Colors.YELLOW))


@patch("gptcomet.clis.commit.MessageGenerator")
def test_entry_keyboard_interrupt(mock_message_generator_class, mock_console):
    mock_generator = Mock()
    mock_generator.generate_commit_message.side_effect = KeyboardInterrupt()
    mock_message_generator_class.return_value = mock_generator

    entry()
    mock_console.print.assert_called_with("\n[yellow]Operation cancelled by user.[/yellow]")
