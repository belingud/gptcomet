import os
from pathlib import Path
from unittest.mock import MagicMock, create_autospec, patch

import glom
import pytest
from git import Git

from gptcomet.message_generator import MessageGenerator


@pytest.fixture
def message_generator(tmp_path):
    os.chdir(Path(__file__).parent.parent.parent)
    config_manager = MagicMock()

    config_manager.get.side_effect = lambda key, default=None: glom.glom(
        {
            "provider": "openai",
            "openai": {
                "api_key": "test-key",
                "model": "gpt-3.5-turbo",
                "api_base": "https://api.openai.com/v1",
                "max_tokens": "100",
                "temperature": "0.7",
                "top_p": "0.9",
                "frequency_penalty": "0.5",
                "extra_headers": '{"X-Custom": "value"}',
            },
            "prompt": {
                "system": "You are a helpful assistant.",
                "user": "Hello!",
            },
        },
        key,
        default=default,
    )
    mock_git = create_autospec(Git, instance=True)
    mg = MessageGenerator(config_manager)
    mg.repo.git = mock_git
    return mg


def test_make_ignored_options_empty_ignored_files(message_generator):
    ignored_files = []
    message_generator.repo.git.diff = MagicMock(return_value="")
    assert message_generator.make_ignored_options(ignored_files) == []


def test_make_ignored_options_ignored_files_not_in_staged_files(message_generator, tmp_path):
    ignored_files = ["file1.txt", "file2.txt"]
    message_generator.repo.git.diff = MagicMock(return_value="file3.txt\nfile4.txt")

    assert message_generator.make_ignored_options(ignored_files) == []


def test_make_ignored_options_ignored_files_in_staged_files(message_generator, tmp_path):
    ignored_files = ["file1.txt", "file2.txt"]
    message_generator.repo.git.diff = MagicMock(return_value="file1.txt\nfile2.txt\nfile3.txt")

    (tmp_path / "file1.txt").touch()
    (tmp_path / "file2.txt").touch()
    (tmp_path / "file3.txt").touch()

    message_generator.repo.git.diff.side_effect = (
        lambda option: "file1.txt\nfile2.txt\nfile3.txt"
        if "--name-only" in option
        else "file1.txt\nfile2.txt\nfile3.txt"
    )
    with patch("pathlib.Path.exists", return_value=True):
        assert message_generator.make_ignored_options(ignored_files) == [
            ":!file1.txt",
            ":!file2.txt",
        ]


def test_make_ignored_options_staged_files_empty(message_generator, tmp_path):
    ignored_files = ["file1.txt", "file2.txt"]
    message_generator.repo.git.diff = MagicMock(return_value="")

    assert message_generator.make_ignored_options(ignored_files) == []


def test_make_ignored_options_staged_files_in_ignored_files_pattern(message_generator, tmp_path):
    ignored_files = ["**/pdm.lock", "**/pnpm-lock.yaml"]
    message_generator.repo.git.diff = MagicMock(
        return_value="frontend/pnpm-lock.yaml\nbackend/pdm.lock"
    )
    (tmp_path / "frontend").mkdir(exist_ok=True)
    (tmp_path / "frontend" / "pnpm-lock.yaml").touch()
    (tmp_path / "backend").mkdir(exist_ok=True)
    (tmp_path / "backend" / "pdm.lock").touch()

    message_generator.repo.git.diff.side_effect = (
        lambda option: "frontend/pnpm-lock.yaml\nbackend/pdm.lock"
        if "--name-only" in option
        else "frontend/pnpm-lock.yaml\nbackend/pdm.lock"
    )
    with patch("pathlib.Path.exists", return_value=True):
        assert message_generator.make_ignored_options(ignored_files) == [
            ":!frontend/pnpm-lock.yaml",
            ":!backend/pdm.lock",
        ]
