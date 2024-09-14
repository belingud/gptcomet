from unittest.mock import MagicMock

import pytest

from tests import ActiveBranch, Author, MockGit, MockRepo


@pytest.fixture
def config_manager(tmp_path):
    from gptcomet.config_manager import ConfigManager

    # Mock tmp config file for all fixture
    runtime_config_file = tmp_path / "gptcomet.yaml"
    config_manager = ConfigManager(config_path=runtime_config_file)

    # Mock runtime config file path for CliRunner
    ConfigManager.get_config_path = MagicMock(return_value=config_manager.current_config_path)
    with config_manager.default_config_file.open() as f1, runtime_config_file.open("w") as f2:
        f2.write(f1.read())

    yield config_manager


@pytest.fixture
def repo():
    return MockRepo(
        active_branch=ActiveBranch(name="master"),
        git=MockGit(name="master", show=MagicMock(return_value="git show")),
    )


@pytest.fixture
def commit():
    return MagicMock(
        hexsha="123456",
        message="test commit message",
        author=Author(name="John Doe", email="john@example.com"),
    )
