import os
from pathlib import Path
from unittest.mock import MagicMock, create_autospec, patch

import pytest
from git import Git

from gptcomet.message_generator import MessageGenerator


@pytest.fixture
def message_generator():
    os.chdir(Path(__file__).parent.parent.parent)
    config_manager = MagicMock()
    mock_git = create_autospec(Git, instance=True)
    mg = MessageGenerator(config_manager)
    mg.repo.git = mock_git
    return mg


def test_make_ignored_options_empty_ignored_files(message_generator):
    ignored_files = []
    message_generator.repo.git.diff = MagicMock(return_value="")
    assert message_generator.make_ignored_options(ignored_files) == []


def test_make_ignored_options_ignored_files_not_in_staged_files(message_generator):
    ignored_files = ["file1.txt", "file2.txt"]
    message_generator.repo.git.diff = MagicMock(return_value="file3.txt\nfile4.txt")
    with patch("pathlib.Path.exists", return_value=True):
        assert message_generator.make_ignored_options(ignored_files) == []


def test_make_ignored_options_ignored_files_in_staged_files(message_generator):
    ignored_files = ["file1.txt", "file2.txt"]
    message_generator.repo.git.diff = MagicMock(return_value="file1.txt\nfile2.txt\nfile3.txt")
    with patch("pathlib.Path.exists", return_value=True):
        assert message_generator.make_ignored_options(ignored_files) == [
            ":!file1.txt",
            ":!file2.txt",
        ]


def test_make_ignored_options_staged_files_empty(message_generator):
    ignored_files = ["file1.txt", "file2.txt"]
    message_generator.repo.git.diff = MagicMock(return_value="")
    with patch("pathlib.Path.exists", return_value=True):
        assert message_generator.make_ignored_options(ignored_files) == []


def test_make_ignored_options_staged_files_in_ignored_files_pattern(message_generator):
    ignored_files = ["**/pdm.lock", "**/pnpm-lock.yaml"]
    message_generator.repo.git.diff = MagicMock(
        return_value="frontend/pnpm-lock.yaml\nbackend/pdm.lock"
    )
    with patch("pathlib.Path.exists", return_value=True):
        assert message_generator.make_ignored_options(ignored_files) == [
            ":!frontend/pnpm-lock.yaml",
            ":!backend/pdm.lock",
        ]
